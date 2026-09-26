// ABOUTME: IsTrivia is the one authority on which kinds carry no syntax.
// ABOUTME: Pins the set against every Kind, so a new kind cannot join silently.
package lexer

import "testing"

// TestIsTriviaCoversEveryKind pins the trivia set against the whole Kind
// enumeration, so adding a Kind forces a decision about it here.
//
// This set was restated in four places and diverged: when `DataSection` joined
// it, three copies were updated and `use_test.go`'s was not. A TIER_2 gate
// found that, not the suite -- the drift was latent because `firstWordOf`'s
// walk cannot reach a data section today.
//
// Enumerating rather than spot-checking is the point. A test naming only the
// kinds that are trivia passes by accident when a kind is added, which is the
// failure mode the four copies already demonstrated once.
func TestIsTriviaCoversEveryKind(t *testing.T) {
	// Every kind, and what it is. A new Kind added to the lexer fails this
	// test until someone writes its answer down.
	want := map[Kind]bool{
		Whitespace:   true,
		Comment:      true,
		Pod:          true,
		DataSection:  true,
		Error:        false,
		UnknownRest:  false,
		Quote:        false,
		Variable:     false,
		Word:         false,
		Number:       false,
		Operator:     false,
		Semicolon:    false,
		Readline:     false,
		FuncSigil:    false,
		DerefSigil:   false,
		Prototype:    false,
		CloseBracket: false,
		HeredocOpen:  false,
		HeredocBody:  false,
		FormatBody:   false,
	}

	for k, isTrivia := range want {
		if got := IsTrivia(k); got != isTrivia {
			t.Errorf("IsTrivia(%s) = %v, want %v", k, got, isTrivia)
		}
	}

	// And the enumeration itself must be complete: walk every kind value the
	// String() method names and require an entry above. A kind with no entry
	// is a kind nobody decided about.
	for k := Kind(0); k < Kind(64); k++ {
		name := k.String()
		if name == "Kind(?)" {
			continue
		}
		if _, ok := want[k]; !ok {
			t.Errorf("kind %s (%d) has no entry in this test; decide whether "+
				"it is trivia and write it down", name, int(k))
		}
	}
}
