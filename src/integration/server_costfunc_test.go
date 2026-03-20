package integration_test

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/elevatorserver"
	"elevatorproject/src/orderdistributor"
	"elevatorproject/src/orders"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func orderStateName(s orders.OrderState) string {
	switch s {
	case orders.UnknownOrderState:
		return "unknown"
	case orders.RemovedOrderState:
		return "removed"
	case orders.UnconfirmedOrderState:
		return "unconfirmed"
	case orders.ConfirmedOrderState:
		return "confirmed"
	case orders.CompletedOrderState:
		return "completed"
	default:
		return fmt.Sprintf("state(%d)", s)
	}
}

func formatHallOrders(h *orders.HallOrders) string {
	if h == nil {
		return "nil"
	}
	s := "["
	for f := 0; f < config.NumFloors; f++ {
		if f > 0 {
			s += ", "
		}
		s += fmt.Sprintf("[up=%s down=%s]",
			orderStateName(h.GetOrderState(f, 0)),
			orderStateName(h.GetOrderState(f, 1)))
	}
	return s + "]"
}

func formatCabOrders(c *orders.CabOrders) string {
	if c == nil {
		return "nil"
	}
	s := "["
	for f := 0; f < config.NumFloors; f++ {
		if f > 0 {
			s += ", "
		}
		s += orderStateName(c.GetOrderState(f))
	}
	return s + "]"
}

func formatActiveOrders(o [config.NumFloors][config.NumButtons]bool, received bool) string {
	if !received {
		return "(none)"
	}
	s := "["
	for f, row := range o {
		if f > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%v", row)
	}
	return s + "]"
}

