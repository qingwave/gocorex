package redislock

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/qingwave/gocorex/syncx"
	"github.com/qingwave/gocorex/utils/wait"

	"github.com/go-redis/redis/v8"
)

const (
	Jitter = 1.2
)

func New(config RedisLockConfig) (syncx.Locker, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("redis client must not be nil")
	}

	if config.Key == "" {
		return nil, fmt.Errorf("redis key must be set")
	}

	if config.ID == "" {
		return nil, fmt.Errorf("id must be set")
	}

	if config.Expiration == 0 {
		return nil, fmt.Errorf("expiration must great than zero")
	}

	return &RedisLock{
		RedisLockConfig: config,
	}, nil
}

type RedisLockConfig struct {
	Client     *redis.Client
	Key        string
	ID         string
	Expiration time.Duration

	LockRetryDuration time.Duration
}

type RedisLock struct {
	RedisLockConfig
}

func (l *RedisLock) TryLock(ctx context.Context) (bool, error) {
	return l.Client.SetNX(ctx, l.Key, l.ID, l.Expiration).Result()
}

func (l *RedisLock) Lock(ctx context.Context) error {
	backoff := wait.Backoff{
		Duration: l.LockRetryDuration,
		Jitter:   Jitter,
		Steps:    math.MaxUint32,
	}
	return wait.ExponentialBackoffWithContext(ctx, backoff, func() (bool, error) {
		return l.TryLock(ctx)
	})
}

const (
	unLockScript = `
if (redis.call("get", KEYS[1]) == KEYS[2]) then
	redis.call("del", KEYS[1])
	return true
end
return false
`
)

func (l *RedisLock) UnLock(ctx context.Context) error {
	_, err := l.Client.Eval(ctx, unLockScript, []string{l.Key, l.ID}).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	return nil
}

func (l *RedisLock) Close() error {
	return l.UnLock(context.Background())
}
