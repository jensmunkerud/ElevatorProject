package callhandler

// In:	ElevatorEvent chan [controller], cost func chan [costfunc]
// Out:	ElevatorServer.RequestOrder(order data) [elevatorServer], controller.setMotorDirection [controller]

// When order is sent out to controller, start eg. a 10sec timer and if no ANYTYPE EVENT is received,
// restart the controller

import (
	"elevatorproject/src/config"
	"elevatorproject/src/controller"
	es "elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"elevatorproject/src/orders"
	"fmt"
	"time"
)

func RequestUpdateCabOrder(floor int, button es.ButtonType, completed bool, cabOrderUpdate chan<- elevatorserver.CabOrderUpdate) {
	myID := config.MyID()

	if cabOrderUpdate == nil || button != es.Cab {
		return
	}

	state := orders.UnconfirmedOrderState
	if completed {
		state = orders.CompletedOrderState
	}

	cabOrderUpdate <- elevatorserver.CabOrderUpdate{
		SenderID: myID,
		Floor:    floor,
		State:    state,
	}
}

func RequestUpdateHallOrder(floor int, button es.ButtonType, completed bool, hallOrderUpdate chan<- elevatorserver.HallOrderUpdate) {
	if hallOrderUpdate == nil || button == es.Cab {
		return
	}
	myID := config.MyID()

	state := orders.UnconfirmedOrderState
	if completed {
		state = orders.CompletedOrderState
	}

	hallOrderUpdate <- elevatorserver.HallOrderUpdate{
		SenderID:  myID,
		Floor:     floor,
		Direction: int(button),
		State:     state,
	}
}

func RunCallHandler(
	ready chan<- struct{},
	elevatorEvent <-chan es.ElevatorEvent,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
	elevatorStateLocal chan<- es.Elevator,
	callHandlerMessage <-chan elevatorserver.CallHandlerMessage, // FOR LIGHTS CONTROL
	activeLocalOrders <-chan [config.NumFloors][config.NumButtons]bool) {

	doorTimer := time.NewTimer(config.DoorOpenDuration)
	doorTimer.Stop()
	serviceTimer := time.NewTimer(config.ServiceTimeout)
	stopTimer(serviceTimer)
	myID := config.MyID()
	localElevator := es.CreateElevator(myID, -1, es.Stop, es.Idle)
	elevators := make(map[string]*es.Elevator)
	elevators[localElevator.Id()] = localElevator
	// updateElevatorState(localElevator)
	fsmInit(localElevator)
	emitLocalState(localElevator, elevatorStateLocal)
	close(ready)

	event := <-elevatorEvent
	go refreshElevatorLights(callHandlerMessage)

	// orderChan := make(chan [config.NumFloors][config.NumButtons]bool, 10)
	// var localOrders [config.NumFloors][config.NumButtons]bool

	fmt.Println("Starting callhandler loop")
	for {
		select {
		case order := <-event.OrderEvent:
			fmt.Printf("%+v\n", order)
			if order.Button == es.Cab {
				RequestUpdateCabOrder(order.Floor, order.Button, false, cabOrderUpdate)
			} else {
				RequestUpdateHallOrder(order.Floor, order.Button, false, hallOrderUpdate)
			}
			// Don't call fsmOnRequestButtonPress: the cost function (via activeLocalOrders)
			// is the sole authority for setting requests. Setting them directly here races
			// with UpdateRequestTotal and causes stops without lights.

		case floor := <-event.FloorEvent:
			fmt.Printf("%+v\n", floor)
			fsmOnFloorArrival(localElevator, floor, doorTimer, serviceTimer, hallOrderUpdate, cabOrderUpdate)
			emitLocalState(localElevator, elevatorStateLocal)

		case obstruction := <-event.ObstructionEvent:
			fmt.Printf("%+v\n", obstruction)
			localElevator.UpdateObstruction(obstruction)
			if localElevator.StopPressed() {
				controller.StopElevator()
			} else if localElevator.Behaviour() == es.Moving {
				switch localElevator.CurrentDirection() {
				case es.Up:
					controller.MoveElevatorUp()
				case es.Down:
					controller.MoveElevatorDown()
				}
			}

		case stop := <-event.StopEvent:
			fmt.Printf("%+v\n", stop)
			localElevator.UpdateStopPressed(stop)
			controller.SetStopLamp(stop)

			if localElevator.StopPressed() {
				controller.StopElevator()
			} else if localElevator.Behaviour() == es.Moving {
				switch localElevator.CurrentDirection() {
				case es.Up:
					controller.MoveElevatorUp()
				case es.Down:
					controller.MoveElevatorDown()
				}
			}

		case newOrders := <-activeLocalOrders:
			localElevator.UpdateRequestTotal(newOrders)
			fsmOnNewOrders(localElevator, doorTimer, serviceTimer, hallOrderUpdate, cabOrderUpdate)
			emitLocalState(localElevator, elevatorStateLocal)

		case <-doorTimer.C:
			fsmOnDoorTimeout(localElevator, doorTimer, serviceTimer, hallOrderUpdate, cabOrderUpdate)
			emitLocalState(localElevator, elevatorStateLocal)

		case <-serviceTimer.C:
			localElevator.UpdateInService(false)
			emitLocalState(localElevator, elevatorStateLocal)
		}
	}
}

// callHandlerMessageChanged returns true if a and b differ in any hall or cab order state.
func callHandlerMessageChanged(a, b elevatorserver.CallHandlerMessage) bool {
	hallA, cabA := a.UnpackForCallHandler()
	hallB, cabB := b.UnpackForCallHandler()
	for f := 0; f < config.NumFloors; f++ {
		for d := 0; d < 2; d++ {
			if hallA.GetOrderState(f, d) != hallB.GetOrderState(f, d) {
				return true
			}
		}
		if cabA.GetOrderState(f) != cabB.GetOrderState(f) {
			return true
		}
	}
	return false
}

// Repurpose this function to instead edit the localElevator.requests, then call setAllLights in fsm.go that
// serves the intended purpose of this function
func refreshElevatorLights(callHandlerMessage <-chan elevatorserver.CallHandlerMessage) {
	var last elevatorserver.CallHandlerMessage
	first := true
	for msg := range callHandlerMessage {
		if !first && !callHandlerMessageChanged(last, msg) {
			continue
		}
		first = false
		last = msg
		hallOrders, cabOrders := msg.UnpackForCallHandler()
		for floorIndex := range hallOrders.Orders {
			for button := es.HallUp; button <= es.HallDown; button++ {
				orderState := hallOrders.GetOrderState(floorIndex, int(button))
				controller.SetButtonLamp(button, floorIndex, orderState == orders.ConfirmedOrderState)
			}

			orderState := cabOrders.GetOrderState(floorIndex)
			controller.SetButtonLamp(es.Cab, floorIndex, orderState == orders.ConfirmedOrderState)
		}
	}
}

func handleActiveLocalOrders(
	localElevator *es.Elevator,
	activeLocalOrders <-chan [config.NumFloors][config.NumButtons]bool,
	elevatorStateLocal chan<- es.Elevator,
) {
	for newActiveOrders := range activeLocalOrders {
		localElevator.UpdateRequestTotal(newActiveOrders)
		emitLocalState(localElevator, elevatorStateLocal)
		fmt.Printf("RECEIVED FROM COSTFUNC")
	}
}

func emitLocalState(current *es.Elevator, elevatorStateLocal chan<- es.Elevator) {
	if elevatorStateLocal == nil || current == nil {
		return
	}
	elevatorStateLocal <- current.Copy()
}
