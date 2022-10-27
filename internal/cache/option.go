package cache

import "time"

type Option func(*ReportCache)

func WithTTL(ttl time.Duration) Option {
	return func(c *ReportCache) {
		c.ttl = ttl
	}
}

func WithCleanInterval(cleanInterval time.Duration) Option {
	return func(c *ReportCache) {
		c.cleanInterval = cleanInterval
	}
}
