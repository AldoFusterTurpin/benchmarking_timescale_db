package worker

import (
	"math/rand"
	"time"
)

// Processer is a contract to process a request, to decouple the data of the request
// with the actual processing of it.
type Processer interface {
	Process() (*Result, error)
}

func NewFakeProcesser() *FakeProcesser {
	return &FakeProcesser{}
}

// FakeProcesser used for testing purposes.
type FakeProcesser struct {
}

func (FakeProcesser *FakeProcesser) Process() (*Result, error) {
	now := time.Now()

	d := time.Duration(rand.Intn(5))
	time.Sleep(d * time.Second)

	result := &Result{
		timeSpend: time.Since(now),
	}
	return result, nil
}
