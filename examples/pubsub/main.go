package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/qingwave/gocorex/pubsub"
	"github.com/qingwave/gocorex/pubsub/etcdpubsub"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func etcdPubSubDemo() {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		log.Fatalf("failed to create etcd lock: %v", err)
	}
	defer client.Close()

	ps, _ := etcdpubsub.New(etcdpubsub.Config{
		Client: client,
		Prefix: "/pubsub",
	})

	topic := "demo-topic"

	ps.Reset(context.Background(), topic)

	producer := func(id int, name, msg string) {
		err := ps.Publish(context.Background(), topic, etcdpubsub.Msg{
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

func pubsubDemo() {
	b := pubsub.NewBroker(8)
	go b.Run()

	consumer := func(name string) pubsub.ConsumeFunc {
		return func(topic pubsub.Topic) {
			log.Printf("[%v] receive topic: %+v", name, topic)
		}
	}

	s1 := pubsub.NewSubscriber("s1", b)
	s1.Subscribe("t1", consumer("s1"))
	s1.Subscribe("t2", consumer("s1"))

	s2 := pubsub.NewSubscriber("s2", b)
	s2.Subscribe("t2", consumer("s2"))

	s3 := pubsub.NewSubscriber("s3", b)
	s3.Subscribe("t3", consumer("s3"))

	p1 := pubsub.NewPublisher("p1", b)
	p2 := pubsub.NewPublisher("p2", b)

	p1.Publish(pubsub.Topic{ID: "t1", Msg: "hello xxx"})
	p1.Publish(pubsub.Topic{ID: "t2", Msg: "hello xxx2"})
	p1.Publish(pubsub.Topic{ID: "t3", Msg: "hello xxx3"})

	p2.Publish(pubsub.Topic{ID: "t2", Msg: "p2 hello xxx"})

	s1.Unsubscribe("t2")
	p1.Publish(pubsub.Topic{ID: "t2", Msg: "hello again xxx"})

	time.Sleep(2 * time.Second)
}

func main() {
	// run pubsub
	// pubsubDemo()

	// run etcd pubsub
	etcdPubSubDemo()
}
