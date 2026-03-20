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
	elevatorStateUpdate chan<- es.Elevator,
	ordersOnNetwork <-chan elevatorserver.CallHandlerMessage, // FOR LIGHTS CONTROL
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
	emitLocalState(localElevator, elevatorStateUpdate)
	close(ready)

	event := <-elevatorEvent
	go refreshElevatorLights(ordersOnNetwork)

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
			emitLocalState(localElevator, elevatorStateUpdate)

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
			emitLocalState(localElevator, elevatorStateUpdate)

		case <-doorTimer.C:
			fsmOnDoorTimeout(localElevator, doorTimer, serviceTimer, hallOrderUpdate, cabOrderUpdate)
			emitLocalState(localElevator, elevatorStateUpdate)

		case <-serviceTimer.C:
			localElevator.UpdateInService(false)
			fsmInit(localElevator)
			restartTimer(serviceTimer, config.ServiceTimeout)
			emitLocalState(localElevator, elevatorStateUpdate)
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

// Overwrites all buttonlamps of elevator if callHandlerMessage has new, different data.
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
		localElevator.UpdateRequest(newActiveOrders)
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
