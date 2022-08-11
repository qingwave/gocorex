package main

import (
	"context"
	"log"

	"github.com/qingwave/gocorex/bloom"
	"github.com/qingwave/gocorex/bloom/redisbitset"

	"github.com/go-redis/redis/v8"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456",
		DB:       0,
	})

	redisBitSet, err := redisbitset.New(client, "test-bloom")
	if err != nil {
		log.Fatalf("faield to create redis bitset, %v", err)
	}

	filter, err := bloom.New(bloom.BloomFilterConfig{
		BitSet: redisBitSet,
		Bits:   2 ^ 16,
	})

	defer filter.Reset()

	if err != nil {
		log.Fatalf("failed to create bloom filter: %v", err)
	}

	check := func(key string) {
		ok, err := filter.Exists([]byte(key))
		if err != nil {
			log.Printf("failed to call exists: %v", err)
		}
		log.Printf("key [%s] exists: %t", key, ok)
	}

	add := func(key string) {
		if err := filter.Add([]byte(key), bloom.WithContext(context.TODO())); err != nil {
			log.Printf("failed to add key %s: %v", key, err)
		}
		log.Printf("add key %s successfully", key)
	}

	check("key1")

	add("key1")

	check("key1")
}
