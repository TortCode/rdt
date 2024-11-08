package gbn

import "time"

// TimeoutTimer represents a timer that will timeout after a fixed duration
type TimeoutTimer struct {
	timer *time.Timer
	d     time.Duration
}

// NewTimeoutTimer creates a stopped TimeoutTimer with duration d
func NewTimeoutTimer(d time.Duration) *TimeoutTimer {
	t := &TimeoutTimer{
		timer: time.NewTimer(0),
		d:     d,
	}
	t.Stop()
	return t
}

// Start starts the timer with deadline that is a duration of t.d in the future
func (t *TimeoutTimer) Start() {
	t.Stop()
	t.timer.Reset(t.d)
}

// Stop stops the timer, draining the underlying channel
func (t *TimeoutTimer) Stop() {
	if t.timer.Stop() {
		return
	}
	select {
	case <-t.timer.C:
	default:
	}
}

// Channel obtains a channel where the current time will be sent upon timeout
func (t *TimeoutTimer) Channel() <-chan time.Time {
	return t.timer.C
}
