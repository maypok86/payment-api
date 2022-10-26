package order

type CreateDTO struct {
	OrderID   int64
	AccountID int64
	ServiceID int64
	Amount    int64
}

type PayForDTO struct {
	OrderID   int64
	AccountID int64
	ServiceID int64
	Amount    int64
}

type CancelDTO struct {
	OrderID   int64
	AccountID int64
	ServiceID int64
	Amount    int64
}
