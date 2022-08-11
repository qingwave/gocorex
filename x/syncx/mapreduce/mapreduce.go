package mapreduce

import (
	"context"
	"fmt"

	"github.com/qingwave/gocorex/x/syncx/group"
	"github.com/qingwave/gocorex/x/syncx/workqueue"
)

type (
	Writer interface {
		Write(any)
	}

	Reader interface {
		Read() (any, bool)
	}

	Producer func(w Writer) error

	Mapper func(item any) (any, error)

	Reducer func(r Reader) (any, error)

	Filter func(any) (bool, error)

	MapReducer interface {
		From(w Writer) error
		Map(item any) (any, error)
		Reduce(r Reader) (any, error)
	}

	Interface interface {
		From(producer Producer) Interface
		Map(mapper Mapper) Interface
		Filter(filters ...Filter) Interface
		Reduce(reducer Reducer) Interface
		Do() (any, error)
		Error() error
	}
)

func New(opts ...Option) Interface {
	options := newOptions()
	for _, opt := range opts {
		opt(options)
	}

	_, cancel := context.WithCancel(options.ctx)

	return &mapreduce{
		options: options,
		cancel:  cancel,
	}
}

func NewFromMapReducer(mr MapReducer, opts ...Option) Interface {
	return New(opts...).
		From(mr.From).
		Map(mr.Map).
		Reduce(mr.Reduce)
}

type mapreduce struct {
	options *MapReduceOption
	cancel  context.CancelFunc

	producer Producer
	mapper   Mapper
	filters  []Filter
	reducer  Reducer

	err error
}

func (mr *mapreduce) Error() error {
	return mr.err
}

func (mr *mapreduce) From(producer Producer) Interface {
	if mr.err != nil {
		return mr
	}

	if mr.producer != nil {
		mr.err = fmt.Errorf("cannot call From twice")
		return mr
	}

	mr.producer = producer
	return mr
}

func (mr *mapreduce) Map(mapper Mapper) Interface {
	if mr.err != nil {
		return mr
	}

	if mr.producer == nil {
		mr.err = fmt.Errorf("cannot call Map before From")
		return mr
	}

	if mr.mapper != nil {
		mr.err = fmt.Errorf("cannot call Map twice")
		return mr
	}

	mr.mapper = mapper
	return mr
}

func (mr *mapreduce) Filter(filters ...Filter) Interface {
	if mr.err != nil {
		return mr
	}

	if mr.filters != nil {
		mr.err = fmt.Errorf("cannot call Filters twice")
		return mr
	}

	mr.filters = filters

	return mr
}

func (mr *mapreduce) Reduce(reducer Reducer) Interface {
	if mr.err != nil {
		return mr
	}

	if mr.mapper == nil {
		mr.err = fmt.Errorf("cannot call Reduce before Map")
		return mr
	}

	if mr.reducer != nil {
		mr.err = fmt.Errorf("cannot call Reduce twice")
		return mr
	}
	mr.reducer = reducer

	return mr
}

func (mr *mapreduce) Do() (any, error) {
	if mr.reducer == nil {
		mr.err = fmt.Errorf("cannot call Do before Reduce")
		return mr.err, nil
	}

	source := workqueue.NewChannelQueue(make(chan any))
	errCh := make(chan error)
	defer close(errCh)

	// produce data
	go func() {
		defer func() {
			source.Stop()
		}()

		errCh <- mr.producer(source)
	}()

	// run worker
	output := workqueue.NewChannelQueue(make(chan any, mr.options.workers))

	go func() {
		defer output.Stop()

		g := group.NewErrGroup(mr.options.ctx)
		for i := 0; i < mr.options.workers; i++ {
			g.Go(func(ctx context.Context) error {
				for {
					select {
					case <-ctx.Done():
						return nil
					default:
					}

					item, ok := source.Read()
					if item == nil && !ok {
						return nil
					}

					if item == nil && ok {
						continue
					}

					var skip bool
					for _, filter := range mr.filters {
						ok, err := filter(item)
						if err != nil {
							return err
						}
						if !ok {
							skip = true
							break
						}
					}

					if skip {
						continue
					}

					val, err := mr.mapper(item)
					if err != nil {
						return err
					}

					output.Write(val)
				}
			})
		}

		errCh <- g.Wait()
	}()

	resp := make(chan any)
	defer close(resp)

	go func() {
		var out any
		var err error

		defer func() {
			errCh <- err
			resp <- out
		}()

		out, err = mr.reducer(output)
	}()

	for {
		select {
		case <-mr.options.ctx.Done():
			return nil, context.DeadlineExceeded
		case v, ok := <-resp:
			if !ok {
				return nil, fmt.Errorf("resp channel closed")
			}
			return v, mr.err
		case err, ok := <-errCh:
			if !ok {
				return nil, fmt.Errorf("error channel closed")
			}
			if err != nil {
				mr.err = err
				mr.cancel()
				return nil, err
			}
		}
	}
}
