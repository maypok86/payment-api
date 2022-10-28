package transaction

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

//go:generate mockgen -source=service.go -destination=mock_test.go -package=transaction_test

type Repository interface {
	GetTransactionsByAccountID(ctx context.Context, senderID int64, listParams ListParams) ([]Transaction, int, error)
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

func (s *Service) GetTransactionsByAccountID(
	ctx context.Context,
	accountID int64,
	listParams ListParams,
) ([]Transaction, int, error) {
	transactions, count, err := s.repository.GetTransactionsByAccountID(ctx, accountID, listParams)
	if err != nil {
		return nil, 0, fmt.Errorf("get transactions by sender accountID: %w", err)
	}

	return transactions, count, nil
}
