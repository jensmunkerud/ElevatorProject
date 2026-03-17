package callhandler

import (
	"elevatorproject/src/config"
	es "elevatorproject/src/elevator"
	"time"
)

func resetTimer(t *time.Timer) {
	t.Stop()
	select {
	case <-t.C:
	default:
	}
	t.Reset(config.ServiceTimeout)
}

func stopTimer(t *time.Timer) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}
}

func syncServiceWatchdog(e *es.Elevator, watchdog *time.Timer) {
	if e.Behaviour() == es.Moving {
		resetTimer(watchdog)
		return
	}
	stopTimer(watchdog)
}
