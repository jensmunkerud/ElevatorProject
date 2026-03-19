.PHONY: build submodules hall main

build: submodules hall main

submodules:
    git submodule update --init --recursive

hall:
    go build -C hall_request_assigner -o hall_request_assigner .

main:
    go build -o elevator ./main.go