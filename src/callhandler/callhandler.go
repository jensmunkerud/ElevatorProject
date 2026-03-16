package callhandler

// In:	ElevatorEvent chan [controller], cost func chan [costfunc]
// Out:	ElevatorServer.RequestOrder(order data) [elevatorServer], controller.setMotorDirection [controller]

// When order is sent out to controller, start eg. a 10sec timer and if no ANYTYPE EVENT is received,
// restart the controller

import (
	"driver-go/elevio"
	"elevatorproject/src/config"
	es "elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"elevatorproject/src/orders"
	"fmt"
	"time"
)

func RequestUpdateCabOrder(floor int, button es.ButtonType, completed bool, cabOrderUpdate chan elevatorserver.CabOrderUpdate) {
	if cabOrderUpdate == nil || button != es.Cab {
		return
	}

	state := orders.UnconfirmedOrderState
	if completed {
		state = orders.CompletedOrderState
	}

	cabOrderUpdate <- elevatorserver.CabOrderUpdate{
		SenderID: config.MyID(),
		Floor:    floor,
		State:    state,
	}
}

func RequestUpdateHallOrder(floor int, button es.ButtonType, completed bool, hallOrderUpdate chan elevatorserver.HallOrderUpdate) {
	if hallOrderUpdate == nil || button == es.Cab {
		return
	}

	state := orders.UnconfirmedOrderState
	if completed {
		state = orders.CompletedOrderState
	}

	hallOrderUpdate <- elevatorserver.HallOrderUpdate{
		SenderID:  config.MyID(),
		Floor:     floor,
		Direction: int(button),
		State:     state,
	}
}

func RunCallHandler(
	ready chan struct{},
	elevatorEvent chan es.ElevatorEvent,
	hallOrderUpdate chan elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan elevatorserver.CabOrderUpdate,
	elevatorStateLocal chan es.Elevator,
	callHandlerMessage chan elevatorserver.CallHandlerMessage, // FOR LIGHTS CONTROL
	activeLocalOrders chan [config.NumFloors][config.NumButtons]bool) {
	// hallOrderUpdateOut = hallOrderUpdate
	// cabOrderUpdateOut = cabOrderUpdate

	doorTimer := time.NewTimer(config.DoorOpenDuration)
	doorTimer.Stop()

	localElevator := es.CreateElevator(config.MyID(), -1, es.Stop, es.Idle)
	elevators := make(map[string]*es.Elevator)
	elevators[localElevator.Id()] = localElevator
	// updateElevatorState(localElevator)
	fsmOnInitBetweenFloors(localElevator)
	emitLocalState(localElevator, elevatorStateLocal)
	close(ready)

	event := <-elevatorEvent
	go refreshElevatorLights(callHandlerMessage)

	// orderChan := make(chan [config.NumFloors][config.NumButtons]bool, 10)
	// var localOrders [config.NumFloors][config.NumButtons]bool

	for {
		select {
		case order := <-event.OrderEvent:
			fmt.Printf("%+v\n", order)
			if order.Button == es.Cab {
				RequestUpdateCabOrder(order.Floor, order.Button, false, cabOrderUpdate)
			} else {
				RequestUpdateHallOrder(order.Floor, order.Button, false, hallOrderUpdate)
			}
			fsmOnRequestButtonPress(localElevator, order.Floor, order.Button, doorTimer, hallOrderUpdate, cabOrderUpdate)
			emitLocalState(localElevator, elevatorStateLocal)

		case floor := <-event.FloorEvent:
			fmt.Printf("%+v\n", floor)
			fsmOnFloorArrival(localElevator, floor, doorTimer, hallOrderUpdate, cabOrderUpdate)
			emitLocalState(localElevator, elevatorStateLocal)

		case obstruction := <-event.ObstructionEvent:
			fmt.Printf("%+v\n", obstruction)
			localElevator.UpdateObstruction(obstruction)
			if localElevator.StopPressed() {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else if localElevator.Behaviour() == es.Moving {
				elevio.SetMotorDirection(elevio.MotorDirection(localElevator.CurrentDirection()))
			}

		case stop := <-event.StopEvent:
			fmt.Printf("%+v\n", stop)
			localElevator.UpdateStopPressed(stop)
			elevio.SetStopLamp(stop)

			if localElevator.StopPressed() {
				elevio.SetMotorDirection(elevio.MD_Stop)
			} else if localElevator.Behaviour() == es.Moving {
				elevio.SetMotorDirection(elevio.MotorDirection(localElevator.CurrentDirection()))
			}

		// case newOrders := <-orderChan:
		// break
		case <-doorTimer.C:
			fsmOnDoorTimeout(localElevator, doorTimer, hallOrderUpdate, cabOrderUpdate)
			emitLocalState(localElevator, elevatorStateLocal)

		case newActiveOrders := <-activeLocalOrders:
			localElevator.UpdateRequestTotal(newActiveOrders)
			emitLocalState(localElevator, elevatorStateLocal)
		}
	}
}

// Repurpose this function to instead edit the localElevator.requests, then call setAllLights in fsm.go that
// serves the intended purpose of this function
func refreshElevatorLights(callHandlerMessage chan elevatorserver.CallHandlerMessage) {
	select {
	case msg := <-callHandlerMessage:
		hallOrders, cabOrders := msg.UnpackForCallHandler()
		for floorIndex := range hallOrders.Orders {
			for b := range []elevio.ButtonType{elevio.BT_HallUp, elevio.BT_HallDown} { // b = 0 (hall up), b = 1 (hall down)
				orderState := hallOrders.GetOrderState(floorIndex, b)
				if orderState == orders.ConfirmedOrderState || orderState == orders.CompletedOrderState {
					elevio.SetButtonLamp(elevio.ButtonType(b), floorIndex, true)
				} else {
					elevio.SetButtonLamp(elevio.ButtonType(b), floorIndex, false)
				}
			}
			// Assuming one cab button per floor (b = 0)
			orderState := cabOrders.GetOrderState(floorIndex)
			if orderState == orders.ConfirmedOrderState || orderState == orders.CompletedOrderState {
				elevio.SetButtonLamp(elevio.BT_Cab, floorIndex, true)
			} else {
				elevio.SetButtonLamp(elevio.BT_Cab, floorIndex, false)
			}
		}
	}
}

func handleActiveOrdersFromOrderDistributor(e *es.Elevator, orders <-chan [][]bool) {
	go func() {
		for newOrder := range orders {
			e.UpdateActiveOrder(newOrder)
		}
	}()
}

func emitLocalState(current *es.Elevator, elevatorStateLocal chan es.Elevator) {
	if elevatorStateLocal == nil || current == nil {
		return
	}
	elevatorStateLocal <- current.Copy()
}
