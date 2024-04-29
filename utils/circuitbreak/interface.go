package circuitbreak

import "time"

// Panel is what users should use.
type Panel interface {
	Succeed(key string)
	Fail(key string)
	FailWithTrip(key string, f TripFunc)
	Timeout(key string)
	TimeoutWithTrip(key string, f TripFunc)
	IsAllowed(key string) bool
	RemoveBreaker(key string)
	DumpBreakers() map[string]Breaker
	// Close should be called if Panel is not used anymore. Or may lead to resource leak.
	// If Panel is used after Close is called, behavior is undefined.
	Close()
	GetMetricer(key string) Metricer
}

// Breaker is the base of a circuit breaker.
type Breaker interface {
	Succeed()
	Fail()
	FailWithTrip(TripFunc)
	Timeout()
	TimeoutWithTrip(TripFunc)
	IsAllowed() bool
	State() State
	Metricer() Metricer
	Reset()
}

// Metricer metrics errors, timeouts and successes
type Metricer interface {
	Failures() int64          // return the number of failures
	Successes() int64         // return the number of successes
	Timeouts() int64          // return the number of timeouts
	ConseErrors() int64       // return the consecutive errors recently
	ConseTime() time.Duration // return the consecutive error time
	ErrorRate() float64       // rate = (timeouts + failures) / (timeouts + failures + successes)
	Samples() int64           // (timeouts + failures + successes)
	Counts() (successes, failures, timeouts int64)
}

// mutable Metricer
type metricer interface {
	Metricer

	Fail()    // records a failure
	Succeed() // records a success
	Timeout() // records a timeout

	Reset()
	tick()
}
