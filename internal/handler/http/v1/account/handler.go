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
		balanceGroup.GET("/:account_id", h.getBalance)
		balanceGroup.POST("/", h.addBalance)
		balanceGroup.POST("/transfer", h.transferBalance)
	}
}

func (h *Handler) getBalance(c *gin.Context) {
	accountID, err := h.ParseIDFromPath(c, "account_id")
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Balance not found. id is not valid")
		return
	}

	balance, err := h.service.GetBalanceByID(c, accountID)
	if err != nil {
		if errors.Is(err, account.ErrNotFound) {
			h.ErrorResponse(c, http.StatusNotFound, err, "Account not found")
			return
		}

		h.ErrorResponse(c, http.StatusInternalServerError, err, "Get balance by id error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": balance,
	})
}

func (h *Handler) addBalance(c *gin.Context) {
	var request addBalanceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Amount not added. request is not valid")
		return
	}

	balance, err := h.service.AddBalance(c, request.toDTO())
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, err, "Add balance error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": balance,
	})
}

type transferBalanceResponse struct {
	SenderBalance   int64 `json:"sender_balance"`
	ReceiverBalance int64 `json:"receiver_balance"`
}

func (h *Handler) transferBalance(c *gin.Context) {
	var request transferBalanceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Amount not transferred. request is not valid")
		return
	}

	senderBalance, receiverBalance, err := h.service.TransferBalance(c, request.toDTO())
	if err != nil {
		if errors.Is(err, account.ErrNotFound) {
			h.ErrorResponse(c, http.StatusNotFound, err, "Sender or receiver not found")
			return
		}

		h.ErrorResponse(c, http.StatusInternalServerError, err, "Transfer balance error")
		return
	}

	c.JSON(http.StatusOK, transferBalanceResponse{
		SenderBalance:   senderBalance,
		ReceiverBalance: receiverBalance,
	})
}
