package worker

import (
	"fmt"
	"log"
	"time"
)

// result contains the result of each worker process.
type Result struct {
	timeSpend time.Duration
}

func (result *Result) Print() {
	log.Println(result.String())
}

func (result *Result) String() string {
	return fmt.Sprintf("result -> time spent is: %v\n", result.timeSpend)
}

// TODO: create a unit test for this to test business logic
// independently from the worker pool.
func consumAllResults(resultCh chan *Result) {
	nQueriesProcessed := 0
	var totalProcessingTime time.Duration
	var minQueryTime time.Duration
	var maxQueryTime time.Duration
	var medianQueryTime time.Duration
	var avarageQueryTime time.Duration
	i := 0

	for result := range resultCh {
		first := i == 0
		if first {
			minQueryTime = result.timeSpend
			maxQueryTime = result.timeSpend
		}

		if result.timeSpend > maxQueryTime {
			maxQueryTime = result.timeSpend
		}
		if result.timeSpend < minQueryTime {
			minQueryTime = result.timeSpend
		}

		nQueriesProcessed++
		totalProcessingTime += result.timeSpend

		result.Print()

		i++
	}

	avarageQueryTime = totalProcessingTime / time.Duration(nQueriesProcessed)

	format := "nQueriesProcessed: %v\ntotalProcessingTime: %v\nminQueryTime: %v\nmaxQueryTime: %v\nmedianQueryTime: %v\navarageQueryTime: %v"
	log.Printf(format,
		nQueriesProcessed, totalProcessingTime, minQueryTime, maxQueryTime, medianQueryTime, avarageQueryTime)
}
