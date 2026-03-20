#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT" || exit

# Build first so os.Executable() returns a stable path for process pairs
echo "Building elevator..."
go build -o elevator main.go || { echo "Build failed!"; exit 1; }

# Start simulators
gnome-terminal --title="Simulator 15657" -- bash -ic "cd \"$PROJECT_ROOT\" && SimElevatorServer --port 15657"
gnome-terminal --title="Simulator 15658" -- bash -ic "cd \"$PROJECT_ROOT\" && SimElevatorServer --port 15658"
gnome-terminal --title="Simulator 15659" -- bash -ic "cd \"$PROJECT_ROOT\" && SimElevatorServer --port 15659"

# Wait 2 seconds
sleep 2

# Start elevators
gnome-terminal --title="Elevator 15657" -- bash -ic "cd \"$PROJECT_ROOT\" && ./elevator -port 15657"
gnome-terminal --title="Elevator 15658" -- bash -ic "cd \"$PROJECT_ROOT\" && ./elevator -port 15658"
gnome-terminal --title="Elevator 15659" -- bash -ic "cd \"$PROJECT_ROOT\" && ./elevator -port 15659"