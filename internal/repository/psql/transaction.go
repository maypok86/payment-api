package psql

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/maypok86/payment-api/internal/domain/transaction"
	"github.com/maypok86/payment-api/internal/pkg/postgres"
	"go.uber.org/zap"
)

type TransactionRepository struct {
	tableName string
	db        *postgres.Client
	logger    *zap.Logger
}

func NewTransactionRepository(db *postgres.Client, logger *zap.Logger) *TransactionRepository {
	return &TransactionRepository{
		tableName: "transactions",
		db:        db,
		logger:    logger,
	}
}

func (tr *TransactionRepository) GetTransactionsBySenderID(
	ctx context.Context,
	senderID int64,
	listParams transaction.ListParams,
) ([]transaction.Transaction, int, error) {
	query := tr.db.Builder.Select(
		"id",
		"type",
		"sender_id",
		"receiver_id",
		"amount",
		"description",
		"created_at", "COUNT(*) OVER () AS total").
		From(tr.tableName).
		Where(sq.Eq{"sender_id": senderID})

	if listParams.Sort != nil {
		query = listParams.Sort.UseSelectBuilder(query)
	}

	sql, args, err := query.Limit(listParams.Pagination.Limit).Offset(listParams.Pagination.Offset).ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build get all transactions query: %w", err)
	}

	tr.logger.Debug("get all transactions query", zap.String("sql", sql), zap.Any("args", args))

	rows, err := tr.db.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("run get all transactions query: %w", err)
	}
	defer rows.Close()

	var entities []transaction.Transaction
	var count int
	for rows.Next() {
		var entity transaction.Transaction
		if err := rows.Scan(
			&entity.ID,
			&entity.Type,
			&entity.SenderID,
			&entity.ReceiverID,
			&entity.Amount,
			&entity.Description,
			&entity.CreatedAt,
			&count,
		); err != nil {
			return nil, 0, fmt.Errorf("scan transaction: %w", err)
		}

		entities = append(entities, entity)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("read all transactions: %w", err)
	}

	return entities, count, nil
}
