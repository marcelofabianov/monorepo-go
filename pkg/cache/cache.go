package cache

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/marcelofabianov/fault"
	"github.com/marcelofabianov/retry"
	"github.com/redis/go-redis/v9"
)

var (
	ErrConnectionFailed = fault.New(
		"redis connection failed after retries",
		fault.WithCode(fault.InfraError),
	)

	ErrInvalidConfig = fault.New(
		"invalid redis configuration",
		fault.WithCode(fault.Invalid),
	)

	ErrAlreadyConnected = fault.New(
		"redis already connected",
		fault.WithCode(fault.Conflict),
	)

	ErrNotConnected = fault.New(
		"redis not connected",
		fault.WithCode(fault.NotFound),
	)

	ErrPingFailed = fault.New(
		"failed to ping redis",
		fault.WithCode(fault.InfraError),
	)

	ErrCloseFailed = fault.New(
		"failed to close redis connection",
		fault.WithCode(fault.Internal),
	)

	ErrOperationFailed = fault.New(
		"redis operation failed",
		fault.WithCode(fault.Internal),
	)

	ErrKeyNotFound = fault.New(
		"key not found in cache",
		fault.WithCode(fault.NotFound),
	)
)

type Cache struct {
	client *redis.Client
	config ConfigProvider
	logger *slog.Logger
}

func New(cfg ConfigProvider) (*Cache, error) {
	if cfg == nil {
		return nil, ErrInvalidConfig
	}

	return &Cache{
		config: cfg,
		logger: slog.Default(),
	}, nil
}

func (c *Cache) SetLogger(logger *slog.Logger) {
	if logger != nil {
		c.logger = logger
	}
}

func (c *Cache) Connect(ctx context.Context) error {
	if c.client != nil {
		return ErrAlreadyConnected
	}

	c.logger.InfoContext(ctx, "Connecting to Redis",
		"host", c.config.GetHost(),
		"port", c.config.GetPort(),
		"db", c.config.GetDB(),
		"max_retries", c.config.GetBackoffRetries(),
	)

	retryConfig := c.getRetryConfig()
	retryConfig.Logger = c.logger

	err := retry.Do(ctx, retryConfig, func(ctx context.Context) error {
		return c.connect(ctx)
	})
	if err != nil {
		c.logger.ErrorContext(ctx, "Failed to connect to Redis",
			"host", c.config.GetHost(),
			"port", c.config.GetPort(),
			"error", err.Error(),
		)

		if fault.IsCode(err, fault.Invalid) {
			return fault.Wrap(ErrConnectionFailed, "connection failed after all retries",
				fault.WithWrappedErr(err),
				fault.WithContext("host", c.config.GetHost()),
				fault.WithContext("port", c.config.GetPort()),
				fault.WithContext("retries", c.config.GetBackoffRetries()),
			)
		}
		return fault.Wrap(err, "redis connection error",
			fault.WithContext("host", c.config.GetHost()),
		)
	}

	c.logger.InfoContext(ctx, "Redis connected successfully",
		"host", c.config.GetHost(),
		"port", c.config.GetPort(),
		"db", c.config.GetDB(),
		"pool_max_idle", c.config.GetMaxIdleConns(),
		"pool_max_active", c.config.GetMaxActiveConns(),
	)

	return nil
}

func (c *Cache) connect(ctx context.Context) error {
	opts := &redis.Options{
		Addr:         fmt.Sprintf("%s:%d", c.config.GetHost(), c.config.GetPort()),
		Password:     c.config.GetPassword(),
		DB:           c.config.GetDB(),
		MaxIdleConns: c.config.GetMaxIdleConns(),
		MinIdleConns: c.config.GetMaxIdleConns() / 2,
		PoolSize:     c.config.GetMaxActiveConns(),
		DialTimeout:  5 * time.Second,
		ReadTimeout:  c.config.GetQueryTimeout(),
		WriteTimeout: c.config.GetExecTimeout(),
	}

	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(ctx, c.config.GetQueryTimeout())
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return fault.Wrap(ErrPingFailed, "ping failed",
			fault.WithWrappedErr(err),
			fault.WithContext("timeout", c.config.GetQueryTimeout().String()),
		)
	}

	c.client = client
	return nil
}

