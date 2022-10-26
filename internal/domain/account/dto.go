package account

type AddBalanceDTO struct {
	AccountID int64
	Amount    int64
}

type TransferBalanceDTO struct {
	SenderID   int64
	ReceiverID int64
	Amount     int64
}

type ReserveBalanceDTO struct {
	AccountID int64
	Amount    int64
}

type ReturnBalanceDTO struct {
	AccountID int64
	Amount    int64
}
