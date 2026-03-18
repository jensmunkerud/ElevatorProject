#!/bin/bash

cd ~/ElevatorProject || exit

# Start simulators
gnome-terminal --title="Simulator 15657" -- bash -ic "SimElevatorServer --port 15657"
gnome-terminal --title="Simulator 15658" -- bash -ic "SimElevatorServer --port 15658"
gnome-terminal --title="Simulator 15659" -- bash -ic "SimElevatorServer --port 15659"

# Wait 2 seconds
sleep 2

# Start elevators
gnome-terminal --title="Elevator 15657" -- bash -ic "go run main.go -port 15657"
gnome-terminal --title="Elevator 15658" -- bash -ic "go run main.go -port 15658"
gnome-terminal --title="Elevator 15659" -- bash -ic "go run main.go -port 15659"