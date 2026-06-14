// Package ordersvc provides the OrderService gRPC server, service descriptor,
// and client — structured identically to usersvc; see internal/codec for the
// JSON codec that removes the protoc dependency.
package ordersvc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	orderpb "github.com/nsalunke729/go-grpc-gateway/internal/pb/order"
)

// OrderServiceServer is the interface a gRPC server must implement.
type OrderServiceServer interface {
	GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error)
	ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error)
	CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error)
}

// ServiceDesc mirrors what protoc-gen-go-grpc would generate.
var ServiceDesc = grpc.ServiceDesc{
	ServiceName: "order.OrderService",
	HandlerType: (*OrderServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "GetOrder", Handler: getOrderHandler},
		{MethodName: "ListOrders", Handler: listOrdersHandler},
		{MethodName: "CreateOrder", Handler: createOrderHandler},
	},
	Streams: []grpc.StreamDesc{},
}

func getOrderHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(orderpb.GetOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServiceServer).GetOrder(ctx, in)
	}
	return interceptor(ctx, in,
		&grpc.UnaryServerInfo{Server: srv, FullMethod: "/order.OrderService/GetOrder"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.(OrderServiceServer).GetOrder(ctx, req.(*orderpb.GetOrderRequest))
		})
}

func listOrdersHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(orderpb.ListOrdersRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServiceServer).ListOrders(ctx, in)
	}
	return interceptor(ctx, in,
		&grpc.UnaryServerInfo{Server: srv, FullMethod: "/order.OrderService/ListOrders"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.(OrderServiceServer).ListOrders(ctx, req.(*orderpb.ListOrdersRequest))
		})
}

func createOrderHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(orderpb.CreateOrderRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(OrderServiceServer).CreateOrder(ctx, in)
	}
	return interceptor(ctx, in,
		&grpc.UnaryServerInfo{Server: srv, FullMethod: "/order.OrderService/CreateOrder"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.(OrderServiceServer).CreateOrder(ctx, req.(*orderpb.CreateOrderRequest))
		})
}

// server is an in-memory OrderService implementation.
type server struct {
	mu     sync.RWMutex
	orders map[string]*orderpb.Order
}

// NewServer returns a ready-to-register OrderService implementation.
func NewServer() OrderServiceServer {
	return &server{orders: make(map[string]*orderpb.Order)}
}

func (s *server) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest) (*orderpb.GetOrderResponse, error) {
	if req.OrderID == "" {
		return nil, status.Error(codes.InvalidArgument, "order_id is required")
	}
	s.mu.RLock()
	o, ok := s.orders[req.OrderID]
	s.mu.RUnlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "order %q not found", req.OrderID)
	}
	return &orderpb.GetOrderResponse{Order: o}, nil
}

func (s *server) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest) (*orderpb.ListOrdersResponse, error) {
	if req.UserID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var orders []*orderpb.Order
	for _, o := range s.orders {
		if o.UserID == req.UserID {
			orders = append(orders, o)
		}
	}
	return &orderpb.ListOrdersResponse{Orders: orders}, nil
}

func (s *server) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest) (*orderpb.CreateOrderResponse, error) {
	if req.UserID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	if req.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be positive")
	}
	o := &orderpb.Order{
		ID:        fmt.Sprintf("ord_%d", time.Now().UnixNano()),
		UserID:    req.UserID,
		Amount:    req.Amount,
		Status:    "pending",
		CreatedAt: time.Now().Unix(),
	}
	s.mu.Lock()
	s.orders[o.ID] = o
	s.mu.Unlock()
	return &orderpb.CreateOrderResponse{Order: o}, nil
}
