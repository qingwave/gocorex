package pubsub

import (
	"context"
	"strings"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const DefaultTTLSeconds = 7 * 24 * 3600

func New(config EtcdPubSubConfig) (*EtcdPubSub, error) {
	if config.TTLSeconds == 0 {
		config.TTLSeconds = DefaultTTLSeconds
	}
	config.Prefix = strings.TrimSuffix(config.Prefix, "/") + "/"
	return &EtcdPubSub{config}, nil
}

type EtcdPubSubConfig struct {
	Client     *clientv3.Client
	Prefix     string
	TTLSeconds int
}

type Msg struct {
	Name string
	Val  string
}

type EtcdPubSub struct {
	EtcdPubSubConfig
}

func (ps *EtcdPubSub) SubscribeFromRev(ctx context.Context, topic string, rev int64) (<-chan Msg, error) {
	wch := ps.Client.Watch(ctx, ps.Prefix+topic, clientv3.WithPrefix(), clientv3.WithFilterDelete(), clientv3.WithRev(rev))

	msg := make(chan Msg)
	go func() {
		defer close(msg)

		for {
			wc, ok := <-wch
			if !ok {
				return
			}

			for _, ev := range wc.Events {
				if ev.Type != mvccpb.PUT {
					break
				}
				name := strings.TrimPrefix(string(ev.Kv.Key), ps.Prefix+topic+"/")
				msg <- Msg{Name: name, Val: string(ev.Kv.Value)}
			}
		}
	}()

	return msg, nil
}

// Subscribe a topic from start
func (ps *EtcdPubSub) SubscribeFromStart(ctx context.Context, topic string) (<-chan Msg, error) {
	return ps.SubscribeFromRev(ctx, topic, 1)
}

// Subscribe a topic from now
func (ps *EtcdPubSub) Subscribe(ctx context.Context, topic string) (<-chan Msg, error) {
	return ps.SubscribeFromRev(ctx, topic, 0)
}

func (ps *EtcdPubSub) Publish(ctx context.Context, topic string, msg Msg) error {
	le, err := ps.Client.Lease.Grant(ctx, int64(ps.TTLSeconds))
	if err != nil {
		return err
	}

	_, err = ps.Client.Put(ctx, ps.Prefix+topic+"/"+msg.Name, msg.Val, clientv3.WithLease(le.ID))
	return err
}

func (ps *EtcdPubSub) Reset(ctx context.Context, topic string) error {
	_, err := ps.Client.Delete(ctx, ps.Prefix+topic+"/", clientv3.WithPrefix())
	return err
}
