package main

import (
	"log"

	"github.com/qingwave/gocorex/x/bloom"

	"github.com/go-redis/redis/v8"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456",
		DB:       0,
	})

	filter, err := bloom.New(bloom.BloomFilterConfig{
		Client: client,
		Key:    "test-bloom",
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
		if err := filter.Add([]byte(key)); err != nil {
			log.Printf("failed to add key %s: %v", key, err)
		}
		log.Printf("add key %s successfully", key)
	}

	check("key1")

	add("key1")

	check("key1")
}
