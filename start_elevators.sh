#!/bin/bash

cd ~/Desktop/ElevatorProject || exit

# Start simulators in separate terminals
gnome-terminal --title="Simulator 15657" -- bash -c "SimElevatorServer --port 15657; exec bash"
gnome-terminal --title="Simulator 15658" -- bash -c "SimElevatorServer --port 15658; exec bash"
gnome-terminal --title="Simulator 15659" -- bash -c "SimElevatorServer --port 15659; exec bash"

# Wait 2 seconds
sleep 2

# Start elevator programs
gnome-terminal --title="Elevator 15657" -- bash -c "go run main.go -port 15657; exec bash"
gnome-terminal --title="Elevator 15658" -- bash -c "go run main.go -port 15658; exec bash"
gnome-terminal --title="Elevator 15659" -- bash -c "go run main.go -port 15659; exec bash"