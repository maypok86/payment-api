package api

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

type App struct {
	logger *zap.Logger
}

func New(ctx context.Context, logger *zap.Logger) (App, error) {
	return App{
		logger: logger,
	}, nil
}

func (a App) Run(ctx context.Context) error {
	eChan := make(chan error)
	interrupt := make(chan os.Signal, 1)

	a.logger.Info("Http server is starting")

	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-eChan:
		return fmt.Errorf("payment-api started failed: %w", err)
	case <-interrupt:
	}

	return nil
}
