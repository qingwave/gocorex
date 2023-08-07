package group

import (
	"context"
	"sync"
)

func NewGroup() *Group {
	return &Group{}
}

type Group struct {
	wg sync.WaitGroup
}

func (g *Group) Wait() {
	g.wg.Wait()
}

func (g *Group) Go(f func()) {
	g.wg.Add(1)
	go func() {
		defer g.wg.Done()
		f()
	}()
}

func (g *Group) GoN(num int, f func()) {
	for i := 0; i < num; i++ {
		g.Go(f)
	}
}

func (g *Group) StartWithChannel(stopCh <-chan struct{}, f func(stopCh <-chan struct{})) {
	g.Go(func() {
		f(stopCh)
	})
}

func (g *Group) StartWithContext(ctx context.Context, f func(context.Context)) {
	g.Go(func() {
		f(ctx)
	})
}