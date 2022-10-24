package transaction

import (
	"database/sql/driver"
	"errors"
	"time"
)

var (
	ErrCreate          = errors.New("create transaction")
	ErrAlreadyExist    = errors.New("transaction with given id already exist")
	ErrAccountNotFound = errors.New("account with given account_id not found")
)

type Type struct {
	string
}

var Enrollment = Type{"enrollment"}

func (t *Type) Scan(value interface{}) error {
	asBytes, ok := value.([]byte)
	if !ok {
		return errors.New("scan source is not []byte")
	}

	values := map[string]Type{
		"enrollment": Enrollment,
	}
	if v, ok := values[string(asBytes)]; ok {
		*t = v
		return nil
	}

	return errors.New("wrong value for Type")
}

func (t Type) Value() (driver.Value, error) {
	values := map[Type]string{
		Enrollment: "enrollment",
	}

	s, ok := values[t]
	if !ok {
		return nil, errors.New("wrong value for Type")
	}

	return s, nil
}

type Transaction struct {
	ID          int64
	Type        Type
	SenderID    int64
	ReceiverID  int64
	Amount      int64
	Description string
	CreatedAt   time.Time
}
