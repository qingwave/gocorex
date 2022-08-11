package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/qingwave/gocorex/discovery/etcdiscovery"
	"github.com/qingwave/gocorex/syncx/group"
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

	worker := func(i int, run bool) {
		id := fmt.Sprintf("worker-%d", i)
		val := fmt.Sprintf("10.0.0.%d", i)

		sd, err := etcdiscovery.New(etcdiscovery.EtcdDiscoveryConfig{
			Client:     client,
			Prefix:     "/services",
			Key:        id,
			Val:        val,
			TTLSeconds: 2,
			Callbacks: etcdiscovery.DiscoveryCallbacks{
				OnStartedDiscovering: func(services []etcdiscovery.Service) {
					log.Printf("[%s], onstarted, services: %v", id, services)
				},
				OnStoppedDiscovering: func() {
					log.Printf("[%s], onstoped", id)
				},
				OnServiceChanged: func(services []etcdiscovery.Service, event etcdiscovery.DiscoveryEvent) {
					log.Printf("[%s], onchanged, services: %v, event: %v", id, services, event)
				},
			},
		})

		if err != nil {
			log.Fatalf("failed to create service etcdiscovery: %v", err)
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

		if err := sd.Watch(context.Background()); err != nil {
			log.Fatalf("failed to watch service: %v", err)
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
		worker(2, false)
	}()

	wg.Wait()
}
