package psql

import (
	"github.com/maypok86/payment-api/internal/pkg/postgres"
	"go.uber.org/zap"
)

type Repositories struct {
	Account *AccountRepository
}

func NewRepositories(db *postgres.Client, logger *zap.Logger) *Repositories {
	return &Repositories{
		Account: NewAccountRepository(db, logger),
	}
}
