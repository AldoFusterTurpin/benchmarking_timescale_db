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

type WorkerPool struct {
	nWorkers       int
	measurementsCh <-chan *util_csv.Measurement
}

func NewWorkerPool(nWorkers int, measurementsCh <-chan *util_csv.Measurement) *WorkerPool {
	return &WorkerPool{
		nWorkers:       nWorkers,
		measurementsCh: measurementsCh,
	}
}

func (wp *WorkerPool) ProcessMeasurements() {
	// resultCh tells where each worker should put the requests
	resultCh := make(chan *Result)

	workersCh := make(ChanelOfChanels)

	// key: workerId.
	// value: at which chanel we send the work.
	// To meet "with the constraint that queries for the same hostname
	// be executed by the same worker each time".
	mapOfChannels := make(map[int]ChanelOfRequests)

	go wp.sendWorkToWorkers(mapOfChannels, workersCh, resultCh)
	go wp.workersDoTheWork(workersCh)

	// TODO: instead of hardcoding what to do with the result,
	// return in this method a chanel of results that the clients
	// can use to perform the business logic.
	consumAllResults(resultCh)
}

func (wp *WorkerPool) sendWorkToWorkers(mapOfWorkerAndChannels map[int]ChanelOfRequests, workersCh ChanelOfChanels, resultCh chan *Result) {
	var wg sync.WaitGroup

	for row := range wp.measurementsCh {
		workerId, err := getWorkerIdForHostname(wp.nWorkers, row.Hostname)
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
		// TODO: I think this can be simplified and avoid that extra chanel of chanels,
		// by starting the worker directly here.
		workersCh <- mapOfWorkerAndChannels[workerId]

		request := &Request{
			measurement: row,
			resultCh:    resultCh,
			wg:          &wg,
		}
		// send the request to the worker to do some work
		mapOfWorkerAndChannels[workerId] <- request
	}

	fmt.Printf("mapOfChannels has length: %v\n", len(mapOfWorkerAndChannels))

	// wait in a different Go routine to avoid blocking, common mistake
	go func() {
		wg.Wait()
		close(resultCh)
	}()
}

func (wp *WorkerPool) workersDoTheWork(workersCh ChanelOfChanels) {
	for requestCh := range workersCh {
		go wp.worker(requestCh)
	}
}

// worker reads work to do from "requestCh" and performs some work,
// and sends the result to the corresponding channel specified in the request.
func (wp *WorkerPool) worker(requestsCh ChanelOfRequests) {
	for request := range requestsCh {
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
