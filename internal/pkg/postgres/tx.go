package postgres

import (
	"context"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
)

type txKey struct{}

func injectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func extractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}

func (c *Client) WithTx(ctx context.Context, txFunc func(ctx context.Context) error) error {
	return c.Pool.BeginFunc(ctx, func(tx pgx.Tx) error {
		return txFunc(injectTx(ctx, tx))
	})
}

func (c *Client) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if tx := extractTx(ctx); tx != nil {
		return tx.QueryRow(ctx, sql, args...)
	}

	return c.Pool.QueryRow(ctx, sql, args...)
}

func (c *Client) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if tx := extractTx(ctx); tx != nil {
		return tx.Query(ctx, sql, args...)
	}

	return c.Pool.Query(ctx, sql, args...)
}

func (c *Client) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	if tx := extractTx(ctx); tx != nil {
		return tx.Exec(ctx, sql, args...)
	}

	return c.Pool.Exec(ctx, sql, args...)
}
