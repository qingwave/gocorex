package mutex

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/qingwave/gocorex/x/utils/wait"

	"github.com/go-redis/redis/v8"
)

func NewRedisLock(config RedisLockConfig) (*RedisLock, error) {
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
}

type RedisLock struct {
	RedisLockConfig
}

func (l *RedisLock) TryLock(ctx context.Context) (bool, error) {
	return l.Client.SetNX(ctx, l.Key, l.ID, l.Expiration).Result()
}

func (l *RedisLock) Lock(ctx context.Context, period time.Duration, factor, jitter float64) error {
	backoff := wait.Backoff{
		Duration: period,
		Factor:   factor,
		Jitter:   jitter,
		Steps:    math.MaxUint32,
		Cap:      l.Expiration,
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
