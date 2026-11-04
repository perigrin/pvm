// ABOUTME: Tests that the x operator's classes are derived from the classes that complete a term, not from neighbour lists.
// ABOUTME: toke.c's KEY_x reads x as repetition exactly when PL_expect is XOPERATOR, which a term-completing token leaves.

package grammar

import (
	"slices"
	"testing"

	"tamarou.com/pvm/internal/lexer"
)

func repeatClasses() (out []TokenClass) {
	for _, d := range perlyClasses {
		if d.Terminal == "MULOP" && d.Key.Text != nil && *d.Key.Text == "x" {
			out = append(out, d.Key)
		}
	}
	return out
}

func hasRepeat(prev1, prev2 *Neighbour) bool {
	return slices.ContainsFunc(repeatClasses(), func(c TokenClass) bool {
		eq := func(a, b *Neighbour) bool { return a == nil && b == nil || a != nil && b != nil && *a == *b }
		return eq(c.Prev1, prev1) && eq(c.Prev2, prev2)
	})
}

func TestRepeatClassesFollowCompletes(t *testing.T) {
	for _, c := range repeatClasses() {
		if c.NotPrev1 != nil {
			t.Errorf("a MULOP x class names %d words it must not follow", len(c.NotPrev1))
		}
	}
	// What completes a term, from toke.c: a variable, number or string; a closing bracket; a
	// no-argument builtin; a bareword; a method name; a name called with &.
	for name, want := range map[string]struct{ prev1, prev2 *Neighbour }{
		"variable":    {&Neighbour{Kind: lexer.Variable}, nil},
		"number":      {&Neighbour{Kind: lexer.Number}, nil},
		"close paren": {&Neighbour{Kind: lexer.CloseBracket, Text: ")"}, nil},
		"post-inc":    {&Neighbour{Kind: lexer.Operator, Text: "++"}, nil},
		"time":        {&Neighbour{Kind: lexer.Word, Text: "time"}, nil},
		"__PACKAGE__": {&Neighbour{Kind: lexer.Word, Text: "__PACKAGE__"}, nil},
		"bareword":    {&Neighbour{Kind: lexer.Word, Text: OtherWord}, nil},
		"method name": {nil, &Neighbour{Kind: lexer.Operator, Text: "->"}},
		"& name":      {nil, &Neighbour{Kind: lexer.FuncSigil}},
		"glob name":   {nil, &Neighbour{Kind: lexer.Operator, Text: "*"}},
	} {
		if !hasRepeat(want.prev1, want.prev2) {
			t.Errorf("no MULOP x class after a %s", name)
		}
	}
	// What waits for its argument does not: x after it is a word.
	for _, w := range []string{"ref", "shift", "print", "push", "last", "not", "require", "return"} {
		if hasRepeat(&Neighbour{Kind: lexer.Word, Text: w}, nil) {
			t.Errorf("a MULOP x class follows %s, which waits for a term", w)
		}
	}
	for _, d := range perlyClasses {
		if d.Key.Completes && slices.Contains([]string{"UNIOP", "FUNC1", "LSTOP", "BLKLSTOP", "LOOPEX", "NOTOP", "KW_REQUIRE", "PERLY_COMMA"}, d.Terminal) {
			t.Errorf("%s is marked as completing a term", d.Terminal)
		}
	}
}

// A { class conditioned on its neighbours keeps its reach: only the one that follows anything is cut
// off from the list operators that read a { as a hash, so a later class can restore `sub open {`.
func TestLimitBlockBraceSparesConditionedClasses(t *testing.T) {
	brace := TokenClass{Kind: ptr(lexer.Operator), Text: ptr("{"), OpensBlock: ptr(true)}
	afterSub := brace
	afterSub.Prev2 = &Neighbour{Kind: lexer.Word, Text: "sub"}
	afterWord := brace
	afterWord.Prev1 = &Neighbour{Kind: lexer.Word, Text: "do"}
	beforeWord := brace
	beforeWord.Next = &Neighbour{Kind: lexer.Word, Text: "x"}
	classes := []ClassDecl{
		{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("open")}},
		{"LSTOP", TokenClass{Kind: ptr(lexer.Word), Text: ptr("print")}},
		{"PERLY_BRACE_OPEN", brace}, {"PERLY_BRACE_OPEN", afterSub}, {"PERLY_BRACE_OPEN", afterWord}, {"PERLY_BRACE_OPEN", beforeWord},
	}
	limitBlockBrace(classes)
	if got := classes[2].Key.NotPrev1; !slices.Equal(got, []Neighbour{{Kind: lexer.Word, Text: "open"}}) {
		t.Errorf("the unconditioned { excludes %v, want open alone (print takes a block)", got)
	}
	for _, i := range []int{3, 4, 5} {
		if classes[i].Key.NotPrev1 != nil {
			t.Errorf("the conditioned { class %d excludes %v", i-2, classes[i].Key.NotPrev1)
		}
	}
}
