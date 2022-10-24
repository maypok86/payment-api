package transaction

type CreateDTO struct {
	Type        Type
	SenderID    int64
	ReceiverID  int64
	Amount      int64
	Description string
}
