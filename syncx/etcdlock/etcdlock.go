package etcdlock

import (
	"context"

	"github.com/qingwave/gocorex/syncx"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

type EtcdLockConfig struct {
	Client     *clientv3.Client
	Prefix     string
	TTLSeconds int
}

type EtcdLock struct {
	EtcdLockConfig
	session *concurrency.Session
	mutex   *concurrency.Mutex
}

func New(config EtcdLockConfig) (syncx.Locker, error) {
	session, err := concurrency.NewSession(config.Client, concurrency.WithTTL(config.TTLSeconds))
	if err != nil {
		return nil, err
	}

	mutex := concurrency.NewMutex(session, config.Prefix)

	return &EtcdLock{
		EtcdLockConfig: config,
		session:        session,
		mutex:          mutex,
	}, nil
}

func (l *EtcdLock) TryLock(ctx context.Context) (bool, error) {
	err := l.mutex.TryLock(ctx)
	if err == nil {
		return true, nil
	}
	if err == concurrency.ErrLocked {
		return false, nil
	}
	return false, err
}

func (l *EtcdLock) Lock(ctx context.Context) error {
	return l.mutex.Lock(ctx)
}

func (l *EtcdLock) UnLock(ctx context.Context) error {
	return l.mutex.Unlock(ctx)
}

func (l *EtcdLock) Close() error {
	return l.session.Close()
}
