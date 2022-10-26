package domain

import (
	"github.com/maypok86/payment-api/internal/domain/account"
	"github.com/maypok86/payment-api/internal/domain/order"
	"github.com/maypok86/payment-api/internal/domain/transaction"
	"github.com/maypok86/payment-api/internal/repository/psql"
	"go.uber.org/zap"
)

type Services struct {
	Account     *account.Service
	Transaction *transaction.Service
	Order       *order.Service
}

func NewServices(repositories *psql.Repositories, logger *zap.Logger) *Services {
	return &Services{
		Account:     account.NewService(repositories.Account, repositories.Transaction, logger),
		Transaction: transaction.NewService(repositories.Transaction, logger),
		Order:       order.NewService(repositories.Order, repositories.Transaction, repositories.Account, repositories.Report, logger),
	}
}
