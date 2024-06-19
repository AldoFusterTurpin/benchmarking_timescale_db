package main

import (
	"fmt"
	"log"
	"os"
	"time"

	util_csv "github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/csv"
	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/db"
	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/domain"
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

	t := time.Now()

	rowsCh := getRowsToConsum("/data/query_params.csv")
	nWorkers := 10
	processer := worker.NewRandomProcesser()
	workerPool := worker.NewWorkerPool(nWorkers, rowsCh, processer)

	statsResult := workerPool.ProcessMeasurements()
	log.Println(statsResult)
	fmt.Println("total execution time elapsed", time.Since(t))
}

// getRowsToConsum returns a channel that returns a row every time we read from that chanel.
func getRowsToConsum(csvPath string) <-chan *domain.Measurement {
	f, err := os.Open(csvPath)
	if err != nil {
		log.Fatal(err)
	}

	rowsToConsum := util_csv.ReadMeasurements(f)
	return rowsToConsum
}
