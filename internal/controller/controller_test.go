package controller

import (
	"fmt"
	"testing"
)

func TestController(t *testing.T) {
	orderEvent, floorEvent, obstructionEvent, stopEvent := InitController(4)

	for {
		select {
		case a := <-orderEvent:
			fmt.Printf("%+v\n", a)

		case a := <-floorEvent:
			fmt.Printf("%+v\n", a)

		case a := <-obstructionEvent:
			fmt.Printf("%+v\n", a)

		case a := <-stopEvent:
			fmt.Printf("%+v\n", a)
		}
	}
}
