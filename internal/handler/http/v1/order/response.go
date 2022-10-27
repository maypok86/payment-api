package order

import (
	"time"

	"github.com/maypok86/payment-api/internal/domain/order"
)

type Response struct {
	OrderID     int64     `json:"order_id"`
	AccountID   int64     `json:"account_id"`
	ServiceID   int64     `json:"service_id"`
	Amount      int64     `json:"amount"`
	IsPaid      bool      `json:"is_paid"`
	IsCancelled bool      `json:"is_cancelled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func NewResponse(entity order.Order) Response {
	return Response{
		OrderID:     entity.OrderID,
		AccountID:   entity.AccountID,
		ServiceID:   entity.ServiceID,
		Amount:      entity.Amount,
		IsPaid:      entity.IsPaid,
		IsCancelled: entity.IsCancelled,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}

type CreateOrderResponse struct {
	Order   Response `json:"order"`
	Balance int64    `json:"balance"`
}

func NewCreateOrderResponse(entity order.Order, balance int64) CreateOrderResponse {
	return CreateOrderResponse{
		Order:   NewResponse(entity),
		Balance: balance,
	}
}

type CancelOrderResponse struct {
	Balance int64 `json:"balance"`
}
