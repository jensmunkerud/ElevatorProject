@echo off
set "SCRIPT_DIR=%~dp0"
for %%I in ("%SCRIPT_DIR%..") do set "PROJECT_ROOT=%%~fI"
cd /d "%PROJECT_ROOT%"

echo Building elevator...
go build -o elevator.exe main.go
if errorlevel 1 (
    echo Build failed!
    pause
    exit /b 1
)

start "Simulator 15657" cmd /k "cd /d ""%PROJECT_ROOT%"" && SimElevatorServer --port 15657"
start "Simulator 15658" cmd /k "cd /d ""%PROJECT_ROOT%"" && SimElevatorServer --port 15658"
start "Simulator 15659" cmd /k "cd /d ""%PROJECT_ROOT%"" && SimElevatorServer --port 15659"

timeout /t 2 /nobreak >nul

start "Elevator 15657" cmd /k "cd /d ""%PROJECT_ROOT%"" && elevator.exe -port 15657"
start "Elevator 15658" cmd /k "cd /d ""%PROJECT_ROOT%"" && elevator.exe -port 15658"
start "Elevator 15659" cmd /k "cd /d ""%PROJECT_ROOT%"" && elevator.exe -port 15659"

