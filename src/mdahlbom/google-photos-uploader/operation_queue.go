// Implements an operation queue with configurable concurrency.
package main

import (
	"errors"
	"fmt"
	"sync"
)

// Errors
var (
	errShutdown = errors.New("Queue has been shutdown")
)

// The main queue type. Operations to be performed are functions without
// parameters. In-order execution is not guaranteed unless maxConcurrency
// is set to 1.
type OperationQueue struct {
	// The main queue buffer
	bufferChan chan func()

	// Maximum number of concurrent goroutines processing items in the queue
	maxConcurrency int

	// Indicates whether the queue is shutting down; if true, no more operations
	// are accepted to the queue.
	shutdown bool

	// Used to signal that graceful shutdown has completed and all tasks
	// have been processed
	shutdownDoneChan chan bool

	// Controls concurrent access to critical resources (buffer, shutdown)
	lock sync.RWMutex

	// Number of items in buffer (or waiting to be put into the buffer)
	// plus the number of items being processed currently.
	itemsLeft int32
}

// Creates a new queue. The parameter maxConcurrency specifies how many
// concurrent goroutines are deployed to work on the queue, and bufferSize
// defines the size of the queue; after the queue buffer is full, Add()
// will block until there is room in the buffer.
func NewOperationQueue(maxConcurrency,
	bufferSize int) (*OperationQueue, error) {

	if maxConcurrency <= 0 {
		return nil, fmt.Errorf("Invalid maxConcurrency value: %v",
			maxConcurrency)
	}

	if bufferSize <= 0 {
		return nil, fmt.Errorf("Invalid bufferSize value: %v", bufferSize)
	}

	q := &OperationQueue{
		bufferChan:       make(chan func(), bufferSize),
		maxConcurrency:   maxConcurrency,
		shutdown:         false,
		itemsLeft:        0,
		shutdownDoneChan: make(chan bool), //TODO add buffer size 1
	}

	q.start()

	return q, nil
}

// Processing loop; run in a dedicated goroutine. Calls the callback every
// time an operation execution has completed.
func (q *OperationQueue) run(callback func()) {
	for {
		op, ok := <-q.bufferChan
		if !ok {
			log.Debugf("Channel closed - exiting worker goroutine!")
			return
		}
		op()
		callback()
	}
}

// Starts the queue's message processing mechanism
func (q *OperationQueue) start() {
	callback := func() {
		q.lock.Lock()
		defer q.lock.Unlock()

		q.itemsLeft--
		if q.shutdown && q.itemsLeft == 0 {
			q.shutdownDoneChan <- true
			close(q.shutdownDoneChan)
		}
	}

	// Create a gorotine to match the set maxConcurrency
	for i := 0; i < q.maxConcurrency; i++ {
		go q.run(callback)
	}
}

// Gracefully shuts down the queue, waiting for all operations to be completed.
// No new operations can be Add()ed after this method has been called.
func (q *OperationQueue) GracefulShutdown() {
	done := func() bool {
		q.lock.Lock()
		defer q.lock.Unlock()

		if q.shutdown {
			log.Debugf("Already shut down.")
			return true
		}

		q.shutdown = true
		if q.itemsLeft == 0 {
			log.Debugf("No items left - shutdown done w/o waiting")
			close(q.bufferChan)
			return true
		}

		return false
	}()

	if done {
		return
	}

	_, _ = <-q.shutdownDoneChan

	close(q.bufferChan)
}

// Adds a new operation to the queue; may block if the buffer is full. If
// the queue has been shut down, returns errShutdown.
func (q *OperationQueue) Add(op func()) error {
	q.lock.Lock()
	if q.shutdown {
		log.Errorf("Trying to Add() to a queue that has been shut down")
		q.lock.Unlock()
		return errShutdown
	}
	q.itemsLeft++
	q.lock.Unlock()

	// Insert operation into the buffer; this may block
	q.bufferChan <- op

	return nil
}
