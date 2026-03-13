package networking

import (
	"fmt"
	"testing"
	"time"

	"Network-go/network/bcast"
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
)

func TestMessageFromOrders_RoundTrip(t *testing.T) {
	hall := orders.CreateHallOrders()
	hall.UpdateOrderState(0, 0, orders.UnconfirmedOrderState)
	hall.UpdateOrderState(2, 1, orders.ConfirmedOrderState)

	cab := orders.CreateCabOrders()
	cab.UpdateOrderState(1, orders.ConfirmedOrderState)
	allCab := map[string]*orders.CabOrders{"elev-1": cab}

	localElev := *elevator.CreateElevator("elev-1", 2, elevator.Up, elevator.Moving)
	msg := messageFromOrders("elev-1", hall, allCab, localElev)

	if msg.SenderID != "elev-1" {
		t.Errorf("SenderID: got %q, want %q", msg.SenderID, "elev-1")
	}
	if got := orders.OrderState(msg.HallOrders[0][0]); got != orders.UnconfirmedOrderState {
		t.Errorf("HallOrders[0][0]: got %v, want %v", got, orders.UnconfirmedOrderState)
	}
	if got := orders.OrderState(msg.HallOrders[2][1]); got != orders.ConfirmedOrderState {
		t.Errorf("HallOrders[2][1]: got %v, want %v", got, orders.ConfirmedOrderState)
	}
	if got := orders.OrderState(msg.AllCabOrders["elev-1"][1]); got != orders.ConfirmedOrderState {
		t.Errorf("AllCabOrders[elev-1][1]: got %v, want %v", got, orders.ConfirmedOrderState)
	}
	state := msg.ElevatorStates["elev-1"]
	if state.Behaviour != "moving" {
		t.Errorf("ElevatorStates behaviour: got %q, want %q", state.Behaviour, "moving")
	}
	if state.Floor != 2 {
		t.Errorf("ElevatorStates floor: got %d, want %d", state.Floor, 2)
	}
	if state.Direction != "up" {
		t.Errorf("ElevatorStates direction: got %q, want %q", state.Direction, "up")
	}
}

// TestBroadcastAndReceiveManual is an integration-style test intended to be run
// on two machines on the same network. It sends a Message on the UDP broadcast
// channel and prints any received Messages to the terminal.
func TestBroadcastAndReceiveManual(t *testing.T) {
	sendCh := make(chan Message)
	recvCh := make(chan Message)

	// Start broadcaster and receiver on the same port as networking.RunNetworking.
	go bcast.Transmitter(16569, sendCh)
	go bcast.Receiver(16569, recvCh)

	// Build a simple message with a recognizable SenderID and no orders.
	msg := Message{
		SenderID: fmt.Sprintf("network-test-%d", time.Now().UnixNano()),
	}

	fmt.Printf("Networking test: sending message with SenderID=%s\n", msg.SenderID)
	sendCh <- msg

	// Listen for a short period and print anything we receive.
	timeout := time.After(10 * time.Second)
	for {
		select {
		case m := <-recvCh:
			fmt.Printf("Networking test: received message from SenderID=%s: %+v\n", m.SenderID, m)
		case <-timeout:
			fmt.Println("Networking test: timeout reached, ending test")
			return
		}
	}
}
