package main

import (
	"fmt"
)

func main() {
	drv_buttons, drv_floors, drv_obstr, drv_stop := InitSingleElevator(4)

	for {
		select {
		case a := <-drv_buttons:
			fmt.Printf("%+v\n", a)

		case a := <-drv_floors:
			fmt.Printf("%+v\n", a)

		case a := <-drv_obstr:
			fmt.Printf("%+v\n", a)

		case a := <-drv_stop:
			fmt.Printf("%+v\n", a)
		}
	}
}
