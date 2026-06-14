// Package usersvc provides the UserService gRPC server implementation,
// service descriptor, and client — written by hand in lieu of protoc generation
// because our JSON codec (internal/codec) avoids a protoc dependency.
package usersvc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	userpb "github.com/nsalunke729/go-grpc-gateway/internal/pb/user"
)

// UserServiceServer is the interface a gRPC server must implement.
type UserServiceServer interface {
	GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error)
	CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error)
}

// ServiceDesc mirrors what protoc-gen-go-grpc would generate.
var ServiceDesc = grpc.ServiceDesc{
	ServiceName: "user.UserService",
	HandlerType: (*UserServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "GetUser", Handler: getUserHandler},
		{MethodName: "CreateUser", Handler: createUserHandler},
	},
	Streams: []grpc.StreamDesc{},
}

func getUserHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(userpb.GetUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).GetUser(ctx, in)
	}
	return interceptor(ctx, in,
		&grpc.UnaryServerInfo{Server: srv, FullMethod: "/user.UserService/GetUser"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.(UserServiceServer).GetUser(ctx, req.(*userpb.GetUserRequest))
		})
}

func createUserHandler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(userpb.CreateUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(UserServiceServer).CreateUser(ctx, in)
	}
	return interceptor(ctx, in,
		&grpc.UnaryServerInfo{Server: srv, FullMethod: "/user.UserService/CreateUser"},
		func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.(UserServiceServer).CreateUser(ctx, req.(*userpb.CreateUserRequest))
		})
}

// server is an in-memory UserService implementation.
type server struct {
	mu    sync.RWMutex
	users map[string]*userpb.User
}

// NewServer returns a ready-to-register UserService implementation.
func NewServer() UserServiceServer {
	return &server{users: make(map[string]*userpb.User)}
}

func (s *server) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.GetUserResponse, error) {
	if req.UserID == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	s.mu.RLock()
	u, ok := s.users[req.UserID]
	s.mu.RUnlock()
	if !ok {
		return nil, status.Errorf(codes.NotFound, "user %q not found", req.UserID)
	}
	return &userpb.GetUserResponse{User: u}, nil
}

func (s *server) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	if req.Name == "" || req.Email == "" {
		return nil, status.Error(codes.InvalidArgument, "name and email are required")
	}
	u := &userpb.User{
		ID:        fmt.Sprintf("usr_%d", time.Now().UnixNano()),
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: time.Now().Unix(),
	}
	s.mu.Lock()
	s.users[u.ID] = u
	s.mu.Unlock()
	return &userpb.CreateUserResponse{User: u}, nil
}
