package worker

import (
	"math/rand"
	"sync"
	"time"

	util_csv "github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/csv"
)

// a queue is a FIFO list of requests that a worker will read from it to do some work.
type Queue chan *Request

// Request contains what each worker should process.
type Request struct {
	// what to process
	measurement *util_csv.Measurement

	// where to put the result after the processing
	resultCh chan *Result

	// to indicate this request has finished processing
	wg *sync.WaitGroup
}

// fake processing to explore
func (request *Request) Process() (*Result, error) {
	now := time.Now()

	d := time.Duration(rand.Intn(5))
	time.Sleep(d * time.Second)

	result := &Result{
		timeSpend: time.Since(now),
	}
	return result, nil
}
