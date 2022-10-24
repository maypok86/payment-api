package psql

import (
	"context"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v4"
	"github.com/maypok86/payment-api/internal/domain/account"
	"github.com/maypok86/payment-api/internal/pkg/postgres"
	"go.uber.org/zap"
)

type AccountRepository struct {
	tableName string
	db        *postgres.Client
	logger    *zap.Logger
}

func NewAccountRepository(db *postgres.Client, logger *zap.Logger) *AccountRepository {
	return &AccountRepository{
		tableName: "accounts",
		db:        db,
		logger:    logger,
	}
}

func (ar *AccountRepository) CreateAccount(ctx context.Context, id int64) (*account.Account, error) {
	sql, args, err := ar.db.Builder.Insert(ar.tableName).
		Columns("id").
		Values(id).
		Suffix("RETURNING id, balance").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build create account query: %w", err)
	}

	ar.logger.Debug("create account query", zap.String("sql", sql), zap.Any("args", args))

	entity := &account.Account{}
	if err = ar.db.Pool.QueryRow(ctx, sql, args...).Scan(&entity.ID, &entity.Balance); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return nil, fmt.Errorf("insert account: %w", account.ErrAlreadyExist)
			}
		}

		return nil, fmt.Errorf("insert account: %w", err)
	}

	return entity, nil
}

func (ar *AccountRepository) GetAccountByID(ctx context.Context, id int64) (*account.Account, error) {
	sql, args, err := ar.db.Builder.Select("id", "balance").
		From(ar.tableName).
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("build get account by id query: %w", err)
	}

	ar.logger.Debug("get account by id query", zap.String("sql", sql), zap.Any("args", args))

	entity := &account.Account{}
	if err = ar.db.Pool.QueryRow(ctx, sql, args...).Scan(&entity.ID, &entity.Balance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("get account by id: %w", account.ErrNotFound)
		}

		return nil, fmt.Errorf("get account by id: %w", err)
	}

	return entity, nil
}
