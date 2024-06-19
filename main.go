package main

import (
	"log"
	"os"
	"strconv"

	util_csv "github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/csv"
	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/db"
	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/domain"
	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/worker"
)

const (
	defaultWorkers = 5
	workersEnvName = "WORKERS"
)

func main() {
	connString := db.DefaultConnString()
	dbService, err := db.NewDBService(connString)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully CONNECTED TO TIMESCALE_DB")

	rowsCh := getRowsToConsumFromCsv("/data/query_params.csv")

	nWorkers := getNumberOfWorkers()

	workerPool := worker.NewPool(nWorkers, rowsCh, dbService)
	statsResult := workerPool.ProcessMeasurements()

	log.Printf("executing worker pool with %v workers\n", nWorkers)

	log.Println(statsResult)
}

func getNumberOfWorkers() int {
	nWorkersStr := os.Getenv(workersEnvName)
	if nWorkersStr == "" {
		return defaultWorkers
	}

	nWorkers, err := strconv.Atoi(nWorkersStr)
	if err != nil {
		log.Printf("error pardsing workers variable: %v. using default value instead: %v", err, defaultWorkers)
		return defaultWorkers
	}
	return nWorkers

}

// getRowsToConsumFromCsv returns a channel that returns a measurement every time we read from that chanel.
func getRowsToConsumFromCsv(csvPath string) <-chan *domain.Measurement {
	f, err := os.Open(csvPath)
	if err != nil {
		log.Fatal(err)
	}

	rowsToConsum := util_csv.ReadMeasurements(f)
	return rowsToConsum
}