// TestServerAndCostFuncInteraction is a ~2 minute integration test that connects
// RunElevatorServer and RunCostFunc, spams orders for ~60s, then idles for ~60s.
// It continuously prints hall orders, cab orders, and cost function output every 5 seconds.
func TestServerAndCostFuncInteraction(t *testing.T) {
	const (
		spamDuration  = 60 * time.Second
		idleDuration  = 60 * time.Second
		printInterval = 5 * time.Second
	)

	config.InitConfigTesting()
	myID := config.MyID()
	t.Logf("My elevator ID: %s", myID)

	// --- Set up all channels ---
	hallUpdate := make(chan elevatorserver.HallOrderUpdate, 500)
	cabUpdate := make(chan elevatorserver.CabOrderUpdate, 500)
	elevatorStateUpdate := make(chan elevator.Elevator, 500)
	peersUpdate := make(chan []string, 10)
	toCallHandler := make(chan elevatorserver.CallHandlerMessage, 10)
	toOrderDist := make(chan elevatorserver.OrderDistributorMessage, 10)
	toNetworking := make(chan elevatorserver.NetworkingDistributorMessage, 10)
	fromNetworking := make(chan elevatorserver.NetworkingDistributorMessage, 10)
	activeOrdersCh := make(chan [config.NumFloors][config.NumButtons]bool, 500)

	// RunElevatorServer blocks waiting for an initial elevator state.
	initElev := elevator.CreateElevator(myID, 0, elevator.Up, elevator.Idle)
	elevatorStateUpdate <- *initElev

	// --- Start goroutines ---
	go elevatorserver.Run(
		hallUpdate, cabUpdate, elevatorStateUpdate, peersUpdate,
		toCallHandler, toOrderDist, toNetworking, fromNetworking,
	)
	go orderdistributor.Run(toOrderDist, activeOrdersCh)

	// Register self as the only online elevator.
	peersUpdate <- []string{myID}
	time.Sleep(200 * time.Millisecond)

	// Drain call-handler and networking channels so distributeResultsToUsers never blocks.
	go func() {
		for range toCallHandler {
		}
	}()
	go func() {
		for range toNetworking {
		}
	}()

	// --- Tracking state ---
	var mu sync.Mutex
	localHall := orders.CreateHallOrders()
	localCab := orders.CreateCabOrders()
	var latestActive [config.NumFloors][config.NumButtons]bool
	var receivedActive bool
	var costFuncOK int64
	var hallSent int64
	var cabSent int64

	// Consume cost-function output.
	go func() {
		for o := range activeOrdersCh {
			mu.Lock()
			latestActive = o
			receivedActive = true
			costFuncOK++
			mu.Unlock()
		}
	}()

	// --- Periodic status printer (every 5s) ---
	printDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(printInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				t.Logf("──────────── Status ────────────")
				t.Logf("  Hall orders (local):  %s", formatHallOrders(localHall))
				t.Logf("  Cab  orders (local):  %s", formatCabOrders(localCab))
				t.Logf("  Active orders (cost): %s", formatActiveOrders(latestActive, receivedActive))
				t.Logf("  Msgs sent:  hall=%d  cab=%d", hallSent, cabSent)
				t.Logf("  Cost func:  ok=%d", costFuncOK)
				mu.Unlock()
			case <-printDone:
				return
			}
		}
	}()

	// ══════════════════════════════════════════════════
	// Phase 1 — Spam orders for ~60 seconds
	// ══════════════════════════════════════════════════
	t.Log(">>> Phase 1: Spamming hall/cab orders for ~60 seconds")
	spamEnd := time.Now().Add(spamDuration)

	for time.Now().Before(spamEnd) {
		floor := rand.Intn(config.NumFloors)
		dir := rand.Intn(2)

		// Weighted random state: mostly Unconfirmed (button press), some Completed.
		var state orders.OrderState
		switch r := rand.Intn(10); {
		case r < 6:
			state = orders.UnconfirmedOrderState
		case r < 8:
			state = orders.CompletedOrderState
		case r < 9:
			state = orders.ConfirmedOrderState
		default:
			state = orders.RemovedOrderState
		}

		mu.Lock()
		localHall.UpdateOrderState(floor, dir, state)
		hallSent++
		mu.Unlock()

		hallUpdate <- elevatorserver.HallOrderUpdate{
			SenderID:  myID,
			Floor:     floor,
			OrderType: dir,
			State:     state,
		}

		// Send cab orders ~30% of the time.
		if rand.Intn(10) < 3 {
			cabFloor := rand.Intn(config.NumFloors)
			cabState := orders.UnconfirmedOrderState
			if rand.Intn(3) == 0 {
				cabState = orders.CompletedOrderState
			}

			mu.Lock()
			localCab.UpdateOrderState(cabFloor, cabState)
			cabSent++
			mu.Unlock()

			cabUpdate <- elevatorserver.CabOrderUpdate{
				SenderID: myID,
				Floor:    cabFloor,
				State:    cabState,
			}
		}

		// Also send elevator state updates occasionally (position changes).
		if rand.Intn(20) == 0 {
			newFloor := rand.Intn(config.NumFloors)
			dirs := []elevator.Direction{elevator.Stop, elevator.Up, elevator.Down}
			behaviours := []elevator.Behaviour{elevator.Idle, elevator.Moving, elevator.DoorOpen}
			e := elevator.CreateElevator(myID, newFloor, dirs[rand.Intn(3)], behaviours[rand.Intn(3)])
			elevatorStateUpdate <- *e
		}

		// 10–100ms between sends for ~600–6000 hall updates per phase.
		time.Sleep(time.Duration(10+rand.Intn(90)) * time.Millisecond)
	}

	mu.Lock()
	t.Logf(">>> Phase 1 complete: sent %d hall, %d cab updates", hallSent, cabSent)
	mu.Unlock()

	// ══════════════════════════════════════════════════
	// Phase 2 — Idle for ~60 seconds (heartbeat still ticking)
	// ══════════════════════════════════════════════════
	t.Log(">>> Phase 2: Idling for ~60 seconds")
	time.Sleep(idleDuration)

	// ══════════════════════════════════════════════════
	// Final report
	// ══════════════════════════════════════════════════
	mu.Lock()
	t.Logf("════════════ Final Report ════════════")
	t.Logf("  Hall orders (local):  %s", formatHallOrders(localHall))
	t.Logf("  Cab  orders (local):  %s", formatCabOrders(localCab))
	t.Logf("  Active orders (cost): %s", formatActiveOrders(latestActive, receivedActive))
	t.Logf("  Msgs sent:  hall=%d  cab=%d", hallSent, cabSent)
	t.Logf("  Cost func:  ok=%d", costFuncOK)
	mu.Unlock()

	close(printDone)
	t.Log("Test completed — no panics, deadlocks, or crashes detected")
}
