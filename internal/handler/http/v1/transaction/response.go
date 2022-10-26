package transaction

import (
	"time"

	"github.com/maypok86/payment-api/internal/domain/transaction"
	"github.com/maypok86/payment-api/internal/pkg/pagination"
)

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
