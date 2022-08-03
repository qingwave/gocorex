package group

import (
	"sync"
)

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
