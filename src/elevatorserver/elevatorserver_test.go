package elevatorserver

import (
	"elevatorproject/src/orders"
	"testing"
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
