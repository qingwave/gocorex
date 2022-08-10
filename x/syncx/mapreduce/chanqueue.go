package mapreduce

import "go.uber.org/atomic"

type Queue interface {
	Reader
	Writer
	Stop()
}

func newQueue(ch chan any) Queue {
	if ch == nil {
		ch = make(chan any)
	}

	return &chanQueue{ch: ch, stop: atomic.NewBool(false)}
}

type chanQueue struct {
	ch   chan any
	stop *atomic.Bool
}

func (q *chanQueue) Read() (any, bool) {
	item, ok := <-q.ch

	return item, ok
}

func (q *chanQueue) Write(item any) {
	if q.stop.Load() {
		return
	}

	q.ch <- item
}

func (q *chanQueue) Stop() {
	if q.stop.Load() {
		return
	}

	q.stop.Store(true)
	close(q.ch)
}
