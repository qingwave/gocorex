package controller

import (
	"context"

	"github.com/qingwave/gocorex/x/syncx/workqueue"
)

type Option func(*ControllerOption)

type ControllerOption struct {
	ctx     context.Context
	workers int
	queue   workqueue.WorkQueue
}

const (
	defaultWorkers = 8
	miniWorkers    = 1
)

func WithContext(ctx context.Context) Option {
	return func(opt *ControllerOption) {
		opt.ctx = ctx
	}
}

func WithWorkers(workers int) Option {
	return func(opts *ControllerOption) {
		if workers < miniWorkers {
			opts.workers = miniWorkers
		} else {
			opts.workers = workers
		}
	}
}

func WithQueue(queue workqueue.WorkQueue) Option {
	return func(opts *ControllerOption) {
		if queue != nil {
			opts.queue = queue
		}
	}
}

func newOptions() *ControllerOption {
	return &ControllerOption{
		ctx:     context.Background(),
		workers: defaultWorkers,
		queue:   workqueue.NewChannelQueue(make(chan any)),
	}
}
