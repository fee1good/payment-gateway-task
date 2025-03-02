package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"payment-gateway/configs/logger"
	"payment-gateway/db"
)

type Cache interface {
	GetGatewaysByCountry(ctx context.Context, dbHandler db.Storage, countryID int) ([]db.Gateway, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(client *redis.Client) Cache {
	return &RedisCache{client: client}
}

func InitRedis(ctx context.Context, addr string, password string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	// Test the connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		logger.Error("Failed to connect to Redis", "error", err)
	} else {
		logger.Info("Successfully connected to Redis", "addr", addr)
	}

	return client
}

func (c *RedisCache) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *RedisCache) GetGatewaysByCountry(ctx context.Context, dbHandler db.Storage, countryID int) ([]db.Gateway, error) {
	cacheKey := fmt.Sprintf("gateways:country:%d", countryID)

	// Try to get from cache first
	val, err := c.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit
		var gateways []db.Gateway
		if err := json.Unmarshal([]byte(val), &gateways); err == nil {
			return gateways, nil
		}
	}

	gateways, err := dbHandler.GetGatewaysByCountry(ctx, countryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get gateways: %v", err)
	}

	gatewaysJSON, _ := json.Marshal(gateways)
	err = c.Set(ctx, cacheKey, gatewaysJSON, 5*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to set cache: %v", err)
	}

	return gateways, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}
