package order

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maypok86/payment-api/internal/domain/account"
	"github.com/maypok86/payment-api/internal/domain/order"
	"github.com/maypok86/payment-api/internal/pkg/handler"
	"go.uber.org/zap"
)

//go:generate mockgen -source=handler.go -destination=mock_test.go -package=order_test

type Service interface {
	CreateOrder(ctx context.Context, dto order.CreateDTO) (order.Order, int64, error)
	PayForOrder(ctx context.Context, dto order.PayForDTO) error
	CancelOrder(ctx context.Context, dto order.CancelDTO) (int64, error)
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
		orderGroup.POST("/create", h.CreateOrder)
		orderGroup.POST("/pay", h.PayForOrder)
		orderGroup.POST("/cancel", h.CancelOrder)
	}
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var request CreateOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Create order error. Invalid request")
		return
	}

	entity, balance, err := h.service.CreateOrder(c.Request.Context(), request.ToDTO())
	if err != nil {
		switch {
		case errors.Is(err, account.ErrNotFound):
			h.ErrorResponse(c, http.StatusNotFound, err, "Create order error. Account not found")
			return
		case errors.Is(err, order.ErrAlreadyExist):
			h.ErrorResponse(c, http.StatusConflict, err, "Create order error. Order already exist")
			return
		}

		h.ErrorResponse(c, http.StatusInternalServerError, err, "Create order error")
		return
	}

	c.JSON(http.StatusOK, NewCreateOrderResponse(entity, balance))
}

func (h *Handler) PayForOrder(c *gin.Context) {
	var request PayForOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Pay for order error. Invalid request")
		return
	}

	if err := h.service.PayForOrder(c.Request.Context(), request.ToDTO()); err != nil {
		if errors.Is(err, order.ErrNotFound) {
			h.ErrorResponse(c, http.StatusNotFound, err, "Pay for order error. Order not found")
			return
		}

		h.ErrorResponse(c, http.StatusInternalServerError, err, "Pay for order error")
		return
	}

	c.Status(http.StatusOK)
}

func (h *Handler) CancelOrder(c *gin.Context) {
	var request CancelOrderRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Cancel order error. Invalid request")
		return
	}

	balance, err := h.service.CancelOrder(c.Request.Context(), request.ToDTO())
	if err != nil {
		switch {
		case errors.Is(err, order.ErrNotFound):
			h.ErrorResponse(c, http.StatusNotFound, err, "Cancel order error. Order not found")
			return
		case errors.Is(err, account.ErrNotFound):
			h.ErrorResponse(c, http.StatusNotFound, err, "Cancel order error. Account not found")
			return
		}

		h.ErrorResponse(c, http.StatusInternalServerError, err, "Cancel order error")
		return
	}

	c.JSON(http.StatusOK, CancelOrderResponse{
		Balance: balance,
	})
}
