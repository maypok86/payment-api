package account

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maypok86/payment-api/internal/domain/account"
	"github.com/maypok86/payment-api/internal/pkg/handler"
	"go.uber.org/zap"
)

//go:generate mockgen -source=handler.go -destination=mock_test.go -package=account_test

type Service interface {
	GetBalanceByID(ctx context.Context, id int64) (int64, error)
	AddBalance(ctx context.Context, dto account.AddBalanceDTO) (int64, error)
	TransferBalance(ctx context.Context, dto account.TransferBalanceDTO) (int64, int64, error)
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
	balanceGroup := router.Group("/balance")
	{
		balanceGroup.GET("/:account_id", h.GetBalance)
		balanceGroup.POST("/add", h.AddBalance)
		balanceGroup.POST("/transfer", h.TransferBalance)
	}
}

func (h *Handler) GetBalance(c *gin.Context) {
	accountID, err := h.ParseIDFromPath(c, "account_id")
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Balance not found. id is not valid")
		return
	}

	balance, err := h.service.GetBalanceByID(c.Request.Context(), accountID)
	if err != nil {
		if errors.Is(err, account.ErrNotFound) {
			h.ErrorResponse(c, http.StatusNotFound, err, "Get balance by error. Account not found")
			return
		}

		h.ErrorResponse(c, http.StatusInternalServerError, err, "Get balance by id error")
		return
	}

	c.JSON(http.StatusOK, GetBalanceResponse{
		Balance: balance,
	})
}

func (h *Handler) AddBalance(c *gin.Context) {
	var request AddBalanceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Amount not added. request is not valid")
		return
	}

	balance, err := h.service.AddBalance(c.Request.Context(), request.ToDTO())
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, err, "Add balance error")
		return
	}

	c.JSON(http.StatusOK, AddBalanceResponse{
		Balance: balance,
	})
}

func (h *Handler) TransferBalance(c *gin.Context) {
	var request TransferBalanceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Amount not transferred. request is not valid")
		return
	}

	senderBalance, receiverBalance, err := h.service.TransferBalance(c.Request.Context(), request.ToDTO())
	if err != nil {
		if errors.Is(err, account.ErrNotFound) {
			h.ErrorResponse(c, http.StatusNotFound, err, "Sender or receiver not found")
			return
		}

		h.ErrorResponse(c, http.StatusInternalServerError, err, "Transfer balance error")
		return
	}

	c.JSON(http.StatusOK, TransferBalanceResponse{
		SenderBalance:   senderBalance,
		ReceiverBalance: receiverBalance,
	})
}
