package workqueue

import (
	"sync"
)

type WorkQueue interface {
	Write(item any)
	Read() (any, bool)
	Stop()
	Len() int
}

func New() *Type {
	t := &Type{
		cond: sync.NewCond(&sync.Mutex{}),
	}

	return t
}

type Type struct {
	queue []any

	cond *sync.Cond

	shuttingDown bool
}

func (q *Type) Write(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.shuttingDown {
		return
	}

	q.queue = append(q.queue, item)
	q.cond.Signal()
}

func (q *Type) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return len(q.queue)
}

func (q *Type) Read() (item interface{}, ok bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	for len(q.queue) == 0 && !q.shuttingDown {
		q.cond.Wait()
	}
	if len(q.queue) == 0 {
		// We must be shutting down.
		return nil, false
	}

	item = q.queue[0]
	// The underlying array still exists and reference this object, so the object will not be garbage collected.
	q.queue[0] = nil
	q.queue = q.queue[1:]

	return item, true
}

func (q *Type) Stop() {
	q.shutdown()
}

func (q *Type) shutdown() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.shuttingDown = true
	q.cond.Broadcast()
}

func (q *Type) Stopped() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	return q.shuttingDown
}
