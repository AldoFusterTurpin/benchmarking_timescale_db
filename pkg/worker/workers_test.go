package worker

import (
	"errors"
	"testing"
)

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

// func TestProcessMeasurements(t *testing.T) {
// 	type testCase struct {
// 		name           string
// 		measurementsCh <-chan *util_csv.Measurement
// 		nWorkers       int
// 	}
// 	tests := []testCase{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ProcessMeasurements(tt.measurementsCh, tt.nWorkers)
// 		})
// 	}
// }
