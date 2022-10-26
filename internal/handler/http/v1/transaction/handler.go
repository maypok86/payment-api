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

type Service interface {
	GetTransactionsBySenderID(
		ctx context.Context,
		senderID int64,
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
		transactionsGroup.GET("/:sender_id", h.getTransactionsBySenderID)
	}
}

func (h *Handler) getTransactionsBySenderID(c *gin.Context) {
	senderID, err := h.ParseIDFromPath(c, "sender_id")
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

	transactions, count, err := h.service.GetTransactionsBySenderID(c, senderID, listParams)
	if err != nil {
		h.ErrorResponse(c, http.StatusInternalServerError, err, "Get transactions by sender id error")
		return
	}

	c.JSON(http.StatusOK, newTransactionListResponse(transactions, params, count))
}
