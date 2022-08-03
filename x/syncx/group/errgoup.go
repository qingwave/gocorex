package group

import (
	"context"
	"sync"
)

func NewErrGroup(ctx context.Context) *ErrGroup {
	ctx, cancel := context.WithCancel(ctx)
	return &ErrGroup{
		ctx:    ctx,
		cancel: cancel,
	}
}

type ErrGroup struct {
	ctx    context.Context
	cancel func()

	wg sync.WaitGroup

	errOnce sync.Once
	err     error
}

func (g *ErrGroup) Wait() error {
	g.wg.Wait()

	if g.cancel != nil {
		g.cancel()
	}

	return g.err
}

func (g *ErrGroup) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()
		if err := f(); err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}
