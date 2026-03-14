package callhandler

import (
	"driver-go/elevio"
	"elevatorproject/src/config"
	"elevatorproject/src/elevatorserver"
	"elevatorproject/src/orders"
	"fmt"
	"testing"
	"time"
)

func TestCallHandler(t *testing.T) {
	go InitCallHandler()
	select {}
}

// TestSetElevatorLightsBlinkInHardware drives setElevatorLights with artificial
// CallHandlerMessages. It prints to the terminal which light should be on for
// one-second intervals so the physical panel can be inspected manually while
// elevatorserver is running.
func TestSetElevatorLightsBlinkInHardware(t *testing.T) {
	go InitCallHandler()
	time.Sleep(1 * time.Second)
	hall := orders.CreateHallOrders()
	cab := orders.CreateCabOrders()

	buttonNames := map[elevio.ButtonType]string{
		elevio.BT_HallUp:   "HallUp",
		elevio.BT_HallDown: "HallDown",
		elevio.BT_Cab:      "Cab",
	}

	for floor := 0; floor < config.NumFloors; floor++ {
		for _, btn := range []elevio.ButtonType{elevio.BT_HallUp, elevio.BT_HallDown, elevio.BT_Cab} {
			// Clear all orders before setting a single light.
			for f := 0; f < config.NumFloors; f++ {
				for dir := 0; dir < 2; dir++ {
					hall.UpdateOrderState(f, dir, orders.RemovedOrderState)
				}
				cab.UpdateOrderState(f, orders.RemovedOrderState)
			}

			// Set the specific light we want to blink.
			switch btn {
			case elevio.BT_HallUp:
				hall.UpdateOrderState(floor, 0, orders.ConfirmedOrderState)
			case elevio.BT_HallDown:
				hall.UpdateOrderState(floor, 1, orders.ConfirmedOrderState)
			case elevio.BT_Cab:
				cab.UpdateOrderState(floor, orders.ConfirmedOrderState)
			}

			msgOn := elevatorserver.NewCallHandlerMessage(hall, cab)

			fmt.Printf("Turning ON %s at floor %d for 1s\n", buttonNames[btn], floor)
			setElevatorLights(msgOn)
			time.Sleep(1 * time.Second)

			// Clear again (turn everything off).
			for f := 0; f < config.NumFloors; f++ {
				for dir := 0; dir < 2; dir++ {
					hall.UpdateOrderState(f, dir, orders.RemovedOrderState)
				}
				cab.UpdateOrderState(f, orders.RemovedOrderState)
			}
			msgOff := elevatorserver.NewCallHandlerMessage(hall, cab)

			fmt.Printf("Turning OFF %s at floor %d\n", buttonNames[btn], floor)
			setElevatorLights(msgOff)

			// Small pause between lamps to make it easier to see transitions.
			time.Sleep(200 * time.Millisecond)
		}
	}

	fmt.Println("Finished blinking all hall and cab lights on all floors.")
}
