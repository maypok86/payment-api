package order

import "github.com/maypok86/payment-api/internal/domain/order"

type CreateOrderRequest struct {
	OrderID   int64 `json:"order_id"   binding:"required,gte=1"`
	AccountID int64 `json:"account_id" binding:"required,gte=1"`
	ServiceID int64 `json:"service_id" binding:"required,gte=1"`
	Amount    int64 `json:"amount"     binding:"required,gt=0"`
}

func (r CreateOrderRequest) ToDTO() order.CreateDTO {
	return order.CreateDTO{
		OrderID:   r.OrderID,
		AccountID: r.AccountID,
		ServiceID: r.ServiceID,
		Amount:    r.Amount,
	}
}

type PayForOrderRequest struct {
	OrderID   int64 `json:"order_id"   binding:"required,gte=1"`
	AccountID int64 `json:"account_id" binding:"required,gte=1"`
	ServiceID int64 `json:"service_id" binding:"required,gte=1"`
	Amount    int64 `json:"amount"     binding:"required,gt=0"`
}

func (r PayForOrderRequest) ToDTO() order.PayForDTO {
	return order.PayForDTO{
		OrderID:   r.OrderID,
		AccountID: r.AccountID,
		ServiceID: r.ServiceID,
		Amount:    r.Amount,
	}
}

type CancelOrderRequest struct {
	OrderID   int64 `json:"order_id"   binding:"required,gte=1"`
	AccountID int64 `json:"account_id" binding:"required,gte=1"`
	ServiceID int64 `json:"service_id" binding:"required,gte=1"`
	Amount    int64 `json:"amount"     binding:"required,gt=0"`
}

func (r CancelOrderRequest) ToDTO() order.CancelDTO {
	return order.CancelDTO{
		OrderID:   r.OrderID,
		AccountID: r.AccountID,
		ServiceID: r.ServiceID,
		Amount:    r.Amount,
	}
}
