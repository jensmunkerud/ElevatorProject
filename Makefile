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
	go build -o elevator main.go

# Optional: clean
clean:
	go clean