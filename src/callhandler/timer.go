package callhandler

import (
	"time"
)

func restartTimer(timer *time.Timer, duration time.Duration) {
	timer.Stop()
	select {
	case <-timer.C:
	default:
	}
	timer.Reset(duration)
}

func stopTimer(timer *time.Timer) {
	if !timer.Stop() {
		select {
		case <-timer.C:
		default:
		}
	}
}
