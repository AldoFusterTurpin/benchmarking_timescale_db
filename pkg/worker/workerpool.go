package worker

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

	"github.com/AldoFusterTurpin/benchmarking_timescale_db/pkg/domain"
)

const (
	hostnamePrefix = "host_"
)

// key: workerId.
// value: at which chanel we send the work.
// To meet with the constraint that "queries for the same hostname
// be executed by the same worker each time".
type mapOfChannels map[int]chan *Request

type WorkerPool struct {
	nWorkers    int
	inputChanel <-chan *domain.Measurement
	processer   Processer
}

func NewWorkerPool(nWorkers int, inputChanel <-chan *domain.Measurement, processer Processer) *WorkerPool {
	return &WorkerPool{
		nWorkers:    nWorkers,
		inputChanel: inputChanel,
		processer:   processer,
	}
}

func (wp *WorkerPool) ProcessMeasurements() *StatsResult {
	// resultCh tells where each worker should put the result.
	// it will be shared in our case.
	resultCh := make(chan *domain.QueryResultWithTime)

	// queueOfWorkers acts as a multiplexer/scheduler of jobs.
	// each value of queueOfWorkers is a chanel that one of the workers
	// will be reading from to do some work.
	queueOfWorkers := make(chan chan *Request)

	go wp.sendWorkToWorkers(queueOfWorkers, resultCh, wp.processer)
	go wp.runWorkers(queueOfWorkers)

	return consumAllResults(resultCh)
}

// sendWorkToWorkers reads from the input chanel of the workerPool and creates a queueOfWorkers that is used
// to track the work that each worker should perform. Each of the workers will be waiting for work in its
// associated chanel.
func (wp *WorkerPool) sendWorkToWorkers(queueOfWorkers chan chan *Request, resultCh chan *domain.QueryResultWithTime, processer Processer) {
	var wg sync.WaitGroup

	mapOfChannels := make(mapOfChannels)

	for row := range wp.inputChanel {
		workerId, err := getWorkerIdForHostname(wp.nWorkers, row.Hostname)
		if err != nil {
			log.Printf("got error when getting worker id for hostname, skipping this row: %v\n", err)
			continue
		}

		// if the worker with id "workerId" does not have a chanel to read requests from ...
		if mapOfChannels[workerId] == nil {
			// we create a new chanel for the worker with id "workerId" to receive requests from
			mapOfChannels[workerId] = make(chan *Request)

			// and send the chanel to the queue of workers indicating that there is a new chanel that
			// the worker will read from to receive and process requests
			queueOfWorkers <- mapOfChannels[workerId]
		}

		// and we send the request to the appropiate chanel of the specific "workerId"
		wg.Add(1)
		request := NewRequest(row, resultCh, &wg, processer)
		mapOfChannels[workerId] <- request
	}

	// we can close the channels that the workers are reading from as we
	// have already sent all the work to them and no more workers will be added
	for _, v := range mapOfChannels {
		close(v)
	}

	// we can also close the queue of workers
	close(queueOfWorkers)

	// fmt.Printf("mapOfChannels has length: %v\n", len(mapOfChannels))

	// wait in a different goroutine to avoid blocking the main goroutine
	go func() {
		// wait until we have processed all the requests
		wg.Wait()

		// we can now close the chanel of results as we have processed all the requests
		close(resultCh)
	}()
}

// runWorkers iterates over a chanel of chanels and invokes a new worker in each
// chanel, where each chanel contains the jobs (requests) to process for the same worker.
func (wp *WorkerPool) runWorkers(queueOfWorkers chan chan *Request) {
	for requestCh := range queueOfWorkers {
		go wp.runSingleWorker(requestCh)
	}
}

// runSingleWorker receives work to do from "requestCh", does some work
// and sends the result to the corresponding channel specified in the request.
func (wp *WorkerPool) runSingleWorker(requestsCh chan *Request) {
	for request := range requestsCh {
		// log.Printf("worker started to process %v\n", request)

		ctx := context.Background()
		resultOfRequest, err := wp.processer.Process(ctx, request.measurement)
		if err != nil {
			log.Printf("received error processing request %v: %v\n", request, err)
			continue
		}

		request.resultCh <- resultOfRequest
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
