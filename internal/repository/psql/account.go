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

type updateBalanceDTO struct {
	accountID int64
	amount    int64
}

func (ar *AccountRepository) updateBalance(
	ctx context.Context,
	tx pgx.Tx,
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

	if err := tx.QueryRow(ctx, sql, args...).Scan(&accountBalance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, account.ErrNotFound
		}

		return 0, err
	}

	return accountBalance, nil
}

func (ar *AccountRepository) createAccount(ctx context.Context, tx pgx.Tx, id int64) (account.Account, error) {
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
	if err := tx.QueryRow(ctx, sql, args...).Scan(&entity.ID, &entity.Balance); err != nil {
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
) (accountBalance int64, err error) {
	err = ar.db.Pool.BeginFunc(ctx, func(tx pgx.Tx) error {
		accountBalance, err = ar.updateBalance(ctx, tx, "add", updateBalanceDTO{
			accountID: dto.ID,
			amount:    dto.Balance,
		})
		if err == nil {
			transactionDTO := transaction.CreateDTO{
				Type:        transaction.Enrollment,
				SenderID:    dto.ID,
				ReceiverID:  dto.ID,
				Amount:      dto.Balance,
				Description: fmt.Sprintf("Add %d kopecks to account with id = %d", dto.Balance, dto.ID),
			}

			return ar.createTransaction(ctx, tx, transactionDTO)
		}

		if errors.Is(err, account.ErrNotFound) {
			if _, err := ar.createAccount(ctx, tx, dto.ID); err != nil {
				return err
			}

			accountBalance, err = ar.updateBalance(ctx, tx, "add", updateBalanceDTO{
				accountID: dto.ID,
				amount:    dto.Balance,
			})
		}

		return err
	})
	if err != nil {
		return accountBalance, fmt.Errorf("add balance: %w", err)
	}

	return accountBalance, nil
}

func (ar *AccountRepository) TransferBalance(
	ctx context.Context,
	dto account.TransferBalanceDTO,
) (senderBalance int64, receiverBalance int64, err error) {
	err = ar.db.Pool.BeginFunc(ctx, func(tx pgx.Tx) error {
		senderBalance, err = ar.updateBalance(ctx, tx, "send", updateBalanceDTO{
			accountID: dto.SenderID,
			amount:    -dto.Amount,
		})
		if err != nil {
			return err
		}

		receiverBalance, err = ar.updateBalance(ctx, tx, "receive", updateBalanceDTO{
			accountID: dto.ReceiverID,
			amount:    dto.Amount,
		})
		if err != nil {
			return err
		}

		transactionDTO := transaction.CreateDTO{
			Type:       transaction.Transfer,
			SenderID:   dto.SenderID,
			ReceiverID: dto.ReceiverID,
			Amount:     dto.Amount,
			Description: fmt.Sprintf(
				"Transfer %d kopecks from account with id = %d to account with id = %d",
				dto.Amount,
				dto.SenderID,
				dto.ReceiverID,
			),
		}

		return ar.createTransaction(ctx, tx, transactionDTO)
	})
	if err != nil {
		return 0, 0, fmt.Errorf("transfer balance: %w", err)
	}

	return senderBalance, receiverBalance, nil
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
