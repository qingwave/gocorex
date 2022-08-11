package rate

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	Inf = 1<<63 - 1

	reserveNScript = `
	local limit_key = KEYS[1]

	local qps = tonumber(ARGV[1])
	local burst = tonumber(ARGV[2])
	local now = ARGV[3]
	local cost = tonumber(ARGV[4])
	local max_wait = tonumber(ARGV[5])

	local tokens = redis.call("hget", limit_key, "token")
	if not tokens then
		tokens = burst
	end

	local last_time = redis.call("hget", limit_key, "last_time")
	if not last_time then
		last_time = 0
	end

	local last_event = redis.call("hget", limit_key, "last_event")
	if not last_event then
		last_event = 0
	end

	local delta = math.max(0, now-last_time)
	local new_tokens = math.min(burst, delta * qps + tokens)
	new_tokens = new_tokens - cost

	local wait_period = 0
	if new_tokens < 0 and qps > 0 then
		wait_period = wait_period - new_tokens / qps
	end

	wait_period = math.ceil(wait_period)

	local time_act = now + wait_period

	local ok = (cost <= burst and wait_period <= max_wait and qps > 0) or (qps == 0 and new_tokens >= 0)

	if ok then
		redis.call("hset", limit_key, "token", new_tokens, "last_time", now, "last_event", time_act)
	end

	return {ok, wait_period}
	`
	cancelAtScript = `
	local limit_key = KEYS[1]

	local qps = tonumber(ARGV[1])
	local burst = tonumber(ARGV[2])
	local now = ARGV[3]
	local cost = tonumber(ARGV[4])
	local event_act = tonumber(ARGV[5])

	local tokens = redis.call("hget", limit_key, "token")
	if not tokens then
		tokens = burst
	end

	local last_time = redis.call("hget", limit_key, "last_time")
	if not last_time then
		last_time = 0
	end

	local last_event = redis.call("hget", limit_key, "last_event")
	if not last_event then
		last_event = 0
	end

	local restore_tokens = cost - qps * math.max(0, last_event - event_act)
	if restore_tokens <= 0 then
		return false
	end

	local delta = math.max(0, now-last_time)
	local new_tokens = math.min(burst, delta * qps + tokens)
	new_tokens = math.min(burst, new_tokens + restore_tokens)
	if new_tokens > 0 then
		return {new_tokens, delta, restore_tokens}
	end

	redis.call("hset", limit_key, "token", new_tokens, "last_time", now)
	if last_event == event_act and qps > 0 then
		pre_event = time_act - math.ceil(new_tokens / qps)
		if pre_event >= now then
			redis.call("hset", limit_key, "last_event", pre_event)
		end
	end

	return true
	`
)

func NewRedisLimiter(config RedisLimiterConfig) (*RedisLimiter, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("redis client must not be nil")
	}

	if config.QPS > config.Burst {
		return nil, fmt.Errorf("qps must less than burst")
	}

	if config.QPS < 0 || config.Burst < 0 {
		return nil, fmt.Errorf("invaild qps or burst")
	}

	return &RedisLimiter{RedisLimiterConfig: config}, nil
}

type RedisLimiterConfig struct {
	Client *redis.Client
	Key    string
	Burst  int
	QPS    int
}

type RedisLimiter struct {
	RedisLimiterConfig
}

func (lim *RedisLimiter) Reset() error {
	err := lim.Client.Del(context.Background(), lim.Key).Err()
	if err != nil && err != redis.Nil {
		return err
	}
	return nil
}

// Allow is shorthand for AllowN(time.Now(), 1).
func (lim *RedisLimiter) Allow() (bool, error) {
	return lim.AllowN(time.Now(), 1)
}

// AllowN reports whether n events may happen at time now.
// Use this method if you intend to drop / skip events that exceed the rate limit.
// Otherwise use Reserve or Wait.
func (lim *RedisLimiter) AllowN(now time.Time, n int) (bool, error) {
	r, err := lim.reserveN(now, n, 0)
	if err != nil {
		return false, err
	}
	return r.OK(), nil
}

func (lim *RedisLimiter) ReserveN(now time.Time, n int) (*Reservation, error) {
	return lim.reserveN(now, n, Inf)
}

