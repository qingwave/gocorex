package main

import (
	"context"
	"log"
	"time"

	"github.com/qingwave/gocorex/x/syncx/etcdlock"
	"github.com/qingwave/gocorex/x/syncx/group"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		log.Fatalf("failed to create etcd lock: %v", err)
	}
	defer client.Close()

	counter := 0
	prefix := "/worker/lock"
	worker := func(i int) {
		m, err := etcdlock.New(etcdlock.EtcdLockConfig{
			Client:     client,
			Prefix:     prefix,
			TTLSeconds: 5,
		})
		if err != nil {
			log.Fatalf("failed to create lock: %v", err)
		}
		defer m.Close()

		err = m.Lock(context.Background())
		log.Printf("worker %d obtain lock, err: %v", i, err)

		defer m.UnLock(context.Background())

		counter++
		log.Printf("worker %d, add counter %d", i, counter)
		time.Sleep(2 * time.Second)
	}

	wg := group.Group{}
	for i := 1; i <= 5; i++ {
		id := i
		wg.Go(func() {
			worker(id)
		})
	}

	wg.Wait()
}
