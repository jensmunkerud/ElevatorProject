package config

import (
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	NumFloors  = 4
	NumButtons = 3

	DoorOpenDuration = 3000 * time.Millisecond
	TravelDuration   = 2500 * time.Millisecond
	ServiceTimeout   = 4000 * time.Millisecond

	HeartbeatInterval       = 500 * time.Millisecond
	ProcessPairRestartDelay = 10 * time.Second
	// Decouple the cost-function from heartbeat to avoid spamming
	CostFuncInterval = 200 * time.Millisecond

	ElevatorPort  = 15657
	PeersPort     = 52317
	BroadcastPort = 53491
)

var (
	myID = ""
	once sync.Once
)

func MyID() string {
	if myID == "" {
		panic("Config not initialized: call InitConfig before using MyID")
	}
	return myID
}

// InitConfig sets the local elevator ID once.
// In simulator mode: ID = port.
// In production mode: ID = MAC address.
func InitConfig(port int, simulatorMode bool) {
	once.Do(func() {
		if simulatorMode {
			myID = strconv.Itoa(port)
		} else {
			mac, err := getMacAddr()
			for attempt := 1; err != nil; attempt++ {
				if attempt > 10 {
					panic("Failed to find MAC address after 10 attempts")
				}
				fmt.Printf("Error: %v, retrying MAC address...\n", err)
				time.Sleep(1 * time.Second)
				mac, err = getMacAddr()
			}
			myID = mac
		}
	})
}

func getMacAddr() (string, error) {
	ifas, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, ifa := range ifas {
		a := ifa.HardwareAddr.String()
		if a != "" {
			return a, nil
		}
	}
	return "", fmt.Errorf("no MAC address found")
}
