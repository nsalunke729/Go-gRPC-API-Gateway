package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "gateway_http_requests_total",
		Help: "Total HTTP requests handled by the gateway, labelled by method, route, and status.",
	}, []string{"method", "route", "status"})

	httpDurationSeconds = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "gateway_http_request_duration_seconds",
		Help:    "HTTP request latency in seconds.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route", "status"})
)

// Metrics is a chi-compatible middleware that records per-route Prometheus metrics.
// It must be added to the router AFTER the chi routing is set up so that
// chi.RouteContext can resolve the route pattern on the way out.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
		start := time.Now()

		next.ServeHTTP(ww, r)

		route := chi.RouteContext(r.Context()).RoutePattern()
		if route == "" {
			route = "unmatched"
		}
		statusStr := strconv.Itoa(ww.Status())
		elapsed := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(r.Method, route, statusStr).Inc()
		httpDurationSeconds.WithLabelValues(r.Method, route, statusStr).Observe(elapsed)
	})
}
