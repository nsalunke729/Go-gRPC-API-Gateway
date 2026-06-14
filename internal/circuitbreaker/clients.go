package circuitbreaker

import (
	"context"

	"google.golang.org/grpc"

	orderpb "github.com/nsalunke729/go-grpc-gateway/internal/pb/order"
	userpb "github.com/nsalunke729/go-grpc-gateway/internal/pb/user"
	"github.com/nsalunke729/go-grpc-gateway/internal/ordersvc"
	"github.com/nsalunke729/go-grpc-gateway/internal/usersvc"
)

// ── User service ──────────────────────────────────────────────────────────────

type userClientBreaker struct {
	inner usersvc.UserServiceClient
	b     *Breaker
}

// WrapUserClient returns a UserServiceClient whose calls run through b.
func WrapUserClient(c usersvc.UserServiceClient, b *Breaker) usersvc.UserServiceClient {
	return &userClientBreaker{inner: c, b: b}
}

func (u *userClientBreaker) GetUser(ctx context.Context, req *userpb.GetUserRequest, opts ...grpc.CallOption) (*userpb.GetUserResponse, error) {
	var resp *userpb.GetUserResponse
	err := u.b.Do(func() error {
		var e error
		resp, e = u.inner.GetUser(ctx, req, opts...)
		return e
	})
	return resp, err
}

func (u *userClientBreaker) CreateUser(ctx context.Context, req *userpb.CreateUserRequest, opts ...grpc.CallOption) (*userpb.CreateUserResponse, error) {
	var resp *userpb.CreateUserResponse
	err := u.b.Do(func() error {
		var e error
		resp, e = u.inner.CreateUser(ctx, req, opts...)
		return e
	})
	return resp, err
}

// ── Order service ─────────────────────────────────────────────────────────────

type orderClientBreaker struct {
	inner ordersvc.OrderServiceClient
	b     *Breaker
}

// WrapOrderClient returns an OrderServiceClient whose calls run through b.
func WrapOrderClient(c ordersvc.OrderServiceClient, b *Breaker) ordersvc.OrderServiceClient {
	return &orderClientBreaker{inner: c, b: b}
}

func (o *orderClientBreaker) GetOrder(ctx context.Context, req *orderpb.GetOrderRequest, opts ...grpc.CallOption) (*orderpb.GetOrderResponse, error) {
	var resp *orderpb.GetOrderResponse
	err := o.b.Do(func() error {
		var e error
		resp, e = o.inner.GetOrder(ctx, req, opts...)
		return e
	})
	return resp, err
}

func (o *orderClientBreaker) ListOrders(ctx context.Context, req *orderpb.ListOrdersRequest, opts ...grpc.CallOption) (*orderpb.ListOrdersResponse, error) {
	var resp *orderpb.ListOrdersResponse
	err := o.b.Do(func() error {
		var e error
		resp, e = o.inner.ListOrders(ctx, req, opts...)
		return e
	})
	return resp, err
}

func (o *orderClientBreaker) CreateOrder(ctx context.Context, req *orderpb.CreateOrderRequest, opts ...grpc.CallOption) (*orderpb.CreateOrderResponse, error) {
	var resp *orderpb.CreateOrderResponse
	err := o.b.Do(func() error {
		var e error
		resp, e = o.inner.CreateOrder(ctx, req, opts...)
		return e
	})
	return resp, err
}
