package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type Cache struct {
	client *redis.Client
}

func New(addr string) (*Cache, error) {
	// TODO redis configuration
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	_, err := rdb.Ping(context.TODO()).Result()
	if err != nil {
		return nil, err
	}

	return &Cache{
		client: rdb,
	}, nil
}

// SaveURL stores the alias and original URL in the cache with an expiration time
func (c *Cache) SaveURL(ctx context.Context, originalURL string, alias string, expiration time.Duration) error {
	return c.client.Set(ctx, alias, originalURL, expiration).Err()
}

// GetURL retrieves the original URL from the cache by the alias
func (c *Cache) GetURL(ctx context.Context, alias string) (string, error) {
	result, err := c.client.Get(ctx, alias).Result()
	if err == redis.Nil {
		return "", nil // Alias not found in cache
	} else if err != nil {
		return "", err
	}

	return result, nil
}

// Close closes the Redis client connection
func (c *Cache) Close() error {
	return c.client.Close()
}
