package worker

import (
	"context"
	"math/rand"
	"time"

	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/domain"
)

// Processer is a contract to process a request, to decouple the data of the request
// with the actual processing of it.
type Processer interface {
	Process(ctx context.Context, m *domain.Measurement) (*domain.QueryResultWithTime, error)
}

func NewRandomProcesser() *RandomProcesser {
	return &RandomProcesser{}
}

// RandomProcesser used for playing without touching the DB
type RandomProcesser struct {
}

func (randomProcesser *RandomProcesser) Process(ctx context.Context, measurement *domain.Measurement) (*domain.QueryResultWithTime, error) {
	randomDuration := time.Duration(rand.Intn(100)) * time.Millisecond
	time.Sleep(randomDuration)

	queryResults := []*domain.QueryResult{
		{
			Timestamp:   time.Now(),
			MaxCpuUsage: 3000,
			MinCpuUsage: 1000,
		},
	}

	return &domain.QueryResultWithTime{
		Queryresults:       queryResults,
		QueryExecutionTime: randomDuration,
	}, nil
}
