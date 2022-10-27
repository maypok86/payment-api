package transaction

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

//go:generate mockgen -source=service.go -destination=mock_test.go -package=transaction_test

type Repository interface {
	GetTransactionsBySenderID(ctx context.Context, senderID int64, listParams ListParams) ([]Transaction, int, error)
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

func (s *Service) GetTransactionsBySenderID(
	ctx context.Context,
	senderID int64,
	listParams ListParams,
) ([]Transaction, int, error) {
	transactions, count, err := s.repository.GetTransactionsBySenderID(ctx, senderID, listParams)
	if err != nil {
		return nil, 0, fmt.Errorf("get transactions by sender id: %w", err)
	}

	return transactions, count, nil
}
