package cache

import (
	"context"
	"fmt"
	"time"
)

type CacheInterface interface {
	Set(context.Context, string, []byte, time.Duration) error
	Get(context.Context, string) ([]byte, error)
	Exists(context.Context, string) (int64, error)
	Del(context.Context, string) (int64, error)
	Ping(ctx context.Context) (string, error)
	TTL(time.Duration) time.Duration
}

func (c *CacheConfig) Set(ctx context.Context, key string, val []byte, ttl time.Duration) error {

	_, err := c.Client.Set(ctx, fmt.Sprintf("%s_%s", c.Prefix, key), string(val), ttl).Result()
	if err != nil {
		return err
	}

	return nil
}

func (c *CacheConfig) Get(ctx context.Context, key string) ([]byte, error) {
	result, err := c.Client.Get(ctx, fmt.Sprintf("%s_%s", c.Prefix, key)).Result()
	if err != nil {
		return nil, err
	}

	return []byte(result), err
}
func (c *CacheConfig) Exists(ctx context.Context, key string) (int64, error) {

	return c.Client.Exists(ctx, fmt.Sprintf("%s_%s", c.Prefix, key)).Result()
}
func (c *CacheConfig) Del(ctx context.Context, key string) (int64, error) {
	return c.Client.Del(ctx, fmt.Sprintf("%s_%s", c.Prefix, key)).Result()
}

func (c *CacheConfig) Ping(ctx context.Context) (string, error) {
	return c.Client.Ping(ctx).Result()
}

func (c *CacheConfig) TTL(t time.Duration) time.Duration {
	return time.Duration(c.Ttl) * t
}
