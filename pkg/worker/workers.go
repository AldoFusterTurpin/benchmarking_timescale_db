package worker

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	util_csv "github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/csv"
)

const (
	hostnamePrefix = "host_"
)

// key: workerId.
// value: at which queue we send the work.
// To meet with the constraint that "queries for the same hostname
// be executed by the same worker each time".
type mapOfQueues map[int]Queue

// each element of QueuesCh is the Queue of a specific worker.
// So len(QueuesCh) will be equal to the number of workers we have.
type QueuesCh chan Queue

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
	// resultCh tells where each worker should put the result
	resultCh := make(chan *Result)

	queueOfWorkers := make(QueuesCh)

	go wp.sendWorkToWorkers(queueOfWorkers, resultCh)
	go wp.workers(queueOfWorkers)

	consumAllResults(resultCh)
}

func (wp *WorkerPool) sendWorkToWorkers(queueOfWorkers QueuesCh, resultCh chan *Result) {
	var wg sync.WaitGroup

	mapOfChannels := make(mapOfQueues)

	for row := range wp.measurementsCh {
		workerId, err := getWorkerIdForHostname(wp.nWorkers, row.Hostname)
		if err != nil {
			log.Printf("got error when getting worker id for hostname, skipping this row: %v\n", err)
			continue
		}

		if mapOfChannels[workerId] == nil {
			mapOfChannels[workerId] = make(Queue)
		}

		wg.Add(1)

		queueOfWorkers <- mapOfChannels[workerId]

		request := &Request{
			measurement: row,
			resultCh:    resultCh,
			wg:          &wg,
		}

		mapOfChannels[workerId] <- request
	}

	fmt.Printf("mapOfChannels has length: %v\n", len(mapOfChannels))

	// wait in a different goroutine to avoid blocking the main goroutine
	go func() {
		wg.Wait()
		close(resultCh)
	}()
}

// workers iterates over a queue of workers and invokes a new worker in each
// go routine to process the tasks concurrently.
func (wp *WorkerPool) workers(queueOfWorkers QueuesCh) {
	for requestCh := range queueOfWorkers {
		go wp.worker(requestCh)
	}
}

// worker receives work to do from "requestCh",
// and sends the result to the corresponding channel specified in the request.
func (wp *WorkerPool) worker(requestsCh Queue) {
	for request := range requestsCh {
		fmt.Printf("worker started to process %v\n", request)

		result, err := request.Process()
		if err != nil {
			log.Printf("received error processing request %v: %v\n", request, err)
			continue
		}

		request.resultCh <- result
		request.wg.Done()
	}
}

// getWorkerIdForHostname returns the worker ID associated to the given hostname given nWorkers.
// i,e: all the work of the "hostname" will be sent to the worker
// with id "id" where "id" is this function's return value.
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
