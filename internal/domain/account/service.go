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

func (s *Service) AddBalance(ctx context.Context, dto UpdateBalanceDTO) (Account, error) {
	account, err := s.repository.GetAccountByID(ctx, dto.ID)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			return Account{}, fmt.Errorf("add balance: %w", err)
		}
	}

	if account == nil {
		// What if something goes wrong?
		account, err = s.repository.CreateAccount(ctx, dto.ID)
		if err != nil {
			return Account{}, fmt.Errorf("add balance: %w", err)
		}
	}

	return Account{
		ID:      dto.ID,
		Balance: account.Balance + dto.Balance,
	}, nil
}
