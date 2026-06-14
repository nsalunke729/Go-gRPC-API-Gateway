package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go.uber.org/zap"

	_ "github.com/nsalunke729/go-grpc-gateway/internal/codec" // register JSON gRPC codec
	"github.com/nsalunke729/go-grpc-gateway/internal/gateway"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync() //nolint:errcheck

	rateLimit, _ := strconv.ParseFloat(getenv("RATE_LIMIT", "100"), 64)
	rateBurst, _ := strconv.Atoi(getenv("RATE_BURST", "200"))

	cfg := gateway.Config{
		Addr:         getenv("GATEWAY_ADDR", ":8080"),
		JWTSecret:    getenv("JWT_SECRET", "dev-secret-change-me"),
		UserSvcAddr:  getenv("USER_SVC_ADDR", "localhost:9001"),
		OrderSvcAddr: getenv("ORDER_SVC_ADDR", "localhost:9002"),
		RateLimit:    rateLimit,
		RateBurst:    rateBurst,
	}

	srv, err := gateway.New(cfg, log)
	if err != nil {
		log.Fatal("failed to build gateway", zap.Error(err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("gateway error", zap.Error(err))
		}
	}()

	<-quit
	log.Info("shutting down gateway")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error", zap.Error(err))
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
