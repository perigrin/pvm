// ABOUTME: Pins on perly.y that the slipper's zero-width ; is never judged away: no reduce before it, and no shift
// ABOUTME: of it, has a forbidding verdict in the state-free tables (ReduceAt, ShiftForbidden).

package prectab

import (
	"slices"
	"testing"
)

// semicolon is PERLY_SEMICOLON's symbol id in p's tables.
func semicolon(t *testing.T, p *Precedence) int {
	t.Helper()
	for id, s := range p.Tab.Symbols {
		if s.Name == "PERLY_SEMICOLON" && s.Terminal {
			return id
		}
	}
	t.Fatal("perly.y's tables have no PERLY_SEMICOLON terminal")
	return -1
}

// TestSlipperSemicolonReduceNeverForbidden: the slipper's inserted ; closes a set's items over completion, and any
// rule ending before it would be judged by ReducesFree(r, ;, pi). Considered: every rule of the tables, the rules a
// forest can derive, unfiltered by lookahead, since ReducesFree reads ReduceAt for any rule; the refused rules are
// left out because no forest derives them. Not allow: a ReduceAt of 1 (forbidden) or 2 (the parent item decides,
// so some parent forbids it).
func TestSlipperSemicolonReduceNeverForbidden(t *testing.T) {
	_, tab, p := committed(t)
	semi := semicolon(t, p)
	forbidden, considered := 0, 0
	for r := range tab.Rules {
		considered++ // counted as judged, so a narrowed loop is seen rather than reported as the whole
		if p.ReduceAt[r*p.NT+semi] != 0 {
			forbidden++
			t.Errorf("reducing rule %d (%s) before PERLY_SEMICOLON has verdict %d, not allow", r, tab.Symbols[tab.Rules[r].LHS].Name, p.ReduceAt[r*p.NT+semi])
		}
	}
	t.Logf("semicolon reduces forbidden: %d of %d", forbidden, considered)
	// Pinned exactly, as TestPrectabItemMapping pins its populations: a narrowed population is seen, not passed.
	if considered != 350 {
		t.Errorf("%d rules considered; want perly.y's 350", considered)
	}
}

// TestSlipperSemicolonShiftNeverForbidden: the slipper advances every item of a set whose state has a goto on ;,
// directly (x: a . ; b) or through a predicted rule that begins with it, which the forest judges as
// ShiftForbidden(top, at, ;) at the kernel item at whose next symbol derives a string beginning with ;.
// Can shift: a (state, item) pair of the tables' LR(0) states whose item has its dot at 1 or past and ; in FIRST of
// the symbol after it (a terminal's FIRST is itself); a dot-0 item is left out because the judgement never asks
// ShiftForbidden there (it shifts under the item that predicted it). Forbidden: ShiftForbidden under any context an
// item can carry, none or any spine top with a row, which over-approximates the tops that reach that state.
func TestSlipperSemicolonShiftNeverForbidden(t *testing.T) {
	_, tab, p := committed(t)
	semi := semicolon(t, p)
	var tops []int32
	for top, row := range p.TopAt {
		if row >= 0 {
			tops = append(tops, int32(top))
		}
	}
	forbidden, canShift := 0, 0
	for s, st := range tab.States {
		for _, i := range st.Items {
			it := tab.Items[i]
			rhs := tab.Rules[it.Rule].RHS
			if it.Dot == 0 || it.Dot >= len(rhs) || !slices.Contains(tab.First[rhs[it.Dot]], semi) {
				continue
			}
			canShift++
			at := p.Item(i)
			if slices.ContainsFunc(tops, func(top int32) bool { return p.ShiftForbidden(top, at, int32(semi)) }) {
				forbidden++
				t.Errorf("state %d: shifting PERLY_SEMICOLON after item %d (rule %d, dot %d) is forbidden", s, i, it.Rule, it.Dot)
			}
		}
	}
	t.Logf("semicolon shifts forbidden: %d of %d", forbidden, canShift)
	t.Logf("spine tops judged under: %d", len(tops))
	// Pinned exactly, as TestPrectabItemMapping pins its populations: a narrowed population is seen, not passed.
	if canShift != 61 || len(tops) != 46 {
		t.Errorf("%d (state, item) pairs can shift PERLY_SEMICOLON under %d spine tops; want perly.y's 61 under 46", canShift, len(tops))
	}
}
