package order

import (
	"time"

	"github.com/maypok86/payment-api/internal/domain/order"
)

type orderResponse struct {
	OrderID     int64     `json:"order_id"`
	AccountID   int64     `json:"account_id"`
	ServiceID   int64     `json:"service_id"`
	Amount      int64     `json:"amount"`
	IsPaid      bool      `json:"is_paid"`
	IsCancelled bool      `json:"is_cancelled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func newOrderResponse(entity order.Order) orderResponse {
	return orderResponse{
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

type createOrderResponse struct {
	Order   orderResponse `json:"order"`
	Balance int64         `json:"balance"`
}

func newCreateOrderResponse(entity order.Order, balance int64) createOrderResponse {
	return createOrderResponse{
		Order:   newOrderResponse(entity),
		Balance: balance,
	}
}
