package usersvc

import (
	"context"

	"google.golang.org/grpc"

	userpb "github.com/nsalunke729/go-grpc-gateway/internal/pb/user"
)

// UserServiceClient is the client-side interface for the user gRPC service.
type UserServiceClient interface {
	GetUser(ctx context.Context, req *userpb.GetUserRequest, opts ...grpc.CallOption) (*userpb.GetUserResponse, error)
	CreateUser(ctx context.Context, req *userpb.CreateUserRequest, opts ...grpc.CallOption) (*userpb.CreateUserResponse, error)
}

type userServiceClient struct {
	cc grpc.ClientConnInterface
}

// NewClient wraps a gRPC connection and returns a UserServiceClient.
func NewClient(cc grpc.ClientConnInterface) UserServiceClient {
	return &userServiceClient{cc}
}

func (c *userServiceClient) GetUser(ctx context.Context, req *userpb.GetUserRequest, opts ...grpc.CallOption) (*userpb.GetUserResponse, error) {
	out := new(userpb.GetUserResponse)
	if err := c.cc.Invoke(ctx, "/user.UserService/GetUser", req, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *userServiceClient) CreateUser(ctx context.Context, req *userpb.CreateUserRequest, opts ...grpc.CallOption) (*userpb.CreateUserResponse, error) {
	out := new(userpb.CreateUserResponse)
	if err := c.cc.Invoke(ctx, "/user.UserService/CreateUser", req, out, opts...); err != nil {
		return nil, err
	}
	return out, nil
}
