package domain

import (
	"fmt"
	"log"
	"time"
)

const (
	timeFormat = time.DateTime
)

type Measurement struct {
	Hostname  string
	StartTime time.Time
	EndTime   time.Time
}

func (measurement *Measurement) Print() {
	log.Println(measurement.String())
}

func (measurement *Measurement) String() string {
	return fmt.Sprintf("CSVRow -> hostname: %v, start_time: %v, end_time: %v\n", measurement.Hostname,
		measurement.StartTime.Format(timeFormat),
		measurement.EndTime.Format(timeFormat))
}

type QueryResult struct {
	Timestamp   time.Time
	MaxCPUUsage float64
	MinCPUUsage float64
}

type QueryResultWithTime struct {
	Queryresults       []*QueryResult
	QueryExecutionTime time.Duration
}
