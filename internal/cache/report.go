package cache

import (
	"sync"
	"time"

	"github.com/maypok86/payment-api/internal/domain/report"
)

const (
	defaultTTL           = 1 * time.Hour
	defaultCleanInterval = 1 * time.Hour
)

type item struct {
	value      []byte
	lastAccess int64
}

type ReportCache struct {
	ttl           time.Duration
	cleanInterval time.Duration
	mutex         sync.Mutex
	cache         map[string]*item
}

func NewReportCache(opts ...Option) *ReportCache {
	reportCache := &ReportCache{
		cache:         make(map[string]*item),
		ttl:           defaultTTL,
		cleanInterval: defaultCleanInterval,
	}

	for _, opt := range opts {
		opt(reportCache)
	}

	go reportCache.clean()

	return reportCache
}

func (c *ReportCache) clean() {
	for {
		<-time.After(c.cleanInterval)

		expire := time.Now().Unix() - int64(c.ttl)
		c.mutex.Lock()
		for key, reportItem := range c.cache {
			if expire > reportItem.lastAccess {
				delete(c.cache, key)
			}
		}
		c.mutex.Unlock()
	}
}

func (c *ReportCache) Set(key string, value []byte) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache[key] = &item{
		value:      value,
		lastAccess: time.Now().Unix(),
	}

	return nil
}

func (c *ReportCache) Get(key string) ([]byte, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	reportItem, ok := c.cache[key]

	if !ok {
		return nil, report.ErrNotFound
	}

	reportItem.lastAccess = time.Now().Unix()

	return reportItem.value, nil
}

func (c *ReportCache) IsExist(key string) bool {
	_, err := c.Get(key)

	return err == nil
}
