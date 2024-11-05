package gbn

import "time"

type TimeoutTimer struct {
	timer *time.Timer
	d     time.Duration
}

func NewTimeoutTimer(d time.Duration) *TimeoutTimer {
	t := &TimeoutTimer{
		timer: time.NewTimer(0),
		d:     d,
	}
	t.Stop()
	return t
}

func (t *TimeoutTimer) Channel() <-chan time.Time {
	return t.timer.C
}

func (t *TimeoutTimer) Start() {
	// clean stop
	t.Stop()
	// restart with new deadline
	t.timer.Reset(t.d)
}

func (t *TimeoutTimer) Stop() {
	if t.timer.Stop() {
		return
	}
	// drain channel if timer fired already
	select {
	case <-t.timer.C:
	default:
	}
}
