package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	_ "github.com/nsalunke729/go-grpc-gateway/internal/codec" // register JSON gRPC codec
	"github.com/nsalunke729/go-grpc-gateway/internal/ordersvc"
)

func main() {
	log, _ := zap.NewProduction()
	defer log.Sync() //nolint:errcheck

	addr := getenv("ORDER_SVC_ADDR", ":9002")
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal("failed to listen", zap.String("addr", addr), zap.Error(err))
	}

	s := grpc.NewServer()
	s.RegisterService(&ordersvc.ServiceDesc, ordersvc.NewServer())

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info("order-svc listening", zap.String("addr", addr))
		if err := s.Serve(lis); err != nil {
			log.Fatal("order-svc error", zap.Error(err))
		}
	}()

	<-quit
	log.Info("shutting down order-svc")
	s.GracefulStop()
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
