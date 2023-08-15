package etcdlock

import (
	"context"
	"errors"
	"time"

	"github.com/qingwave/gocorex/syncx"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

var (
	ErrInvaildClient = errors.New("invaild etcd client")
	ErrTimeout       = errors.New("connect to etcd timeout")
	DefaultTimeout   = 5 * time.Second
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
	if config.Client == nil {
		return nil, ErrInvaildClient
	}

	timeout := DefaultTimeout
	if config.TTLSeconds > 0 {
		timeout = time.Duration(config.TTLSeconds) * time.Second
	}

	locker := &EtcdLock{
		EtcdLockConfig: config,
	}

	errCh := make(chan error)
	defer close(errCh)

	go func() {
		var err error
		locker.session, err = concurrency.NewSession(config.Client, concurrency.WithTTL(config.TTLSeconds))
		if err != nil {
			errCh <- err
			return
		}

		locker.mutex = concurrency.NewMutex(locker.session, config.Prefix)
		errCh <- nil
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	case <-time.After(timeout):
		return nil, ErrTimeout
	}

	return locker, nil
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
	if l.session != nil {
		return l.session.Close()
	}
	return nil
}
