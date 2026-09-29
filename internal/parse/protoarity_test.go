// ABOUTME: A prototype is unary when it has exactly ONE slot, mandatory or optional;
// ABOUTME: `($;$)` and `(;$$)` take the whole list, as perl parses them.
package parse

import "testing"

// TestPrototypeArity holds perl's line between a named unary and a list
// operator. ShapeOf counted only the MANDATORY slots, so `($;$)` read as unary
// and a parenless call lost every argument past the first. Measured on 5.42.0
// with `perl -MO=Deparse,-p -e 'sub f (P) {} my ($a,$b); f $a, $b;'`:
//
//	$  ;$  _  *  \@  ;\@        (f($a), $b)     unary
//	$;$  ;$$  \@;$              f($a, $b)       list
func TestPrototypeArity(t *testing.T) {
	cases := []struct {
		proto string
		want  Shape
	}{
		{"($)", ShapeUnary},
		{"(;$)", ShapeUnary},
		{"(_)", ShapeUnary},
		{"(*)", ShapeUnary},
		{`(\@)`, ShapeUnary},
		{`(;\@)`, ShapeUnary},
		{"($;$)", ShapeList},
		{"(;$$)", ShapeList},
		{`(\@;$)`, ShapeList},
		{`(\[$%@];$)`, ShapeList},
		{"()", ShapeNiladic},
	}
	for _, c := range cases {
		if got := ShapeOf(c.proto); got != c.want {
			t.Errorf("ShapeOf(%q) = %v, want %v", c.proto, got, c.want)
		}
	}
}
