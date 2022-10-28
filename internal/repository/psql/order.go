package psql

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/maypok86/payment-api/internal/domain/order"
	"github.com/maypok86/payment-api/internal/pkg/postgres"
	"go.uber.org/zap"
)

type OrderRepository struct {
	tableName string
	db        *postgres.Client
	logger    *zap.Logger
}

func NewOrderRepository(db *postgres.Client, logger *zap.Logger) *OrderRepository {
	return &OrderRepository{
		tableName: "orders",
		db:        db,
		logger:    logger,
	}
}

func (or *OrderRepository) CreateOrder(ctx context.Context, dto order.CreateDTO) (order.Order, error) {
	sql, args, err := or.db.Builder.Insert(or.tableName).
		Columns("order_id", "account_id", "service_id", "amount").
		Values(dto.OrderID, dto.AccountID, dto.ServiceID, dto.Amount).
		Suffix("RETURNING is_paid, is_cancelled, created_at, updated_at").
		ToSql()
	if err != nil {
		return order.Order{}, fmt.Errorf("build create order query: %w", err)
	}

	or.logger.Debug("create order query", zap.String("sql", sql), zap.Any("args", args))

	entity := order.Order{
		OrderID:   dto.OrderID,
		AccountID: dto.AccountID,
		ServiceID: dto.ServiceID,
		Amount:    dto.Amount,
	}
	if err := or.db.QueryRow(ctx, sql, args...).Scan(
		&entity.IsPaid,
		&entity.IsCancelled,
		&entity.CreatedAt,
		&entity.UpdatedAt,
	); err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return order.Order{}, fmt.Errorf("insert order: %w", order.ErrAlreadyExist)
			case pgerrcode.ForeignKeyViolation:
				return order.Order{}, fmt.Errorf("insert order: %w", order.ErrAccountNotFound)
			}
		}

		return order.Order{}, fmt.Errorf("insert order: %w", err)
	}

	return entity, nil
}

func (or *OrderRepository) PayForOrder(ctx context.Context, dto order.PayForDTO) error { //nolint:dupl
	sql, args, err := or.db.Builder.Update(or.tableName).
		Set("is_paid", true).
		Where(sq.And{
			sq.Eq{"order_id": dto.OrderID},
			sq.Eq{"account_id": dto.AccountID},
			sq.Eq{"service_id": dto.ServiceID},
			sq.Eq{"amount": dto.Amount},
			sq.Eq{"is_paid": false},
			sq.Eq{"is_cancelled": false},
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build pay for order query: %w", err)
	}

	or.logger.Debug("pay for order query", zap.String("sql", sql), zap.Any("args", args))

	result, err := or.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("pay for order: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("pay for order: %w", order.ErrNotFound)
	}

	return nil
}

func (or *OrderRepository) CancelOrder(ctx context.Context, dto order.CancelDTO) error { //nolint:dupl
	sql, args, err := or.db.Builder.Update(or.tableName).
		Set("is_cancelled", true).
		Where(sq.And{
			sq.Eq{"order_id": dto.OrderID},
			sq.Eq{"account_id": dto.AccountID},
			sq.Eq{"service_id": dto.ServiceID},
			sq.Eq{"amount": dto.Amount},
			sq.Eq{"is_paid": false},
			sq.Eq{"is_cancelled": false},
		}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build cancel order query: %w", err)
	}

	or.logger.Debug("cancel order query", zap.String("sql", sql), zap.Any("args", args))

	result, err := or.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("cancel order: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("cancel order: %w", order.ErrNotFound)
	}

	return nil
}
