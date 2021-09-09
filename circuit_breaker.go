package circuitbreaker

import (
	"sync/atomic"
	"time"
)

const (
	OPEN uint32 = iota
	CLOSED
	HALF_CLOSED
)

type Counters struct {
	requestsContinuous uint64
	successContinuous  uint64
	failureContinuous  uint64

	requestsTotal  uint64
	successTotal   uint64
	failureTotal   uint64
	throttledTotal uint64
}

func (c *Counters) onFailure() {
	atomic.AddUint64(&c.requestsTotal, 1)
	atomic.AddUint64(&c.requestsContinuous, 1)
	atomic.AddUint64(&c.failureTotal, 1)
	atomic.StoreUint64(&c.successContinuous, 0)
	atomic.AddUint64(&c.failureContinuous, 1)
}

func (c *Counters) onSuccess() {
	atomic.AddUint64(&c.requestsTotal, 1)
	atomic.AddUint64(&c.requestsContinuous, 1)
	atomic.AddUint64(&c.successTotal, 1)
	atomic.AddUint64(&c.successContinuous, 1)
	atomic.StoreUint64(&c.failureContinuous, 0)
}

func (c *Counters) clear() {
	atomic.StoreUint64(&c.requestsTotal, 0)
	atomic.StoreUint64(&c.throttledTotal, 0)
	atomic.StoreUint64(&c.failureTotal, 0)
	atomic.StoreUint64(&c.successTotal, 0)
}

type CircuitBreaker struct {
	state       uint32 // state is state of CircuitBreaker (e.g open, closed, half-closed).
	interval    time.Duration
	timeout     time.Duration
	threshold   uint32
	maxRequests uint32
	counters    Counters // counters holds counters (e.g success, failed).
}

// Allow returns true if request allowed or false if not. Use it in your code.
func (cb *CircuitBreaker) Allow() bool {
	st := atomic.LoadUint32(&cb.state)
	if st == CLOSED {
		return true
	}
	if st == HALF_CLOSED {
		if atomic.LoadUint64(&cb.counters.requestsContinuous) < uint64(cb.maxRequests) {
			return true
		}
	}
	atomic.AddUint64(&cb.counters.throttledTotal, 1)
	return false
}

// Settings use it for create new CircuitBreaker object.
type Settings struct {
	Interval    time.Duration // Interval after which all counters will be reset if state closed. Default disabled.
	Timeout     time.Duration // Timeout is time after which CircuitBreaker will enter the half-closed state. Default 1 second.
	Threshold   int           // Threshold is the value after which CircuitBreaker will switch to the Open state. Default 1.
	MaxRequests int           // MaxRequests is the number of requests that will be sent in the Half Open state. Default 1.
}

func (s *Settings) validate() {
	if s.Threshold <= 0 {
		s.Threshold = 1
	}
	if s.Timeout <= 0 {
		s.Timeout = 1 * time.Second
	}
	if s.MaxRequests <= 0 {
		s.MaxRequests = 1
	}
}

func New(settings Settings) *CircuitBreaker {
	settings.validate()

	cb := &CircuitBreaker{
		state:       1,
		interval:    settings.Interval,
		timeout:     settings.Timeout,
		threshold:   uint32(settings.Threshold),
		maxRequests: uint32(settings.MaxRequests),
	}
	if cb.interval != 0 {
		go cb.clearCounters()
	}
	return cb
}

// Counters return Counters object with different counters.
func (cb *CircuitBreaker) Counters() Counters {
	return cb.counters
}

func (cb *CircuitBreaker) RegisterError() {
	cb.counters.onFailure()
	if atomic.LoadUint32(&cb.state) == HALF_CLOSED {
		cb.updateState(OPEN)
		go cb.startWaitTimer()
	}
	if atomic.LoadUint64(&cb.counters.failureContinuous) > uint64(cb.threshold) {
		if atomic.LoadUint32(&cb.state) == CLOSED {
			cb.updateState(OPEN)
			go cb.startWaitTimer()
		}
	}
}

func (cb *CircuitBreaker) RegisterOK() {
	cb.counters.onSuccess()
	if atomic.LoadUint32(&cb.state) == HALF_CLOSED {
		if atomic.LoadUint64(&cb.counters.successContinuous) == uint64(cb.maxRequests) {
			cb.updateState(CLOSED)
		}
	}
}

func (cb *CircuitBreaker) clearCounters() {
	ticker := time.NewTicker(cb.interval)
	for {
		<-ticker.C
		if atomic.LoadUint32(&cb.state) == CLOSED {
			cb.counters.clear()
		}
	}
}

func (cb *CircuitBreaker) startWaitTimer() {
	<-time.After(cb.timeout)
	cb.updateState(HALF_CLOSED)
}

func (cb *CircuitBreaker) updateState(state uint32) {
	atomic.StoreUint64(&cb.counters.requestsContinuous, 0)
	atomic.StoreUint32(&cb.state, state)
}