func (c *Cache) Close() error {
	if c.client == nil {
		return ErrNotConnected
	}

	c.logger.Info("Closing Redis connection")

	if err := c.client.Close(); err != nil {
		return fault.Wrap(ErrCloseFailed, "close failed",
			fault.WithWrappedErr(err),
		)
	}

	c.client = nil
	return nil
}

// getRetryConfig converts the config to a retry.Config
func (c *Cache) getRetryConfig() *retry.Config {
	strategy := retry.NewExponentialBackoff(retry.ExponentialBackoffConfig{
		Min:    c.config.GetBackoffMin(),
		Max:    c.config.GetBackoffMax(),
		Factor: float64(c.config.GetBackoffFactor()),
		Jitter: c.config.GetBackoffJitter(),
	})

	return &retry.Config{
		MaxAttempts: c.config.GetBackoffRetries(),
		Strategy:    strategy,
	}
}

func (c *Cache) Ping(ctx context.Context) error {
	if c.client == nil {
		return ErrNotConnected
	}

	pingCtx, cancel := context.WithTimeout(ctx, c.config.GetQueryTimeout())
	defer cancel()

	if err := c.client.Ping(pingCtx).Err(); err != nil {
		return fault.Wrap(ErrPingFailed, "ping failed",
			fault.WithWrappedErr(err),
			fault.WithContext("timeout", c.config.GetQueryTimeout().String()),
		)
	}

	return nil
}

func (c *Cache) HealthCheck(ctx context.Context) error {
	if c.client == nil {
		return ErrNotConnected
	}

	if err := c.Ping(ctx); err != nil {
		return err
	}

	stats := c.client.PoolStats()

	// Validate MaxActiveConns before converting to uint32 to prevent overflow
	maxActive := c.config.GetMaxActiveConns()
	if maxActive < 0 {
		c.logger.WarnContext(ctx, "Invalid max_active_conns, treating as 0", "value", maxActive)
		maxActive = 0
	}
	if maxActive > math.MaxUint32 {
		c.logger.WarnContext(ctx, "max_active_conns exceeds uint32 max, capping",
			"value", maxActive, "max", math.MaxUint32)
		maxActive = math.MaxUint32
	}

	//nolint:gosec // G115: Safe conversion after validation above
	maxActiveUint32 := uint32(maxActive)

	if stats.IdleConns == 0 && stats.TotalConns >= maxActiveUint32 {
		c.logger.WarnContext(ctx, "All Redis connections are in use",
			"total_conns", stats.TotalConns,
			"max_active", maxActive,
		)
	}

	return nil
}

func (c *Cache) IsConnected() bool {
	return c.client != nil
}

func (c *Cache) Client() *redis.Client {
	return c.client
}

func (c *Cache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if c.client == nil {
		return ErrNotConnected
	}

	execCtx, cancel := context.WithTimeout(ctx, c.config.GetExecTimeout())
	defer cancel()

	if err := c.client.Set(execCtx, key, value, expiration).Err(); err != nil {
		c.logger.ErrorContext(ctx, "Redis SET failed",
			"key", key,
			"expiration", expiration.String(),
			"error", err.Error(),
		)
		return fault.Wrap(ErrOperationFailed, "set operation failed",
			fault.WithWrappedErr(err),
			fault.WithContext("key", key),
			fault.WithContext("expiration", expiration.String()),
		)
	}

	return nil
}

func (c *Cache) Get(ctx context.Context, key string) (string, error) {
	if c.client == nil {
		return "", ErrNotConnected
	}

	queryCtx, cancel := context.WithTimeout(ctx, c.config.GetQueryTimeout())
	defer cancel()

	val, err := c.client.Get(queryCtx, key).Result()
	if err == redis.Nil {
		return "", fault.Wrap(ErrKeyNotFound, "key does not exist",
			fault.WithContext("key", key),
		)
	}
	if err != nil {
		c.logger.ErrorContext(ctx, "Redis GET failed",
			"key", key,
			"error", err.Error(),
		)
		return "", fault.Wrap(ErrOperationFailed, "get operation failed",
			fault.WithWrappedErr(err),
			fault.WithContext("key", key),
		)
	}

	return val, nil
}

func (c *Cache) Delete(ctx context.Context, keys ...string) error {
	if c.client == nil {
		return ErrNotConnected
	}

	execCtx, cancel := context.WithTimeout(ctx, c.config.GetExecTimeout())
	defer cancel()

	if err := c.client.Del(execCtx, keys...).Err(); err != nil {
		c.logger.ErrorContext(ctx, "Redis DEL failed",
			"keys", keys,
			"error", err.Error(),
		)
		return fault.Wrap(ErrOperationFailed, "delete operation failed",
			fault.WithWrappedErr(err),
			fault.WithContext("keys", keys),
		)
	}

	return nil
}

