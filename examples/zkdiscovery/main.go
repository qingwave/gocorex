package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/qingwave/gocorex/discovery/zkdiscovery"
	"github.com/qingwave/gocorex/syncx/group"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	worker := func(i int, run bool) {
		id := fmt.Sprintf("10.0.0.%d", i)
		val := fmt.Sprintf("10.0.0.%d", i)

		sd, err := zkdiscovery.New(zkdiscovery.ZkDiscoveryConfig{
			Endpoints:      []string{"127.0.0.1"},
			Path:           "/zk/services",
			SessionTimeout: 2 * time.Second,
			Key:            id,
			Val:            val,
			Callbacks: zkdiscovery.DiscoveryCallbacks{
				OnStartedDiscovering: func(services []zkdiscovery.Service) {
					log.Printf("[%s] onstarted, services: %v", id, services)
				},
				OnStoppedDiscovering: func() {
					log.Printf("[%s] onstoped", id)
				},
				OnServiceChanged: func(services []zkdiscovery.Service) {
					log.Printf("[%s] onchanged, services: %v", id, services)
				},
			},
		})

		if err != nil {
			log.Fatalf("failed to create service discovery: %v", err)
		}
		defer sd.Close()

		if !run {
			if sd.UnRegister(context.Background()); err != nil {
				log.Fatalf("failed to unregister service [%s]: %v", id, err)
			}
			return
		}

		if err := sd.Register(context.Background()); err != nil {
			log.Fatalf("failed to register service [%s]: %v", id, err)
		}

		if err := sd.Watch(ctx); err != nil {
			log.Printf("[%s] failed to watch service: %v", id, err)
		}
	}

	wg := group.NewGroup()
	for i := 0; i < 3; i++ {
		id := i
		wg.Go(func() { worker(id, true) })
	}

	go func() {
		time.Sleep(2 * time.Second)
		worker(3, true)
	}()

	// unregister
	go func() {
		time.Sleep(4 * time.Second)
		worker(1, false)
	}()

	// wg.Wait()

	time.Sleep(5 * time.Second)
	cancel()
	time.Sleep(1 * time.Second)
}
