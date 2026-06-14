package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	userpb "github.com/nsalunke729/go-grpc-gateway/internal/pb/user"
	"github.com/nsalunke729/go-grpc-gateway/internal/usersvc"
)

// UserHandler translates REST requests into UserService gRPC calls.
type UserHandler struct {
	client usersvc.UserServiceClient
	log    *zap.Logger
}

func NewUserHandler(client usersvc.UserServiceClient, log *zap.Logger) *UserHandler {
	return &UserHandler{client: client, log: log}
}

// GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.client.GetUser(r.Context(), &userpb.GetUserRequest{UserID: id})
	if err != nil {
		h.log.Warn("GetUser upstream error", zap.String("user_id", id), zap.Error(err))
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// POST /users
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req userpb.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	resp, err := h.client.CreateUser(r.Context(), &req)
	if err != nil {
		h.log.Warn("CreateUser upstream error", zap.Error(err))
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}
