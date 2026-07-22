package httpapi

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type RouterOptions struct {
	Environment string
	Version     string
	Commit      string
	Logger      *slog.Logger
}

func NewRouter(options RouterOptions) http.Handler {
	if options.Environment == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(requestLogger(options.Logger))

	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/readyz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	api := router.Group("/api/v1")
	{
		api.GET("/greetings", func(c *gin.Context) {
			name := strings.TrimSpace(c.Query("name"))
			if name == "" {
				name = "world"
			}

			hostname, err := os.Hostname()
			if err != nil {
				hostname = "unknown"
			}

			c.JSON(http.StatusOK, gin.H{
				"message":     "Hello, " + name + "!",
				"environment": options.Environment,
				"hostname":    hostname,
				"version":     options.Version,
				"commit":      options.Commit,
			})
		})
	}

	return router
}

func requestLogger(logger *slog.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = slog.Default()
	}

	return func(c *gin.Context) {
		startedAt := time.Now()
		c.Next()
		logger.Info("request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency_ms", time.Since(startedAt).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
