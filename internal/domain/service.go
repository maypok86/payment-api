package domain

import (
	"context"

	"github.com/maypok86/payment-api/internal/cache"
	"github.com/maypok86/payment-api/internal/domain/account"
	"github.com/maypok86/payment-api/internal/domain/order"
	"github.com/maypok86/payment-api/internal/domain/report"
	"github.com/maypok86/payment-api/internal/domain/transaction"
	"github.com/maypok86/payment-api/internal/repository/psql"
	"go.uber.org/zap"
)

type Transactor interface {
	WithTx(ctx context.Context, txFunc func(ctx context.Context) error) error
}

type Services struct {
	Account     *account.Service
	Transaction *transaction.Service
	Order       *order.Service
	Report      *report.Service
}

func NewServices(
	transactor Transactor,
	repositories *psql.Repositories,
	reportCache *cache.ReportCache,
	logger *zap.Logger,
) *Services {
	return &Services{
		Account:     account.NewService(transactor, repositories.Account, repositories.Transaction, logger),
		Transaction: transaction.NewService(repositories.Transaction, logger),
		Order: order.NewService(
			transactor,
			repositories.Order,
			repositories.Transaction,
			repositories.Account,
			logger,
		),
		Report: report.NewService(repositories.Report, reportCache, logger),
	}
}
