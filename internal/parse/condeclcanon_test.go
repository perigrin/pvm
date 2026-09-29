// ABOUTME: A declaration in a condition keeps its parentheses in canon; only a
// ABOUTME: foreach's loop variable, which the parser marks, sits outside them.

package parse_test

import (
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestConditionDeclarationCanon: canon took ANY Declaration child of a
// Conditional or Loop for a foreach loop variable and wrote it without
// parentheses, so a condition that declares -- and a foreach list whose one
// element is an anonymous sub -- lost them, at Unknown=0. Measured on 5.42.0:
//
//	$ perl -MO=Deparse -e 'while (my $l = <F>) { 1 } if (my $x = f()) { 1 }
//	      for my $y (sub {}) { 6 }'
//	while (defined(my $l = readline F)) {
//	if (my $x = f()) {
//	foreach my $y (sub {
func TestConditionDeclarationCanon(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"while (my $l = <F>) { 1 }", "while (my $l = <F>) {1;}"},
		{"if (my $x = f()) { 1 }", "if (my $x = f()) {1;}"},
		{"for my $y (sub {}) { 6 }", "for my $y (sub {}) {6;}"},
		{"for $y (sub {}) { 6 }", "for $y (sub {}) {6;}"},
		{"for my $z (@l) { 6 }", "for my $z (@l) {6;}"},
	} {
		root := parse.Parse([]byte(tc.src))
		if containsKind(root, parse.Unknown) {
			t.Errorf("%q: perl accepts this; got %s", tc.src, shape(root))
			continue
		}
		if got := strings.TrimSpace(parse.Canon(root, []byte(tc.src))); got != tc.want {
			t.Errorf("%q: canon %q, want %q", tc.src, got, tc.want)
		}
	}
}
