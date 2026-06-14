package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	orderpb "github.com/nsalunke729/go-grpc-gateway/internal/pb/order"
	"github.com/nsalunke729/go-grpc-gateway/internal/ordersvc"
)

// OrderHandler translates REST requests into OrderService gRPC calls.
type OrderHandler struct {
	client ordersvc.OrderServiceClient
	log    *zap.Logger
}

func NewOrderHandler(client ordersvc.OrderServiceClient, log *zap.Logger) *OrderHandler {
	return &OrderHandler{client: client, log: log}
}

// GET /orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.client.GetOrder(r.Context(), &orderpb.GetOrderRequest{OrderID: id})
	if err != nil {
		h.log.Warn("GetOrder upstream error", zap.String("order_id", id), zap.Error(err))
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// GET /users/{userID}/orders
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	resp, err := h.client.ListOrders(r.Context(), &orderpb.ListOrdersRequest{UserID: userID})
	if err != nil {
		h.log.Warn("ListOrders upstream error", zap.String("user_id", userID), zap.Error(err))
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// POST /orders
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var req orderpb.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	resp, err := h.client.CreateOrder(r.Context(), &req)
	if err != nil {
		h.log.Warn("CreateOrder upstream error", zap.Error(err))
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}
