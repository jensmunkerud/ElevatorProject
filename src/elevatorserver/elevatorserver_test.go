package elevatorserver

import (
	"elevatorproject/src/config"
	"elevatorproject/src/elevator"
	"elevatorproject/src/orders"
	"testing"
	"time"
)

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
	origID := config.MyID
	config.MyID = myID
	t.Cleanup(func() { config.MyID = origID })

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
	hallUpdate, _, _, peersUpdate, channelFromNetworking := startServer(t, "A")

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

	// The goroutine unpacks the message and feeds hall/cab/elevator updates back
	// into the main loop channels. Drain hallUpdate to verify forwarding happened.
	timeout := time.After(500 * time.Millisecond)
	foundHall := false
	for !foundHall {
		select {
		case u := <-hallUpdate:
			if u.SenderID == "B" && u.Floor == 1 && u.Direction == 0 && u.State == orders.ConfirmedOrderState {
				foundHall = true
			}
		case <-timeout:
			t.Fatal("timeout waiting for hall update forwarded from networking goroutine")
		}
	}
}

// --- distributeResultsToUsers ---

func TestDistributeResultsToUsers_PublishesAllOnStateUpdates(t *testing.T) {
	origID := config.MyID
	config.MyID = "me"
	t.Cleanup(func() { config.MyID = origID })

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
		"me":    meCab,
		"other": otherCab,
	}

	meElev := elevator.CreateElevator("me", 2, elevator.Down, elevator.Moving)
	elevState := map[string]*elevator.Elevator{"me": meElev}

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
	gotMeCab, ok := orderMsg.allCabOrders["me"]
	if !ok {
		t.Fatalf("expected allCabOrders to include key %q", "me")
	}
	if (&gotMeCab).GetOrderState(2) != orders.ConfirmedOrderState {
		t.Fatalf("expected allCabOrders[me][2] Confirmed")
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
	gotElev, ok := netMsg.elevatorState["me"]
	if !ok {
		t.Fatalf("expected elevatorState to include key %q", "me")
	}
	if (&gotElev).CurrentFloor() != 2 || (&gotElev).CurrentDirection() != elevator.Down || (&gotElev).Behaviour() != elevator.Moving {
		t.Fatalf("elevatorState not propagated correctly")
	}

}
