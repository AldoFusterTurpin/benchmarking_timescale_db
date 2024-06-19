package worker

import (
	"sync"

	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/domain"
)

func NewRequest(measurement *domain.Measurement, resultCh chan *domain.QueryResultWithTime, wg *sync.WaitGroup) *Request {
	return &Request{
		measurement: measurement,
		resultCh:    resultCh,
		wg:          wg,
	}
}

// Request contains what each worker should process and a Processer to process the request.
type Request struct {
	// what to process
	measurement *domain.Measurement

	// where to put the result after the processing
	resultCh chan *domain.QueryResultWithTime

	// to indicate this request has finished processing
	wg *sync.WaitGroup
}