// Wait is shorthand for WaitN(ctx, 1).
func (lim *RedisLimiter) Wait(ctx context.Context) (err error) {
	return lim.WaitN(ctx, 1)
}

func (lim *RedisLimiter) WaitN(ctx context.Context, n int) (err error) {
	if n > lim.Burst {
		return fmt.Errorf("rate: Wait(n=%d) exceeds limiter's burst %d", n, lim.Burst)
	}
	// Check if ctx is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	// Determine wait limit
	now := time.Now()
	waitLimit := Inf
	if deadline, ok := ctx.Deadline(); ok {
		waitLimit = int(deadline.Sub(now))
	}
	// Reserve
	r, err := lim.reserveN(now, n, waitLimit)
	if err != nil {
		return err
	}
	if !r.OK() {
		return fmt.Errorf("rate: Wait(n=%d) would exceed context deadline", n)
	}

	// Wait if necessary
	delay := r.DelayFrom(now)
	if delay == 0 {
		return nil
	}
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-t.C:
		// We can proceed.
		return nil
	case <-ctx.Done():
		// Context was canceled before we could proceed.  Cancel the
		// reservation, which may permit other events to proceed sooner.
		if err := r.Cancel(); err != nil {
			return err
		}
		return ctx.Err()
	}
}

type Reservation struct {
	ok        bool
	lim       *RedisLimiter
	tokens    int
	timeToAct time.Time
}

func (r *Reservation) OK() bool {
	return r.ok
}

// Delay is shorthand for DelayFrom(time.Now()).
func (r *Reservation) Delay() time.Duration {
	return r.DelayFrom(time.Now())
}

// DelayFrom returns the duration for which the reservation holder must wait
// before taking the reserved action.  Zero duration means act immediately.
// InfDuration means the limiter cannot grant the tokens requested in this
// Reservation within the maximum wait time.
func (r *Reservation) DelayFrom(now time.Time) time.Duration {
	if !r.ok {
		return Inf
	}
	delay := r.timeToAct.Sub(now)
	if delay < 0 {
		return 0
	}
	return delay
}

// Cancel is shorthand for CancelAt(time.Now()).
func (r *Reservation) Cancel() error {
	return r.CancelAt(time.Now())
}

// CancelAt indicates that the reservation holder will not perform the reserved action
// and reverses the effects of this Reservation on the rate limit as much as possible,
// considering that other reservations may have already been made.
func (r *Reservation) CancelAt(now time.Time) error {
	if !r.ok {
		return nil
	}

	if r.lim.QPS == Inf || r.tokens == 0 || r.timeToAct.Before(now) {
		return nil
	}

	lim := r.lim
	res, err := lim.Client.Eval(context.Background(), cancelAtScript, []string{lim.Key}, lim.QPS, lim.Burst, now.Unix(), r.tokens, r.timeToAct.Unix()).Result()
	if err != nil && err != redis.Nil {
		return nil
	}
	fmt.Println(res)

	return nil
}

// reserveN is a helper method for AllowN, ReserveN, and WaitN.
// maxFutureReserve specifies the maximum reservation wait duration allowed.
// reserveN returns Reservation, not *Reservation, to avoid allocation in AllowN and WaitN.
func (lim *RedisLimiter) reserveN(now time.Time, n int, maxFutureReserveSecond int) (*Reservation, error) {
	if lim.QPS == Inf {
		return &Reservation{
			ok:        true,
			lim:       lim,
			tokens:    n,
			timeToAct: now,
		}, nil
	}

	res, err := lim.Client.Eval(context.Background(), reserveNScript, []string{lim.Key}, lim.QPS, lim.Burst, now.Unix(), n, maxFutureReserveSecond).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	values, ok := res.([]interface{})
	if !ok {
		return nil, fmt.Errorf("invaild response, excepted []interface{}, got %v", res)
	}

	if len(values) != 2 {
		return nil, fmt.Errorf("invaild response length, excepted %d, got %d", 2, len(values))
	}

	allow, allowOK := values[0].(int64)
	wait, waitOK := values[1].(int64)
	if !allowOK || !waitOK {
		return nil, fmt.Errorf("invaild response type, excepted int64")
	}

	return &Reservation{
		ok:        allow == 1,
		lim:       lim,
		tokens:    n,
		timeToAct: now.Add(time.Duration(wait) * time.Second),
	}, nil
}
