package circuitbreaker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewCircuitBreaker(t *testing.T) {
	// test default
	cb := New(Settings{})
	assert.Equal(t, uint32(1), cb.state)
	assert.Equal(t, time.Duration(0), cb.interval)
	assert.Equal(t, time.Second, cb.timeout)
	assert.Equal(t, uint32(1), cb.maxRequests)
	assert.Equal(t, uint32(1), cb.threshold)

	// test custom
	cb = New(Settings{
		Interval:    10 * time.Second,
		Timeout:     3 * time.Second,
		Threshold:   2,
		MaxRequests: 2,
	})
	assert.Equal(t, uint32(1), cb.state)
	assert.Equal(t, 10*time.Second, cb.interval)
	assert.Equal(t, 3*time.Second, cb.timeout)
	assert.Equal(t, uint32(2), cb.maxRequests)
	assert.Equal(t, uint32(2), cb.threshold)
}
