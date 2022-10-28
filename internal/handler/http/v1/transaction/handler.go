package transaction

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/maypok86/payment-api/internal/domain/transaction"
	"github.com/maypok86/payment-api/internal/pkg/handler"
	"go.uber.org/zap"
)

//go:generate mockgen -source=handler.go -destination=mock_test.go -package=transaction_test

type Service interface {
	GetTransactionsByAccountID(
		ctx context.Context,
		accountID int64,
		listParams transaction.ListParams,
	) ([]transaction.Transaction, int, error)
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
	transactionsGroup := router.Group("/transaction")
	{
		transactionsGroup.GET("/:account_id", h.GetTransactionsByAccountID)
	}
}

func (h *Handler) GetTransactionsByAccountID(c *gin.Context) {
	accountID, err := h.ParseIDFromPath(c, "account_id")
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Transactions not found. id is not valid")
		return
	}

	params, err := h.ParsePaginationParams(c)
	if err != nil {
		h.ErrorResponse(c, http.StatusBadRequest, err, "Transactions not found. Pagination params is not valid")
		return
	}

	listParams, err := transaction.NewListParams(c.Query("sort"), c.Query("direction"), params)
	if err != nil {
		if errors.Is(err, transaction.ErrInvalidSortParam) {
			h.ErrorResponse(c, http.StatusBadRequest, err, "Transactions not found. Sort param is not valid")
			return
		}

		if errors.Is(err, transaction.ErrInvalidDirectionParam) {
			h.ErrorResponse(c, http.StatusBadRequest, err, "Transactions not found. Direction param is not valid")
			return
		}

		h.ErrorResponse(c, http.StatusBadRequest, err, "Transactions not found. Can not create list params")
		return
	}

	transactions, count, err := h.service.GetTransactionsByAccountID(c.Request.Context(), accountID, listParams)
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, err, "Get transactions by account id error")
		return
	}

	c.JSON(http.StatusOK, NewListResponse(transactions, params, count))
}
