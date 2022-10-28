package integration

import (
	"context"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
	"github.com/maypok86/payment-api/internal/pkg/postgres"
	"github.com/stretchr/testify/suite"
)

const (
	host       = "backend:8080"
	healthPath = "http://" + host + "/health"
	attempts   = 20
	basePath   = "http://" + host + "/api/v1"
)

type APISuite struct {
	suite.Suite

	db *postgres.Client
}

func (as *APISuite) SetupSuite() {
	cfg := postgres.NewConnectionConfig(
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DBNAME"),
		os.Getenv("POSTGRES_SSLMODE"),
	)
	client, err := postgres.NewClient(context.Background(), cfg)
	as.Require().NoError(err)

	as.db = client
}

func (as *APISuite) TearDownTest() {
	_, err := as.db.Pool.Exec(context.Background(), "TRUNCATE TABLE accounts, transactions, orders CASCADE")
	as.Require().NoError(err)
}

func (as *APISuite) TearDownSuite() {
	as.db.Close()
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APISuite))
}

func TestMain(m *testing.M) {
	if err := healthCheck(attempts); err != nil {
		log.Fatalf("Integration tests: host %s is not available: %s", host, err)
	}

	log.Printf("Integration tests: host %s is available", host)

	code := m.Run()
	os.Exit(code)
}

func healthCheck(attempts int) error {
	var err error

	for attempts > 0 {
		err = Do(Get(healthPath), Expect().Status().Equal(http.StatusOK))
		if err == nil {
			return nil
		}

		log.Printf("Integration tests: url %s is not available, attempts left: %d", healthPath, attempts)

		time.Sleep(time.Second)

		attempts--
	}

	return err
}
