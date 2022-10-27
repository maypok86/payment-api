package account

type GetBalanceResponse struct {
	Balance int64 `json:"balance"`
}

type AddBalanceResponse struct {
	Balance int64 `json:"balance"`
}

type TransferBalanceResponse struct {
	SenderBalance   int64 `json:"sender_balance"`
	ReceiverBalance int64 `json:"receiver_balance"`
}
