package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/domain"
)

// CSV columns indexes
const (
	hostname = iota
	startTime
	endTime
)

const (
	excpectedColumns = 3
	timeFormat       = time.DateTime
)

func ConvertLineToMeasurement(line []string) (*domain.Measurement, error) {
	n := len(line)
	if n != excpectedColumns {
		return nil, fmt.Errorf("unexpected number of columns: %v", n)
	}

	hostname := line[hostname]

	startTimeStr := line[startTime]
	startTime, err := time.Parse(timeFormat, startTimeStr)
	if err != nil {
		return nil, err
	}

	endTimeStr := line[endTime]
	endTime, err := time.Parse(timeFormat, endTimeStr)
	if err != nil {
		return nil, err
	}

	return &domain.Measurement{
		Hostname:  hostname,
		StartTime: startTime,
		EndTime:   endTime,
	}, nil
}

// ReadMeasurements reads rows from a reader r and returns a chanel to consume those rows one by one.
// The returned chanel produces values of type Measurement to decouple the business logic
// with the CSV specific format.
func ReadMeasurements(reader io.Reader) <-chan *domain.Measurement {
	r := csv.NewReader(reader)
	measurementsToConsum := make(chan *domain.Measurement)

	// produce values concurrently
	go func() {
		/*csvHeader*/ _, err := r.Read()
		if err != nil {
			log.Fatalf("aborting, not able to read CSV header: %v", err)
		}
		// log.Println("CSV header is: ", csvHeader)

		for {
			line, err := r.Read()
			if err == io.EOF {
				break
			}

			if err != nil {
				log.Printf("obtained error '%v' while reading line '%v', skipping line\n", err, line)
				continue
			}

			measurement, err := ConvertLineToMeasurement(line)
			if err != nil {
				log.Printf("skiping line '%v' due to error: %v", line, err)
				continue
			}

			measurementsToConsum <- measurement
		}

		close(measurementsToConsum)
	}()

	return measurementsToConsum
}
