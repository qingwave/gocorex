package syncx

import "context"

type Locker interface {
	Lock(ctx context.Context) error
	UnLock(ctx context.Context) error
	TryLock(ctx context.Context) (bool, error)
	Close() error
}
