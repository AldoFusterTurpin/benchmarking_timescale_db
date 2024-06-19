package worker

import (
	"context"
	_ "embed"
	"errors"
	"math/rand"
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

	sum1 := (0 + 10 + 20)
	avarage1 := float64(0+10+20) / 3
	median1 := 10

	sum2 := (0 + 10 + 20 + 30 + 40 + 50)
	avarage2 := float64(0+10+20+30+40+50) / 6
	median2 := 25

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
				TotalProcessingTime: time.Duration(sum1) * time.Millisecond,
				MinQueryTime:        0 * time.Millisecond,
				MaxQueryTime:        20 * time.Millisecond,
				AvarageQueryTime:    time.Duration(avarage1) * time.Millisecond,
				MedianQueryTime:     time.Duration(median1) * time.Millisecond,
			},
		},
		{
			name:        "csv with 6 rows",
			nWorkers:    3,
			inputChanel: util_csv.ReadMeasurements(strings.NewReader(testCsv_6_rows)),
			processer:   NewIncrementalProcesser(),
			expected: &StatsResult{
				NQueriesProcessed:   6,
				TotalProcessingTime: time.Duration(sum2) * time.Millisecond,
				MinQueryTime:        0 * time.Millisecond,
				MaxQueryTime:        50 * time.Millisecond,
				AvarageQueryTime:    time.Duration(avarage2) * time.Millisecond,
				MedianQueryTime:     time.Duration(median2) * time.Millisecond,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wp := &Pool{
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

func NewIncrementalProcesser() *IncrementalProcesser {
	return &IncrementalProcesser{}
}

// IncrementalProcesser is a processer used for testing purposes where each time that
// the Process() method is called it will take it on more second to process the next
// time the Process() method is called. It can not be used to test the concurrent nature
// of the worker pool as it contains a mutex that will be the bottleneck/wait point of the go routines.
// To test the concurrent nature of the worker pool check "TestWorkerPool_ProcessMeasurementsIsConcurrentSafe"
// which uses a non blocking RandomProcesser
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

	f.i += 10

	return &domain.QueryResultWithTime{
		Queryresults:       nil, // we don't care about this field for testing
		QueryExecutionTime: d,
	}, nil
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
		// already done in the dockerfile
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
			wp := &Pool{
				nWorkers:    tt.nWorkers,
				inputChanel: tt.inputChanel,
				processer:   tt.processer,
			}
			wp.ProcessMeasurements()
		})
	}
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
			MaxCPUUsage: 3000,
			MinCPUUsage: 1000,
		},
	}

	return &domain.QueryResultWithTime{
		Queryresults:       queryResults,
		QueryExecutionTime: randomDuration,
	}, nil
}

func Test_getMedian(t *testing.T) {
	type testCase struct {
		name string
		s    []time.Duration
		want time.Duration
	}

	f1 := 0
	f2 := 6.5
	f3 := 8

	tests := []testCase{
		{
			name: "median of slice with 0 elements",
			s:    nil,
			want: time.Duration(f1),
		},
		{
			name: "median of slice with even number of elements",
			s:    []time.Duration{5, 2, 95, 2, 44, 5, 8, 9, 2, 46},
			want: time.Duration(f2),
		},
		{
			name: "median of slice with odd number of elements",
			s:    []time.Duration{109, 5, 5, 8, 9, 2, 46},
			want: time.Duration(f3),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getMedian(tt.s); got != tt.want {
				t.Errorf("getMedian() = %v, want %v", got, tt.want)
			}
		})
	}
}
