package account

type AddBalanceDTO struct {
	ID      int64
	Balance int64
}

type TransferBalanceDTO struct {
	SenderID   int64
	ReceiverID int64
	Amount     int64
}
