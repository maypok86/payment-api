package report

type CreateDTO struct {
	ServiceID int64
	Amount    int64
}

type AddAmountDTO struct {
	ServiceID int64
	Amount    int64
}

type GetDTO struct {
	Month int64
	Year  int64
}
