package pubsub

import (
	"sync"
)

type Topic struct {
	ID   string
	From string
	Msg  string
}

type Broker struct {
	mu sync.RWMutex

	topicBuffer chan Topic
	subscribers map[string][]*Subscriber
}

func NewBroker(bufSize int) *Broker {
	return &Broker{
		topicBuffer: make(chan Topic, bufSize),
		subscribers: make(map[string][]*Subscriber),
	}
}

func (b *Broker) Subscribe(topicID string, subscriber *Subscriber) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	b.subscribers[topicID] = append(b.subscribers[topicID], subscriber)
}

func (b *Broker) Unsubscribe(topicID string, subscriber *Subscriber) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for i, sub := range b.subscribers[topicID] {
		if sub.name == subscriber.name {
			b.subscribers[topicID] = append(b.subscribers[topicID][:i], b.subscribers[topicID][i+1:]...)
			return
		}
	}
}

func (b *Broker) Publish(topic Topic) {
	b.topicBuffer <- topic
}

func (b *Broker) Stop() {
	for _, subs := range b.subscribers {
		for _, sub := range subs {
			sub.Stop()
		}
	}
	close(b.topicBuffer)
}

func (b *Broker) Run() {
	for {
		topic, ok := <-b.topicBuffer
		if !ok {
			return
		}
		for _, sub := range b.subscribers[topic.ID] {
			sub.Receive(topic)
		}
	}
}

type ConsumeFunc func(topic Topic)

type Subscriber struct {
	name   string
	buf    chan Topic
	topics map[string]ConsumeFunc

	broker *Broker
}

func NewSubscriber(name string, broker *Broker) *Subscriber {
	sub := &Subscriber{
		name:   name,
		buf:    make(chan Topic, 1),
		topics: make(map[string]ConsumeFunc),
		broker: broker,
	}
	go sub.Consume()
	return sub
}

func (s *Subscriber) Subscribe(topicID string, consume ConsumeFunc) {
	s.topics[topicID] = consume
	s.broker.Subscribe(topicID, s)
}

func (s *Subscriber) Unsubscribe(topicID string) {
	delete(s.topics, topicID)
	s.broker.Unsubscribe(topicID, s)
}

func (s *Subscriber) Receive(topic Topic) {
	s.buf <- topic
}

func (s *Subscriber) Consume() {
	for topic := range s.buf {
		consume, ok := s.topics[topic.ID]
		if !ok {
			continue
		}
		consume(topic)
	}
}

func (s *Subscriber) Stop() {
	for topic := range s.topics {
		s.Unsubscribe(topic)
	}
	close(s.buf)
}

type Publisher struct {
	name   string
	broker *Broker
}

func NewPublisher(name string, broker *Broker) *Publisher {
	return &Publisher{
		name:   name,
		broker: broker,
	}
}

func (p *Publisher) Publish(topic Topic) {
	topic.From = p.name
	p.broker.Publish(topic)
}
