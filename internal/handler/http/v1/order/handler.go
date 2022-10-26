package order

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maypok86/payment-api/internal/domain/order"
	"github.com/maypok86/payment-api/internal/pkg/handler"
	"go.uber.org/zap"
)

type Service interface {
	CreateOrder(ctx context.Context, dto order.CreateDTO) (order.Order, int64, error)
	PayForOrder(ctx context.Context, id int64) error
	CancelOrder(ctx context.Context, id int64) (int64, error)
}

type Handler struct {
	*handler.BaseHandler
	service Service
	logger  *zap.Logger
}

func NewHandler(service Service, logger *zap.Logger) *Handler {
	return &Handler{
		BaseHandler: handler.NewBaseHandler(logger),
		service:     service,
		logger:      logger,
	}
}

func (h *Handler) InitAPI(router *gin.RouterGroup) {
	orderGroup := router.Group("/order")
	{
		orderGroup.POST("/create", h.createOrder)
		orderGroup.POST("/pay/:order_id", h.payForOrder)
		orderGroup.POST("/cancel/:order_id", h.cancelOrder)
	}
}

type createOrderRequest struct {
	OrderID   int64 `json:"order_id"`
	AccountID int64 `json:"account_id" binding:"required"`
	ServiceID int64 `json:"service_id" binding:"required"`
	Amount    int64 `json:"amount"     binding:"required,gt=0"`
}

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

type createOrderResponse struct {
	Order   orderResponse `json:"order"`
	Balance int64         `json:"balance"`
}

func (h *Handler) createOrder(c *gin.Context) {
	var request createOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Create order error. Invalid request")
		return
	}

	entity, balance, err := h.service.CreateOrder(c, order.CreateDTO{
		OrderID:   request.OrderID,
		AccountID: request.AccountID,
		ServiceID: request.ServiceID,
		Amount:    request.Amount,
	})
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, err, "Create order error")
		return
	}

	c.JSON(http.StatusOK, createOrderResponse{
		Order: orderResponse{
			OrderID:     entity.OrderID,
			AccountID:   entity.AccountID,
			ServiceID:   entity.ServiceID,
			Amount:      entity.Amount,
			IsPaid:      entity.IsPaid,
			IsCancelled: entity.IsCancelled,
			CreatedAt:   entity.CreatedAt,
			UpdatedAt:   entity.UpdatedAt,
		},
		Balance: balance,
	})
}

func (h *Handler) payForOrder(c *gin.Context) {
	orderID, err := h.ParseIDFromPath(c, "order_id")
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Pay for order error. Invalid order_id")
		return
	}

	if err := h.service.PayForOrder(c, orderID); err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, err, "Pay for order error")
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) cancelOrder(c *gin.Context) {
	orderID, err := h.ParseIDFromPath(c, "order_id")
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Cancel order error. Invalid order_id")
		return
	}

	balance, err := h.service.CancelOrder(c, orderID)
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, err, "Cancel order error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": balance,
	})
}
