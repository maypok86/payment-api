package domain

import (
	"github.com/maypok86/payment-api/internal/domain/account"
	"github.com/maypok86/payment-api/internal/domain/transaction"
	"github.com/maypok86/payment-api/internal/repository/psql"
	"go.uber.org/zap"
)

type Services struct {
	Account     *account.Service
	Transaction *transaction.Service
}

func NewServices(repositories *psql.Repositories, logger *zap.Logger) *Services {
	return &Services{
		Account:     account.NewService(repositories.Account, logger),
		Transaction: transaction.NewService(repositories.Transaction, logger),
	}
}
