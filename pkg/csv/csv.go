package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"time"
)

// CSV columns indexes
const (
	hostname = iota
	start_time
	end_time
)

const (
	excpectedColumns = 3
	timeFormat       = time.DateTime
)

// ReadMeasurements reads rows from a reader r and returns a chanel to consume those rows one by one.
// The returned chanel produces values of type SingleMeasurement to decouple the business logic
// with the CSV specific format.
func ReadMeasurements(r *csv.Reader) <-chan *Measurement {
	measurementsToConsum := make(chan *Measurement)

	go func() {
		// read headaer
		record, err := r.Read()
		if err != nil {
			log.Fatalf("aborting, not able to read CSV header: %v", err)
		}

		log.Println("CSV header is: ", record)

		for {
			line, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Printf("obtained error '%v' while reading line '%v', skipping line\n", err, line)
			}

			measurement, err := convertLineToMeasurement(line)
			if err != nil {
				log.Printf("skiping line '%v' due to error: %v", line, err)
			}

			measurementsToConsum <- measurement
		}
		close(measurementsToConsum)
	}()

	return measurementsToConsum
}

type Measurement struct {
	Hostname   string
	Start_time time.Time
	End_time   time.Time
}

func (measurement *Measurement) Print() {
	log.Println(measurement.String())
}

func (measurement *Measurement) String() string {
	return fmt.Sprintf("CSVRow -> hostname: %v, start_time: %v, end_time: %v\n", measurement.Hostname,
		measurement.Start_time.Format(timeFormat),
		measurement.End_time.Format(timeFormat))
}

func convertLineToMeasurement(line []string) (*Measurement, error) {
	n := len(line)
	if n != excpectedColumns {
		return nil, fmt.Errorf("unexpected number of columns: %v", n)
	}

	hostname := line[hostname]

	startTimeStr := line[start_time]
	startTime, err := time.Parse(timeFormat, startTimeStr)
	if err != nil {
		return nil, err
	}

	endTimeStr := line[end_time]
	endTime, err := time.Parse(timeFormat, endTimeStr)
	if err != nil {
		return nil, err
	}

	return &Measurement{
		Hostname:   hostname,
		Start_time: startTime,
		End_time:   endTime,
	}, nil
}
