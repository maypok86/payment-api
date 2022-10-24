package account

import "errors"

var (
	ErrAlreadyExist = errors.New("account with given fields already exist")
	ErrNotFound     = errors.New("account not found")
)

type Account struct {
	ID      int64
	Balance int64
}
