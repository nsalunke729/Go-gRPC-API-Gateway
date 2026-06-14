package orderpb

// Order is the core domain model returned by the order service.
type Order struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
	CreatedAt int64   `json:"created_at"`
}

type GetOrderRequest struct {
	OrderID string `json:"order_id"`
}

type GetOrderResponse struct {
	Order *Order `json:"order"`
}

type ListOrdersRequest struct {
	UserID string `json:"user_id"`
}

type ListOrdersResponse struct {
	Orders []*Order `json:"orders"`
}

type CreateOrderRequest struct {
	UserID string  `json:"user_id"`
	Amount float64 `json:"amount"`
}

type CreateOrderResponse struct {
	Order *Order `json:"order"`
}
