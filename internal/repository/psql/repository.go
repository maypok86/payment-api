package psql

import (
	"github.com/maypok86/payment-api/internal/pkg/postgres"
	"go.uber.org/zap"
)

type Repositories struct {
	Account     *AccountRepository
	Transaction *TransactionRepository
	Order       *OrderRepository
	Report      *ReportRepository
}

func NewRepositories(db *postgres.Client, logger *zap.Logger) *Repositories {
	return &Repositories{
		Account:     NewAccountRepository(db, logger),
		Transaction: NewTransactionRepository(db, logger),
		Order:       NewOrderRepository(db, logger),
		Report:      NewReportRepository(db, logger),
	}
}
