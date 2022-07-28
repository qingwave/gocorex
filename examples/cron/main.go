package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/go-logr/stdr"
	"github.com/qingwave/gocorex/x/cron"
)

func TestCron(c cron.Interface) {
	go c.Run()
	defer c.Stop()

	addTask := func(id int, schedule cron.Schedule) {
		name := fmt.Sprintf("job%d", id)
		c.Add(&cron.Task{
			Job: cron.SimpleJob(name, func() error {
				log.Printf("job: [%s] ==> running", name)
				return nil
			}),
			Schedule: schedule,
		})
	}

	now := time.Now()
	addTask(1, cron.Every(5*time.Second))
	addTask(2, cron.Once(now.Add(8*time.Second)))
	addTask(3, cron.At(now.Add(2*time.Second), 3*time.Second, 3))
	addTask(4, cron.EveryAt(now.Add(10*time.Second), 2*time.Second))
	addTask(5, cron.Once(now.Add(time.Second)))
	addTask(6, cron.Once(now.Add(-time.Second)))

	log.Printf("start at %v", time.Now())
	for i := 0; i <= 60; i++ {
		log.Printf("current tasks %d", c.Len())
		time.Sleep(1 * time.Second)
	}
}

var runTimeWheel = flag.Bool("timewheel", false, "run timingwheel cron, default run mini heap cron")

func main() {
	var c cron.Interface
	logger := stdr.New(log.Default())
	if *runTimeWheel {
		c = cron.NewTimeWheel(1*time.Second, 60, logger)
	} else {
		c = cron.NewCron(logger)
	}

	TestCron(c)
}
