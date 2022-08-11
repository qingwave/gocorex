package main

import (
	"fmt"
	"log"
	"time"

	"github.com/qingwave/gocorex/syncx/mapreduce"
)

func main() {
	num := 1000000

	// demo 1, chain create MapReduce
	res, err := mapreduce.New(mapreduce.WithWorkers(16)).
		From(func(r mapreduce.Writer) error {
			for i := 1; i < num; i++ {
				r.Write(i)
			}
			return nil
		}).
		Map(func(item any) (any, error) {
			v, ok := item.(int)
			if !ok {
				return nil, fmt.Errorf("invaild type")
			}

			time.Sleep(200 * time.Nanosecond)
			resp := v * v

			return resp, nil
		}).
		Reduce(func(r mapreduce.Reader) (any, error) {
			sum := 0
			for {
				item, ok := r.Read()
				if !ok {
					break
				}

				v, ok := item.(int)
				if !ok {
					return nil, fmt.Errorf("invaild type")
				}

				sum += v
			}
			return sum, nil
		}).
		Do()

	log.Println(res, err)

	// create from MapReducer interface
	res, err = mapreduce.NewFromMapReducer(&mr{num}).Do()
	log.Println(res, err)
}

type mr struct {
	num int
}

func (m *mr) From(w mapreduce.Writer) error {
	for i := 1; i < m.num; i++ {
		w.Write(i)
	}
	return nil
}

func (m *mr) Map(item any) (any, error) {
	v, ok := item.(int)
	if !ok {
		return nil, fmt.Errorf("invaild type")
	}

	time.Sleep(200 * time.Nanosecond)
	resp := v * v

	return resp, nil
}

func (m *mr) Reduce(r mapreduce.Reader) (any, error) {
	sum := 0
	for {
		item, ok := r.Read()
		if !ok {
			break
		}

		v, ok := item.(int)
		if !ok {
			return nil, fmt.Errorf("invaild type")
		}

		sum += v
	}
	return sum, nil
}
