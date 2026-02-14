package main

import "elevatorproject/internal/callhandler"
import "elevatorproject/internal/networking"
import "elevatorproject/internal/controller"

func main() {
	callhandler.Test()
	networking.Test()
	controller.Test()
}