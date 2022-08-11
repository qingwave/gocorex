package cron

import (
	"container/list"
	"fmt"
	"time"

	"github.com/qingwave/gocorex/containerx"

	"github.com/go-logr/logr"
)

type TimeWheel struct {
	interval    time.Duration
	slots       int
	currentSlot int
	tasks       []*list.List
	set         containerx.Set[string]

	tricker *time.Ticker

	logger logr.Logger
}

type TimeWheelTask struct {
	Task

	initialized bool
	slot        int
	circle      int
}

func NewTimeWheel(interval time.Duration, slots int, logger logr.Logger) Interface {
	return &TimeWheel{
		interval: interval,
		slots:    slots,
		tasks:    make([]*list.List, slots),
		set:      containerx.NewSet[string](),
		logger:   logger,
	}
}

func (tw *TimeWheel) Run() error {
	tw.tricker = time.NewTicker(tw.interval)

	for {
		now, ok := <-tw.tricker.C
		if !ok {
			break
		}
		tw.RunTask(now, tw.currentSlot)
		tw.currentSlot = (tw.currentSlot + 1) % tw.slots
	}

	return nil
}

func (tw *TimeWheel) RunTask(now time.Time, slot int) {
	taskList := tw.tasks[slot]
	if taskList == nil {
		return
	}

	for item := taskList.Front(); item != nil; {
		task, ok := item.Value.(*TimeWheelTask)
		if !ok || task == nil {
			item = item.Next()
			continue
		}

		if task.circle > 0 {
			task.circle--
			item = item.Next()
			continue
		}

		// run task
		go func() {
			start := time.Now()
			if err := task.Exec(); err != nil {
				tw.logger.Info(fmt.Sprintf("Run job [%s] failed: %v", task.Name(), err))
				return
			}
			tw.logger.Info(fmt.Sprintf("Run job [%s] successfully, duration %v", task.Name(), time.Since(start)))
		}()

		// delete or update task
		next := item.Next()
		taskList.Remove(item)
		item = next

		task.next = task.Next(now)
		if !task.next.IsZero() {
			tw.add(now, task)
		} else {
			tw.Remove(task.Name())
		}
	}
}

func (tw *TimeWheel) Stop() {
	tw.tricker.Stop()
}

func (tw *TimeWheel) Len() int {
	return tw.set.Len()
}

func (tw *TimeWheel) Add(task *Task) {
	tw.add(time.Now(), &TimeWheelTask{
		Task: *task,
	})
}

func (tw *TimeWheel) add(now time.Time, task *TimeWheelTask) {
	if !task.initialized {
		task.next = task.Next(now)
		task.initialized = true
	}

	duration := task.next.Sub(now)
	if duration <= 0 {
		task.slot = tw.currentSlot + 1
		task.circle = 0
	} else {
		mult := int(duration / tw.interval)
		task.slot = (tw.currentSlot + mult) % tw.slots
		task.circle = mult / tw.slots
	}

	if tw.tasks[task.slot] == nil {
		tw.tasks[task.slot] = list.New()
	}

	tw.tasks[task.slot].PushBack(task)
	tw.set.Insert(task.Name())
}

func (tw *TimeWheel) Remove(name string) {
	tw.set.Delete(name)
}
