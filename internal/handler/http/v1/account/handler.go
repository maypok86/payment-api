package account

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maypok86/payment-api/internal/domain/account"
	"github.com/maypok86/payment-api/internal/pkg/handler"
	"go.uber.org/zap"
)

type Service interface {
	GetBalanceByID(ctx context.Context, id int64) (int64, error)
	AddBalance(ctx context.Context, dto account.AddBalanceDTO) (int64, error)
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
		h.ErrorResponse(c, http.StatusInternalServerError, err, "Get balance by id error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": balance,
	})
}

type addBalanceRequest struct {
	ID      int64 `json:"id"      binding:"required"`
	Balance int64 `json:"balance" binding:"gte=0"`
}

func (h *Handler) addBalance(c *gin.Context) {
	var request addBalanceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Balance not added. request is not valid")
		return
	}

	balance, err := h.service.AddBalance(c, account.AddBalanceDTO{
		ID:      request.ID,
		Balance: request.Balance,
	})
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, err, "Add balance error")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"balance": balance,
	})
}
