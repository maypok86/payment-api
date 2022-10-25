package postgres

import "context"

type Repository struct {
	c *Client
}

func NewRepository(c *Client) *Repository {
	return &Repository{c: c}
}

func (r *Repository) WithTx(ctx context.Context, txFunc func(ctx context.Context) error) error {
	return r.c.WithTx(ctx, txFunc)
}
