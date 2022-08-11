package group

import "sync"

func NewCtrlGroup(number int) *CtrlGroup {
	return &CtrlGroup{
		ch: make(chan struct{}, number),
	}
}

type CtrlGroup struct {
	ch chan struct{}
	wg sync.WaitGroup
}

func (g *CtrlGroup) Enter() {
	g.ch <- struct{}{}
}

func (g *CtrlGroup) Leave() {
	<-g.ch
}

func (g *CtrlGroup) Go(f func()) {
	g.Enter()
	g.wg.Add(1)

	go func() {
		defer g.Leave()
		defer g.wg.Done()
		f()
	}()
}

func (g *CtrlGroup) Wait() {
	g.wg.Wait()
}
