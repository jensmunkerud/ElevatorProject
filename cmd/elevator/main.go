package main

import "elevatorproject/internal/callhandler"
import "elevatorproject/internal/networking"
import "elevatorproject/internal/controller"

func main() {
	callhandler.Test()
	networking.Test()
	controller.Test()

	// HallRequest
	// Elevator 1
	// Elevator 2
	// Elevator 3

	//Send and recieve elevator state through four channels (HallRequest, Elevator 1, Elevator 2, Elevator 3)

}