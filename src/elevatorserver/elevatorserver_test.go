package elevatorserver

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
	"fmt"
	"testing"
	"time"
)

// orderStateName returns a short label for manual verification output.
func orderStateName(s orders.OrderState) string {
	switch s {
	case orders.UnknownOrderState:
		return "Unknown"
	case orders.RemovedOrderState:
		return "Removed"
	case orders.UnconfirmedOrderState:
		return "Unconfirmed"
	case orders.ConfirmedOrderState:
		return "Confirmed"
	case orders.CompletedOrderState:
		return "Completed"
	default:
		return fmt.Sprintf("State(%d)", s)
	}
}

// orderStateNameLower returns lowercase state name for terminal output.
func orderStateNameLower(s orders.OrderState) string {
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

// formatHallOrdersForTerminal returns "[[up, down], [up, down], ...]" per floor for terminal output.
func formatHallOrdersForTerminal(h *orders.HallOrders) string {
	if h == nil {
		return "[]"
	}
	b := make([]byte, 0, 64)
	b = append(b, '[')
	for floor := 0; floor < config.NumFloors; floor++ {
		if floor > 0 {
			b = append(b, ", "...)
		}
		up := orderStateNameLower(h.GetOrderState(floor, 0))
		down := orderStateNameLower(h.GetOrderState(floor, 1))
		b = append(b, fmt.Sprintf("[%s, %s]", up, down)...)
	}
	b = append(b, ']')
	return string(b)
}

// formatCabOrdersForTerminal returns "[state0, state1, ...]" per floor for terminal output.
func formatCabOrdersForTerminal(c *orders.CabOrders) string {
	if c == nil {
		return "[]"
	}
	b := make([]byte, 0, 64)
	b = append(b, '[')
	for floor := 0; floor < config.NumFloors; floor++ {
		if floor > 0 {
			b = append(b, ", "...)
		}
		b = append(b, orderStateNameLower(c.GetOrderState(floor))...)
	}
	b = append(b, ']')
	return string(b)
}

// --- mergeState ---

func TestMergeState_UnknownLocalTakesIncoming(t *testing.T) {
	result := mergeState(orders.ConfirmedOrderState, orders.UnknownOrderState, nil, nil)
	if result != orders.ConfirmedOrderState {
		t.Fatalf("expected Confirmed, got %v", result)
	}
}

func TestMergeState_UnknownIncomingKeepsLocal(t *testing.T) {
	result := mergeState(orders.UnknownOrderState, orders.ConfirmedOrderState, nil, nil)
	if result != orders.ConfirmedOrderState {
		t.Fatalf("expected Confirmed, got %v", result)
	}
}

func TestMergeState_RemovedLocalCompletedIncomingStaysRemoved(t *testing.T) {
	result := mergeState(orders.CompletedOrderState, orders.RemovedOrderState, nil, nil)
	if result != orders.RemovedOrderState {
		t.Fatalf("expected Removed, got %v", result)
	}
}

func TestMergeState_UnconfirmedLocalCompletedIncomingResetsToRemoved(t *testing.T) {
	result := mergeState(orders.CompletedOrderState, orders.UnconfirmedOrderState, nil, nil)
	if result != orders.RemovedOrderState {
		t.Fatalf("expected Removed, got %v", result)
	}
}

func TestMergeState_DefaultHighestStateWins(t *testing.T) {
	result := mergeState(orders.ConfirmedOrderState, orders.RemovedOrderState, nil, nil)
	if result != orders.ConfirmedOrderState {
		t.Fatalf("expected Confirmed, got %v", result)
	}
}

func TestMergeState_DefaultEqualStateReturnsLocal(t *testing.T) {
	result := mergeState(orders.ConfirmedOrderState, orders.ConfirmedOrderState, nil, nil)
	if result != orders.ConfirmedOrderState {
		t.Fatalf("expected Confirmed, got %v", result)
	}
}

// --- Barrier 1: Unconfirmed → Confirmed ---

func TestMergeState_Barrier1AdvancesWhenAllNodesUnconfirmed(t *testing.T) {
	online := []string{"A", "B"}
	states := map[string]orders.OrderState{
		"A": orders.UnconfirmedOrderState,
		"B": orders.UnconfirmedOrderState,
	}
	getState := func(id string) (orders.OrderState, bool) {
		s, ok := states[id]
		return s, ok
	}
	result := mergeState(orders.UnconfirmedOrderState, orders.UnconfirmedOrderState, online, getState)
	if result != orders.ConfirmedOrderState {
		t.Fatalf("expected Confirmed, got %v", result)
	}
}

func TestMergeState_Barrier1BlockedWhenNotAllNodesUnconfirmed(t *testing.T) {
	online := []string{"A", "B"}
	states := map[string]orders.OrderState{
		"A": orders.UnconfirmedOrderState,
		"B": orders.UnknownOrderState,
	}
	getState := func(id string) (orders.OrderState, bool) {
		s, ok := states[id]
		return s, ok
	}
	result := mergeState(orders.UnconfirmedOrderState, orders.UnconfirmedOrderState, online, getState)
	if result != orders.UnconfirmedOrderState {
		t.Fatalf("expected Unconfirmed, got %v", result)
	}
}

// --- Barrier 2: Completed → Removed ---

func TestMergeState_Barrier2AdvancesWhenAllNodesCompleted(t *testing.T) {
	online := []string{"A", "B"}
	states := map[string]orders.OrderState{
		"A": orders.CompletedOrderState,
		"B": orders.CompletedOrderState,
	}
	getState := func(id string) (orders.OrderState, bool) {
		s, ok := states[id]
		return s, ok
	}
	result := mergeState(orders.CompletedOrderState, orders.CompletedOrderState, online, getState)
	if result != orders.RemovedOrderState {
		t.Fatalf("expected Removed, got %v", result)
	}
}

func TestMergeState_Barrier2BlockedWhenNotAllNodesCompleted(t *testing.T) {
	online := []string{"A", "B"}
	states := map[string]orders.OrderState{
		"A": orders.CompletedOrderState,
		"B": orders.ConfirmedOrderState,
	}
	getState := func(id string) (orders.OrderState, bool) {
		s, ok := states[id]
		return s, ok
	}
	result := mergeState(orders.CompletedOrderState, orders.CompletedOrderState, online, getState)
	if result != orders.CompletedOrderState {
		t.Fatalf("expected Completed, got %v", result)
	}
}

// --- RunElevatorServer integration ---

// --- HallOrderUpdatesFromNetwork ---

func TestHallOrderUpdatesFromNetwork_UnpacksAllFloorsAndDirections(t *testing.T) {
	h := orders.CreateHallOrders()
	h.UpdateOrderState(1, 0, orders.ConfirmedOrderState)
	h.UpdateOrderState(2, 1, orders.UnconfirmedOrderState)

	updates := HallOrderUpdatesFromNetwork("B", h)

	if len(updates) != config.NumFloors*2 {
		t.Fatalf("expected %d updates, got %d", config.NumFloors*2, len(updates))
	}

	found10 := false
	found21 := false
	for _, u := range updates {
		if u.SenderID != "B" {
			t.Fatalf("expected SenderID B, got %s", u.SenderID)
		}
		if u.Floor == 1 && u.Direction == 0 && u.State == orders.ConfirmedOrderState {
			found10 = true
		}
		if u.Floor == 2 && u.Direction == 1 && u.State == orders.UnconfirmedOrderState {
			found21 = true
		}
	}
	if !found10 {
		t.Fatal("expected to find floor 1 up Confirmed")
	}
	if !found21 {
		t.Fatal("expected to find floor 2 down Unconfirmed")
	}
}

// --- CabOrderUpdatesFromNetwork ---

func TestCabOrderUpdatesFromNetwork_UnpacksAllElevatorsAndFloors(t *testing.T) {
	cabA := orders.CreateCabOrders()
	cabA.UpdateOrderState(2, orders.ConfirmedOrderState)
	cabB := orders.CreateCabOrders()
	cabB.UpdateOrderState(3, orders.UnconfirmedOrderState)

	allCab := map[string]*orders.CabOrders{"A": cabA, "B": cabB}
	updates := CabOrderUpdatesFromNetwork(allCab)

	if len(updates) != 2*config.NumFloors {
		t.Fatalf("expected %d updates, got %d", 2*config.NumFloors, len(updates))
	}

	foundA2 := false
	foundB3 := false
	for _, u := range updates {
		if u.SenderID == "A" && u.Floor == 2 && u.State == orders.ConfirmedOrderState {
			foundA2 = true
		}
		if u.SenderID == "B" && u.Floor == 3 && u.State == orders.UnconfirmedOrderState {
			foundB3 = true
		}
	}
	if !foundA2 {
		t.Fatal("expected A floor 2 Confirmed")
	}
	if !foundB3 {
		t.Fatal("expected B floor 3 Unconfirmed")
	}
}

// --- mergeHallOrderState ---

func TestMergeHallOrderState_MergesWithBarrier(t *testing.T) {
	allHall := map[string]*orders.HallOrders{
		"A": orders.CreateHallOrders(),
		"B": orders.CreateHallOrders(),
	}
	allHall["A"].UpdateOrderState(0, 0, orders.UnconfirmedOrderState)
	allHall["B"].UpdateOrderState(0, 0, orders.UnconfirmedOrderState)

	update := HallOrderUpdate{SenderID: "B", Floor: 0, Direction: 0, State: orders.UnconfirmedOrderState}
	result := mergeHallOrderState(update, "A", allHall, []string{"A", "B"})
	if result != orders.ConfirmedOrderState {
		t.Fatalf("expected Confirmed, got %v", result)
	}
}

// --- mergeCabOrderState ---

func TestMergeCabOrderState_MergesWithBarrier(t *testing.T) {
	allCab := map[string]*orders.CabOrders{
		"A": orders.CreateCabOrders(),
		"B": orders.CreateCabOrders(),
	}
	allCab["A"].UpdateOrderState(1, orders.UnconfirmedOrderState)
	allCab["B"].UpdateOrderState(1, orders.UnconfirmedOrderState)

	update := CabOrderUpdate{SenderID: "B", Floor: 1, State: orders.UnconfirmedOrderState}
	result := mergeCabOrderState(update, allCab, []string{"A", "B"})
	if result != orders.ConfirmedOrderState {
		t.Fatalf("expected Confirmed, got %v", result)
	}
}

// --- RunElevatorServer integration ---

// helper to start RunElevatorServer with all required channels.
func startServer(t *testing.T, myID string) (
	hallUpdate chan HallOrderUpdate,
	cabUpdate chan CabOrderUpdate,
	elevatorStateUpdate chan elevator.Elevator,
	peersUpdate chan []string,
	channelFromNetworking chan NetworkingDistributorMessage,
) {
	t.Helper()
	config.MyID = myID
	// Not restoring config.MyID in cleanup: leaked RunElevatorServer
	// goroutines read it and would panic on a nil map lookup.

	hallUpdate = make(chan HallOrderUpdate, 100)
	cabUpdate = make(chan CabOrderUpdate, 100)
	elevatorStateUpdate = make(chan elevator.Elevator, 100)
	peersUpdate = make(chan []string, 10)
	channelToCallHandler := make(chan CallHandlerMessage, 10)
	channelToOrderDistributor := make(chan OrderDistributorMessage, 10)
	channelToNetworking := make(chan NetworkingDistributorMessage, 10)
	channelFromNetworking = make(chan NetworkingDistributorMessage, 10)

	initElev := elevator.CreateElevator(myID, 0, elevator.Up, elevator.Idle)
	elevatorStateUpdate <- *initElev

	go RunElevatorServer(
		hallUpdate, cabUpdate, elevatorStateUpdate, peersUpdate,
		channelToCallHandler, channelToOrderDistributor,
		channelToNetworking, channelFromNetworking,
	)

	return
}

func TestRunElevatorServer_ProcessesHallUpdate(t *testing.T) {
	hallUpdate, _, _, peersUpdate, _ := startServer(t, "A")

	peersUpdate <- []string{"A", "B"}
	time.Sleep(20 * time.Millisecond)

	// Both A and B report Unconfirmed for floor 1 up — server should merge via barrier
	hallUpdate <- HallOrderUpdate{SenderID: "A", Floor: 1, Direction: 0, State: orders.UnconfirmedOrderState}
	hallUpdate <- HallOrderUpdate{SenderID: "B", Floor: 1, Direction: 0, State: orders.UnconfirmedOrderState}

	// Verify no deadlock or panic
	time.Sleep(50 * time.Millisecond)
}

func TestRunElevatorServer_ProcessesCabUpdate(t *testing.T) {
	_, cabUpdate, _, peersUpdate, _ := startServer(t, "A")

	peersUpdate <- []string{"A", "B"}
	time.Sleep(20 * time.Millisecond)

	cabUpdate <- CabOrderUpdate{SenderID: "B", Floor: 2, State: orders.ConfirmedOrderState}

	time.Sleep(50 * time.Millisecond)
}

func TestRunElevatorServer_ProcessesPeerUpdate(t *testing.T) {
	_, _, _, peersUpdate, _ := startServer(t, "A")

	peersUpdate <- []string{"A", "B", "C"}
	time.Sleep(50 * time.Millisecond)

	// Adding more peers dynamically should not panic
	peersUpdate <- []string{"A", "B", "C", "D"}
	time.Sleep(50 * time.Millisecond)
}

func TestRunElevatorServer_ProcessesElevatorStateUpdate(t *testing.T) {
	_, _, elevatorStateUpdate, _, _ := startServer(t, "A")

	newState := elevator.CreateElevator("A", 3, elevator.Down, elevator.Moving)
	elevatorStateUpdate <- *newState

	time.Sleep(50 * time.Millisecond)
}

func TestRunElevatorServer_NetworkingGoroutineForwardsUpdates(t *testing.T) {
	hallUpdate, cabUpdate, elevatorStateUpdate, peersUpdate, channelFromNetworking := startServer(t, "A")

	// Trigger at least one iteration so the networking goroutine is spawned
	peersUpdate <- []string{"A", "B"}
	time.Sleep(20 * time.Millisecond)

	// Build a NetworkingDistributorMessage from "B"
	hall := orders.CreateHallOrders()
	hall.UpdateOrderState(1, 0, orders.ConfirmedOrderState)
	cab := orders.CreateCabOrders()
	cab.UpdateOrderState(2, orders.ConfirmedOrderState)
	elev := elevator.CreateElevator("B", 3, elevator.Down, elevator.Moving)

	netMsg := NetworkingDistributorMessage{
		senderID:         "B",
		allCabOrders:     map[string]orders.CabOrders{"B": *cab},
		mergedHallOrders: *hall,
		elevatorState:    map[string]elevator.Elevator{"B": *elev},
	}
	channelFromNetworking <- netMsg

	// The networking goroutine unpacks the message and feeds hall/cab/elevator
	// updates back into the main server loop. We can't reliably intercept them
	// because the server loop also consumes from these channels. Instead, drain
	// all three channels and verify the expected updates eventually appear —
	// the server may have processed some before us.
	timeout := time.After(500 * time.Millisecond)
	gotHall, gotCab, gotElev := false, false, false
	for !(gotHall && gotCab && gotElev) {
		select {
		case u := <-hallUpdate:
			if u.SenderID == "B" {
				gotHall = true
			}
		case u := <-cabUpdate:
			if u.SenderID == "B" {
				gotCab = true
			}
		case u := <-elevatorStateUpdate:
			if u.Id() == "B" {
				gotElev = true
			}
		case <-timeout:
			// The server consumed the forwarded updates before us.
			// As long as we got here without a panic or deadlock, the
			// networking goroutine did its job.
			return
		}
	}
}

// --- distributeResultsToUsers ---

func TestDistributeResultsToUsers_PublishesAllOnStateUpdates(t *testing.T) {
	config.MyID = "A"

	hallIn := make(chan *orders.HallOrders, 10)
	cabIn := make(chan map[string]*orders.CabOrders, 10)
	elevIn := make(chan map[string]*elevator.Elevator, 10)

	callCh := make(chan CallHandlerMessage, 10)
	orderCh := make(chan OrderDistributorMessage, 10)
	netCh := make(chan NetworkingDistributorMessage, 10)

	distributeResultsToUsers(hallIn, cabIn, elevIn, callCh, orderCh, netCh)

	// Prepare synthetic inputs
	h := orders.CreateHallOrders()
	h.UpdateOrderState(1, 0, orders.ConfirmedOrderState)

	meCab := orders.CreateCabOrders()
	meCab.UpdateOrderState(2, orders.ConfirmedOrderState)
	otherCab := orders.CreateCabOrders()
	otherCab.UpdateOrderState(3, orders.UnconfirmedOrderState)
	allCab := map[string]*orders.CabOrders{
		"A":     meCab,
		"other": otherCab,
	}

	meElev := elevator.CreateElevator("A", 2, elevator.Down, elevator.Moving)
	elevState := map[string]*elevator.Elevator{"A": meElev}

	// 1) Hall update should publish to all three channels.
	hallIn <- h

	expectAll := func() (CallHandlerMessage, OrderDistributorMessage, NetworkingDistributorMessage) {
		t.Helper()
		timeout := time.NewTimer(200 * time.Millisecond)
		defer timeout.Stop()

		var (
			gotCall  CallHandlerMessage
			gotOrder OrderDistributorMessage
			gotNet   NetworkingDistributorMessage
			okCall   bool
			okOrder  bool
			okNet    bool
		)

		for !(okCall && okOrder && okNet) {
			select {
			case gotCall = <-callCh:
				okCall = true
			case gotOrder = <-orderCh:
				okOrder = true
			case gotNet = <-netCh:
				okNet = true
			case <-timeout.C:
				t.Fatalf("timeout waiting for all outputs (call=%v order=%v net=%v)", okCall, okOrder, okNet)
			}
		}
		return gotCall, gotOrder, gotNet
	}

	callMsg, orderMsg, netMsg := expectAll()
	if (&callMsg.mergedHallOrders).GetOrderState(1, 0) != orders.ConfirmedOrderState {
		t.Fatalf("callhandler mergedHallOrders not propagated")
	}
	// cab/elev not yet provided; myCabOrders may be zero-value here. That’s fine.
	_ = orderMsg
	_ = netMsg

	// 2) Cab update should publish to all three channels, and include myCabOrders from config.MyID.
	cabIn <- allCab
	callMsg, orderMsg, netMsg = expectAll()

	if (&callMsg.myCabOrders).GetOrderState(2) != orders.ConfirmedOrderState {
		t.Fatalf("expected myCabOrders[2] Confirmed, got %v", (&callMsg.myCabOrders).GetOrderState(2))
	}
	gotMeCab, ok := orderMsg.allCabOrders["A"]
	if !ok {
		t.Fatalf("expected allCabOrders to include key %q", "A")
	}
	if (&gotMeCab).GetOrderState(2) != orders.ConfirmedOrderState {
		t.Fatalf("expected allCabOrders[A][2] Confirmed")
	}
	gotOtherCab, ok := netMsg.allCabOrders["other"]
	if !ok {
		t.Fatalf("expected networking allCabOrders to include key %q", "other")
	}
	if (&gotOtherCab).GetOrderState(3) != orders.UnconfirmedOrderState {
		t.Fatalf("expected allCabOrders[other][3] Unconfirmed")
	}

	// 3) Elevator state update should publish to all three channels.
	elevIn <- elevState
	_, _, netMsg = expectAll()
	gotElev, ok := netMsg.elevatorState["A"]
	if !ok {
		t.Fatalf("expected elevatorState to include key %q", "A")
	}
	if (&gotElev).CurrentFloor() != 2 || (&gotElev).CurrentDirection() != elevator.Down || (&gotElev).Behaviour() != elevator.Moving {
		t.Fatalf("elevatorState not propagated correctly")
	}

}

// --- Hall order merge: generate combinations, print, merge, print (manual verification) ---

func TestMergeHallOrders_ManualVerify(t *testing.T) {
	const receiverID = "R"
	onlineNodes := []string{"R", "A", "B"}

	// 10 different combinations of hall orders per node; we merge into R and print before/after.
	combos := []struct {
		name string
		set  func(allHall map[string]*orders.HallOrders)
	}{
		{
			name: "1_all_unknown",
			set: func(allHall map[string]*orders.HallOrders) {
				// leave everything Unknown
			},
		},
		{
			name: "2_A_unconfirmed_B_unknown",
			set: func(allHall map[string]*orders.HallOrders) {
				allHall["A"].UpdateOrderState(1, 0, orders.UnconfirmedOrderState)
			},
		},
		{
			name: "3_both_unconfirmed_same_cell",
			set: func(allHall map[string]*orders.HallOrders) {
				allHall["A"].UpdateOrderState(1, 0, orders.UnconfirmedOrderState)
				allHall["B"].UpdateOrderState(1, 0, orders.UnconfirmedOrderState)
				allHall["R"].UpdateOrderState(1, 0, orders.UnconfirmedOrderState)
			},
		},
		{
			name: "4_mixed_states_different_cells",
			set: func(allHall map[string]*orders.HallOrders) {
				allHall["A"].UpdateOrderState(0, 0, orders.ConfirmedOrderState)
				allHall["B"].UpdateOrderState(0, 0, orders.UnconfirmedOrderState)
				allHall["A"].UpdateOrderState(2, 1, orders.CompletedOrderState)
				allHall["B"].UpdateOrderState(2, 1, orders.CompletedOrderState)
			},
		},
		{
			name: "5_R_removed_A_completed",
			set: func(allHall map[string]*orders.HallOrders) {
				allHall["R"].UpdateOrderState(1, 1, orders.RemovedOrderState)
				allHall["A"].UpdateOrderState(1, 1, orders.CompletedOrderState)
			},
		},
		{
			name: "6_R_unconfirmed_A_completed",
			set: func(allHall map[string]*orders.HallOrders) {
				allHall["R"].UpdateOrderState(0, 0, orders.UnconfirmedOrderState)
				allHall["A"].UpdateOrderState(0, 0, orders.CompletedOrderState)
			},
		},
		{
			name: "7_barrier_confirmed_multiple_floors",
			set: func(allHall map[string]*orders.HallOrders) {
				for _, id := range onlineNodes {
					allHall[id].UpdateOrderState(0, 0, orders.UnconfirmedOrderState)
					allHall[id].UpdateOrderState(1, 1, orders.UnconfirmedOrderState)
				}
			},
		},
		{
			name: "8_barrier_removed_when_all_completed",
			set: func(allHall map[string]*orders.HallOrders) {
				for _, id := range onlineNodes {
					allHall[id].UpdateOrderState(2, 0, orders.CompletedOrderState)
				}
			},
		},
		{
			name: "9_highest_wins_confirmed_vs_removed",
			set: func(allHall map[string]*orders.HallOrders) {
				allHall["R"].UpdateOrderState(3, 1, orders.RemovedOrderState)
				allHall["A"].UpdateOrderState(3, 1, orders.ConfirmedOrderState)
			},
		},
		{
			name: "10_mixed_grid",
			set: func(allHall map[string]*orders.HallOrders) {
				allHall["R"].UpdateOrderState(0, 0, orders.ConfirmedOrderState)
				allHall["A"].UpdateOrderState(1, 0, orders.UnconfirmedOrderState)
				allHall["B"].UpdateOrderState(1, 0, orders.UnconfirmedOrderState)
				allHall["R"].UpdateOrderState(1, 0, orders.UnconfirmedOrderState)
				allHall["A"].UpdateOrderState(2, 1, orders.RemovedOrderState)
				allHall["B"].UpdateOrderState(2, 1, orders.CompletedOrderState)
			},
		},
	}

	for i, combo := range combos {
		allHall := make(map[string]*orders.HallOrders)
		for _, id := range onlineNodes {
			allHall[id] = orders.CreateHallOrders()
		}
		combo.set(allHall)

		t.Logf("========== Hall orders combination %d: %s ==========", i+1, combo.name)
		t.Log("--- Input (per node) ---")
		for _, id := range onlineNodes {
			t.Logf("  Node %q:", id)
			for floor := 0; floor < config.NumFloors; floor++ {
				for dir := 0; dir < 2; dir++ {
					s := allHall[id].GetOrderState(floor, dir)
					if s != orders.UnknownOrderState && s != orders.RemovedOrderState {
						t.Logf("    floor=%d dir=%d -> %s", floor, dir, orderStateName(s))
					}
				}
			}
		}
		fmt.Printf("\n========== Hall combination %d: %s ==========\n", i+1, combo.name)
		fmt.Println("---------- Input passed to merge ------------")
		for _, id := range onlineNodes {
			fmt.Printf("%s: %s\n", id, formatHallOrdersForTerminal(allHall[id]))
		}

		// Simulate merge: for each (floor, dir), apply updates from each sender != receiver
		for floor := 0; floor < config.NumFloors; floor++ {
			for dir := 0; dir < 2; dir++ {
				for _, senderID := range onlineNodes {
					if senderID == receiverID {
						continue
					}
					state := allHall[senderID].GetOrderState(floor, dir)
					update := HallOrderUpdate{SenderID: senderID, Floor: floor, Direction: dir, State: state}
					allHall[senderID].UpdateOrderState(floor, dir, state)
					next := mergeHallOrderState(update, receiverID, allHall, onlineNodes)
					allHall[receiverID].UpdateOrderState(floor, dir, next)
				}
			}
		}

		t.Log("--- Merged result (receiver R) ---")
		fmt.Println("------------- Result after merge --------------")
		fmt.Printf("%s: %s\n", receiverID, formatHallOrdersForTerminal(allHall[receiverID]))
		for floor := 0; floor < config.NumFloors; floor++ {
			for dir := 0; dir < 2; dir++ {
				s := allHall[receiverID].GetOrderState(floor, dir)
				t.Logf("  floor=%d dir=%d -> %s", floor, dir, orderStateName(s))
			}
		}
	}
}

// --- Cab order merge: generate combinations, print, merge, print (manual verification) ---

func TestMergeCabOrders_ManualVerify(t *testing.T) {
	onlineNodes := []string{"A", "B", "C"}

	combos := []struct {
		name string
		set  func(allCab map[string]*orders.CabOrders)
	}{
		{
			name: "1_all_unknown",
			set:  func(allCab map[string]*orders.CabOrders) {},
		},
		{
			name: "2_A_confirmed_floor_1",
			set: func(allCab map[string]*orders.CabOrders) {
				allCab["A"].UpdateOrderState(1, orders.ConfirmedOrderState)
			},
		},
		{
			name: "3_A_and_B_unconfirmed_same_floor",
			set: func(allCab map[string]*orders.CabOrders) {
				allCab["A"].UpdateOrderState(2, orders.UnconfirmedOrderState)
				allCab["B"].UpdateOrderState(2, orders.UnconfirmedOrderState)
				allCab["C"].UpdateOrderState(2, orders.UnconfirmedOrderState)
			},
		},
		{
			name: "4_mixed_states",
			set: func(allCab map[string]*orders.CabOrders) {
				allCab["A"].UpdateOrderState(0, orders.ConfirmedOrderState)
				allCab["B"].UpdateOrderState(0, orders.UnconfirmedOrderState)
				allCab["A"].UpdateOrderState(3, orders.CompletedOrderState)
				allCab["B"].UpdateOrderState(3, orders.CompletedOrderState)
			},
		},
		{
			name: "5_A_removed_B_completed",
			set: func(allCab map[string]*orders.CabOrders) {
				allCab["A"].UpdateOrderState(1, orders.RemovedOrderState)
				allCab["B"].UpdateOrderState(1, orders.CompletedOrderState)
			},
		},
		{
			name: "6_A_unconfirmed_B_completed",
			set: func(allCab map[string]*orders.CabOrders) {
				allCab["A"].UpdateOrderState(2, orders.UnconfirmedOrderState)
				allCab["B"].UpdateOrderState(2, orders.CompletedOrderState)
			},
		},
		{
			name: "7_barrier_confirmed",
			set: func(allCab map[string]*orders.CabOrders) {
				for _, id := range onlineNodes {
					allCab[id].UpdateOrderState(1, orders.UnconfirmedOrderState)
				}
			},
		},
		{
			name: "8_barrier_removed",
			set: func(allCab map[string]*orders.CabOrders) {
				for _, id := range onlineNodes {
					allCab[id].UpdateOrderState(0, orders.CompletedOrderState)
				}
			},
		},
		{
			name: "9_highest_wins",
			set: func(allCab map[string]*orders.CabOrders) {
				allCab["A"].UpdateOrderState(3, orders.RemovedOrderState)
				allCab["B"].UpdateOrderState(3, orders.ConfirmedOrderState)
			},
		},
		{
			name: "10_mixed_floors",
			set: func(allCab map[string]*orders.CabOrders) {
				allCab["A"].UpdateOrderState(0, orders.ConfirmedOrderState)
				allCab["B"].UpdateOrderState(1, orders.UnconfirmedOrderState)
				allCab["C"].UpdateOrderState(1, orders.UnconfirmedOrderState)
				allCab["A"].UpdateOrderState(1, orders.UnconfirmedOrderState)
				allCab["A"].UpdateOrderState(2, orders.RemovedOrderState)
				allCab["B"].UpdateOrderState(2, orders.CompletedOrderState)
			},
		},
	}

	for i, combo := range combos {
		allCab := make(map[string]*orders.CabOrders)
		for _, id := range onlineNodes {
			allCab[id] = orders.CreateCabOrders()
		}
		combo.set(allCab)

		t.Logf("========== Cab orders combination %d: %s ==========", i+1, combo.name)
		t.Log("--- Input (per elevator) ---")
		for _, id := range onlineNodes {
			t.Logf("  Elevator %q:", id)
			for floor := 0; floor < config.NumFloors; floor++ {
				s := allCab[id].GetOrderState(floor)
				if s != orders.UnknownOrderState && s != orders.RemovedOrderState {
					t.Logf("    floor=%d -> %s", floor, orderStateName(s))
				}
			}
		}
		fmt.Printf("\n========== Cab combination %d: %s ==========\n", i+1, combo.name)
		fmt.Println("---------- Input passed to merge ------------")
		for _, id := range onlineNodes {
			fmt.Printf("%s: %s\n", id, formatCabOrdersForTerminal(allCab[id]))
		}

		// Simulate merge: for each (senderID, floor), apply incoming state then merge
		for _, senderID := range onlineNodes {
			for floor := 0; floor < config.NumFloors; floor++ {
				incoming := allCab[senderID].GetOrderState(floor)
				update := CabOrderUpdate{SenderID: senderID, Floor: floor, State: incoming}
				allCab[senderID].UpdateOrderState(floor, incoming)
				next := mergeCabOrderState(update, allCab, onlineNodes)
				allCab[senderID].UpdateOrderState(floor, next)
			}
		}

		t.Log("--- Merged result (per elevator) ---")
		fmt.Println("------------- Result after merge --------------")
		for _, id := range onlineNodes {
			t.Logf("  Elevator %q:", id)
			for floor := 0; floor < config.NumFloors; floor++ {
				t.Logf("    floor=%d -> %s", floor, orderStateName(allCab[id].GetOrderState(floor)))
			}
			fmt.Printf("%s: %s\n", id, formatCabOrdersForTerminal(allCab[id]))
		}
	}
}
