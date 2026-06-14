package ui

import (
	_ "embed"
	"net/http"
)

//go:embed index.html
var page []byte

// Handler serves the landing page / API playground.
func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(page) //nolint:errcheck
	}
}
