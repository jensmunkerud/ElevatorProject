package main



func main() {
	// Initialize config, get my id, then lock to read only.
	// Channels for communication between goroutines.
	// Input channels:
	// ElevatorStateLocal := make(chan elevator.Elevator)
	// CurrentOrders := make(chan elevatorserver.CallHandlerMessage)
	// SendWorldviewToNetwork := make(chan elevatorserver.NetworkingDistributorMessage)

	// PeerUpdate := make(chan []string) MAYBE?
	// WorldviewToOrderDistributor:= make(chan elevatorserver.OrderDistributorMessage)
	// SendWorldviewToNetwork:= make(chan elevatorserver.NetworkingDistributorMessage)
	// ReceiveWorldviewFromNetwork := make(chan elevatorserver.NetworkingDistributorMessage)

	// ActiveOrders := make(chan [][]bool)

	// Start goroutines:
	// go controller(HallOrderUpdate, CabOrderUpdate, ElevatorStateLocal)
	// go callHandler(HallOrderUpdate, CabOrderUpdate, ElevatorStateLocal, CurrentOrders)
	// go ElevatorServer(HallOrderUpdate, CabOrderUpdate, ElevatorStateLocal, WorldviewToCostFunction, SendWorldviewToNetwork, ReceiveWorldviewFromNetwork, CurrentOrders)
	// go RunNetworking(SendWorldviewToNetwork, PeerUpdate, ReceiveWorldviewFromNetwork)
	// go costFunction(WorldviewToCostFunction, ActiveOrders)
}