func (c *Cache) Exists(ctx context.Context, keys ...string) (int64, error) {
	if c.client == nil {
		return 0, ErrNotConnected
	}

	queryCtx, cancel := context.WithTimeout(ctx, c.config.GetQueryTimeout())
	defer cancel()

	count, err := c.client.Exists(queryCtx, keys...).Result()
	if err != nil {
		c.logger.ErrorContext(ctx, "Redis EXISTS failed",
			"keys", keys,
			"error", err.Error(),
		)
		return 0, fault.Wrap(ErrOperationFailed, "exists operation failed",
			fault.WithWrappedErr(err),
			fault.WithContext("keys", keys),
		)
	}

	return count, nil
}

func (c *Cache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if c.client == nil {
		return ErrNotConnected
	}

	execCtx, cancel := context.WithTimeout(ctx, c.config.GetExecTimeout())
	defer cancel()

	if err := c.client.Expire(execCtx, key, expiration).Err(); err != nil {
		c.logger.ErrorContext(ctx, "Redis EXPIRE failed",
			"key", key,
			"expiration", expiration.String(),
			"error", err.Error(),
		)
		return fault.Wrap(ErrOperationFailed, "expire operation failed",
			fault.WithWrappedErr(err),
			fault.WithContext("key", key),
			fault.WithContext("expiration", expiration.String()),
		)
	}

	return nil
}

func (c *Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	if c.client == nil {
		return 0, ErrNotConnected
	}

	queryCtx, cancel := context.WithTimeout(ctx, c.config.GetQueryTimeout())
	defer cancel()

	ttl, err := c.client.TTL(queryCtx, key).Result()
	if err != nil {
		c.logger.ErrorContext(ctx, "Redis TTL failed",
			"key", key,
			"error", err.Error(),
		)
		return 0, fault.Wrap(ErrOperationFailed, "ttl operation failed",
			fault.WithWrappedErr(err),
			fault.WithContext("key", key),
		)
	}

	return ttl, nil
}

func (c *Cache) Increment(ctx context.Context, key string) (int64, error) {
	if c.client == nil {
		return 0, ErrNotConnected
	}

	execCtx, cancel := context.WithTimeout(ctx, c.config.GetExecTimeout())
	defer cancel()

	val, err := c.client.Incr(execCtx, key).Result()
	if err != nil {
		c.logger.ErrorContext(ctx, "Redis INCR failed",
			"key", key,
			"error", err.Error(),
		)
		return 0, fault.Wrap(ErrOperationFailed, "increment operation failed",
			fault.WithWrappedErr(err),
			fault.WithContext("key", key),
		)
	}

	return val, nil
}

func (c *Cache) Decrement(ctx context.Context, key string) (int64, error) {
	if c.client == nil {
		return 0, ErrNotConnected
	}

	execCtx, cancel := context.WithTimeout(ctx, c.config.GetExecTimeout())
	defer cancel()

	val, err := c.client.Decr(execCtx, key).Result()
	if err != nil {
		c.logger.ErrorContext(ctx, "Redis DECR failed",
			"key", key,
			"error", err.Error(),
		)
		return 0, fault.Wrap(ErrOperationFailed, "decrement operation failed",
			fault.WithWrappedErr(err),
			fault.WithContext("key", key),
		)
	}

	return val, nil
}

func (c *Cache) FlushDB(ctx context.Context) error {
	if c.client == nil {
		return ErrNotConnected
	}

	execCtx, cancel := context.WithTimeout(ctx, c.config.GetExecTimeout())
	defer cancel()

	if err := c.client.FlushDB(execCtx).Err(); err != nil {
		c.logger.ErrorContext(ctx, "Redis FLUSHDB failed", "error", err.Error())
		return fault.Wrap(ErrOperationFailed, "flush db operation failed",
			fault.WithWrappedErr(err),
		)
	}

	c.logger.WarnContext(ctx, "Redis database flushed")
	return nil
}

func (c *Cache) Stats() *redis.PoolStats {
	if c.client == nil {
		return &redis.PoolStats{}
	}
	stats := c.client.PoolStats()
	return stats
}
