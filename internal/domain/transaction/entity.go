package transaction

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/maypok86/payment-api/internal/pkg/pagination"
	"github.com/maypok86/payment-api/internal/pkg/sort"
)

var (
	ErrCreate          = errors.New("create transaction")
	ErrAlreadyExist    = errors.New("transaction with given id already exist")
	ErrAccountNotFound = errors.New("account with given account_id not found")
)

type Type struct {
	string
}

var (
	Enrollment        = Type{"enrollment"}
	Transfer          = Type{"transfer"}
	Reservation       = Type{"reservation"}
	CancelReservation = Type{"cancel_reservation"}
)

var transactionTypeToString = map[Type]string{
	Enrollment:        "enrollment",
	Transfer:          "transfer",
	Reservation:       "reservation",
	CancelReservation: "cancel_reservation",
}

var stringToTransactionType = map[string]Type{
	"enrollment":         Enrollment,
	"transfer":           Transfer,
	"reservation":        Reservation,
	"cancel_reservation": CancelReservation,
}

func (t Type) String() string {
	return t.string
}

func (t *Type) Scan(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return errors.New("scan source is not string")
	}

	if v, ok := stringToTransactionType[s]; ok {
		*t = v
		return nil
	}

	return errors.New("wrong value for Type")
}

func (t Type) Value() (driver.Value, error) {
	s, ok := transactionTypeToString[t]
	if !ok {
		return nil, errors.New("wrong value for Type")
	}

	return s, nil
}

type Transaction struct {
	TransactionID int64
	Type          Type
	SenderID      int64
	ReceiverID    int64
	Amount        int64
	Description   string
	CreatedAt     time.Time
}

var (
	ErrInvalidSortParam      = errors.New("invalid sort param")
	ErrInvalidDirectionParam = errors.New("invalid direction param")
)

type ListParams struct {
	Pagination pagination.Params
	Sort       *sort.Sort
}

func NewListParams(sortParam, directionParam string, params pagination.Params) (ListParams, error) {
	if params.Limit == 0 || params.Limit > pagination.MaxLimit {
		params.Limit = pagination.DefaultLimit
	}
	if sortParam == "" {
		return ListParams{Pagination: params}, nil
	}

	sortMap := map[string]string{
		"date": "created_at",
		"sum":  "amount",
	}

	column, ok := sortMap[sortParam]
	if !ok {
		return ListParams{}, ErrInvalidSortParam
	}

	directionSet := map[string]struct{}{
		"asc":  {},
		"desc": {},
	}

	if _, ok := directionSet[directionParam]; !ok && directionParam != "" {
		return ListParams{}, ErrInvalidDirectionParam
	}

	return ListParams{
		Pagination: params,
		Sort:       sort.New(column, directionParam),
	}, nil
}
