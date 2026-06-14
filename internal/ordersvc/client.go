package ordersvc

import (
	"context"

	"google.golang.org/grpc"

	orderpb "github.com/nsalunke729/go-grpc-gateway/internal/pb/order"
)

// OrderServiceClient is the client-side interface for the order gRPC service.
type OrderServiceClient interface {
	GetOrder(ctx context.Context, req *orderpb.GetOrderRequest, opts ...grpc.CallOption) (*orderpb.GetOrderResponse, error)
	ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest, opts ...grpc.CallOption) (*orderpb.ListOrdersResponse, error)
	CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest, opts ...grpc.CallOption) (*orderpb.CreateOrderResponse, error)
}

type orderServiceClient struct {
	cc grpc.ClientConnInterface
}

// NewClient wraps a gRPC connection and returns an OrderServiceClient.
func NewClient(cc grpc.ClientConnInterface) OrderServiceClient {
	return &orderServiceClient{cc}
}

func (c *orderServiceClient) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest, opts ...grpc.CallOption) (*orderpb.GetOrderResponse, error) {
	out := new(orderpb.GetOrderResponse)
	if err := c.cc.Invoke(ctx, "/order.OrderService/GetOrder", req, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderServiceClient) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest, opts ...grpc.CallOption) (*orderpb.ListOrdersResponse, error) {
	out := new(orderpb.ListOrdersResponse)
	if err := c.cc.Invoke(ctx, "/order.OrderService/ListOrders", req, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *orderServiceClient) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest, opts ...grpc.CallOption) (*orderpb.CreateOrderResponse, error) {
	out := new(orderpb.CreateOrderResponse)
	if err := c.cc.Invoke(ctx, "/order.OrderService/CreateOrder", req, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}
