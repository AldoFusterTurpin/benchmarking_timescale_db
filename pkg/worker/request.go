package worker

import (
	"sync"

	util_csv "github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/csv"
)

type ChanelOfRequests chan *Request
type ChanelOfChanels chan ChanelOfRequests

// request contains what each worker should process.
type Request struct {
	// what to process
	arg *util_csv.Measurement

	// where to put the result after the processing
	resultCh chan *Result

	wg *sync.WaitGroup
}
