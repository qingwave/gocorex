package bloom

import (
	"context"
	"fmt"

	"github.com/spaolacci/murmur3"
)

const defaultMaps = 14

type Option func(opt *FilterOption)

type FilterOption struct {
	Context context.Context
}

func WithContext(ctx context.Context) Option {
	return func(opt *FilterOption) {
		opt.Context = ctx
	}
}

func NewFilterOptions(opts ...Option) *FilterOption {
	fo := &FilterOption{Context: context.Background()}
	for _, opt := range opts {
		opt(fo)
	}
	return fo
}

type BloomFilterConfig struct {
	BitSet BitSet
	Key    string
	Bits   int
	Maps   int
}

func New(config BloomFilterConfig) (*BloomFilter, error) {
	if config.Bits <= 0 {
		return nil, fmt.Errorf("bits must great than zero")
	}

	if config.Maps <= 0 {
		config.Maps = defaultMaps
	}

	if config.BitSet == nil {
		config.BitSet = newBitSet(config.Bits)
	}

	return &BloomFilter{
		BloomFilterConfig: config,
	}, nil
}

type BloomFilter struct {
	BloomFilterConfig
}

func (f *BloomFilter) Reset(opts ...Option) error {
	return f.BitSet.Reset(opts...)
}

func (f *BloomFilter) Add(data []byte, opts ...Option) error {
	if len(data) == 0 {
		return nil
	}

	locations := f.getLocations(data)

	return f.BitSet.Add(locations, opts...)
}

func (f *BloomFilter) Exists(data []byte, opts ...Option) (bool, error) {
	if len(data) == 0 {
		return false, nil
	}

	locations := f.getLocations(data)

	return f.BitSet.Exists(locations, opts...)
}

func (f *BloomFilter) getLocations(data []byte) []int {
	locations := make([]int, f.Maps)
	for i := 0; i < int(f.Maps); i++ {
		val := murmur3.Sum64(append(data, byte(i)))
		locations[i] = int(val % uint64(f.Bits))
	}
	return locations
}
