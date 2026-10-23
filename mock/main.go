package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

const testToken = "local-test-token"

type RouterOptions struct{ Scenario string }

type Transaction struct {
	ID          string `json:"id"`
	AccountID   string `json:"account_id"`
	Status      string `json:"status"`
	AmountCents int64  `json:"amount_cents"`
	Currency    string `json:"currency"`
	CreatedAt   string `json:"created_at"`
}

type transactionSummary struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type pageResponse struct {
	Items      []transactionSummary `json:"items"`
	NextCursor *string              `json:"next_cursor"`
}

type requestRecord struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Status int    `json:"status"`
}

type mockState struct {
	mu       sync.Mutex
	requests []requestRecord
	failures map[string]int
}

func NewRouter(options RouterOptions) http.Handler {
	state := &mockState{failures: make(map[string]int)}
	transactions, accounts := fixtures()
	byID := make(map[string]Transaction)
	byAccount := make(map[string][]Transaction)
	for id := range accounts {
		byAccount[id] = nil
	}
	for _, tx := range transactions {
		byID[tx.ID] = tx
		byAccount[tx.AccountID] = append(byAccount[tx.AccountID], tx)
	}
	// Repeat a listing entry on a later page; detail lookup stays canonical.
	original := byAccount["acct_001"]
	listing := append([]Transaction{}, original[:8]...)
	listing = append(listing, byID["tx_005"])
	byAccount["acct_001"] = append(listing, original[8:]...)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /_admin/reset", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		state.failures = make(map[string]int)
		state.requests = nil
		state.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]string{"status": "reset"})
	})
	mux.HandleFunc("GET /_admin/requests", func(w http.ResponseWriter, r *http.Request) {
		state.mu.Lock()
		items := append([]requestRecord{}, state.requests...)
		state.mu.Unlock()
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	})

	api := http.NewServeMux()
	api.HandleFunc("GET /v1/accounts/{account_id}/transactions", func(w http.ResponseWriter, r *http.Request) {
		items, ok := byAccount[r.PathValue("account_id")]
		if !ok {
			writeError(w, http.StatusNotFound, "unknown account")
			return
		}
		limit := 4
		if raw, present := r.URL.Query()["limit"]; present {
			value, err := strconv.Atoi(raw[0])
			if len(raw) != 1 || err != nil || value < 1 || value > 20 {
				writeError(w, http.StatusBadRequest, "limit must be an integer from 1 to 20")
				return
			}
			limit = value
		}
		start := 0
		if cursors, present := r.URL.Query()["cursor"]; present {
			cursor := cursors[0]
			offset, err := strconv.Atoi(strings.TrimPrefix(cursor, "cursor_"))
			if len(cursors) != 1 || !strings.HasPrefix(cursor, "cursor_") || err != nil || offset <= 0 || offset >= len(items) || offset%limit != 0 {
				writeError(w, http.StatusBadRequest, "invalid cursor")
				return
			}
			start = offset
		}
		end := start + limit
		if end > len(items) {
			end = len(items)
		}
		page := make([]transactionSummary, 0, end-start)
		for _, tx := range items[start:end] {
			page = append(page, transactionSummary{ID: tx.ID, Status: tx.Status, CreatedAt: tx.CreatedAt})
		}
		var next *string
		if end < len(items) {
			value := "cursor_" + strconv.Itoa(end)
			next = &value
		}
		writeJSON(w, http.StatusOK, pageResponse{Items: page, NextCursor: next})
	})
	api.HandleFunc("GET /v1/transactions/{transaction_id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("transaction_id")
		tx, ok := byID[id]
		if !ok {
			writeError(w, http.StatusNotFound, "unknown transaction")
			return
		}
		if options.Scenario == "transient" {
			status := 0
			if id == "tx_003" {
				status = http.StatusServiceUnavailable
			}
			if id == "tx_004" {
				status = http.StatusTooManyRequests
			}
			if status != 0 {
				state.mu.Lock()
				state.failures[id]++
				attempt := state.failures[id]
				state.mu.Unlock()
				if attempt == 1 {
					if status == http.StatusTooManyRequests {
						w.Header().Set("Retry-After", "1")
					}
					writeError(w, status, http.StatusText(status))
					return
				}
			}
		}
		writeJSON(w, http.StatusOK, tx)
	})
	mux.Handle("/v1/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+testToken {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		api.ServeHTTP(w, r)
	}))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorded := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		mux.ServeHTTP(recorded, r)
		if r.Method == http.MethodPost && r.URL.Path == "/_admin/reset" {
			return
		}
		state.mu.Lock()
		defer state.mu.Unlock()
		state.requests = append(state.requests, requestRecord{Method: r.Method, Path: r.URL.RequestURI(), Status: recorded.status})
		if len(state.requests) > 100 {
			state.requests = state.requests[len(state.requests)-100:]
		}
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "18080"
	}
	value, err := strconv.Atoi(port)
	if err != nil || value < 1 || value > 65535 {
		log.Fatal("PORT must be an integer from 1 to 65535")
	}
	scenario := os.Getenv("MOCK_SCENARIO")
	if scenario == "" {
		scenario = "normal"
	}
	if scenario != "normal" && scenario != "transient" {
		log.Fatal("MOCK_SCENARIO must be normal or transient")
	}
	server := &http.Server{
		Addr:              "127.0.0.1:" + port,
		Handler:           NewRouter(RouterOptions{Scenario: scenario}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			_ = server.Close()
		}
	}()
	log.Printf("mock API listening on http://%s (scenario=%s)", server.Addr, scenario)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
