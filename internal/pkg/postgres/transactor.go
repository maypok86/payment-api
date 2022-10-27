package postgres

import (
	"context"
)

type Transactor struct {
	db *Client
}

func NewTransactor(db *Client) *Transactor {
	return &Transactor{db: db}
}

func (t *Transactor) WithTx(ctx context.Context, txFunc func(ctx context.Context) error) error {
	return t.db.WithTx(ctx, txFunc)
}
