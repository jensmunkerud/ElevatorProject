# Detect OS
UNAME_S := $(shell uname -s)

# Default target
all: submodules hall_request_assigner main

# Step 1: Update submodules
submodules:
	git submodule update --init --recursive

# Step 2: Run correct platform-specific script
hall_request_assigner:
ifeq ($(UNAME_S),Linux)
	chmod +x build_hall_request_assigner_linux.sh
	./build_hall_request_assigner_linux.sh
endif
ifeq ($(UNAME_S),Darwin)
	chmod +x build_hall_request_assigner_darwin.sh
	./build_hall_request_assigner_darwin.sh
endif
ifeq ($(OS),Windows_NT)
	build_hall_request_assigner_windows.bat
endif

# Step 3: Build Go project
main:
ifeq ($(UNAME_S),Linux)
	GOOS=linux GOARCH=amd64 go build -o elevator main.go
endif
ifeq ($(UNAME_S),Darwin)
	GOOS=darwin GOARCH=amd64 go build -o elevator main.go
endif
ifeq ($(OS),Windows_NT)
	set GOOS=windows&& set GOARCH=amd64&& go build -o elevator.exe main.go
endif

# Optional: clean
clean:
	go clean