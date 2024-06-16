package worker

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	util_csv "github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/csv"
)

const (
	hostnamePrefix = "host_"
)

// getWorkerIdForHostname returns the worker ID associated to the given hostname given nWorkers.
// i,e: all the work of the "hostname" will be sent to the worker with id
// where id is this function's return value.
// We need this function as we can have more hostnames than workers.
func getWorkerIdForHostname(nWorkers int, hostname string) (int, error) {
	hostId, err := getHostIdFromHostName(hostname)
	if err != nil {
		return 0, err
	}

	return hostId % nWorkers, nil
}

func getHostIdFromHostName(hostname string) (int, error) {
	after, found := strings.CutPrefix(hostname, hostnamePrefix)
	if !found {
		return 0, fmt.Errorf("prefix %v not found in hostname: %v", hostnamePrefix, hostname)
	}

	hostId, err := strconv.Atoi(after)
	if err != nil {
		return 0, err
	}
	return hostId, nil
}

func ProcessMeasurements(measurementsCh <-chan *util_csv.Measurement, nWorkers int) {
	// resultCh tells us where each worker should put the requests, in our case it is the same
	// but could change.
	resultCh := make(chan *Result)

	workersCh := make(ChanelOfChanels)

	// key: workerId.
	// value: at which chanel we send the work.
	// To meet "with the constraint that queries for the same hostname
	// be executed by the same worker each time".
	mapOfChannels := make(map[int]ChanelOfRequests)

	go sendWorkToWorkers(nWorkers, measurementsCh, mapOfChannels, workersCh, resultCh)
	go workersDoTheWork(workersCh)
	consumAllResults(resultCh)
}

func sendWorkToWorkers(nWorkers int, mesaurementsCh <-chan *util_csv.Measurement,
	mapOfWorkerAndChannels map[int]ChanelOfRequests, workersCh ChanelOfChanels, resultCh chan *Result) {
	var wg sync.WaitGroup

	for row := range mesaurementsCh {
		workerId, err := getWorkerIdForHostname(nWorkers, row.Hostname)
		if err != nil {
			log.Printf("got error when getting worker id from hostname, skipping this row: %v\n", err)
			continue
		}

		// first time it will be nil for sure as problem statement wants us to process each line as soon
		// as we read it. If we could load all input in memory at once to check for errors,
		// we could instead first build the map and just pass the map to this function initialized.
		if mapOfWorkerAndChannels[workerId] == nil {
			mapOfWorkerAndChannels[workerId] = make(chan *Request)
		}

		wg.Add(1)

		// send new workerCh to the chanel of workers
		// TODO: I think this can be simplified
		workersCh <- mapOfWorkerAndChannels[workerId]
		request := &Request{
			arg:      row,
			resultCh: resultCh,
			wg:       &wg,
		}

		// and send the request to the worker
		mapOfWorkerAndChannels[workerId] <- request
	}

	fmt.Printf("mapOfChannels has length: %v\n", len(mapOfWorkerAndChannels))

	go func() {
		wg.Wait()
		close(resultCh)
	}()

}

func workersDoTheWork(allChanels ChanelOfChanels) {
	for requestCh := range allChanels {
		go worker(requestCh)
	}
}

// worker reads work to do from "requestCh" and performs some work,
// and sends the result to the corresponding channel specified in the request.
// This function is intended to be called concurrently (using "go ...")
func worker(requestCh ChanelOfRequests) {
	for request := range requestCh {
		now := time.Now()
		fmt.Printf("worker started to process %v\n", *request)

		//random stuff to play for now:
		result := &Result{
			timeSpend: time.Since(now),
		}
		request.resultCh <- result
		request.wg.Done()
	}
}
