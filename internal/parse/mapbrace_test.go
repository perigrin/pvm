// ABOUTME: map's and grep's brace selects a variant: the comma ending the first argument makes it an expression.
// ABOUTME: perl's first-tokens guess picks the same variant on every program perl accepts, and refuses the rest.
package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// mapGrepCall returns the map or grep call of a one-statement source, or
// nil when the statement is not one.
func mapGrepCall(n *parse.Node) *parse.Node {
	if len(n.Children) != 1 || n.Children[0].Kind != parse.Statement ||
		len(n.Children[0].Children) != 1 {
		return nil
	}
	call := n.Children[0].Children[0]
	if call.Kind != parse.Call || (call.Text != "map" && call.Text != "grep") {
		return nil
	}
	return call
}

// firstArgIsAnonHash reports whether the leftmost term of the call's
// argument is an anonymous hash: the expression variant, whose first
// argument a hash constructor begins.
func firstArgIsAnonHash(call *parse.Node) bool {
	n := call
	for len(n.Children) > 0 {
		n = n.Children[0]
		if n.Kind == parse.AnonHash {
			return true
		}
		if n.Kind == parse.Block {
			return false
		}
	}
	return false
}

// TestMapGrepCommaSelectsVariant: the comma that ends the first argument
// selects the expression variant, its absence the block variant, across
// brace contents perl guesses either way. Measured on 5.42.0 with
// -MO=Deparse, for map and grep alike:
//
//	map { $_ => 1 } @a              map {$_, 1;} @a          a block holding pairs
//	map { lc($_) => 1 } @a          map {lc $_, 1;} @a
//	map {; "a" => 1 } @a            map {'a', 1;} @a
//	map { "a" => 1 }, @a            map {'a', 1}, @a         a hash constructor
//	map {}, @a                      map {}, @a
//	map { foo => 1 }->{foo}, @a     map {'foo', 1}->{'foo'}, @a
//	map { "a" => 1 } ? 1 : 2, @a    map(({'a', 1} ? 1 : 2), @a)
//	map {}->{a}, @a                 map {}->{'a'}, @a
//
// The last three hold the comma back past a postfix or an operator: the
// comma that decides is the one ENDING the first argument, not one that
// must follow `}` directly.
func TestMapGrepCommaSelectsVariant(t *testing.T) {
	blocks := []string{
		`{ $_ => 1 } @a`,
		`{ lc($_) => 1 } @a`,
		`{ 1 } @a`,
		`{; "a" => 1 } @a`,
		`{ $_, 1 } @a`,
	}
	exprs := []string{
		`{ "a" => 1 }, @a`,
		`{ a => 1 }, @a`,
		`{ "\L$_" => 1 }, @a`,
		`{ "a", 1 }, @a`,
		`{}, @a`,
		`{} => @a`,
		`{ foo => 1 }->{foo}, @a`,
		`{ "a" => 1 } ? 1 : 2, @a`,
		`{}->{a}, @a`,
	}
	for _, op := range []string{"map", "grep"} {
		for _, body := range blocks {
			src := op + " " + body + ";"
			t.Run(src, func(t *testing.T) {
				n := parse.Parse([]byte(src))
				if got := countUnknown(n); got != 0 {
					t.Fatalf("Parse(%q) has %d Unknown, want 0", src, got)
				}
				call := mapGrepCall(n)
				if call == nil || len(call.Children) != 2 || call.Children[0].Kind != parse.Block {
					t.Errorf("Parse(%q): want the block variant, a Block then the list", src)
				}
			})
		}
		for _, body := range exprs {
			src := op + " " + body + ";"
			t.Run(src, func(t *testing.T) {
				n := parse.Parse([]byte(src))
				if got := countUnknown(n); got != 0 {
					t.Fatalf("Parse(%q) has %d Unknown, want 0", src, got)
				}
				call := mapGrepCall(n)
				if call == nil || len(call.Children) != 1 || !firstArgIsAnonHash(call) {
					t.Errorf("Parse(%q): want the expression variant, an anon hash first argument", src)
				}
			})
		}
	}
}

// TestMapGrepAnonHashBeforeComma: forms perl accepts as the expression
// variant stay accepted, an empty brace among them. Measured on 5.42.0
// with -MO=Deparse:
//
//	map {}, @a                map {}, @a
//	grep { "a" => 1 }, @a     grep {'a', 1}, @a
//	map({}, @a)               map {}, @a
//	grep({}, @a)              grep {}, @a
func TestMapGrepAnonHashBeforeComma(t *testing.T) {
	for _, src := range []string{
		`map {}, @a;`,
		`grep {}, @a;`,
		`grep { "a" => 1 }, @a;`,
		`map({}, @a);`,
		`grep({}, @a);`,
	} {
		t.Run(src, func(t *testing.T) {
			n := parse.Parse([]byte(src))
			if got := countUnknown(n); got != 0 {
				t.Fatalf("Parse(%q) has %d Unknown, want 0", src, got)
			}
			call := mapGrepCall(n)
			if call == nil || !firstArgIsAnonHash(call) {
				t.Errorf("Parse(%q): want an anon hash first argument", src)
			}
		})
	}
}
