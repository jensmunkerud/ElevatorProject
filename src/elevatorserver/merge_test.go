package elevatorserver

import (
	"fmt"
	"strings"
	"testing"

	"elevatorproject/src/config"
	"elevatorproject/src/orders"
)

func mergeTestStateName(s orders.OrderState) string {
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

func formatCabForMergeTest(c *orders.CabOrders) string {
	values := make([]string, 0, config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++ {
		values = append(values, mergeTestStateName(c.GetOrderState(floor)))
	}
	return "[" + strings.Join(values, ", ") + "]"
}

func formatHallForMergeTest(h *orders.HallOrders) string {
	rows := make([]string, 0, config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++ {
		up := mergeTestStateName(h.GetOrderState(floor, 0))
		down := mergeTestStateName(h.GetOrderState(floor, 1))
		rows = append(rows, fmt.Sprintf("[%s, %s]", up, down))
	}
	return "[" + strings.Join(rows, ", ") + "]"
}

func setCabStates(c *orders.CabOrders, states []orders.OrderState) {
	for floor, state := range states {
		c.UpdateOrderState(floor, state)
	}
}

func setHallStates(h *orders.HallOrders, states [][2]orders.OrderState) {
	for floor, statePair := range states {
		h.UpdateOrderState(floor, 0, statePair[0])
		h.UpdateOrderState(floor, 1, statePair[1])
	}
}

func assertCabEquals(t *testing.T, got *orders.CabOrders, expected []orders.OrderState) {
	t.Helper()
	for floor, want := range expected {
		if got.GetOrderState(floor) != want {
			t.Fatalf(
				"cab mismatch at floor %d: got %s, want %s",
				floor,
				mergeTestStateName(got.GetOrderState(floor)),
				mergeTestStateName(want),
			)
		}
	}
}

func assertHallEquals(t *testing.T, got *orders.HallOrders, expected [][2]orders.OrderState) {
	t.Helper()
	for floor, pair := range expected {
		gotUp := got.GetOrderState(floor, 0)
		gotDown := got.GetOrderState(floor, 1)
		if gotUp != pair[0] || gotDown != pair[1] {
			t.Fatalf(
				"hall mismatch at floor %d: got [%s, %s], want [%s, %s]",
				floor,
				mergeTestStateName(gotUp),
				mergeTestStateName(gotDown),
				mergeTestStateName(pair[0]),
				mergeTestStateName(pair[1]),
			)
		}
	}
}

func unknownHall() [][2]orders.OrderState {
	states := make([][2]orders.OrderState, config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++ {
		states[floor] = [2]orders.OrderState{orders.UnknownOrderState, orders.UnknownOrderState}
	}
	return states
}

func unknownCab() []orders.OrderState {
	states := make([]orders.OrderState, config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++ {
		states[floor] = orders.UnknownOrderState
	}
	return states
}

func copyHallStates(in [][2]orders.OrderState) [][2]orders.OrderState {
	cp := make([][2]orders.OrderState, len(in))
	copy(cp, in)
	return cp
}

func copyCabStates(in []orders.OrderState) []orders.OrderState {
	cp := make([]orders.OrderState, len(in))
	copy(cp, in)
	return cp
}

func TestMergeState_ThreeElevators_TenScenarios_PrintInputAndOutput(t *testing.T) {
	U := orders.UnknownOrderState
	R := orders.RemovedOrderState
	UN := orders.UnconfirmedOrderState
	C := orders.ConfirmedOrderState
	CP := orders.CompletedOrderState

	type scenario struct {
		name          string
		elev1Hall     [][2]orders.OrderState
		elev2Hall     [][2]orders.OrderState
		elev3Hall     [][2]orders.OrderState
		elev1Cab      []orders.OrderState
		elev2Cab      []orders.OrderState
		elev3Cab      []orders.OrderState
		incomingCabE1 []orders.OrderState
		wantHallE1    [][2]orders.OrderState
		wantCabE1     []orders.OrderState
	}

	build := func() (h [][2]orders.OrderState, c []orders.OrderState) {
		return unknownHall(), unknownCab()
	}

	h0, c0 := build()
	h1, c1 := build()
	h2, c2 := build()
	h3, c3 := build()
	h4, c4 := build()
	h5, c5 := build()
	h6, c6 := build()
	h7, c7 := build()
	h8, c8 := build()
	h9, c9 := build()

	// 1) Original mixed sample.
	h0[0] = [2]orders.OrderState{R, R}
	h0[1] = [2]orders.OrderState{UN, R}
	h0[3] = [2]orders.OrderState{U, C}
	c0 = []orders.OrderState{UN, UN, C, R}
	in0 := []orders.OrderState{UN, UN, CP, C}
	wH0 := [][2]orders.OrderState{{R, R}, {C, R}, {CP, U}, {U, CP}}
	wC0 := []orders.OrderState{C, C, CP, R}
	h0b := copyHallStates(h0)
	h0b[2] = [2]orders.OrderState{C, U}
	h0b[3] = [2]orders.OrderState{U, C}
	h0c := copyHallStates(h0)
	h0c[2] = [2]orders.OrderState{CP, U}
	h0c[3] = [2]orders.OrderState{U, CP}

	// 2) All unknown.
	in1 := []orders.OrderState{U, U, U, U}
	wH1 := copyHallStates(h1)
	wC1 := []orders.OrderState{U, U, U, U}

	// 3) Hall unconfirmed barrier and cab unconfirmed->confirmed.
	h2[0] = [2]orders.OrderState{UN, U}
	h2b := copyHallStates(h2)
	h2c := copyHallStates(h2)
	c2[0] = UN
	in2 := []orders.OrderState{U, U, U, U}
	wH2 := copyHallStates(h2)
	wH2[0] = [2]orders.OrderState{C, U}
	wC2 := []orders.OrderState{C, U, U, U}

	// 4) Hall confirmed receiving completed -> completed.
	h3[2] = [2]orders.OrderState{U, C}
	h3b := copyHallStates(h3)
	h3b[2] = [2]orders.OrderState{U, CP}
	h3c := copyHallStates(h3)
	c3[2] = C
	in3 := []orders.OrderState{U, U, CP, U}
	wH3 := copyHallStates(h3)
	wH3[2] = [2]orders.OrderState{U, CP}
	wC3 := []orders.OrderState{U, U, CP, U}

	// 5) Hall completed barrier reached -> removed.
	h4[1] = [2]orders.OrderState{CP, U}
	h4b := copyHallStates(h4)
	h4c := copyHallStates(h4)
	h4c[1] = [2]orders.OrderState{R, U}
	in4 := []orders.OrderState{U, U, U, U}
	wH4 := copyHallStates(h4)
	wH4[1] = [2]orders.OrderState{R, U}
	wC4 := []orders.OrderState{U, U, U, U}

	// 6) Hall removed + unconfirmed resurrects, then barrier blocks at unconfirmed.
	h5[3] = [2]orders.OrderState{U, R}
	h5b := copyHallStates(h5)
	h5b[3] = [2]orders.OrderState{U, UN}
	h5c := copyHallStates(h5)
	in5 := []orders.OrderState{U, U, U, U}
	wH5 := copyHallStates(h5)
	wH5[3] = [2]orders.OrderState{U, UN}
	wC5 := []orders.OrderState{U, U, U, U}

	// 7) Cab removed should ignore stale confirmed.
	c6[3] = R
	in6 := []orders.OrderState{U, U, U, C}
	wH6 := copyHallStates(h6)
	wC6 := []orders.OrderState{U, U, U, R}

	// 8) Cab completed should move to removed (owner barrier).
	c7[1] = CP
	in7 := []orders.OrderState{U, U, U, U}
	wH7 := copyHallStates(h7)
	wC7 := []orders.OrderState{U, R, U, U}

	// 9) Cab removed + unconfirmed means new order.
	c8[0] = R
	in8 := []orders.OrderState{UN, U, U, U}
	wH8 := copyHallStates(h8)
	wC8 := []orders.OrderState{UN, U, U, U}

	// 10) Mixed stress scenario.
	h9[0] = [2]orders.OrderState{UN, R}
	h9[1] = [2]orders.OrderState{CP, U}
	h9[2] = [2]orders.OrderState{U, C}
	h9b := copyHallStates(h9)
	h9b[0] = [2]orders.OrderState{UN, R}
	h9b[1] = [2]orders.OrderState{R, U}
	h9b[2] = [2]orders.OrderState{U, CP}
	h9c := copyHallStates(h9)
	h9c[0] = [2]orders.OrderState{UN, R}
	h9c[1] = [2]orders.OrderState{CP, U}
	h9c[2] = [2]orders.OrderState{U, CP}
	c9 = []orders.OrderState{UN, C, CP, R}
	in9 := []orders.OrderState{UN, CP, U, C}
	wH9 := [][2]orders.OrderState{
		{C, R},
		{R, U},
		{U, R},
		{U, U},
	}
	wC9 := []orders.OrderState{C, CP, R, R}

	scenarios := []scenario{
		{
			name:          "01_original_sample",
			elev1Hall:     h0,
			elev2Hall:     h0b,
			elev3Hall:     h0c,
			elev1Cab:      c0,
			elev2Cab:      unknownCab(),
			elev3Cab:      unknownCab(),
			incomingCabE1: in0,
			wantHallE1:    wH0,
			wantCabE1:     wC0,
		},
		{
			name:          "02_all_unknown",
			elev1Hall:     h1,
			elev2Hall:     copyHallStates(h1),
			elev3Hall:     copyHallStates(h1),
			elev1Cab:      c1,
			elev2Cab:      unknownCab(),
			elev3Cab:      unknownCab(),
			incomingCabE1: in1,
			wantHallE1:    wH1,
			wantCabE1:     wC1,
		},
		{
			name:          "03_hall_unconfirmed_barrier",
			elev1Hall:     h2,
			elev2Hall:     h2b,
			elev3Hall:     h2c,
			elev1Cab:      c2,
			elev2Cab:      unknownCab(),
			elev3Cab:      unknownCab(),
			incomingCabE1: in2,
			wantHallE1:    wH2,
			wantCabE1:     wC2,
		},
		{
			name:          "04_hall_confirmed_to_completed",
			elev1Hall:     h3,
			elev2Hall:     h3b,
			elev3Hall:     h3c,
			elev1Cab:      c3,
			elev2Cab:      unknownCab(),
			elev3Cab:      unknownCab(),
			incomingCabE1: in3,
			wantHallE1:    wH3,
			wantCabE1:     wC3,
		},
		{
			name:          "05_hall_completed_to_removed",
			elev1Hall:     h4,
			elev2Hall:     h4b,
			elev3Hall:     h4c,
			elev1Cab:      c4,
			elev2Cab:      unknownCab(),
			elev3Cab:      unknownCab(),
			incomingCabE1: in4,
			wantHallE1:    wH4,
			wantCabE1:     wC4,
		},
		{
			name:          "06_hall_removed_resurrect_unconfirmed",
			elev1Hall:     h5,
			elev2Hall:     h5b,
			elev3Hall:     h5c,
			elev1Cab:      c5,
			elev2Cab:      unknownCab(),
			elev3Cab:      unknownCab(),
			incomingCabE1: in5,
			wantHallE1:    wH5,
			wantCabE1:     wC5,
		},
		{
			name:          "07_cab_removed_ignores_stale_confirmed",
			elev1Hall:     h6,
			elev2Hall:     copyHallStates(h6),
			elev3Hall:     copyHallStates(h6),
			elev1Cab:      c6,
			elev2Cab:      unknownCab(),
			elev3Cab:      unknownCab(),
			incomingCabE1: in6,
			wantHallE1:    wH6,
			wantCabE1:     wC6,
		},
		{
			name:          "08_cab_completed_to_removed",
			elev1Hall:     h7,
			elev2Hall:     copyHallStates(h7),
			elev3Hall:     copyHallStates(h7),
			elev1Cab:      c7,
			elev2Cab:      unknownCab(),
			elev3Cab:      unknownCab(),
			incomingCabE1: in7,
			wantHallE1:    wH7,
			wantCabE1:     wC7,
		},
		{
			name:          "09_cab_removed_to_unconfirmed_on_new_press",
			elev1Hall:     h8,
			elev2Hall:     copyHallStates(h8),
			elev3Hall:     copyHallStates(h8),
			elev1Cab:      c8,
			elev2Cab:      unknownCab(),
			elev3Cab:      unknownCab(),
			incomingCabE1: in8,
			wantHallE1:    wH8,
			wantCabE1:     wC8,
		},
		{
			name:          "10_mixed_stress",
			elev1Hall:     h9,
			elev2Hall:     h9b,
			elev3Hall:     h9c,
			elev1Cab:      c9,
			elev2Cab:      unknownCab(),
			elev3Cab:      unknownCab(),
			incomingCabE1: in9,
			wantHallE1:    wH9,
			wantCabE1:     wC9,
		},
	}

	onlineNodes := []string{"Elev1", "Elev2", "Elev3"}
	for i, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			allHall := map[string]*orders.HallOrders{
				"Elev1": orders.CreateHallOrders(),
				"Elev2": orders.CreateHallOrders(),
				"Elev3": orders.CreateHallOrders(),
			}
			allCab := map[string]*orders.CabOrders{
				"Elev1": orders.CreateCabOrders(),
				"Elev2": orders.CreateCabOrders(),
				"Elev3": orders.CreateCabOrders(),
			}

			setHallStates(allHall["Elev1"], tc.elev1Hall)
			setHallStates(allHall["Elev2"], tc.elev2Hall)
			setHallStates(allHall["Elev3"], tc.elev3Hall)

			setCabStates(allCab["Elev1"], tc.elev1Cab)
			setCabStates(allCab["Elev2"], tc.elev2Cab)
			setCabStates(allCab["Elev3"], tc.elev3Cab)

			fmt.Printf("\n===== Scenario %d: %s =====\n", i+1, tc.name)
			fmt.Println("======= input ========")
			for _, id := range onlineNodes {
				fmt.Printf("%s: Caborder = %s\n", id, formatCabForMergeTest(allCab[id]))
				fmt.Printf("%s: Hallorder = %s\n", id, formatHallForMergeTest(allHall[id]))
			}

			for floor := 0; floor < config.NumFloors; floor++ {
				for dir := 0; dir < 2; dir++ {
					for _, senderID := range []string{"Elev2", "Elev3"} {
						update := HallOrderUpdate{
							SenderID:  senderID,
							Floor:     floor,
							OrderType: dir,
							State:     allHall[senderID].GetOrderState(floor, dir),
						}
						next := mergeHallOrderState(update, "Elev1", allHall, onlineNodes)
						allHall["Elev1"].UpdateOrderState(floor, dir, next)
					}
				}
			}

			for floor := 0; floor < config.NumFloors; floor++ {
				update := CabOrderUpdate{
					SenderID: "Elev1",
					Floor:    floor,
					State:    tc.incomingCabE1[floor],
				}
				next := mergeCabOrderState(update, allCab, onlineNodes)
				allCab["Elev1"].UpdateOrderState(floor, next)
			}

			fmt.Println("======= Output ======")
			fmt.Printf("Merged hallorder: %s\n", formatHallForMergeTest(allHall["Elev1"]))
			fmt.Printf("Merged caborder: %s\n", formatCabForMergeTest(allCab["Elev1"]))

			assertHallEquals(t, allHall["Elev1"], tc.wantHallE1)
			assertCabEquals(t, allCab["Elev1"], tc.wantCabE1)
		})
	}
}

// stateGetFunc builds a getState callback from a fixed map.
// Any id not in the map returns (UnknownOrderState, false), causing barrierReached to return false.
func stateGetFunc(m map[string]orders.OrderState) func(string) (orders.OrderState, bool) {
	return func(id string) (orders.OrderState, bool) {
		s, ok := m[id]
		return s, ok
	}
}

// TestMergeState_DiagramVerification exhaustively checks that mergeState()
// implements every arrow and self-loop in the state diagram:
//
//	Unknown  ──────────────────→  all states (takes newOrder directly)
//	Removed  + Unconfirmed     →  Unconfirmed
//	Removed  + else            →  Removed  (self-loop)
//	Unconfirmed + barrier      →  Confirmed
//	Unconfirmed + else         →  Unconfirmed  (self-loop)
//	Confirmed   + Completed    →  Completed
//	Confirmed   + else         →  Confirmed  (self-loop)
//	Completed   + barrier      →  Removed
//	Completed   + else         →  Completed  (self-loop)
func TestMergeState_DiagramVerification(t *testing.T) {
	U := orders.UnknownOrderState
	R := orders.RemovedOrderState
	UN := orders.UnconfirmedOrderState
	C := orders.ConfirmedOrderState
	CP := orders.CompletedOrderState

	noBarrier := func(string) (orders.OrderState, bool) { return U, false }

	type tc struct {
		name        string
		local       orders.OrderState
		newOrder    orders.OrderState
		onlineNodes []string
		getState    func(string) (orders.OrderState, bool)
		want        orders.OrderState
	}

	tests := []tc{
		// ── Unknown: always takes newOrder ──────────────────────────────────
		{"Unknown/newOrder=Unknown    → Unknown", U, U, nil, nil, U},
		{"Unknown/newOrder=Removed   → Removed", U, R, nil, nil, R},
		{"Unknown/newOrder=Unconfirmed→Unconfirmed", U, UN, nil, nil, UN},
		{"Unknown/newOrder=Confirmed → Confirmed", U, C, nil, nil, C},
		{"Unknown/newOrder=Completed → Completed", U, CP, nil, nil, CP},

		// ── Removed: only Unconfirmed restarts the cycle ─────────────────
		{"Removed/newOrder=Unconfirmed→Unconfirmed", R, UN, nil, noBarrier, UN},
		{"Removed/newOrder=Unknown   → Removed", R, U, nil, noBarrier, R},
		{"Removed/newOrder=Removed   → Removed", R, R, nil, noBarrier, R},
		{"Removed/newOrder=Confirmed → Removed", R, C, nil, noBarrier, R},
		{"Removed/newOrder=Completed → Removed", R, CP, nil, noBarrier, R},

		// ── Unconfirmed: barrier → Confirmed; no barrier → stay ──────────
		// barrier reached: all nodes at Unconfirmed
		{
			"Unconfirmed/barrier(all=UN)         → Confirmed",
			UN, UN, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": UN, "B": UN}),
			C,
		},
		// barrier reached: all nodes at Confirmed (>= Unconfirmed still passes)
		{
			"Unconfirmed/barrier(all=C)          → Confirmed",
			UN, C, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": C, "B": C}),
			C,
		},
		// barrier reached: one node already at Completed (>= Unconfirmed passes)
		{
			"Unconfirmed/barrier(mix=UN,CP)      → Confirmed",
			UN, UN, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": UN, "B": CP}),
			C,
		},
		// barrier NOT reached: one node at Removed (< Unconfirmed)
		{
			"Unconfirmed/no-barrier(one=R)       → Unconfirmed",
			UN, UN, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": UN, "B": R}),
			UN,
		},
		// barrier NOT reached: one node at Unknown (< Unconfirmed)
		{
			"Unconfirmed/no-barrier(one=U)       → Unconfirmed",
			UN, UN, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": UN, "B": U}),
			UN,
		},
		// barrier NOT reached: one node absent from getState (ok=false)
		{
			"Unconfirmed/no-barrier(B-absent)    → Unconfirmed",
			UN, UN, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": UN}),
			UN,
		},
		// newOrder is irrelevant once barrier is reached
		{
			"Unconfirmed/barrier+newOrder=R      → Confirmed (newOrder ignored)",
			UN, R, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": UN, "B": UN}),
			C,
		},
		{
			"Unconfirmed/barrier+newOrder=CP     → Confirmed (newOrder ignored)",
			UN, CP, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": UN, "B": UN}),
			C,
		},

		// ── Confirmed: only Completed advances the state ──────────────────
		{"Confirmed/newOrder=Completed → Completed", C, CP, nil, noBarrier, CP},
		{"Confirmed/newOrder=Unknown   → Confirmed", C, U, nil, noBarrier, C},
		{"Confirmed/newOrder=Removed   → Confirmed", C, R, nil, noBarrier, C},
		{"Confirmed/newOrder=Unconfirmed→ Confirmed", C, UN, nil, noBarrier, C},
		{"Confirmed/newOrder=Confirmed  → Confirmed", C, C, nil, noBarrier, C},

		// ── Completed: barrier → Removed; no barrier → stay ─────────────
		// barrier reached: all at Completed
		{
			"Completed/barrier(all=CP)           → Removed",
			CP, CP, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": CP, "B": CP}),
			R,
		},
		// barrier reached: Removed counts as "already passed through Completed"
		{
			"Completed/barrier(mix=CP,R)         → Removed",
			CP, CP, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": CP, "B": R}),
			R,
		},
		{
			"Completed/barrier(all=R)            → Removed",
			CP, R, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": R, "B": R}),
			R,
		},
		// barrier NOT reached: one node at Confirmed
		{
			"Completed/no-barrier(one=C)         → Completed",
			CP, CP, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": CP, "B": C}),
			CP,
		},
		// barrier NOT reached: one node at Unconfirmed
		{
			"Completed/no-barrier(one=UN)        → Completed",
			CP, CP, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": CP, "B": UN}),
			CP,
		},
		// barrier NOT reached: one node at Unknown
		{
			"Completed/no-barrier(one=U)         → Completed",
			CP, CP, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": CP, "B": U}),
			CP,
		},
		// newOrder is irrelevant when in Completed state
		{
			"Completed/no-barrier+newOrder=R     → Completed (newOrder ignored)",
			CP, R, []string{"A", "B"},
			stateGetFunc(map[string]orders.OrderState{"A": CP, "B": C}),
			CP,
		},
	}

	allStates := []orders.OrderState{U, R, UN, C, CP}
	_ = allStates

	fmt.Println("\n========== TestMergeState_DiagramVerification ==========")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeState(tt.newOrder, tt.local, tt.onlineNodes, tt.getState)
			status := "PASS"
			if got != tt.want {
				status = "FAIL"
			}
			fmt.Printf("  [%s] local=%-12s newOrder=%-12s → got=%-12s want=%s\n",
				status,
				mergeTestStateName(tt.local),
				mergeTestStateName(tt.newOrder),
				mergeTestStateName(got),
				mergeTestStateName(tt.want),
			)
			if got != tt.want {
				t.Errorf("mergeState(newOrder=%s, local=%s): got %s, want %s",
					mergeTestStateName(tt.newOrder),
					mergeTestStateName(tt.local),
					mergeTestStateName(got),
					mergeTestStateName(tt.want),
				)
			}
		})
	}
}
