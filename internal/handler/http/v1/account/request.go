package account

import "github.com/maypok86/payment-api/internal/domain/account"

type AddBalanceRequest struct {
	AccountID int64 `json:"account_id" binding:"required"`
	Amount    int64 `json:"amount"     binding:"gte=0"`
}

func (r AddBalanceRequest) ToDTO() account.AddBalanceDTO {
	return account.AddBalanceDTO{
		AccountID: r.AccountID,
		Amount:    r.Amount,
	}
}

type TransferBalanceRequest struct {
	SenderID   int64 `json:"sender_id"   binding:"required"`
	ReceiverID int64 `json:"receiver_id" binding:"required"`
	Amount     int64 `json:"amount"      binding:"required,gt=0"`
}

func (r TransferBalanceRequest) ToDTO() account.TransferBalanceDTO {
	return account.TransferBalanceDTO{
		SenderID:   r.SenderID,
		ReceiverID: r.ReceiverID,
		Amount:     r.Amount,
	}
}
