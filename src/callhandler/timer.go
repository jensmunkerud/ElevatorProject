package callhandler

import (
	"time"
)

// Drains a given timer t, then reset (start) it with the duration d
func restartTimer(t *time.Timer, d time.Duration) {
	t.Stop()
	select {
	case <-t.C:
	default:
	}
	t.Reset(d)
}

// Stops and drains a given timer t
func stopTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}
