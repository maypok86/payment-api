package account

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
)

type Repository interface {
	CreateAccount(ctx context.Context, id int64) (*Account, error)
	GetAccountByID(ctx context.Context, id int64) (*Account, error)
	AddBalance(ctx context.Context, dto AddBalanceDTO) (int64, error)
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
	account, err := s.repository.GetAccountByID(ctx, dto.ID)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return 0, fmt.Errorf("add balance: %w", err)
		}
	}

	if account == nil {
		// What if something goes wrong?
		_, err = s.repository.CreateAccount(ctx, dto.ID)
		if err != nil {
			return 0, fmt.Errorf("add balance: %w", err)
		}
	}

	balance, err := s.repository.AddBalance(ctx, dto)
	if err != nil {
		return 0, fmt.Errorf("add balance: %w", err)
	}

	return balance, nil
}
