package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/qingwave/gocorex/x/syncx/leaderelection"
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

	prefix := "/worker/election"
	worker := func(i int) {
		id := fmt.Sprintf("worker-%d", i)

		le, err := leaderelection.New(leaderelection.LeaderElectionConfig{
			Client:       client,
			LeaseSeconds: 15,
			Prefix:       prefix,
			Identity:     id,
			Callbacks: leaderelection.LeaderCallbacks{
				OnStartedLeading: func(ctx context.Context) {
					log.Printf("OnStarted[%s]: acquire new leader", id)
					time.Sleep(3 * time.Second)
					log.Printf("OnStarted[%s]: worker done", id)
				},
				OnStoppedLeading: func() {
					log.Printf("OnStopped[%s]: exit", id)
				},
				OnNewLeader: func(identity string) {
					log.Printf("OnNewLeader[%s]: new leader %s", id, identity)
				},
			},
		})

		if err != nil {
			log.Fatalf("failed to create leader election: %v", err)
		}
		defer le.Close()

		le.Run(context.Background())
	}

	wg := sync.WaitGroup{}
	for i := 1; i <= 3; i++ {
		wg.Add(1)
		id := i
		go func() {
			defer wg.Done()
			worker(id)
		}()
	}

	wg.Wait()
}
