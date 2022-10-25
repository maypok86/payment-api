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
	*postgres.Repository
	tableName string
	db        *postgres.Client
	logger    *zap.Logger
}

func NewAccountRepository(db *postgres.Client, logger *zap.Logger) *AccountRepository {
	return &AccountRepository{
		Repository: postgres.NewRepository(db),
		tableName:  "accounts",
		db:         db,
		logger:     logger,
	}
}

func (ar *AccountRepository) GetAccountByID(ctx context.Context, id int64) (account.Account, error) {
	sql, args, err := ar.db.Builder.Select("id", "balance").
		From(ar.tableName).
		Where(sq.Eq{"id": id}).
		Limit(1).
		ToSql()
	if err != nil {
		return account.Account{}, fmt.Errorf("build get account by id query: %w", err)
	}

	ar.logger.Debug("get account by id query", zap.String("sql", sql), zap.Any("args", args))

	var entity account.Account
	if err = ar.db.QueryRow(ctx, sql, args...).Scan(&entity.ID, &entity.Balance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return account.Account{}, fmt.Errorf("get account by id: %w", account.ErrNotFound)
		}

		return account.Account{}, fmt.Errorf("get account by id: %w", err)
	}

	return entity, nil
}

type updateBalanceDTO struct {
	accountID int64
	amount    int64
}

func (ar *AccountRepository) updateBalance(
	ctx context.Context,
	updateType string,
	dto updateBalanceDTO,
) (int64, error) {
	var accountBalance int64
	sql, args, err := ar.db.Builder.Update(ar.tableName).
		Set("balance", sq.Expr("balance + ?", dto.amount)).
		Where(sq.Eq{"id": dto.accountID}).
		Suffix("RETURNING balance").
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("build %s balance query: %w", updateType, err)
	}

	ar.logger.Debug(fmt.Sprintf("%s balance query", updateType), zap.String("sql", sql), zap.Any("args", args))

	if err := ar.db.QueryRow(ctx, sql, args...).Scan(&accountBalance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, account.ErrNotFound
		}

		return 0, err
	}

	return accountBalance, nil
}

func (ar *AccountRepository) createAccount(ctx context.Context, id int64) (account.Account, error) {
	sql, args, err := ar.db.Builder.Insert(ar.tableName).
		Columns("id").
		Values(id).
		Suffix("RETURNING id, balance").
		ToSql()
	if err != nil {
		return account.Account{}, fmt.Errorf("build create account query: %w", err)
	}

	ar.logger.Debug("create account query", zap.String("sql", sql), zap.Any("args", args))

	var entity account.Account
	if err := ar.db.QueryRow(ctx, sql, args...).Scan(&entity.ID, &entity.Balance); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return account.Account{}, fmt.Errorf("insert account: %w", account.ErrAlreadyExist)
			}
		}

		return account.Account{}, fmt.Errorf("insert account: %w", err)
	}

	return entity, nil
}

func (ar *AccountRepository) AddBalance(
	ctx context.Context,
	dto account.AddBalanceDTO,
) (int64, error) {
	accountBalance, err := ar.updateBalance(ctx, "add", updateBalanceDTO{
		accountID: dto.ID,
		amount:    dto.Balance,
	})
	if err == nil {
		return accountBalance, nil
	}

	if errors.Is(err, account.ErrNotFound) {
		if _, err := ar.createAccount(ctx, dto.ID); err != nil {
			return 0, fmt.Errorf("add balance: %w", err)
		}

		accountBalance, err = ar.updateBalance(ctx, "add", updateBalanceDTO{
			accountID: dto.ID,
			amount:    dto.Balance,
		})
		if err == nil {
			return accountBalance, nil
		}
	}

	return 0, fmt.Errorf("add balance: %w", err)
}

func (ar *AccountRepository) TransferBalance(
	ctx context.Context,
	dto account.TransferBalanceDTO,
) (int64, int64, error) {
	senderBalance, err := ar.updateBalance(ctx, "send", updateBalanceDTO{
		accountID: dto.SenderID,
		amount:    -dto.Amount,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("transfer balance: %w", err)
	}

	receiverBalance, err := ar.updateBalance(ctx, "receive", updateBalanceDTO{
		accountID: dto.ReceiverID,
		amount:    dto.Amount,
	})
	if err != nil {
		return 0, 0, fmt.Errorf("transfer balance: %w", err)
	}

	return senderBalance, receiverBalance, nil
}
