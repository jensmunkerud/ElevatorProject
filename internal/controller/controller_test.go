package controller

import (
	"fmt"
	"testing"
)

func TestController(t *testing.T) {
	ready := make(chan struct{})
	c := InitController(ready)
	<-ready

	for {
		select {
		case btn := <-c.OrderEvent:
			fmt.Printf("%+v\n", btn)

			done := make(chan struct{})
			go c.GoToFloor(btn.Floor, done)

			<-done
		}
	}
}
