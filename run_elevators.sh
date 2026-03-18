#!/bin/bash

cd /Users/jens/Desktop/ElevatorProject || exit

# Build first so os.Executable() returns a stable path for process pairs
echo "Building elevator..."
go build -o elevator main.go || { echo "Build failed!"; exit 1; }

# Start simulators
osascript -e 'tell application "Terminal" to do script "cd /Users/jens/Desktop/ElevatorProject && ./SimElevatorServer --port 15657"'
osascript -e 'tell application "Terminal" to do script "cd /Users/jens/Desktop/ElevatorProject && ./SimElevatorServer --port 15658"'
osascript -e 'tell application "Terminal" to do script "cd /Users/jens/Desktop/ElevatorProject && ./SimElevatorServer --port 15659"'

# Wait 2 seconds
sleep 2

# Start elevators
osascript -e 'tell application "Terminal" to do script "cd /Users/jens/Desktop/ElevatorProject && ./elevator -port 15657"'
osascript -e 'tell application "Terminal" to do script "cd /Users/jens/Desktop/ElevatorProject && ./elevator -port 15658"'
osascript -e 'tell application "Terminal" to do script "cd /Users/jens/Desktop/ElevatorProject && ./elevator -port 15659"'
