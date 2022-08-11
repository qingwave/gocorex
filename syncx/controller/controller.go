package controller

import (
	"fmt"

	"github.com/qingwave/gocorex/syncx/group"
)

func New[T any](opts ...Option) *Controller[T] {
	options := newOptions()
	for _, opt := range opts {
		opt(options)
	}

	return &Controller[T]{
		ControllerOption: options,
	}
}

type Controller[T any] struct {
	*ControllerOption

	err error

	source  <-chan T
	handler func(item T)
}

func (c *Controller[T]) Error() error {
	return c.err
}

func (c *Controller[T]) From(source <-chan T) *Controller[T] {
	c.source = source
	return c
}

func (c *Controller[T]) Handle(handler func(T)) *Controller[T] {
	if c.source == nil {
		c.err = fmt.Errorf("controller source is nil")
		return c
	}

	c.handler = handler

	return c
}

func (c *Controller[T]) Run() error {
	if c.err != nil {
		return c.err
	}

	if c.handler == nil {
		c.err = fmt.Errorf("controller handler is nil")
		return c.err
	}

	defer c.queue.Stop()

	go func() {
		defer c.queue.Stop()

		for v := range c.source {
			c.queue.Write(v)
		}
	}()

	g := group.NewGroup()
	for i := 0; i < c.workers; i++ {
		g.Go(c.runWorker)
	}

	g.Wait()

	return nil
}

func (c *Controller[T]) runWorker() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		val, ok := c.queue.Read()
		if !ok && val == nil {
			return
		}

		item, ok := val.(T)
		if !ok {
			continue
		}

		c.handler(item)
	}
}
