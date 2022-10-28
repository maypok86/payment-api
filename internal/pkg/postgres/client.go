package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v4/pgxpool"
)

const (
	defaultMaxPoolSize  int           = 4
	defaultConnAttempts int           = 10
	defaultConnTimeout  time.Duration = time.Second
)

var ErrUnableToConnect = errors.New("all attempts are exceeded. Unable to connect to instance")

type Client struct {
	maxPoolSize  int
	connAttempts int
	connTimeout  time.Duration
	Builder      sq.StatementBuilderType
	Pool         *pgxpool.Pool
}

func NewClient(ctx context.Context, config ConnectionConfig, opts ...Option) (*Client, error) {
	instance := &Client{
		connAttempts: defaultConnAttempts,
		connTimeout:  defaultConnTimeout,
		maxPoolSize:  defaultMaxPoolSize,
	}

	for _, opt := range opts {
		opt(instance)
	}

	instance.Builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	dsn := config.getDSN()
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	poolCfg.MaxConns = int32(instance.maxPoolSize)

	for instance.connAttempts > 0 {
		instance.Pool, err = pgxpool.ConnectConfig(ctx, poolCfg)
		if err == nil {
			break
		}

		log.Printf("Postgres is trying to connect, attempts left: %d", instance.connAttempts)
		time.Sleep(instance.connTimeout)

		instance.connAttempts--
	}

	if err != nil {
		return nil, ErrUnableToConnect
	}

	return instance, nil
}

func (c *Client) Close() {
	if c.Pool != nil {
		c.Pool.Close()
	}
}
