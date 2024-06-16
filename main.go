package main

import (
	"encoding/csv"
	"log"
	"os"

	util_csv "github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/csv"
	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/db"
	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/worker"
)

func main() {
	connString := db.DefaultConnString()
	dbService, err := db.NewDBService(connString)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Successfully CONNECTED TO TIMESCALE_DB")

	nRows, err := dbService.CountAllRows()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("there are", nRows, "rows")

	rowsCh := getRowsToConsum()

	nWorkers := 5
	worker.ProcessMeasurements(rowsCh, nWorkers)
}

// getRowsToConsum returns a channel that returns a row every time we read from that chanel.
func getRowsToConsum() <-chan *util_csv.Measurement {
	csvPath := "/data/query_params.csv"
	f, err := os.Open(csvPath)
	if err != nil {
		log.Fatal(err)
	}

	r := csv.NewReader(f)
	rowsToConsum := util_csv.ReadMeasurements(r)
	return rowsToConsum
}
