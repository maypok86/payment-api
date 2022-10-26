package order

import "github.com/maypok86/payment-api/internal/domain/order"

type createOrderRequest struct {
	OrderID   int64 `json:"order_id"   binding:"required"`
	AccountID int64 `json:"account_id" binding:"required"`
	ServiceID int64 `json:"service_id" binding:"required"`
	Amount    int64 `json:"amount"     binding:"required,gt=0"`
}

func (r createOrderRequest) toDTO() order.CreateDTO {
	return order.CreateDTO{
		OrderID:   r.OrderID,
		AccountID: r.AccountID,
		ServiceID: r.ServiceID,
		Amount:    r.Amount,
	}
}

type payForOrderRequest struct {
	OrderID   int64 `json:"order_id"   binding:"required"`
	AccountID int64 `json:"account_id" binding:"required"`
	ServiceID int64 `json:"service_id" binding:"required"`
	Amount    int64 `json:"amount"     binding:"required,gt=0"`
}

func (r payForOrderRequest) toDTO() order.PayForDTO {
	return order.PayForDTO{
		OrderID:   r.OrderID,
		AccountID: r.AccountID,
		ServiceID: r.ServiceID,
		Amount:    r.Amount,
	}
}

type cancelOrderRequest struct {
	OrderID   int64 `json:"order_id"   binding:"required"`
	AccountID int64 `json:"account_id" binding:"required"`
	ServiceID int64 `json:"service_id" binding:"required"`
	Amount    int64 `json:"amount"     binding:"required,gt=0"`
}

func (r cancelOrderRequest) toDTO() order.CancelDTO {
	return order.CancelDTO{
		OrderID:   r.OrderID,
		AccountID: r.AccountID,
		ServiceID: r.ServiceID,
		Amount:    r.Amount,
	}
}
