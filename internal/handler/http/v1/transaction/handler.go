package transaction

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/maypok86/payment-api/internal/domain/transaction"
	"github.com/maypok86/payment-api/internal/pkg/handler"
	"github.com/maypok86/payment-api/internal/pkg/pagination"
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

type transactionResponse struct {
	TransactionID int64     `json:"transaction_id"`
	Type          string    `json:"type"`
	SenderID      int64     `json:"sender_id"`
	ReceiverID    int64     `json:"receiver_id"`
	Amount        int64     `json:"amount"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
}

func newTransactionResponse(transaction transaction.Transaction) transactionResponse {
	return transactionResponse{
		TransactionID: transaction.TransactionID,
		Type:          transaction.Type.String(),
		SenderID:      transaction.SenderID,
		ReceiverID:    transaction.ReceiverID,
		Amount:        transaction.Amount,
		Description:   transaction.Description,
		CreatedAt:     transaction.CreatedAt,
	}
}

type transactionListResponse struct {
	Transactions []transactionResponse `json:"transactions"`
	Range        pagination.ListRange  `json:"range"`
}

func newTransactionListResponse(
	transactions []transaction.Transaction,
	params pagination.Params,
	count int,
) transactionListResponse {
	responses := make([]transactionResponse, 0, len(transactions))

	for _, tr := range transactions {
		responses = append(responses, newTransactionResponse(tr))
	}

	return transactionListResponse{
		Transactions: responses,
		Range:        pagination.NewListRange(params, count),
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
