package callhandler

/*
In:		hardwareEvent chan						[controller]
In:		activeLocalOrders chan					[orderdistributor]
In:		ordersOnNetwork chan					[elevatorserver]
=================================================================
Out:	controller.MoveElevator					[controller]
Out:	elevatorStateUpdate						[elevatorserver]
out:	hallOrderUpdate chan					[elevatorserver]
out:	cabOrderUpdate chan						[elevatorserver]
*/

import (
	"elevatorproject/src/config"
	"elevatorproject/src/controller"
	"elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"elevatorproject/src/orders"
	"fmt"
	"time"
)

// requestUpdateOrder sends order updates to the appropriate channel based on order type.
// orderCompleted is true for removing orders, false for adding new orders.
func requestUpdateOrder(
	floor int,
	orderType elevator.OrderType,
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
	case elevator.Cab:
		if cabOrderUpdate != nil {
			cabOrderUpdate <- elevatorserver.CreateCabOrderUpdate(myID, floor, state)
		}
	case elevator.HallUp, elevator.HallDown:
		if hallOrderUpdate != nil {
			hallOrderUpdate <- elevatorserver.CreateHallOrderUpdate(myID, floor, orderType, state)
		}
	default:
		return
	}
}

// This launches the callhandler, which listens for events from the controller
// and sends order updates to the elevatorserver.
// It also listens for new orders from orderdistributor and overwrites the old ones.
func Run(
	ready chan<- struct{},
	hardwareEvent <-chan elevator.HardwareEvent,
	hallOrderUpdate chan<- elevatorserver.HallOrderUpdate,
	cabOrderUpdate chan<- elevatorserver.CabOrderUpdate,
	elevatorStateUpdate chan<- elevator.Elevator,
	ordersOnNetwork <-chan elevatorserver.CallHandlerMessage,
	activeLocalOrders <-chan [config.NumFloors][config.NumButtons]bool) {

	doorTimer := time.NewTimer(config.DoorOpenDuration)
	stopTimer(doorTimer)
	serviceTimer := time.NewTimer(config.ServiceTimeout)
	stopTimer(serviceTimer)
	myID := config.MyID()
	localElevator := elevator.CreateElevator(myID, -1, elevator.Stop, elevator.Idle)
	elevators := make(map[string]*elevator.Elevator)
	elevators[localElevator.Id()] = localElevator
	initializeElevatorToValidFloor(localElevator)
	sendLocalState(localElevator, elevatorStateUpdate)

	event := <-hardwareEvent
	go refreshElevatorLights(ordersOnNetwork)
	close(ready)
	fmt.Println("Starting callhandler loop")
	for {
		select {
		case order := <-event.OrderEvent:
			fmt.Printf("%+v\n", order)
			requestUpdateOrder(order.Floor, order.OrderEvent, false, cabOrderUpdate, hallOrderUpdate)

		case floor := <-event.FloorEvent:
			fmt.Printf("%+v\n", floor)
			elevatorArrivedAtFloor(localElevator, floor, doorTimer, serviceTimer, hallOrderUpdate, cabOrderUpdate)
			sendLocalState(localElevator, elevatorStateUpdate)

		case obstruction := <-event.ObstructionEvent:
			fmt.Printf("%+v\n", obstruction)
			localElevator.UpdateObstruction(obstruction)
			if localElevator.StopPressed() {
				controller.StopElevator()
			} else if localElevator.Behaviour() == elevator.Moving {
				switch localElevator.CurrentDirection() {
				case elevator.Up:
					controller.MoveElevatorUp()
				case elevator.Down:
					controller.MoveElevatorDown()
				}
			}

		case stop := <-event.StopEvent:
			fmt.Printf("%+v\n", stop)
			localElevator.UpdateStopPressed(stop)
			controller.SetStopLamp(stop)

			if localElevator.StopPressed() {
				controller.StopElevator()
			} else if localElevator.Behaviour() == elevator.Moving {
				switch localElevator.CurrentDirection() {
				case elevator.Up:
					controller.MoveElevatorUp()
				case elevator.Down:
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
		for floorIndex := 0; floorIndex < config.NumFloors; floorIndex++ {
			for button := elevator.HallUp; button <= elevator.HallDown; button++ {
				orderState := hallOrders.GetOrderState(floorIndex, button)
				controller.SetButtonLamp(button, floorIndex, orderState == orders.ConfirmedOrderState)
			}

			orderState := cabOrders.GetOrderState(floorIndex)
			controller.SetButtonLamp(elevator.Cab, floorIndex, orderState == orders.ConfirmedOrderState)
		}
	}
}

// handleActiveLocalOrders takes the most recent output from the order distributor and updates the local elevator's request state.
func handleActiveLocalOrders(
	localElevator *elevator.Elevator,
	activeLocalOrders <-chan [config.NumFloors][config.NumButtons]bool,
	elevatorStateLocal chan<- elevator.Elevator,
) {
	for newActiveOrders := range activeLocalOrders {
		localElevator.UpdateRequest(newActiveOrders)
		sendLocalState(localElevator, elevatorStateLocal)
	}
}

func sendLocalState(current *elevator.Elevator, elevatorStateLocal chan<- elevator.Elevator) {
	if elevatorStateLocal == nil || current == nil {
		return
	}
	elevatorStateLocal <- current.Copy()
}
