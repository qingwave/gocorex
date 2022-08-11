package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/qingwave/gocorex/pubsub"
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

	ps, _ := pubsub.New(pubsub.EtcdPubSubConfig{
		Client: client,
		Prefix: "/pubsub",
	})

	topic := "demo-topic"

	ps.Reset(context.Background(), topic)

	producer := func(id int, name, msg string) {
		err := ps.Publish(context.Background(), topic, pubsub.Msg{
			Name: name,
			Val:  msg,
		})
		if err != nil {
			log.Printf("[%v]-> failed to publish: %v", id, err)
		}
		log.Printf("[%v]-> publish success: %v", id, name)
	}

	customer := func(id int) {
		ch, err := ps.Subscribe(context.Background(), topic)
		if err != nil {
			log.Printf("[%v]<- failed to subscribe topic: %v", id, err)
		}
		for {
			msg, ok := <-ch
			if !ok {
				log.Printf("[%v]<- subscribe closed", id)
			}

			log.Printf("[%v]<- receive msg: %v", id, msg)
		}
	}

	producerCount := 2
	customerCount := 3

	for i := 1; i <= customerCount; i++ {
		go customer(i)
	}

	for i := 0; i < producerCount*10; i++ {
		id := i%producerCount + 1
		producer(id, fmt.Sprintf("msg-%d", i), fmt.Sprintf("val-%d", i))
		time.Sleep(1 * time.Second)
	}
}
