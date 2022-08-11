package workqueue

import "sync/atomic"

func NewChannelQueue(ch chan any) WorkQueue {
	if ch == nil {
		ch = make(chan any)
	}

	return &channelQueue{ch: ch, stop: new(atomic.Bool)}
}

type channelQueue struct {
	ch   chan any
	stop *atomic.Bool
}

func (q *channelQueue) Read() (any, bool) {
	item, ok := <-q.ch

	return item, ok
}

func (q *channelQueue) Write(item any) {
	if q.stop.Load() {
		return
	}

	q.ch <- item
}

func (q *channelQueue) Stop() {
	if q.stop.Load() {
		return
	}

	q.stop.Store(true)
	close(q.ch)
}

func (q *channelQueue) Len() int {
	return len(q.ch)
}
