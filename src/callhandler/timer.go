package callhandler

import (
	"elevatorproject/src/config"
	es "elevatorproject/src/elevator"
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

// Restarts timer t elevator is moving, else stop timer
func syncServiceWatchdog(e *es.Elevator, watchdog *time.Timer) {
	if e.Behaviour() == es.Moving {
		restartTimer(watchdog, config.ServiceTimeout)
		return
	}
	stopTimer(watchdog)
}
