// Package embed provides in-process adapters that implement the gRPC client
// interfaces by calling the server implementations directly — no network hop.
// Used by the Vercel serverless handler where persistent gRPC connections
// between processes are not viable.
package embed

import (
	"context"

	"google.golang.org/grpc"

	orderpb "github.com/nsalunke729/go-grpc-gateway/internal/pb/order"
	userpb "github.com/nsalunke729/go-grpc-gateway/internal/pb/user"
	"github.com/nsalunke729/go-grpc-gateway/internal/ordersvc"
	"github.com/nsalunke729/go-grpc-gateway/internal/usersvc"
)

// UserAdapter wraps a UserServiceServer and satisfies UserServiceClient.
type UserAdapter struct{ srv usersvc.UserServiceServer }

func NewUserAdapter(srv usersvc.UserServiceServer) usersvc.UserServiceClient {
	return &UserAdapter{srv}
}

func (a *UserAdapter) GetUser(ctx context.Context, req *userpb.GetUserRequest, _ ...grpc.CallOption) (*userpb.GetUserResponse, error) {
	return a.srv.GetUser(ctx, req)
}
func (a *UserAdapter) CreateUser(ctx context.Context, req *userpb.CreateUserRequest, _ ...grpc.CallOption) (*userpb.CreateUserResponse, error) {
	return a.srv.CreateUser(ctx, req)
}

// OrderAdapter wraps an OrderServiceServer and satisfies OrderServiceClient.
type OrderAdapter struct{ srv ordersvc.OrderServiceServer }

func NewOrderAdapter(srv ordersvc.OrderServiceServer) ordersvc.OrderServiceClient {
	return &OrderAdapter{srv}
}

func (a *OrderAdapter) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest, _ ...grpc.CallOption) (*orderpb.GetOrderResponse, error) {
	return a.srv.GetOrder(ctx, req)
}
func (a *OrderAdapter) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest, _ ...grpc.CallOption) (*orderpb.ListOrdersResponse, error) {
	return a.srv.ListOrders(ctx, req)
}
func (a *OrderAdapter) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest, _ ...grpc.CallOption) (*orderpb.CreateOrderResponse, error) {
	return a.srv.CreateOrder(ctx, req)
}
