package callhandler

// In:	ElevatorEvent chan [controller], cost func chan [costfunc]
// Out:	ElevatorServer.RequestOrder(order data) [elevatorServer], controller.setMotorDirection [controller]

// When order is sent out to controller, start eg. a 10sec timer and if no ANYTYPE EVENT is received,
// restart the controller

import (
	"elevatorproject/src/config"
	"elevatorproject/src/controller"
	"elevatorproject/src/elevator"
	es "elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"elevatorproject/src/orders"
	"fmt"
	"time"
)

// RequestUpdateOrder sends order updates to the appropriate channel based on order type.
// orderCompleted is true for removing orders, false for adding new orders.
func RequestUpdateOrder(
	floor int,
	orderType es.OrderType,
	orderCompleted bool,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate) {

	if floor < 0 || floor >= config.NumFloors {
		return
	}

	myID := config.MyID()
	state := orders.UnconfirmedOrderState
	if orderCompleted {
		state = orders.CompletedOrderState
	}

	switch orderType {
	case es.Cab:
		if cabOrderUpdate != nil {
			cabOrderUpdate <- elevatorserver.NewCabOrderUpdate(myID, floor, state)
		}
	case es.HallUp, es.HallDown:
		if hallOrderUpdate != nil {
			hallOrderUpdate <- elevatorserver.NewHallOrderUpdate(myID, floor, orderType, state)
		}
	default:
		return
	}
}

// This launches the callhandler, which listens for events from the elevator
// and sends order updates to the elevatorserver.
// It also listens for new orders from order distributor and overwrites the old ones.
func Run(
	ready chan<- struct{},
	elevatorEvent <-chan es.ElevatorEvent,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
	elevatorStateUpdate chan<- es.Elevator,
	ordersOnNetwork <-chan elevatorserver.CallHandlerMessage,
	activeLocalOrders <-chan [config.NumFloors][config.NumButtons]bool) {

	doorTimer := time.NewTimer(config.DoorOpenDuration)
	stopTimer(doorTimer)
	serviceTimer := time.NewTimer(config.ServiceTimeout)
	stopTimer(serviceTimer)
	myID := config.MyID()
	localElevator := es.CreateElevator(myID, -1, es.Stop, es.Idle)
	elevators := make(map[string]*es.Elevator)
	elevators[localElevator.Id()] = localElevator
	initializeElevatorToValidFloor(localElevator)
	sendLocalState(localElevator, elevatorStateUpdate)
	close(ready)

	event := <-elevatorEvent
	go refreshElevatorLights(ordersOnNetwork)

	fmt.Println("Starting callhandler loop")
	for {
		select {
		case order := <-event.OrderEvent:
			fmt.Printf("%+v\n", order)
			RequestUpdateOrder(order.Floor, order.OrderType, false, cabOrderUpdate, hallOrderUpdate)

		case floor := <-event.FloorEvent:
			fmt.Printf("%+v\n", floor)
			elevatorArrivedAtFloor(localElevator, floor, doorTimer, serviceTimer, hallOrderUpdate, cabOrderUpdate)
			sendLocalState(localElevator, elevatorStateUpdate)

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
			elevatorOnNewOrders(localElevator, doorTimer, serviceTimer, hallOrderUpdate, cabOrderUpdate)
			sendLocalState(localElevator, elevatorStateUpdate)

		case <-doorTimer.C:
			elevatorOnDoorTimeout(localElevator, doorTimer, serviceTimer, hallOrderUpdate, cabOrderUpdate)
			sendLocalState(localElevator, elevatorStateUpdate)

		case <-serviceTimer.C:
			localElevator.UpdateInService(false)
			initializeElevatorToValidFloor(localElevator)
			restartTimer(serviceTimer, config.ServiceTimeout)
			sendLocalState(localElevator, elevatorStateUpdate)
		}
	}
}

// callHandlerMessageChanged returns true if previous and current differ in any hall or cab order state.
func callHandlerMessageChanged(previous, current elevatorserver.CallHandlerMessage) bool {
	hallPrevious, cabPrevious := previous.UnpackForCallHandler()
	hallCurrent, cabCurrent := current.UnpackForCallHandler()
	for atFloor := 0; atFloor < config.NumFloors; atFloor++ {
		for _, orderType := range elevator.HallOrderTypes {
			if hallPrevious.GetOrderState(atFloor, orderType) != hallCurrent.GetOrderState(atFloor, orderType) {
				return true
			}
		}
		if cabPrevious.GetOrderState(atFloor) != cabCurrent.GetOrderState(atFloor) {
			return true
		}
	}
	return false
}

// refreshElevatorLights updates the elevator lights if orders have changed.
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
				orderState := hallOrders.GetOrderState(floorIndex, button)
				controller.SetButtonLamp(button, floorIndex, orderState == orders.ConfirmedOrderState)
			}

			orderState := cabOrders.GetOrderState(floorIndex)
			controller.SetButtonLamp(es.Cab, floorIndex, orderState == orders.ConfirmedOrderState)
		}
	}
}

// handleActiveLocalOrders takes the most recent output from the order distributor and updates the local elevator's request state.
func handleActiveLocalOrders(
	localElevator *es.Elevator,
	activeLocalOrders <-chan [config.NumFloors][config.NumButtons]bool,
	elevatorStateLocal chan<- es.Elevator,
) {
	for newActiveOrders := range activeLocalOrders {
		localElevator.UpdateRequest(newActiveOrders)
		sendLocalState(localElevator, elevatorStateLocal)
	}
}

func sendLocalState(current *es.Elevator, elevatorStateLocal chan<- es.Elevator) {
	if elevatorStateLocal == nil || current == nil {
		return
	}
	elevatorStateLocal <- current.Copy()
}
