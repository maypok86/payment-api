package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maypok86/payment-api/internal/config"
	"github.com/maypok86/payment-api/internal/pkg/server"
	"go.uber.org/zap"
)

type App struct {
	logger     *zap.Logger
	httpServer *server.Server
}

func New(ctx context.Context, logger *zap.Logger) (*App, error) {
	cfg := config.Get()

	return &App{
		logger: logger,
		httpServer: server.New(
			http.NewServeMux(),
			server.WithHost(cfg.HTTP.Host),
			server.WithPort(cfg.HTTP.Port),
			server.WithMaxHeaderBytes(cfg.HTTP.MaxHeaderBytes),
			server.WithReadTimeout(cfg.HTTP.ReadTimeout),
			server.WithWriteTimeout(cfg.HTTP.WriteTimeout),
		),
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	eChan := make(chan error)
	interrupt := make(chan os.Signal, 1)

	a.logger.Info("Http server is starting")

	go func() {
		if err := a.httpServer.Start(); err != nil {
			eChan <- fmt.Errorf("listen and serve: %w", err)
		}
	}()

	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	select {
	case err := <-eChan:
		return fmt.Errorf("payment-api started failed: %w", err)
	case <-interrupt:
	}

	const httpShutdownTimeout = 5 * time.Second
	if err := a.httpServer.Stop(ctx, httpShutdownTimeout); err != nil {
		return fmt.Errorf("stop http server: %w", err)
	}

	return nil
}
