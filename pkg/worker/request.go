package worker

import (
	"sync"

	util_csv "github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/csv"
)

// a queue is a FIFO list of requests that a worker will read from it to do some work.
type Queue chan *Request

func NewRequest(measurement *util_csv.Measurement, resultCh chan *Result, wg *sync.WaitGroup,
	processer Processer) *Request {
	return &Request{
		measurement: measurement,
		resultCh:    resultCh,
		wg:          wg,
		processer:   processer,
	}
}

// Request contains what each worker should process and a Processer to process the request.
type Request struct {
	// what to process
	measurement *util_csv.Measurement

	// where to put the result after the processing
	resultCh chan *Result

	// to indicate this request has finished processing
	wg *sync.WaitGroup

	// how the request should be processed
	processer Processer
}

func (request *Request) Process() (*Result, error) {
	return request.processer.Process()
}
