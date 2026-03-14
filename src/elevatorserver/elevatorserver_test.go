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

func TestRunElevatorServer_HallUpdateMergesAndBroadcasts(t *testing.T) {
	hallUpdates := make(chan HallOrderUpdate, 10)
	cabUpdates := make(chan CabOrderUpdate, 10)
	peerUpdates := make(chan []string, 1)
	hallOut := make(chan *orders.HallOrders, 10)
	cabOut := make(chan map[string]*orders.CabOrders, 10)

	go RunElevatorServer(hallUpdates, cabUpdates, peerUpdates, hallOut, cabOut, "A")

	// Seed peers so barrier can fire
	peerUpdates <- []string{"A", "B"}

	// Both A and B report Unconfirmed for floor 1 up — barrier should fire → Confirmed
	hallUpdates <- HallOrderUpdate{SenderID: "A", Floor: 1, Direction: 0, State: orders.UnconfirmedOrderState}
	hallUpdates <- HallOrderUpdate{SenderID: "B", Floor: 1, Direction: 0, State: orders.UnconfirmedOrderState}

	// Drain until we see Confirmed at floor 1 up
	for i := 0; i < 20; i++ {
		select {
		case h := <-hallOut:
			if h.GetOrderState(1, 0) == orders.ConfirmedOrderState {
				return
			}
		case <-cabOut:
		}
	}
	t.Fatal("expected hall order at floor 1 up to reach Confirmed")
}

func TestRunElevatorServer_CabUpdateStoresUnderCorrectOwner(t *testing.T) {
	hallUpdates := make(chan HallOrderUpdate, 10)
	cabUpdates := make(chan CabOrderUpdate, 10)
	peerUpdates := make(chan []string, 1)
	hallOut := make(chan *orders.HallOrders, 10)
	cabOut := make(chan map[string]*orders.CabOrders, 10)

	go RunElevatorServer(hallUpdates, cabUpdates, peerUpdates, hallOut, cabOut, "A")

	peerUpdates <- []string{"A", "B"}

	// B reports a Confirmed cab order at floor 2 — it should appear in the broadcast map under "B"
	cabUpdates <- CabOrderUpdate{SenderID: "B", Floor: 2, State: orders.ConfirmedOrderState}

	for i := 0; i < 20; i++ {
		select {
		case m := <-cabOut:
			if b, ok := m["B"]; ok && b.GetOrderState(2) == orders.ConfirmedOrderState {
				return
			}
		case <-hallOut:
		}
	}
	t.Fatal("expected B's cab order at floor 2 to be Confirmed in the broadcast map")
}

func TestDistributeResultsToUsers_PublishesAllOnStateUpdates(t *testing.T) {
	origID := config.MyID
	config.MyID = "me"
	t.Cleanup(func() { config.MyID = origID })

	hallIn := make(chan *orders.HallOrders, 10)
	cabIn := make(chan map[string]*orders.CabOrders, 10)
	elevIn := make(chan map[string]*elevator.Elevator, 10)

	callCh, orderCh, netCh := distributeResultsToUsers(hallIn, cabIn, elevIn)

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
