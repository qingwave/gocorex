package mapreduce

import "context"

type Option func(*MapReduceOption)

type MapReduceOption struct {
	ctx     context.Context
	workers int
}

const (
	defaultWorkers = 8
	miniWorkers    = 1
)

func WithContext(ctx context.Context) Option {
	return func(opt *MapReduceOption) {
		opt.ctx = ctx
	}
}

func WithWorkers(workers int) Option {
	return func(opts *MapReduceOption) {
		if workers < miniWorkers {
			opts.workers = miniWorkers
		} else {
			opts.workers = workers
		}
	}
}

func newOptions() *MapReduceOption {
	return &MapReduceOption{
		ctx:     context.Background(),
		workers: defaultWorkers,
	}
}
