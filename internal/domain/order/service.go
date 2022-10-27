package order

import (
	"context"
	"fmt"

	"github.com/maypok86/payment-api/internal/domain/account"
	"github.com/maypok86/payment-api/internal/domain/transaction"
	"go.uber.org/zap"
)

type Transactor interface {
	WithTx(ctx context.Context, txFunc func(ctx context.Context) error) error
}

type Repository interface {
	CreateOrder(ctx context.Context, dto CreateDTO) (Order, error)
	PayForOrder(ctx context.Context, dto PayForDTO) error
	CancelOrder(ctx context.Context, dto CancelDTO) (int64, int64, error)
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, dto transaction.CreateDTO) error
}

type AccountRepository interface {
	ReserveBalance(ctx context.Context, dto account.ReserveBalanceDTO) (int64, error)
	ReturnBalance(ctx context.Context, dto account.ReturnBalanceDTO) (int64, error)
}

type Service struct {
	transactor            Transactor
	repository            Repository
	transactionRepository TransactionRepository
	accountRepository     AccountRepository
	logger                *zap.Logger
}

func NewService(
	transactor Transactor,
	repository Repository,
	transactionRepository TransactionRepository,
	accountRepository AccountRepository,
	logger *zap.Logger,
) *Service {
	return &Service{
		transactor:            transactor,
		repository:            repository,
		transactionRepository: transactionRepository,
		accountRepository:     accountRepository,
		logger:                logger,
	}
}

func (s *Service) CreateOrder(ctx context.Context, dto CreateDTO) (order Order, balance int64, err error) {
	err = s.transactor.WithTx(ctx, func(ctx context.Context) error {
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

func (s *Service) PayForOrder(ctx context.Context, dto PayForDTO) error {
	if err := s.repository.PayForOrder(ctx, dto); err != nil {
		return fmt.Errorf("pay for order: %w", err)
	}

	return nil
}

func (s *Service) CancelOrder(ctx context.Context, dto CancelDTO) (balance int64, err error) {
	err = s.transactor.WithTx(ctx, func(ctx context.Context) error {
		accountID, amount, err := s.repository.CancelOrder(ctx, dto)
		if err != nil {
			return err
		}

		balance, err = s.accountRepository.ReturnBalance(ctx, account.ReturnBalanceDTO{
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
			Description: fmt.Sprintf("Cancel reservation %d kopecks for order with id = %d", amount, dto.OrderID),
		}

		return s.transactionRepository.CreateTransaction(ctx, transactionDTO)
	})
	if err != nil {
		return 0, fmt.Errorf("cancel order: %w", err)
	}

	return balance, nil
}
