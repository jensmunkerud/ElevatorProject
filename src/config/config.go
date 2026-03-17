package config

import (
	"fmt"
	"net"
	"sync"
	"time"
	"os"
	"strconv"
)

const (
	NumFloors         = 4
	NumButtons        = 3
	NumElevators      = 3
	DoorOpenDuration  = 3000 * time.Millisecond
	Timeout           = 1000 * time.Millisecond
	HeartbeatInterval = 100 * time.Millisecond
	PeersPort = 15647
	BroadcastPort = 16569
	testing = true
)

var (
	// ElevatorIOPort is the address for elevator I/O. In testing mode it is set from user input in SetMyID.
	ElevatorIOPort = "localhost:15658"
	myID           = "placeholder"
	once           sync.Once
)

// MyID returns the local elevator ID (set once from MAC address or via SetMyID).
func MyID() string {
	return myID
}

// SetMyID sets the local ID once. If id is empty, myID is set from the machine's MAC address.
// Only the first call has effect. The application should call SetMyID("") at startup to use MAC-based ID.
func SetMyID() {
	SetMyIDFromPort(0)
}

// SetMyIDFromPort sets the local elevator ID once. If port > 0, myID is set to the port string
// so that the same elevator (same simulator port) keeps a stable ID across restarts. That allows
// other nodes to retain and re-send this elevator's cab orders when it rejoins the network.
// If port is 0, behaviour is the same as SetMyID() (PID when testing, MAC otherwise).
func SetMyIDFromPort(port int) {
	once.Do(func() {
		if port > 0 {
			myID = strconv.Itoa(port)
			return
		}
		if testing {
			myID = strconv.Itoa(os.Getpid())
			fmt.Print("Elevator IO port (e.g. 15658): ")
			var port string
			fmt.Scanln(&port)
			if port != "" {
				ElevatorIOPort = "localhost:" + port
			}
		} else {
			mac, err := getMacAddr()
			attempts := 0
			for err != nil {
				fmt.Printf("Error: %v, retrying MAC address...\n", err)
				mac, err = getMacAddr() // keep "placeholder" on error
				time.Sleep(1 * time.Second)
				attempts++
				if attempts > 10 {
					panic("Failed to find MAC address after 10 attempts")
				}
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


