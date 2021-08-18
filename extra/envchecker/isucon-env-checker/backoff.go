package main

import (
	"time"

	"github.com/cenkalti/backoff/v4"
)

func newBackoff() backoff.BackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     1 * time.Second,
		RandomizationFactor: 0.25,
		Multiplier:          2,
		MaxInterval:         10 * time.Second,
		MaxElapsedTime:      30 * time.Second,
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	b.Reset()
	return b
}
