package postgres

import "time"

type Option func(*Client)

func WithMaxPoolSize(maxPoolSize int) Option {
	return func(c *Client) {
		c.maxPoolSize = maxPoolSize
	}
}

func WithConnAttempts(connAttempts int) Option {
	return func(c *Client) {
		c.connAttempts = connAttempts
	}
}

func WithConnTimeout(connTimeout time.Duration) Option {
	return func(c *Client) {
		c.connTimeout = connTimeout
	}
}
