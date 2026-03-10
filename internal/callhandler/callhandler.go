package callhandler

// In:	ElevatorEvent chan [controller], cost func chan [costfunc]
// Out:	ElevatorServer.RequestOrder(order data) [elevatorServer], controller.setMotorDirection [controller]

// When order is sent out to controller, start eg. a 10sec timer and if no ANYTYPE EVENT is received,
// restart the controller

import (
	"driver-go/elevio"
	"elevatorproject/internal/config"
	controller "elevatorproject/internal/controller"
	es "elevatorproject/internal/elevatorstruct"
	"fmt"
	"net"
	"testing"
)

func TestCallHandler(t *testing.T) {
	go InitCallHandler()
	select {}
}

func InitCallHandler() {
	ready := make(chan struct{})
	c, floor := controller.InitController(ready)
	<-ready

	id, err := getMacAddr()
	if err != nil {
		fmt.Printf("Error finding MAC address")
		return
	}

	localElevator := es.CreateElevator(id, floor, es.Stop, es.Idle)
	elevators := make(map[string]*es.Elevator)
	elevators[localElevator.Id()] = localElevator
	updateElevatorState(localElevator)

	orderChan := make(chan es.ElevatorButtons, 10)
	var localOrders es.ElevatorButtons

	for {
		select {
		case order := <-c.OrderEvent:
			fmt.Printf("%+v\n", order)
			// ElevatorServer.RequestOrder(order data)
			// -> Requests elevatorserver to actually create (or not) a new order,
			// callHandler does not have this authority.
			switch order.Button {
			case elevio.BT_HallUp:
				localOrders.Buttons[order.Floor][0] = true

			case elevio.BT_HallDown:
				localOrders.Buttons[order.Floor][1] = true
			}

			orderChan <- localOrders
			break

		case floor := <-c.FloorEvent:
			fmt.Printf("%+v\n", floor)
			localElevator.UpdateCurrentFloor(floor)
			// CHECK IF ORDER AT FLOOR
			if shouldStop(localElevator, localOrders) {
				clearOrdersAtFloor(floor, &localOrders)
				localElevator.UpdateBehaviour(es.DoorOpen)
			}

			updateElevatorState(localElevator)

		case obstruction := <-c.ObstructionEvent:
			fmt.Printf("%+v\n", obstruction)
			if obstruction {
				localElevator.UpdateBehaviour(es.Idle)
				// localElevator.UpdateCurrentDirection(es.Stop)
			} else {
				localElevator.UpdateBehaviour(es.Moving) // Possibly dangerous?
			}
			updateElevatorState(localElevator)
			break

		case stop := <-c.StopEvent:
			fmt.Printf("%+v\n", stop)
			if stop {
				localElevator.UpdateBehaviour(es.Idle)
			} else {
				localElevator.UpdateBehaviour(es.Moving)
			}
			// localElevator.UpdateCurrentDirection(es.Stop)
			updateElevatorState(localElevator)
			break

		case newOrders := <-orderChan:
			updateElevatorFromOrders(localElevator, newOrders)
			updateElevatorState(localElevator)
		}
	}
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

func updateElevatorState(localElevator *es.Elevator) {
	switch localElevator.Behaviour() {
	case es.Idle:
		elevio.SetDoorOpenLamp(false)
		elevio.SetMotorDirection(elevio.MD_Stop)
		break
	case es.Moving:
		elevio.SetDoorOpenLamp(false)

		switch localElevator.CurrentDirection() {
		case es.Stop:
			elevio.SetMotorDirection(elevio.MD_Stop)
			break
		case es.Up:
			elevio.SetMotorDirection(elevio.MD_Up)
			break
		case es.Down:
			elevio.SetMotorDirection(elevio.MD_Down)
			break
		}

		break
	case es.DoorOpen:
		elevio.SetDoorOpenLamp(true)
		elevio.SetMotorDirection(elevio.MD_Stop)
	}
}

func getLocalOrders(e *es.Elevator, orders es.ElevatorButtons) es.ElevatorButtons {
	return orders
}

func clearOrdersAtFloor(floor int, orders *es.ElevatorButtons) {
	orders.Buttons[floor][0] = false
	orders.Buttons[floor][1] = false
}

func shouldStop(e *es.Elevator, buttons es.ElevatorButtons) bool {
	f := e.CurrentFloor()

	return buttons.Buttons[f][0] || buttons.Buttons[f][1]
}

func updateElevatorFromOrders(
	e *es.Elevator,
	orders es.ElevatorButtons,
) {

	buttons := getLocalOrders(e, orders)

	if shouldStop(e, buttons) {
		e.UpdateBehaviour(es.DoorOpen)
		return
	}

	dir := chooseDirection(e, buttons)

	e.UpdateCurrentDirection(dir)

	if dir == es.Stop {
		e.UpdateBehaviour(es.Idle)
	} else {
		e.UpdateBehaviour(es.Moving)
	}
	updateElevatorState(e)
}

func chooseDirection(e *es.Elevator, buttons es.ElevatorButtons) es.Direction {

	current := e.CurrentFloor()

	for f := current + 1; f < config.NumFloors; f++ {
		if buttons.Buttons[f][0] || buttons.Buttons[f][1] {
			return es.Up
		}
	}

	for f := current - 1; f >= 0; f-- {
		if buttons.Buttons[f][0] || buttons.Buttons[f][1] {
			return es.Down
		}
	}

	return es.Stop
}
