package cache

import (
	"context"
	"fmt"
	"payment/pkg/config"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type CacheDatabase struct {
	client            *redis.Client
	log               *zap.Logger
	expirationMinutes int
}

func (c *CacheDatabase) Get(ctx context.Context, key string) (string, error) {
	value, err := c.client.Get(ctx, key).Result()
	c.log.Info("Get cache", zap.String("key", key), zap.String("value", value), zap.Error(err))
	return value, err
}

func (c *CacheDatabase) Set(ctx context.Context, key string, value []byte) error {
	err := c.client.Set(ctx, key, value, time.Duration(c.expirationMinutes)*time.Minute).Err()
	c.log.Info("Set cache", zap.String("key", key), zap.String("value", string(value)), zap.Error(err))
	return err
}

func (c *CacheDatabase) Del(ctx context.Context, key string) error {
	err := c.client.Del(ctx, key).Err()
	c.log.Info("Del cache", zap.String("key", key), zap.Error(err))
	return err
}

func (c *CacheDatabase) Close() error {
	return c.client.Close()
}

func New(cfg config.CacheConfig, log *zap.Logger) (*CacheDatabase, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return &CacheDatabase{client: client, log: log, expirationMinutes: cfg.ExpirationMinutes}, nil
}
