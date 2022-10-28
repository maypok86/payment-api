package integration

import (
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	. "github.com/Eun/go-hit"
)

const (
	host       = "backend:8080"
	healthPath = "http://" + host + "/health"
	attempts   = 20
	// basePath   = "http://" + host + "/api/v1".
)

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

func TestOK(t *testing.T) {
	Test(t,
		Get(healthPath),
		Expect().Status().Equal(http.StatusOK),
	)
}
