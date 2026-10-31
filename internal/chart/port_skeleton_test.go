// ABOUTME: Tests the port skeleton: the six start forms after the program form, use/no statements, the signature start, optexpr and optlistexpr.
// ABOUTME: Snippets under testdata/port/skeleton are recognized through internal/lexer, chart/lex and the recognizer.

package chart

import (
	"os"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
)

// skeletonLines are the perly.y alternatives the skeleton ports or fits, LEX's core aside.
var skeletonLines = []int{198, 208, 221, 234, 247, 257, 714, 899, 1023, 1170, 1185, 1787, 1788, 1792, 1793}

func TestPortSkeleton(t *testing.T) {
	if err := grammar.CheckFamilyMarked(skeletonLines, nil); err != nil {
		t.Error(err)
	}
	for _, f := range snippets(t, "skeleton") {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		if v := recognizeSource(t, src); !v.accepted {
			t.Errorf("%s is refused at byte %d (expected %v); perl accepts it", f, v.at, v.expected)
		}
	}
}

// The pinned perl accepts the first four and refuses the two nameless ones at their ;.
func TestPortSkeletonUse(t *testing.T) {
	for _, src := range []string{"use strict;", "no warnings;", "use feature 'say';", "use lib 'a', 'b';"} {
		if v := recognizeSource(t, []byte(src)); !v.accepted {
			t.Errorf("%q is refused at byte %d (expected %v)", src, v.at, v.expected)
		}
	}
	for _, src := range []string{"use ;", "no ;"} {
		if v := recognizeSource(t, []byte(src)); v.accepted {
			t.Errorf("%q, with no module name, is accepted", src)
		} else if want := strings.Index(src, ";"); v.at != want {
			t.Errorf("%q is refused at byte %d, want %d", src, v.at, want)
		}
	}
}
