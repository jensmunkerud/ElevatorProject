#!/bin/bash

cd /Users/jensv/Desktop/ElevatorProject || exit

# Start simulators
osascript -e 'tell application "Terminal" to do script "cd /Users/jensv/Desktop/ElevatorProject && SimElevatorServer --port 15657"'
osascript -e 'tell application "Terminal" to do script "cd /Users/jensv/Desktop/ElevatorProject && SimElevatorServer --port 15658"'
osascript -e 'tell application "Terminal" to do script "cd /Users/jensv/Desktop/ElevatorProject && SimElevatorServer --port 15659"'

# Wait 2 seconds
sleep 2

# Start elevators
osascript -e 'tell application "Terminal" to do script "cd /Users/jensv/Desktop/ElevatorProject && go run main.go -port 15657 -processpair"'
osascript -e 'tell application "Terminal" to do script "cd /Users/jensv/Desktop/ElevatorProject && go run main.go -port 15658 -processpair"'
osascript -e 'tell application "Terminal" to do script "cd /Users/jensv/Desktop/ElevatorProject && go run main.go -port 15659 -processpair"'