package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// CircuitBreaker states
const (
	Closed State = iota
	Open
	HalfOpen
)

// State type for the Circuit Breaker states
type State int

type CircuitBreaker struct {
	state            State
	failureCount     int
	successCount     int
	failureThreshold int
	successThreshold int
	stateOpenTimeout time.Duration
	lastFailure      time.Time
}

func NewCircuitBreaker(failureThreshold int, timeout time.Duration, successThreshold int) *CircuitBreaker {
	return &CircuitBreaker{
		state:            Closed,
		failureThreshold: failureThreshold,
		stateOpenTimeout: timeout,
		successThreshold: successThreshold,
	}
}

func (cb *CircuitBreaker) Run(mainFunc func() error, fallbackFunc func() error) error {
	switch cb.state {
	case Open:
		if time.Since(cb.lastFailure) > cb.stateOpenTimeout {
			fmt.Println("State is half open")
			cb.state = HalfOpen
		} else {
			fmt.Println("State is open")
			return fallbackFunc()
		}
	}

	err := mainFunc()
	if err != nil {
		cb.increaseFailureCount()
		return fallbackFunc()
	}

	cb.success()
	fmt.Println("State is closed")
	return nil
}

func (cb *CircuitBreaker) increaseFailureCount() {
	cb.failureCount++
	if cb.failureCount >= cb.failureThreshold {
		cb.state = Open
		cb.lastFailure = time.Now()
	}
}

func (cb *CircuitBreaker) success() {
	if cb.state == HalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successThreshold {
			cb.reset()
		}
	} else {
		cb.reset()
	}
}

func (cb *CircuitBreaker) reset() {
	cb.state = Closed
	cb.failureCount = 0
	cb.successCount = 0
}

func MainCall() error {
	url := "http://localhost:8080/swagger/doc.json"
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New("service call failed")
	}
	fmt.Printf("Service call to %s success!! \n", url)
	return nil
}

func FallbackCall() error {
	fmt.Println("Executing fallback call")
	return nil
}

func main() {
	cb := NewCircuitBreaker(3, 5*time.Second, 2)

	for {
		err := cb.Run(MainCall, FallbackCall)
		if err != nil {
			fmt.Printf("Service call failed: %v\n", err)
		}
		time.Sleep(1 * time.Second)
	}
}
