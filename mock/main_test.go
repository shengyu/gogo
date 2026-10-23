package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func perform(handler http.Handler, method, path, token string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	return res
}

func mockRouter(scenario string) http.Handler { return NewRouter(RouterOptions{Scenario: scenario}) }

func TestTransactionsRequireAuthentication(t *testing.T) {
	handler := mockRouter("normal")
	for _, token := range []string{"", "wrong-token"} {
		res := perform(handler, http.MethodGet, "/v1/transactions/tx_001", token)
		if res.Code != http.StatusUnauthorized {
			t.Fatalf("token %q: expected 401, got %d", token, res.Code)
		}
	}
	res := perform(handler, http.MethodGet, "/v1/transactions/tx_001", testToken)
	if res.Code != http.StatusOK {
		t.Fatalf("valid token: expected 200, got %d", res.Code)
	}
}

func TestAccountPaginationAndEmptyHistory(t *testing.T) {
	handler := mockRouter("normal")
	first := perform(handler, http.MethodGet, "/v1/accounts/acct_001/transactions?limit=4", testToken)
	if first.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", first.Code)
	}
	var page pageResponse
	if err := json.Unmarshal(first.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 4 || page.NextCursor == nil || *page.NextCursor != "cursor_4" {
		t.Fatalf("unexpected first page: %+v", page)
	}
	second := perform(handler, http.MethodGet, "/v1/accounts/acct_001/transactions?limit=4&cursor="+*page.NextCursor, testToken)
	if err := json.Unmarshal(second.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 4 || page.Items[0].ID != "tx_005" {
		t.Fatalf("unexpected second page: %+v", page)
	}
	third := perform(handler, http.MethodGet, "/v1/accounts/acct_001/transactions?limit=4&cursor=cursor_8", testToken)
	if err := json.Unmarshal(third.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 4 || page.Items[0].ID != "tx_005" {
		t.Fatalf("duplicate id should appear on following page: %+v", page)
	}
	empty := perform(handler, http.MethodGet, "/v1/accounts/acct_empty/transactions", testToken)
	if err := json.Unmarshal(empty.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 0 || page.NextCursor != nil {
		t.Fatalf("unexpected empty history: %+v", page)
	}
}

func TestTransactionDetailsAndUnknownIDs(t *testing.T) {
	handler := mockRouter("normal")
	res := perform(handler, http.MethodGet, "/v1/transactions/tx_001", testToken)
	if res.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.Code)
	}
	var tx Transaction
	if err := json.Unmarshal(res.Body.Bytes(), &tx); err != nil {
		t.Fatal(err)
	}
	if tx.AccountID != "acct_001" || tx.AmountCents != 1599 || tx.Currency != "USD" {
		t.Fatalf("unexpected transaction details: %+v", tx)
	}
	missing := perform(handler, http.MethodGet, "/v1/transactions/does_not_exist", testToken)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", missing.Code)
	}
}

func TestInvalidRequests(t *testing.T) {
	handler := mockRouter("normal")
	for _, path := range []string{
		"/v1/accounts/acct_001/transactions?limit=0",
		"/v1/accounts/acct_001/transactions?limit=21",
		"/v1/accounts/acct_001/transactions?limit=oops",
		"/v1/accounts/acct_001/transactions?cursor=nope",
		"/v1/accounts/acct_001/transactions?cursor=cursor_3",
	} {
		res := perform(handler, http.MethodGet, path, testToken)
		if res.Code != http.StatusBadRequest {
			t.Errorf("%s: expected 400, got %d", path, res.Code)
		}
	}
	unknown := perform(handler, http.MethodGet, "/v1/accounts/missing/transactions", testToken)
	if unknown.Code != http.StatusNotFound {
		t.Errorf("unknown account: expected 404, got %d", unknown.Code)
	}
}

func TestTransientFailuresThenSuccess(t *testing.T) {
	handler := mockRouter("transient")
	for _, test := range []struct {
		id     string
		status int
	}{{"tx_003", http.StatusServiceUnavailable}, {"tx_004", http.StatusTooManyRequests}} {
		first := perform(handler, http.MethodGet, "/v1/transactions/"+test.id, testToken)
		if first.Code != test.status {
			t.Fatalf("%s first attempt: expected %d, got %d", test.id, test.status, first.Code)
		}
		if test.status == http.StatusTooManyRequests && first.Header().Get("Retry-After") != "1" {
			t.Fatal("429 response missing Retry-After: 1")
		}
		second := perform(handler, http.MethodGet, "/v1/transactions/"+test.id, testToken)
		if second.Code != http.StatusOK {
			t.Fatalf("%s second attempt: expected 200, got %d", test.id, second.Code)
		}
	}
}

func TestResetClearsFailuresAndRequestLogs(t *testing.T) {
	handler := mockRouter("transient")
	first := perform(handler, http.MethodGet, "/v1/transactions/tx_003", testToken)
	if first.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected initial 503, got %d", first.Code)
	}
	logs := perform(handler, http.MethodGet, "/_admin/requests", "")
	var payload struct {
		Items []requestRecord `json:"items"`
	}
	if err := json.Unmarshal(logs.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Items) == 0 || payload.Items[len(payload.Items)-1].Status != http.StatusServiceUnavailable {
		t.Fatalf("request log missing status: %+v", payload.Items)
	}
	reset := perform(handler, http.MethodPost, "/_admin/reset", "")
	if reset.Code != http.StatusOK {
		t.Fatalf("reset expected 200, got %d", reset.Code)
	}
	logs = perform(handler, http.MethodGet, "/_admin/requests", "")
	if err := json.Unmarshal(logs.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Items) != 0 {
		t.Fatalf("reset should clear request log, got %+v", payload.Items)
	}
	again := perform(handler, http.MethodGet, "/v1/transactions/tx_003", testToken)
	if again.Code != http.StatusServiceUnavailable {
		t.Fatalf("reset should restore first-attempt failure, got %d", again.Code)
	}
}
