package resilience

import (
	"time"

	"github.com/sony/gobreaker"
)

// CircuitBreaker defines the interface for circuit breakers
type CircuitBreaker interface {
	// Execute runs the given function within the circuit breaker
	Execute(req func() (interface{}, error)) (interface{}, error)
	// Name returns the name of the circuit breaker
	Name() string
	// State returns the current state of the circuit breaker
	State() string
}

// SonyBreaker implements CircuitBreaker using sony/gobreaker
type SonyBreaker struct {
	cb *gobreaker.CircuitBreaker
}

// CircuitBreakerConfig holds configuration for the circuit breaker
type CircuitBreakerConfig struct {
	Name        string
	MaxRequests uint32
	Interval    time.Duration
	Timeout     time.Duration
	ReadyToTrip func(counts gobreaker.Counts) bool
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config CircuitBreakerConfig) *SonyBreaker {
	settings := gobreaker.Settings{
		Name:        config.Name,
		MaxRequests: config.MaxRequests,
		Interval:    config.Interval,
		Timeout:     config.Timeout,
		ReadyToTrip: config.ReadyToTrip,
	}

	if settings.ReadyToTrip == nil {
		settings.ReadyToTrip = func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		}
	}

	return &SonyBreaker{
		cb: gobreaker.NewCircuitBreaker(settings),
	}
}

// Execute runs the given function within the circuit breaker
func (b *SonyBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	return b.cb.Execute(req)
}

// Name returns the name of the circuit breaker
func (b *SonyBreaker) Name() string {
	return b.cb.Name()
}

// State returns the current state of the circuit breaker
func (b *SonyBreaker) State() string {
	return b.cb.State().String()
}

// MockCircuitBreaker for testing
type MockCircuitBreaker struct {
	NameVal  string
	StateVal string
}

func (m *MockCircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {
	return req()
}

func (m *MockCircuitBreaker) Name() string {
	return m.NameVal
}

func (m *MockCircuitBreaker) State() string {
	if m.StateVal == "" {
		return "closed"
	}
	return m.StateVal
}
