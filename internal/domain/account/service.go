package account

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type Repository interface {
	GetAccountByID(ctx context.Context, id int64) (*Account, error)
	AddBalance(ctx context.Context, dto AddBalanceDTO) (int64, error)
	TransferBalance(ctx context.Context, dto TransferBalanceDTO) (int64, int64, error)
}

type Service struct {
	repository Repository
	logger     *zap.Logger
}

func NewService(repository Repository, logger *zap.Logger) *Service {
	return &Service{
		repository: repository,
		logger:     logger,
	}
}

func (s *Service) GetBalanceByID(ctx context.Context, id int64) (int64, error) {
	account, err := s.repository.GetAccountByID(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("get balance by id: %w", err)
	}

	return account.Balance, nil
}

func (s *Service) AddBalance(ctx context.Context, dto AddBalanceDTO) (int64, error) {
	balance, err := s.repository.AddBalance(ctx, dto)
	if err != nil {
		return 0, fmt.Errorf("add balance: %w", err)
	}

	return balance, nil
}

func (s *Service) TransferBalance(ctx context.Context, dto TransferBalanceDTO) (int64, int64, error) {
	senderBalance, receiverBalance, err := s.repository.TransferBalance(ctx, dto)
	if err != nil {
		return 0, 0, fmt.Errorf("transfer balance: %w", err)
	}

	return senderBalance, receiverBalance, nil
}
