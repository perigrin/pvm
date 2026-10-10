// ABOUTME: Tests for multi candidates with an invocant colon: how selection and the ambiguity check read the slot.
// ABOUTME: A call either writes the leading slot with no comma after it or does not; candidates compare per kind of call.
package types

import "testing"

// withInvocant is s with the invocant inv, `(Code $c: List @l)`.
func withInvocant(inv Param, s Signature) Signature {
	s.Invocant = &inv
	return s
}

// list is a slurpy `List @name`, which accepts zero arguments.
func list(name string) Param {
	return Param{Name: name, Sigil: '@', Type: List}
}

// TestAmbiguityComparesInvocants: RFC 0001 "Builtins that keep their own
// parse". A call either writes an invocant, `sort byname @x`, or does not,
// `sort @x`. Two candidates share a call only when both take the same
// kind: a required invocant takes only calls that write one, no invocant
// only calls that write none, and an optional one, print's `FileHandle $fh
// = select()`, both. Where both write one, the invocant's type is compared
// as the first argument's.
func TestAmbiguityComparesInvocants(t *testing.T) {
	sortBy := withInvocant(scalar("by", Code|Str), sig(List, list("list")))
	sortPlain := sig(List, Param{Name: "list", Sigil: '@', Type: List, Required: true})
	for name, cands := range map[string][]Signature{
		"disjoint invocants": {
			withInvocant(scalar("c", Code), sig(List, list("l"))),
			withInvocant(scalar("h", Hash), sig(List, list("l"))),
		},
		"a narrower invocant": {
			withInvocant(scalar("a", Int), sig(List, list("l"))),
			withInvocant(scalar("a", Num), sig(List, list("l"))),
		},
		"sort's invocant beside its plain list": {sortBy, sortPlain},
	} {
		if err := Ambiguity(cands); err != nil {
			t.Errorf("%s: got %v, want no ambiguity", name, err)
		}
	}

	same := withInvocant(scalar("c", Code), sig(List, list("l")))
	want := "candidates (Code $c: List @l) and (Code $c: List @l) are ambiguous for (Code)"
	if err := Ambiguity([]Signature{same, same}); err == nil || err.Error() != want {
		t.Errorf("identical invocants: got %v, want %q", err, want)
	}

	optional := Param{Name: "fh", Sigil: '$', Type: FileHandle, Default: "select()"}
	printLike := []Signature{withInvocant(optional, sig(Boolean, list("a"))), sig(Boolean, list("a"))}
	want = "candidates (FileHandle $fh: List @a) and (List @a) are ambiguous for ()"
	if err := Ambiguity(printLike); err == nil || err.Error() != want {
		t.Errorf("optional invocant beside none: got %v, want %q", err, want)
	}
}

// TestSelectRequiredInvocantTakesNoBareCall: Select's arguments are a
// call's comma-separated ones, so the call it is given writes no invocant.
// A candidate whose invocant is required, `(Code|Str $by: List @list)`,
// takes no such call; one whose invocant has a default, print's, does.
func TestSelectRequiredInvocantTakesNoBareCall(t *testing.T) {
	required := []Signature{withInvocant(scalar("by", Code|Str), sig(List, list("list")))}
	wantFailed(t, required, nil)
	wantFailed(t, required, []Type{List})

	optional := Param{Name: "fh", Sigil: '$', Type: FileHandle, Default: "select()"}
	wantSelected(t, []Signature{withInvocant(optional, sig(Boolean, list("a")))}, []Type{Str}, 0, Boolean)
}
