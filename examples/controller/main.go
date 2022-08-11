package main

import (
	"log"
	"time"

	"github.com/qingwave/gocorex/x/syncx/controller"
)

func main() {
	ch := make(chan int)

	go func() {
		defer close(ch)
		for i := 1; i < 10000; i++ {
			ch <- i
		}
	}()

	start := time.Now()
	err := controller.New[int](controller.WithWorkers(8)).
		From(ch).
		Handle(func(i int) {
			time.Sleep(time.Millisecond)
		}).
		Run()

	log.Println(time.Since(start), err)
}
