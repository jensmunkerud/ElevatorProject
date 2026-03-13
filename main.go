package main



func main() {
	// Initialize config, get my id, then lock to read only.
	// Lage channels for communication between goroutines:

	// HallOrderUpdate := make(chan orders.HallOrders)
	// CabOrderUpdate := make(chan map[string]orders.CabOrders)
	// ElevatorStateLocal := make(chan elevator.Elevator)
	// CurrentOrders := make(chan elevatorserver.CallHandlerMessage)

	// PeerUpdate := make(chan []string) MAYBE?
	// WorldviewToCostFunction:= make(chan elevatorserver.OrderDistributorMessage)
	// SendWorldviewToNetwork:= make(chan elevatorserver.NetworkingDistributorMessage)
	// ReceiveWorldviewFromNetwork := make(chan elevatorserver.NetworkingDistributorMessage)

	// ActiveOrders := make(chan [][]bool)

	// Start goroutines:
	// go controller(HallOrderUpdate, CabOrderUpdate, ElevatorStateLocal)
	// go callHandler(HallOrderUpdate, CabOrderUpdate, ElevatorStateLocal, CurrentOrders)
	// go ElevatorServer(HallOrderUpdate, CabOrderUpdate, ElevatorStateLocal, WorldviewToCostFunction, SendWorldviewToNetwork, ReceiveWorldviewFromNetwork, CurrentOrders)
	// go RunNetworking(SendWorldviewToNetwork, ReceiveWorldviewFromNetwork, PeerUpdate)
	// go costFunction(WorldviewToCostFunction, ActiveOrders)
}