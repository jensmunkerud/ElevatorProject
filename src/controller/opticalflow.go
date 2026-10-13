package controller

import "sync"

var (
	opticalFlowPollerMu sync.RWMutex
	opticalFlowPoller   = func(chan<- float64) {}
)

func RegisterOpticalFlowPoller(poller func(chan<- float64)) {
	if poller == nil {
		return
	}
	opticalFlowPollerMu.Lock()
	opticalFlowPoller = poller
	opticalFlowPollerMu.Unlock()
}

func pollOpticalFlowSensor(output chan<- float64) {
	opticalFlowPollerMu.RLock()
	poller := opticalFlowPoller
	opticalFlowPollerMu.RUnlock()
	poller(output)
}
