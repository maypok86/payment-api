package psql

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/maypok86/payment-api/internal/domain/order"
	"github.com/maypok86/payment-api/internal/pkg/postgres"
	"go.uber.org/zap"
)

type OrderRepository struct {
	*postgres.Repository
	tableName string
	db        *postgres.Client
	logger    *zap.Logger
}

func NewOrderRepository(db *postgres.Client, logger *zap.Logger) *OrderRepository {
	return &OrderRepository{
		Repository: postgres.NewRepository(db),
		tableName:  "orders",
		db:         db,
		logger:     logger,
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

func (or *OrderRepository) PayForOrder(ctx context.Context, orderID int64) error {
	sql, args, err := or.db.Builder.Update(or.tableName).
		Set("is_paid", true).
		Where(sq.And{
			sq.Eq{"order_id": orderID},
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

func (or *OrderRepository) CancelOrder(ctx context.Context, orderID int64) (int64, int64, error) {
	sql, args, err := or.db.Builder.Update(or.tableName).
		Set("is_cancelled", true).
		Where(sq.And{
			sq.Eq{"order_id": orderID},
			sq.Eq{"is_paid": false},
		}).
		Suffix("RETURNING account_id, amount").
		ToSql()
	if err != nil {
		return 0, 0, fmt.Errorf("build cancel order query: %w", err)
	}

	or.logger.Debug("cancel order query", zap.String("sql", sql), zap.Any("args", args))

	var accountID int64
	var amount int64
	if err := or.db.QueryRow(ctx, sql, args...).Scan(&accountID, &amount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, fmt.Errorf("cancel order: %w", order.ErrNotFound)
		}

		return 0, 0, fmt.Errorf("cancel order: %w", err)
	}

	return accountID, amount, nil
}
