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
	"github.com/maypok86/payment-api/internal/domain/transaction"
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

func (ar *AccountRepository) AddBalance(ctx context.Context, dto account.AddBalanceDTO) (int64, error) {
	var accountBalance int64
	err := ar.db.Pool.BeginFunc(ctx, func(tx pgx.Tx) error {
		sql, args, err := ar.db.Builder.Update(ar.tableName).
			Set("balance", sq.Expr("balance + ?", dto.Balance)).
			Where(sq.Eq{"id": dto.ID}).
			Suffix("RETURNING balance").
			ToSql()
		if err != nil {
			return fmt.Errorf("build add balance query: %w", err)
		}

		ar.logger.Debug("add balance query", zap.String("sql", sql), zap.Any("args", args))

		if err := tx.QueryRow(ctx, sql, args...).Scan(&accountBalance); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return account.ErrNotFound
			}

			return err
		}

		transactionDTO := transaction.CreateDTO{
			Type:        transaction.Enrollment,
			SenderID:    dto.ID,
			ReceiverID:  dto.ID,
			Amount:      dto.Balance,
			Description: fmt.Sprintf("Add %d kopecks to account with id = %d", dto.Balance, dto.ID),
		}

		return ar.createTransaction(ctx, tx, transactionDTO)
	})
	if err != nil {
		return accountBalance, fmt.Errorf("add balance: %w", err)
	}

	return accountBalance, nil
}

func (ar *AccountRepository) createTransaction(ctx context.Context, tx pgx.Tx, dto transaction.CreateDTO) error {
	sql, args, err := ar.db.Builder.Insert("transactions").
		Columns("type", "sender_id", "receiver_id", "amount", "description").
		Values(dto.Type, dto.SenderID, dto.ReceiverID, dto.Amount, dto.Description).
		ToSql()
	if err != nil {
		return fmt.Errorf("build create transaction query: %w", err)
	}

	ar.logger.Debug("create transaction query", zap.String("sql", sql), zap.Any("args", args))

	result, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return fmt.Errorf("insert transaction: %w", transaction.ErrAlreadyExist)
			case pgerrcode.ForeignKeyViolation:
				return fmt.Errorf("insert transaction: %w", transaction.ErrAccountNotFound)
			}
		}

		return fmt.Errorf("insert transaction: %w", err)
	}

	if result.RowsAffected() != 1 {
		return transaction.ErrCreate
	}

	return nil
}
