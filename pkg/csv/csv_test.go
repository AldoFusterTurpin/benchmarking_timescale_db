package csv

import (
	"errors"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/domain"
)

func safeTimeParse(timeValue string) time.Time {
	t, err := time.Parse(time.DateTime, timeValue)
	if err != nil {
		log.Fatal("safeTimeParse failed", err)
	}
	return t
}

func Test_convertLineToRow(t *testing.T) {
	type testCase struct {
		name          string
		line          []string
		expectedRow   *domain.Measurement
		expectedError error
	}
	tests := []testCase{
		{
			name:          "returns appropiate error when row has no columns",
			line:          nil,
			expectedRow:   nil,
			expectedError: errors.New("unexpected number of columns: 0"),
		},
		{
			name:          "returns appropiate error when row has more columns than expected",
			line:          []string{"host_000008", "2017-01-01 08:59:22", "2017-01-01 09:59:22", "unexpected_column"},
			expectedRow:   nil,
			expectedError: errors.New("unexpected number of columns: 4"),
		},
		{
			name: "can convert to row without errors",
			line: []string{"host_000008", "2017-01-01 08:59:22", "2017-01-01 09:59:22"},
			expectedRow: &domain.Measurement{
				Hostname:  "host_000008",
				StartTime: safeTimeParse("2017-01-01 08:59:22"),
				EndTime:   safeTimeParse("2017-01-01 09:59:22"),
			},
		},
		{
			name:          "returns appropiate error when start_time has invalid format",
			line:          []string{"host_000008", "1234", "2017-01-01 09:59:22"},
			expectedRow:   nil,
			expectedError: errors.New("parsing time \"1234\" as \"2006-01-02 15:04:05\": cannot parse \"\" as \"-\""),
		},
		{
			name:          "returns appropiate error when end_time has invalid format",
			line:          []string{"host_000008", "2017-01-01 08:59:22", "in-2017-valid"},
			expectedRow:   nil,
			expectedError: errors.New("parsing time \"in-2017-valid\" as \"2006-01-02 15:04:05\": cannot parse \"in-2017-valid\" as \"2006\""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row, err := ConvertLineToMeasurement(tt.line)

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

			if !reflect.DeepEqual(tt.expectedRow, row) {
				t.Fatalf("expected and row are different\nexpected: %+v\nrow: %+v", tt.expectedRow, row)
			}
		})
	}
}
