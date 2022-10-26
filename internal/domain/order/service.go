package order

import (
	"context"
	"fmt"

	"github.com/maypok86/payment-api/internal/domain/account"
	"github.com/maypok86/payment-api/internal/domain/transaction"
	"go.uber.org/zap"
)

type Repository interface {
	WithTx(ctx context.Context, txFunc func(ctx context.Context) error) error
	CreateOrder(ctx context.Context, dto CreateDTO) (Order, error)
	PayForOrder(ctx context.Context, orderID int64) error
	CancelOrder(ctx context.Context, orderID int64) (int64, int64, error)
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, dto transaction.CreateDTO) error
}

type AccountRepository interface {
	AddBalance(ctx context.Context, dto account.AddBalanceDTO) (int64, error)
	ReserveBalance(ctx context.Context, dto account.ReserveBalanceDTO) (int64, error)
}

type Service struct {
	repository            Repository
	transactionRepository TransactionRepository
	accountRepository     AccountRepository
	logger                *zap.Logger
}

func NewService(
	repository Repository,
	transactionRepository TransactionRepository,
	accountRepository AccountRepository,
	logger *zap.Logger,
) *Service {
	return &Service{
		repository:            repository,
		transactionRepository: transactionRepository,
		accountRepository:     accountRepository,
		logger:                logger,
	}
}

func (s *Service) CreateOrder(ctx context.Context, dto CreateDTO) (order Order, balance int64, err error) {
	err = s.repository.WithTx(ctx, func(ctx context.Context) error {
		balance, err = s.accountRepository.ReserveBalance(ctx, account.ReserveBalanceDTO{
			AccountID: dto.AccountID,
			Amount:    dto.Amount,
		})
		if err != nil {
			return err
		}

		order, err = s.repository.CreateOrder(ctx, dto)
		if err != nil {
			return err
		}

		transactionDTO := transaction.CreateDTO{
			Type:        transaction.Reservation,
			SenderID:    dto.AccountID,
			ReceiverID:  dto.AccountID,
			Amount:      dto.Amount,
			Description: fmt.Sprintf("Reserve %d kopecks for order with id = %d", dto.Amount, dto.OrderID),
		}

		return s.transactionRepository.CreateTransaction(ctx, transactionDTO)
	})
	if err != nil {
		return Order{}, 0, err
	}

	return order, balance, nil
}

func (s *Service) PayForOrder(ctx context.Context, orderID int64) error {
	if err := s.repository.PayForOrder(ctx, orderID); err != nil {
		return fmt.Errorf("pay for order: %w", err)
	}

	return nil
}

func (s *Service) CancelOrder(ctx context.Context, orderID int64) (balance int64, err error) {
	err = s.repository.WithTx(ctx, func(ctx context.Context) error {
		accountID, amount, err := s.repository.CancelOrder(ctx, orderID)
		if err != nil {
			return err
		}

		balance, err = s.accountRepository.AddBalance(ctx, account.AddBalanceDTO{
			AccountID: accountID,
			Amount:    amount,
		})
		if err != nil {
			return err
		}

		transactionDTO := transaction.CreateDTO{
			Type:        transaction.CancelReservation,
			SenderID:    accountID,
			ReceiverID:  accountID,
			Amount:      amount,
			Description: fmt.Sprintf("Cancel reservation %d kopecks for order with orderID = %d", amount, orderID),
		}

		return s.transactionRepository.CreateTransaction(ctx, transactionDTO)
	})
	if err != nil {
		return 0, fmt.Errorf("cancel order: %w", err)
	}

	return balance, nil
}
