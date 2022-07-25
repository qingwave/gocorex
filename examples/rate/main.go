package main

import (
	"context"
	"log"

	"github.com/qingwave/gocorex/x/rate"

	"github.com/go-redis/redis/v8"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456",
		DB:       0,
	})

	r, err := rate.NewRedisLimiter(rate.RedisLimiterConfig{
		Client: client,
		Key:    "test-rate",
		Burst:  2,
		QPS:    1,
	})
	if err != nil {
		log.Fatalf("failed to create rate limiter: %v", err)
	}

	for i := 0; i < 5; i++ {
		err := r.Wait(context.Background())
		log.Printf("worker %d allowed: %v", i, err)
	}
}
