// Package gateway wires the chi router, middleware, and gRPC clients together
// into an HTTP server that fronts the backend microservices.
package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/nsalunke729/go-grpc-gateway/internal/gateway/handlers"
	"github.com/nsalunke729/go-grpc-gateway/internal/gateway/middleware"
	"github.com/nsalunke729/go-grpc-gateway/internal/gateway/ui"
	"github.com/nsalunke729/go-grpc-gateway/internal/ordersvc"
	"github.com/nsalunke729/go-grpc-gateway/internal/usersvc"
)

// Config holds all gateway configuration derived from environment variables.
type Config struct {
	Addr         string
	JWTSecret    string
	UserSvcAddr  string
	OrderSvcAddr string
	// RateLimit is the per-client token-bucket steady-state rate (req/sec).
	RateLimit float64
	// RateBurst is the maximum burst size per client.
	RateBurst int
}

// Server is an HTTP server that acts as the API gateway.
type Server struct {
	httpSrv *http.Server
	log     *zap.Logger
}

// New dials the backend gRPC services and constructs the gateway HTTP server.
func New(cfg Config, log *zap.Logger) (*Server, error) {
	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())} //nolint:staticcheck

	userConn, err := grpc.Dial(cfg.UserSvcAddr, dialOpts...) //nolint:staticcheck
	if err != nil {
		return nil, fmt.Errorf("dial user-svc at %s: %w", cfg.UserSvcAddr, err)
	}
	orderConn, err := grpc.Dial(cfg.OrderSvcAddr, dialOpts...) //nolint:staticcheck
	if err != nil {
		return nil, fmt.Errorf("dial order-svc at %s: %w", cfg.OrderSvcAddr, err)
	}

	return NewWithClients(cfg, usersvc.NewClient(userConn), ordersvc.NewClient(orderConn), log), nil
}

// NewWithClients builds a gateway using pre-wired service clients.
// Used in serverless deployments (e.g. Vercel) where a separate gRPC dial
// is not viable — callers pass in-process adapters instead.
func NewWithClients(cfg Config, userClient usersvc.UserServiceClient, orderClient ordersvc.OrderServiceClient, log *zap.Logger) *Server {
	rl := middleware.NewRateLimiter(rate.Limit(cfg.RateLimit), cfg.RateBurst, log)
	authMW := middleware.Auth(cfg.JWTSecret, log)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(zapRequestLogger(log))
	r.Use(rl.Middleware)

	// Landing page / API playground
	r.Get("/", ui.Handler())

	// Health check — no auth required
	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`)) //nolint:errcheck
	})

	// Demo token — issues a short-lived JWT for the playground; no auth required
	r.Get("/demo/token", func(w http.ResponseWriter, r *http.Request) {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub": "demo-user",
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
		})
		signed, err := token.SignedString([]byte(cfg.JWTSecret))
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": signed}) //nolint:errcheck
	})

	// Authenticated routes
	r.Group(func(r chi.Router) {
		r.Use(authMW)

		userH := handlers.NewUserHandler(userClient, log)
		r.Get("/users/{id}", userH.GetUser)
		r.Post("/users", userH.CreateUser)

		orderH := handlers.NewOrderHandler(orderClient, log)
		r.Get("/orders/{id}", orderH.GetOrder)
		r.Get("/users/{userID}/orders", orderH.ListOrders)
		r.Post("/orders", orderH.CreateOrder)
	})

	return &Server{
		httpSrv: &http.Server{
			Addr:         cfg.Addr,
			Handler:      r,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		log: log,
	}
}

// HTTPHandler returns the underlying http.Handler, used by the Vercel serverless adapter.
func (s *Server) HTTPHandler() http.Handler { return s.httpSrv.Handler }

// Start begins serving HTTP traffic. It blocks until the server stops.
func (s *Server) Start() error {
	s.log.Info("gateway listening", zap.String("addr", s.httpSrv.Addr))
	return s.httpSrv.ListenAndServe()
}

// Shutdown gracefully drains in-flight requests within the given context deadline.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpSrv.Shutdown(ctx)
}

// zapRequestLogger is a chi-compatible middleware that emits one structured log
// line per request after the response is sent.
func zapRequestLogger(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			log.Info("http",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.Status()),
				zap.Duration("latency", time.Since(start)),
				zap.String("request_id", chimw.GetReqID(r.Context())),
			)
		})
	}
}
