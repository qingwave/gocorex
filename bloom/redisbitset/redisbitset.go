package redisbitset

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/qingwave/gocorex/bloom"
)

const (
	// set bit lua script
	setScript = `
for _, offset in ipairs(ARGV) do
	redis.call("setbit", KEYS[1], offset, 1)
end
`
	// get lua script
	getScript = `
for _, offset in ipairs(ARGV) do
	if tonumber(redis.call("getbit", KEYS[1], offset)) == 0 then
		return false
	end
end
return true
`
)

func New(client *redis.Client, key string) (bloom.BitSet, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client must not be nil")
	}

	if key == "" {
		return nil, fmt.Errorf("key must not be empty")
	}

	return &RedisBitSet{
		client: client,
		key:    key,
	}, nil
}

type RedisBitSet struct {
	client *redis.Client
	key    string
}

func (r *RedisBitSet) Reset(opts ...bloom.Option) error {
	ctx := bloom.NewFilterOptions(opts...).Context
	return r.client.Del(ctx, r.key).Err()
}

func (r *RedisBitSet) Add(items []int, opts ...bloom.Option) error {
	ctx := bloom.NewFilterOptions(opts...).Context

	_, err := r.client.Eval(ctx, setScript, []string{r.key}, getArgs(items)).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	return nil
}

func (r *RedisBitSet) Exists(items []int, opts ...bloom.Option) (bool, error) {
	ctx := bloom.NewFilterOptions(opts...).Context

	resp, err := r.client.Eval(ctx, getScript, []string{r.key}, getArgs(items)).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}

	exists, ok := resp.(int64)
	if !ok {
		return false, nil
	}

	return exists == 1, nil
}

func getArgs(locations []int) []string {
	args := make([]string, 0)
	for _, l := range locations {
		args = append(args, strconv.Itoa(l))
	}
	return args
}
