// Package handler is the Vercel serverless entry point.
// It imports only the public `server` package so that Go's internal-package
// visibility rule is not triggered (Vercel compiles this file under a
// synthetic path "handler/api", outside the module root).
package handler

import (
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/nsalunke729/go-grpc-gateway/server"
)

var (
	once    sync.Once
	h       http.Handler
)

// Handler is the Vercel-compatible entrypoint called for every request.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		rateLimit, _ := strconv.ParseFloat(getenv("RATE_LIMIT", "100"), 64)
		rateBurst, _ := strconv.Atoi(getenv("RATE_BURST", "200"))
		h = server.NewHandler(
			getenv("JWT_SECRET", "dev-secret-change-me"),
			rateLimit,
			rateBurst,
		)
	})
	h.ServeHTTP(w, r)
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
