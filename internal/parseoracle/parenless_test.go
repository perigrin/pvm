// ABOUTME: The declared/undeclared boundary for a parenless call, scored on perl itself.
// ABOUTME: `sub ok {} ok 8, 9` compiles and bare `ok 8, 9` does not — the discriminating pair.

package parseoracle_test

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parseoracle"
)

// TestParenlessCallBoundary asks perl where the line is between a parenless
// call and a syntax error, and pins the answer.
//
// THE PAIR IS THE POINT. The two sources differ by the declaration alone --
// the call site `ok 8, 9;` is byte-identical in both -- and perl compiles one
// and refuses the other. That is the whole rule a parser needs: "declared
// before the call site", and nothing at the call site can answer it.
//
// WHY THIS IS AN ORACLE FIXTURE rather than a parse test. `internal/parse`
// tests what our parser does with these sources, which is only half the
// claim; the other half is that perl draws the line exactly there, and only
// perl can say so. Our parser cannot be scored `wider` or WRONG against a
// boundary nobody measured.
//
// WHY `Facts.OK` AND NOT A Verdict. CompareFacts returns BucketNoAnswer the
// moment `!oracle.OK` -- deliberately, so a subject is never blamed for
// perl's refusal (compare_facts.go:37-41). That makes a Verdict silent on
// exactly the question this test asks, so the assertion is on perl's own
// compile result.
func TestParenlessCallBoundary(t *testing.T) {
	if _, err := exec.LookPath("perl"); err != nil {
		t.Skipf("perl unavailable: %v", err)
	}

	cases := []struct {
		name string
		src  string
		want bool // whether perl compiles it
		// why names the rule the case measures, so a failure says which
		// half of the boundary moved.
		why string
	}{
		{
			// The declaration is ABOVE the call, so `ok` is a known sub and
			// `8, 9` is its argument list.
			name: "declared above the call site",
			src:  "sub ok { 1 }\nok 8, 9;\n",
			want: true,
			why:  "a sub declared earlier in the file makes a parenless call to it a call",
		},
		{
			// The same call site with no declaration at all. perl's message
			// is `Number found where operator expected (Do you need to
			// predeclare "ok"?)`, then a syntax error: there is no call here
			// for an argument extent to be measured on.
			name: "undeclared",
			src:  "ok 8, 9;\n",
			want: false,
			why:  "an undeclared callee with a numeric argument is a syntax error, not a call",
		},
		{
			// The declaration is BELOW the call. perl parses top to bottom,
			// so this refuses exactly as the undeclared case does -- which is
			// why a whole-file pre-pass over declarations would be wrong.
			name: "declared below the call site",
			src:  "ok 8, 9;\nsub ok { 1 }\n",
			want: false,
			why:  "perl reads top to bottom, so a declaration below the call does not reach it",
		},
		{
			// A bodiless forward declaration is a declaration: it predeclares
			// the name and nothing else.
			name: "a forward declaration is enough",
			src:  "sub ok;\nok 8, 9;\n",
			want: true,
			why:  "predeclaration is what the error message asks for, and it suffices",
		},
		{
			// A SCALAR first argument is a different shape entirely, and the
			// surprise of the set: with no declaration perl reads `ok $x, 9`
			// as an INDIRECT OBJECT method call, `$x->ok`. It compiles, and
			// it is not the call the source looks like -- which is why the
			// numeric argument is what the refusing cases use.
			name: "an undeclared callee with a scalar argument is an indirect object",
			src:  "my $x; ok $x, 9;\n",
			want: true,
			why:  "`ok $x, 9` is `$x->ok`, a method call on the scalar, not a call to ok",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			facts, err := parseoracle.Ask(context.Background(),
				[]byte(c.src), parseoracle.Options{})
			if err != nil {
				t.Fatalf("Ask: %v", err)
			}
			if facts.OK != c.want {
				t.Fatalf("perl compiled = %v, want %v\nsource:\n%s\nrule: %s\nstderr: %s",
					facts.OK, c.want, c.src, c.why, facts.Stderr)
			}
			// A refusal must be a SYNTAX error, not an environmental one. A
			// missing module or a failed BEGIN would also set OK false and
			// would measure nothing about the boundary.
			if !c.want && !strings.Contains(facts.Stderr, "syntax error") {
				t.Errorf("want a syntax error, got stderr:\n%s\nrule: %s",
					facts.Stderr, c.why)
			}
		})
	}
}
