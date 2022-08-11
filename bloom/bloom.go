package bloom

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/spaolacci/murmur3"
)

type Filter interface {
	Add(data []byte) error
	AddWithContext(ctx context.Context, data []byte) error
	Exists(data []byte) (bool, error)
	ExistsWithContext(ctx context.Context, data []byte) (bool, error)
	Reset() error
	ResetWithContext(ctx context.Context) error
}

type BloomFilterConfig struct {
	Client *redis.Client
	Key    string
	Bits   uint
	Maps   uint
}

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
	defaultMaps = 14
)

func New(config BloomFilterConfig) (Filter, error) {
	if config.Client == nil {
		return nil, fmt.Errorf("redis client must not be nil")
	}

	if config.Key == "" {
		return nil, fmt.Errorf("redis key must not be nil")
	}

	if config.Bits == 0 {
		return nil, fmt.Errorf("bits must great than zero")
	}

	if config.Maps == 0 {
		config.Maps = defaultMaps
	}

	return &BloomFilter{
		BloomFilterConfig: config,
	}, nil
}

type BloomFilter struct {
	BloomFilterConfig
}

func (f *BloomFilter) Reset() error {
	return f.ResetWithContext(context.Background())
}

func (f *BloomFilter) ResetWithContext(ctx context.Context) error {
	return f.Client.Del(ctx, f.Key).Err()
}

func (f *BloomFilter) Add(data []byte) error {
	return f.AddWithContext(context.Background(), data)
}

func (f *BloomFilter) AddWithContext(ctx context.Context, data []byte) error {
	args := getArgs(f.getLocations(data))
	_, err := f.Client.Eval(ctx, setScript, []string{f.Key}, args).Result()
	if err != nil && err != redis.Nil {
		return err
	}
	return nil
}

func (f *BloomFilter) Exists(data []byte) (bool, error) {
	return f.ExistsWithContext(context.Background(), data)
}

func (f *BloomFilter) ExistsWithContext(ctx context.Context, data []byte) (bool, error) {
	args := getArgs(f.getLocations(data))
	resp, err := f.Client.Eval(ctx, getScript, []string{f.Key}, args).Result()
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

func (f *BloomFilter) getLocations(data []byte) []uint {
	locations := make([]uint, f.Maps)
	for i := 0; i < int(f.Maps); i++ {
		val := murmur3.Sum64(append(data, byte(i)))
		locations[i] = uint(val) % f.Bits
	}
	return locations
}

func getArgs(locations []uint) []string {
	args := make([]string, 0)
	for _, l := range locations {
		args = append(args, strconv.FormatUint(uint64(l), 10))
	}
	return args
}
