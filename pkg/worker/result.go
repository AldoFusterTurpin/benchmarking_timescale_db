package worker

import (
	"fmt"
	"log"
	"time"

	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/domain"
)

// result contains the result of each worker process.
type RequestResult struct {
	timeSpend time.Duration
}

func (result *RequestResult) Print() {
	log.Println(result.String())
}

func (result *RequestResult) String() string {
	return fmt.Sprintf("result -> time spent is: %v\n", result.timeSpend)
}

type StatsResult struct {
	NQueriesProcessed   int
	TotalProcessingTime time.Duration
	MinQueryTime        time.Duration
	MaxQueryTime        time.Duration
	MedianQueryTime     time.Duration
	AvarageQueryTime    time.Duration
}

func (s *StatsResult) String() string {
	format := "nQueriesProcessed: %v\ntotalProcessingTime: %v\nminQueryTime: %v\nmaxQueryTime: %v\nmedianQueryTime: %v\navarageQueryTime: %v\n"
	return fmt.Sprintf(format, s.NQueriesProcessed, s.TotalProcessingTime, s.MinQueryTime, s.MaxQueryTime, s.MedianQueryTime, s.AvarageQueryTime)
}

func consumAllResults(resultCh chan *domain.QueryResultWithTime) *StatsResult {
	nQueriesProcessed := 0
	var totalProcessingTime time.Duration
	var minQueryTime time.Duration
	var maxQueryTime time.Duration
	var medianQueryTime time.Duration // TODO
	var avarageQueryTime time.Duration
	i := 0

	for queryResultWithTime := range resultCh {
		d := queryResultWithTime.QueryExecutionTime

		first := i == 0
		if first {
			minQueryTime = d
			maxQueryTime = d
		}

		if d > maxQueryTime {
			maxQueryTime = d
		}
		if d < minQueryTime {
			minQueryTime = d
		}

		nQueriesProcessed++
		totalProcessingTime += d

		i++
	}

	avarageQueryTime = totalProcessingTime / time.Duration(nQueriesProcessed)
	log.Println("avarageQueryTime is: ", avarageQueryTime)

	statsResult := &StatsResult{
		NQueriesProcessed:   nQueriesProcessed,
		TotalProcessingTime: totalProcessingTime,
		MinQueryTime:        minQueryTime,
		MaxQueryTime:        maxQueryTime,
		MedianQueryTime:     medianQueryTime,
		AvarageQueryTime:    avarageQueryTime,
	}

	return statsResult
}
