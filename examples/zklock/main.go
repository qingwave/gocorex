package main

import (
	"context"
	"log"
	"time"

	"github.com/go-zookeeper/zk"
	"github.com/qingwave/gocorex/x/syncx/group"
	"github.com/qingwave/gocorex/x/syncx/zklock"
)

func main() {
	ccon, _, err := zk.Connect([]string{"127.0.0.1"}, 3*time.Second)
	if err != nil {
		log.Fatalf("failed to create zk ccon: %v", err)
	}
	defer ccon.Close()

	counter := 0
	path := "/worker/lock"
	worker := func(i int) {
		m, err := zklock.New(zklock.ZkLockConfig{
			Conn: ccon,
			Path: path,
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

	wg := group.NewGroup()
	for i := 1; i <= 5; i++ {
		id := i
		wg.Go(func() {
			worker(id)
		})
	}

	wg.Wait()
}
