package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/qingwave/gocorex/syncx/redislock"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "123456",
		DB:       0,
	})

	counter := 0
	worker := func(i int) {
		id := fmt.Sprintf("worker%d", i)

		m, err := redislock.New(redislock.RedisLockConfig{
			Client:            client,
			Key:               "test-lock",
			ID:                id,
			Expiration:        10 * time.Second,
			LockRetryDuration: 1 * time.Second,
		})

		if err != nil {
			log.Fatalf("failed to create lock: %v", err)
		}

		err = m.Lock(context.Background())
		log.Printf("worker %d obtain lock, err: %v", i, err)
		defer m.UnLock(context.Background())

		counter++
		log.Printf("worker %d, add counter %d", i, counter)
		time.Sleep(1 * time.Second)
	}

	wg := sync.WaitGroup{}
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		id := i
		go func() {
			defer wg.Done()
			worker(id)
		}()
	}

	wg.Wait()
}
