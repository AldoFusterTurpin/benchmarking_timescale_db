package worker

import (
	"fmt"
	"log"
	"time"
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
