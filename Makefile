# Detect OS
UNAME_S := $(shell uname -s)

# Default target
all: submodules build_script build

# Step 1: Update submodules
submodules:
	go submodule update --init --recursive

# Step 2: Run correct platform-specific script
build_script:
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
build:
	go build main.go

# Optional: clean
clean:
	go clean