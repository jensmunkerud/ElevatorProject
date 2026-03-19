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

// Adds a new cab order update locally. orderCompleted is true for removing orders, false for adding new orders.
func RequestUpdateCabOrder(floor int, orderType es.ButtonType, orderCompleted bool, cabOrderUpdate chan<- elevatorserver.CabOrderUpdate) {
	myID := config.MyID()

	if cabOrderUpdate == nil || orderType != es.Cab {
		return
	}

	state := orders.UnconfirmedOrderState
	if orderCompleted {
		state = orders.CompletedOrderState
	}

	cabOrderUpdate <- elevatorserver.CabOrderUpdate{
		SenderID: myID,
		Floor:    floor,
		State:    state,
	}
}

// Adds a new hall order update locally. orderCompleted is true for removing orders, false for adding new orders.
func RequestUpdateHallOrder(floor int, orderType es.ButtonType, orderCompleted bool, hallOrderUpdate chan<- elevatorserver.HallOrderUpdate) {
	if hallOrderUpdate == nil || orderType == es.Cab {
		return
	}
	myID := config.MyID()

	state := orders.UnconfirmedOrderState
	if orderCompleted {
		state = orders.CompletedOrderState
	}

	hallOrderUpdate <- elevatorserver.HallOrderUpdate{
		SenderID:  myID,
		Floor:     floor,
		Direction: int(orderType),
		State:     state,
	}
}

// This launches the callhandler, which listens for events from the elevator
// and sends order updates to the elevatorserver.
// It also listens for new orders from the cost function and overwrites the old ones.
func Run(
	ready chan<- struct{},
	elevatorEvent <-chan es.ElevatorEvent,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
	elevatorStateLocal chan<- es.Elevator,
	callHandlerMessage <-chan elevatorserver.CallHandlerMessage, // FOR LIGHTS CONTROL
	activeLocalOrders <-chan [config.NumFloors][config.NumButtons]bool) {

	doorTimer := time.NewTimer(config.DoorOpenDuration)
	stopTimer(doorTimer)
	serviceTimer := time.NewTimer(config.ServiceTimeout)
	stopTimer(serviceTimer)
	myID := config.MyID()
	localElevator := es.CreateElevator(myID, -1, es.Stop, es.Idle)
	elevators := make(map[string]*es.Elevator)
	elevators[localElevator.Id()] = localElevator
	fsmInit(localElevator)
	emitLocalState(localElevator, elevatorStateLocal)
	close(ready)

	event := <-elevatorEvent
	go refreshElevatorLights(callHandlerMessage)

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
			localElevator.UpdateRequest(newOrders)
			fsmOnNewOrders(localElevator, doorTimer, serviceTimer, hallOrderUpdate, cabOrderUpdate)
			emitLocalState(localElevator, elevatorStateLocal)

		case <-doorTimer.C:
			fsmOnDoorTimeout(localElevator, doorTimer, serviceTimer, hallOrderUpdate, cabOrderUpdate)
			emitLocalState(localElevator, elevatorStateLocal)

		case <-serviceTimer.C:
			localElevator.UpdateInService(false)
			fsmInit(localElevator)
			restartTimer(serviceTimer, config.ServiceTimeout)
			emitLocalState(localElevator, elevatorStateLocal)
		}
	}
}

// callHandlerMessageChanged returns true if previous and current differ in any hall or cab order state.
func callHandlerMessageChanged(previous, current elevatorserver.CallHandlerMessage) bool {
	hallPrevious, cabPrevious := previous.UnpackForCallHandler()
	hallCurrent, cabCurrent := current.UnpackForCallHandler()
	for atFloor := range config.NumFloors {
		for direction := 0; direction < 2; direction++ {
			if hallPrevious.GetOrderState(atFloor, direction) != hallCurrent.GetOrderState(atFloor, direction) {
				return true
			}
		}
		if cabPrevious.GetOrderState(atFloor) != cabCurrent.GetOrderState(atFloor) {
			return true
		}
	}
	return false
}

// Overwrites all buttonlamps of elevator if callHandlerMessage has new, different data.
func refreshElevatorLights(callHandlerMessage <-chan elevatorserver.CallHandlerMessage) {
	var lastMessage elevatorserver.CallHandlerMessage
	first := true
	for msg := range callHandlerMessage {
		if !first && !callHandlerMessageChanged(lastMessage, msg) {
			continue
		}
		first = false
		lastMessage = msg
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
		localElevator.UpdateRequest(newActiveOrders)
		emitLocalState(localElevator, elevatorStateLocal)
		fmt.Printf("RECEIVED FROM COSTFUNC\n")
	}
}

func emitLocalState(current *es.Elevator, elevatorStateLocal chan<- es.Elevator) {
	if elevatorStateLocal == nil || current == nil {
		return
	}
	elevatorStateLocal <- current.Copy()
}
