package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/maypok86/payment-api/internal/app/api"
	"github.com/maypok86/payment-api/internal/config"
	"github.com/maypok86/payment-api/internal/pkg/logger"
	"go.uber.org/zap"
)

var (
	version   string
	buildDate string
)

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	cfg := config.Get()
	l := logger.New(os.Stdout, cfg.Logger.Level)
	l.Info("conduit", zap.String("version", version), zap.String("build_date", buildDate))

	app, err := api.New(ctx, l)
	if err != nil {
		return fmt.Errorf("create app: %w", err)
	}

	if err := app.Run(ctx); err != nil {
		return fmt.Errorf("run app: %w", err)
	}

	return nil
}
