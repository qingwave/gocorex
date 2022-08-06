package zklock

import (
	"context"

	"github.com/go-zookeeper/zk"
	"github.com/qingwave/gocorex/x/syncx"
)

func New(config ZkLockConfig) (syncx.Locker, error) {
	if len(config.ACL) == 0 {
		config.ACL = zk.WorldACL(zk.PermAll)
	}
	lock := zk.NewLock(config.Conn, config.Path, config.ACL)
	return &ZkLock{
		ZkLockConfig: config,
		lock:         lock,
	}, nil
}

type ZkLockConfig struct {
	Conn *zk.Conn
	Path string
	ACL  []zk.ACL
}

type ZkLock struct {
	ZkLockConfig
	lock *zk.Lock
}

func (l *ZkLock) Lock(ctx context.Context) error {
	return l.lock.Lock()
}

func (l *ZkLock) UnLock(ctx context.Context) error {
	return l.lock.Unlock()
}

func (l *ZkLock) TryLock(ctx context.Context) (bool, error) {
	_, err := l.Conn.CreateProtectedEphemeralSequential(l.Path, []byte{}, l.ACL)
	if err != nil {
		if err == zk.ErrNodeExists {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (l *ZkLock) Close() error {
	return nil
}
