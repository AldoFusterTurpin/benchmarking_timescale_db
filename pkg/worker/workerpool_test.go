package worker

import (
	"context"
	_ "embed"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	util_csv "github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/csv"
	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/domain"
	"github.com/google/go-cmp/cmp"
)

//go:embed query_params_test.csv
var testCsv_3_rows string

//go:embed query_params_test_2.csv
var testCsv_6_rows string

func Test_getHostIdFromHostName(t *testing.T) {
	type testCase struct {
		name          string
		hostname      string
		expected      int
		expectedError error
	}
	tests := []testCase{
		{
			name:     "host_000007",
			hostname: "host_000007",
			expected: 7,
		},
		{
			name:     "host_000017",
			hostname: "host_000017",
			expected: 17,
		},
		{
			name:     "host_020919",
			hostname: "host_020919",
			expected: 20919,
		},
		{
			name:          "should return appropriate error",
			hostname:      "020919",
			expected:      0,
			expectedError: errors.New("prefix host_ not found in hostname: 020919"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := getHostIdFromHostName(tt.hostname)

			if tt.expectedError != nil && err == nil {
				t.Fatalf("expected error\n%v\n, but got nil error", tt.expectedError)
			}

			if tt.expectedError == nil && err != nil {
				t.Fatalf("expected nil error, but got error\n%v", err)
			}

			if err != nil && tt.expectedError != nil &&
				err.Error() != tt.expectedError.Error() {
				t.Fatalf("expected error\n%v,\nbut got\n%v", tt.expectedError, err)
			}

			if res != tt.expected {
				t.Fatalf("expected %v, but got %v", tt.expected, res)
			}
		})
	}
}

func TestWorkerPool_ProcessMeasurements_HandlesBusinessLogic(t *testing.T) {
	type testCase struct {
		name        string
		nWorkers    int
		inputChanel <-chan *domain.Measurement
		processer   Processer
		expected    *StatsResult
	}

	s1 := (0 + 1000 + 2000)
	f1 := float64(0+1000+2000) / 3

	s2 := (0 + 1000 + 2000 + 3000 + 4000 + 5000)
	f2 := float64(0+1000+2000+3000+4000+5000) / 6

	tests := []testCase{
		// NewIncrementalProcesser() will return a processer that will block when used concurrently,
		// by the worker pool, but we are using it here to control the values returned by the Process() method
		// to test the business logic, NOT to control the concurrent nature of the worker pool.
		{
			name:        "csv with three rows",
			nWorkers:    3,
			inputChanel: util_csv.ReadMeasurements(strings.NewReader(testCsv_3_rows)),
			processer:   NewIncrementalProcesser(),
			expected: &StatsResult{
				NQueriesProcessed:   3,
				TotalProcessingTime: time.Duration(s1) * time.Millisecond,
				MinQueryTime:        0 * time.Millisecond,
				MaxQueryTime:        2000 * time.Millisecond,
				AvarageQueryTime:    time.Duration(f1) * time.Millisecond,
				MedianQueryTime:     0 * time.Millisecond,
			},
		},
		{
			name:        "csv with 6 rows",
			nWorkers:    3,
			inputChanel: util_csv.ReadMeasurements(strings.NewReader(testCsv_6_rows)),
			processer:   NewIncrementalProcesser(),
			expected: &StatsResult{
				NQueriesProcessed:   6,
				TotalProcessingTime: time.Duration(s2) * time.Millisecond,
				MinQueryTime:        0 * time.Millisecond,
				MaxQueryTime:        5000 * time.Millisecond,
				AvarageQueryTime:    time.Duration(f2) * time.Millisecond,
				MedianQueryTime:     0 * time.Millisecond,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := &WorkerPool{
				nWorkers:    tt.nWorkers,
				inputChanel: tt.inputChanel,
				processer:   tt.processer,
			}
			res := wp.ProcessMeasurements()
			if !cmp.Equal(res, tt.expected) {
				// a "-" prefix indicates an element removed from res,
				// a "+" prefix to indicates an element added from tt.expected,
				diff := cmp.Diff(res, tt.expected)
				t.Fatalf("expected different than res:\n%v", diff)
			}
		})
	}
}

func TestWorkerPool_ProcessMeasurementsIsConcurrentSafe(t *testing.T) {
	type testCase struct {
		name        string
		nWorkers    int
		inputChanel <-chan *domain.Measurement
		processer   Processer
		expected    *StatsResult
	}

	tests := []testCase{

		// NewRandomProcesser() will return a processer that can be used concurrently,
		// but it doesn't allow us to test the business logic. It is just used to ensure
		// that the concurrent nature of the worker pool is correct, by doing
		// go test -race -v ./...

		{
			name:        "csv with 6 rows random processer",
			nWorkers:    3,
			inputChanel: util_csv.ReadMeasurements(strings.NewReader(testCsv_6_rows)),
			processer:   NewRandomProcesser(),
			// we don't care about the actual result here, just that there are no race conditions
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := &WorkerPool{
				nWorkers:    tt.nWorkers,
				inputChanel: tt.inputChanel,
				processer:   tt.processer,
			}
			wp.ProcessMeasurements()
		})
	}
}

func NewIncrementalProcesser() *IncrementalProcesser {
	return &IncrementalProcesser{}
}

// IncrementalProcesser is a processer used for testing purposes where each time that
// the Process() method is called it will take it on more second to process the next
// time the Process() method is called.
type IncrementalProcesser struct {
	mu sync.Mutex
	i  float64
}

// Process takes f.i seconds to process and returns that duration. f.i is incresaed by one
// second every time the method is called.
func (f *IncrementalProcesser) Process(ctx context.Context, measurement *domain.Measurement) (*domain.QueryResultWithTime, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	d := time.Duration(f.i) * time.Millisecond

	// fake long running task
	time.Sleep(d)

	f.i += 1000

	return &domain.QueryResultWithTime{
		Queryresults:       nil, // we don't care about this field for testing
		QueryExecutionTime: d,
	}, nil
}
