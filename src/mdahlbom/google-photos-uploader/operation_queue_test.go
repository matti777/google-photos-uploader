package main

import (
	"fmt"
	"testing"
	"time"
)

func NewTimeoutTimer(d time.Duration, t *testing.T) *time.Timer {
	timer := time.AfterFunc(d, func() {
		// Timed out, fail the test
		t.Errorf("Timed out")
	})

	return timer
}

func TestCreationParams(t *testing.T) {
	if _, err := NewOperationQueue(-1, 1); err == nil {
		t.Errorf("Was expecting error")
	}

	if _, err := NewOperationQueue(1, 0); err == nil {
		t.Errorf("Was expecting error")
	}
}

func TestSlowCompletion(t *testing.T) {
	q, err := NewOperationQueue(1, 1)
	if err != nil {
		t.Errorf("Failed to create queue")
	}

	q.Add(func() {
		fmt.Println("Slow operation starting..")
		time.Sleep(time.Millisecond * 250)
		fmt.Println("Slow operation completed.")
	})

	timer := NewTimeoutTimer(time.Millisecond*500, t)
	fmt.Println("Calling GracefulShutdown")
	q.GracefulShutdown()
	fmt.Println("GracefulShutdown returned")
	timer.Stop()
}

func TestFastCompletion(t *testing.T) {
	q, err := NewOperationQueue(1, 1)
	if err != nil {
		t.Errorf("Failed to create queue")
	}

	q.Add(func() {
		fmt.Println("Fast operation being run")
	})

	time.Sleep(time.Millisecond * 100)
	timer := NewTimeoutTimer(time.Millisecond*100, t)

	fmt.Println("Calling GracefulShutdown")
	q.GracefulShutdown()
	fmt.Println("GracefulShutdown returned")
	timer.Stop()
}

func TestFastAndSlowCompletion(t *testing.T) {
	q, err := NewOperationQueue(1, 1)
	if err != nil {
		t.Errorf("Failed to create queue")
	}

	q.Add(func() {
		fmt.Println("Fast operation being run")
	})
	q.Add(func() {
		fmt.Println("Slow operation starting..")
		time.Sleep(time.Millisecond * 250)
		fmt.Println("Slow operation completed.")
	})

	time.Sleep(time.Millisecond * 50)
	timer := NewTimeoutTimer(time.Millisecond*500, t)

	fmt.Println("Calling GracefulShutdown")
	q.GracefulShutdown()
	fmt.Println("GracefulShutdown returned")
	timer.Stop()
}

func TestConcurrentExecution(t *testing.T) {
	numItems := 10
	numConcurrency := 4

	starts := make([]bool, numConcurrency)
	ends := make([]bool, numConcurrency)
	for i := 0; i < numConcurrency; i++ {
		starts[i] = false
		ends[i] = false
	}

	q, err := NewOperationQueue(numConcurrency, numItems-1)
	if err != nil {
		t.Errorf("Failed to create queue")
	}

	for i := 0; i < numItems; i++ {
		index := i

		q.Add(func() {
			fmt.Println("Slow operation starting..")
			if index < numConcurrency {
				starts[index] = true
				for _, b := range ends {
					if b {
						t.Errorf("No operations should have completed by now")
					}
				}
			}
			time.Sleep(time.Millisecond * 500)
			fmt.Println("Slow operation completed.")
			if index < numConcurrency {
				ends[index] = true
			}
		})
	}

	timer := NewTimeoutTimer(time.Millisecond*2500, t)

	fmt.Println("Calling GracefulShutdown")
	q.GracefulShutdown()
	fmt.Println("GracefulShutdown returned")
	timer.Stop()
}
