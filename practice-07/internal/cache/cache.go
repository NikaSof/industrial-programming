package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrMiss = redis.Nil

type Cache struct {
	rdb *redis.Client
}

func New(addr string) *Cache {
	return &Cache{rdb: redis.NewClient(&redis.Options{
		Addr:                  addr,
		DB:                    0,
		DialTimeout:           2 * time.Second,
		ReadTimeout:           2 * time.Second,
		WriteTimeout:          2 * time.Second,
		ContextTimeoutEnabled: true,
	})}
}

func (c *Cache) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Cache) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, value, ttl).Err()
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return c.rdb.TTL(ctx, key).Result()
}

func (c *Cache) Close() error {
	return c.rdb.Close()
}
