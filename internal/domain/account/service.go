package account

import (
	"context"
	"fmt"

	"github.com/maypok86/payment-api/internal/domain/transaction"
	"go.uber.org/zap"
)

type Repository interface {
	WithTx(ctx context.Context, txFunc func(ctx context.Context) error) error
	GetAccountByID(ctx context.Context, accountID int64) (Account, error)
	AddBalance(ctx context.Context, dto AddBalanceDTO) (int64, error)
	TransferBalance(ctx context.Context, dto TransferBalanceDTO) (int64, int64, error)
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, dto transaction.CreateDTO) error
}

type Service struct {
	repository            Repository
	transactionRepository TransactionRepository
	logger                *zap.Logger
}

func NewService(repository Repository, transactionRepository TransactionRepository, logger *zap.Logger) *Service {
	return &Service{
		repository:            repository,
		transactionRepository: transactionRepository,
		logger:                logger,
	}
}

func (s *Service) GetBalanceByID(ctx context.Context, accountID int64) (int64, error) {
	account, err := s.repository.GetAccountByID(ctx, accountID)
	if err != nil {
		return 0, fmt.Errorf("get balance by id: %w", err)
	}

	return account.Balance, nil
}

func (s *Service) AddBalance(ctx context.Context, dto AddBalanceDTO) (balance int64, err error) {
	err = s.repository.WithTx(ctx, func(ctx context.Context) error {
		balance, err = s.repository.AddBalance(ctx, dto)
		if err != nil {
			return err
		}

		transactionDTO := transaction.CreateDTO{
			Type:        transaction.Enrollment,
			SenderID:    dto.AccountID,
			ReceiverID:  dto.AccountID,
			Amount:      dto.Amount,
			Description: fmt.Sprintf("Add %d kopecks to account with id = %d", dto.Amount, dto.AccountID),
		}

		return s.transactionRepository.CreateTransaction(ctx, transactionDTO)
	})
	if err != nil {
		return 0, fmt.Errorf("add balance: %w", err)
	}

	return balance, nil
}

func (s *Service) TransferBalance(
	ctx context.Context,
	dto TransferBalanceDTO,
) (senderBalance int64, receiverBalance int64, err error) {
	err = s.repository.WithTx(ctx, func(ctx context.Context) error {
		senderBalance, receiverBalance, err = s.repository.TransferBalance(ctx, dto)
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

		return s.transactionRepository.CreateTransaction(ctx, transactionDTO)
	})
	if err != nil {
		return 0, 0, fmt.Errorf("transfer balance: %w", err)
	}

	return senderBalance, receiverBalance, nil
}
