package order

import (
	"errors"
	"time"
)

var (
	ErrAlreadyExist    = errors.New("order with given id already exist")
	ErrAccountNotFound = errors.New("account with given account_id not found")
	ErrNotFound        = errors.New("order not found")
)

type Order struct {
	OrderID     int64
	AccountID   int64
	ServiceID   int64
	Amount      int64
	IsPaid      bool
	IsCancelled bool
	UpdatedAt   time.Time
	CreatedAt   time.Time
}
