@echo off
cd /d C:\Users\jensv\Desktop\ElevatorProject

start "Simulator 15657" cmd /k "SimElevatorServer --port 15657"
start "Simulator 15658" cmd /k "SimElevatorServer --port 15658"
start "Simulator 15659" cmd /k "SimElevatorServer --port 15659"

timeout /t 2 /nobreak >nul

start "Elevator 15657" cmd /k "go run main.go -port 15657"
start "Elevator 15658" cmd /k "go run main.go -port 15658"
start "Elevator 15659" cmd /k "go run main.go -port 15659"

