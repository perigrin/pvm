// ABOUTME: Tests that no caller of Parse or ParseWith can turn precedence off (F3-7): no parameter or option is a
// ABOUTME: switch, ParseWith panics on a recognizer judging other than New derives, and Parse never enters Constrain.

package forest

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
	"tamarou.com/pvm/internal/chart/prectab"
	"tamarou.com/pvm/internal/chart/recognize"
)

// parseWithPanics reports whether ParseWith panics with r over tab, on a source the committed grammar accepts.
func parseWithPanics(tab *grammar.Tables, r *recognize.Recognizer) (panicked bool) {
	defer func() { panicked = recover() != nil }()
	ParseWith(tab, r, []byte("my $x = 1 + 2 * 3;\n"), Scope{})
	return false
}

// emptyJudged is a recognizer over tab judged with an explicit empty precedence: no levels, no %prec, nothing refused.
func emptyJudged(t *testing.T, tab *grammar.Tables) *recognize.Recognizer {
	t.Helper()
	p, err := prectab.NewPrecedence(tab, nil, make([]string, len(tab.Rules)), nil)
	if err != nil {
		t.Fatal(err)
	}
	r, _ := recognize.NewJudged(tab, p, recognize.Options{})
	return r
}

// TestParseHasNoPrecedenceSwitch: nothing a caller passes Parse or ParseWith selects the mode or substitutes a
// precedence for perly.y's.
func TestParseHasNoPrecedenceSwitch(t *testing.T) {
	tab := committedTables(t)
	t.Run("no-parameter-or-option", func(t *testing.T) {
		precType := reflect.TypeFor[*prectab.Precedence]()
		for name, fn := range map[string]reflect.Type{"Parse": reflect.TypeOf(Parse), "ParseWith": reflect.TypeOf(ParseWith)} {
			for i := range fn.NumIn() {
				in := fn.In(i)
				if in.Kind() == reflect.Bool || in == precType {
					t.Errorf("%s's parameter %d is a %s, a switch", name, i, in)
				}
				for _, st := range []reflect.Type{in, reflect.TypeFor[recognize.Options]()} {
					if st.Kind() != reflect.Struct {
						continue
					}
					for _, f := range reflect.VisibleFields(st) {
						if f.IsExported() && (f.Type == precType || strings.Contains(strings.ToLower(f.Name), "prec")) {
							t.Errorf("%s: %s.%s is a precedence switch", name, st, f.Name)
						}
					}
				}
			}
		}
	})
	t.Run("parsewith-refuses-no-precedence-recognizer", func(t *testing.T) {
		if !parseWithPanics(tab, recognize.NewNoPrecedence(tab, recognize.Options{})) {
			t.Error("ParseWith accepts the no-precedence constructor's recognizer")
		}
	})
	t.Run("parsewith-accepts-default", func(t *testing.T) {
		if parseWithPanics(tab, recognize.New(tab, recognize.Options{})) {
			t.Fatal("ParseWith panics on recognize.New's recognizer")
		}
		if f, ref := ParseWith(tab, recognize.New(tab, recognize.Options{}), []byte("my $x = 1 + 2 * 3;\n"), Scope{}); f == nil {
			t.Errorf("ParseWith with recognize.New refuses %+v", ref)
		}
	})
	t.Run("parsewith-refuses-empty-explicit-precedence-over-committed-tables", func(t *testing.T) {
		if !parseWithPanics(tab, emptyJudged(t, tab)) {
			t.Error("ParseWith accepts an empty explicit precedence over the committed tables")
		}
	})
	t.Run("parsewith-refuses-empty-explicit-precedence-over-copy-of-committed-tables", func(t *testing.T) {
		cp := new(grammar.Tables)
		*cp = *tab // a value of its own, whose rules are still perly.y's
		if !parseWithPanics(cp, emptyJudged(t, cp)) {
			t.Error("ParseWith accepts an empty explicit precedence over a copy of the committed tables")
		}
		if parseWithPanics(cp, recognize.New(cp, recognize.Options{})) {
			t.Error("ParseWith panics on recognize.New's recognizer over the copy")
		}
	})
}

// constrainFixtures are sources on which V1's Constrain drops alternatives: forbidden reductions (`1 + 2 * 3;`, and
// `_ - $h{1} - $fh;`'s UNIOPSUB term before ADDOP) and forbidden shifts (`1 - 2 - 3;`, `$x = $y = $z;`, `!$x == 1;`).
var constrainFixtures = []string{"1 + 2 * 3;", "_ - $h{1} - $fh;", "1 - 2 - 3;", "$x = $y = $z;", "!$x == 1;", "foo -$x * $y;"}

// droppedByV1 is each alternative, by node key, that the no-precedence forest of src after LibraryOperators reaches
// and V1's Constrain drops.
func droppedByV1(t *testing.T, tab *grammar.Tables, src string) map[NodeKey][]altKey {
	t.Helper()
	built := builtBy(t, recognize.NewNoPrecedence, src)
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	if built == nil || LibraryOperators(tab, built, s, []byte(src), Scope{}) >= 0 {
		t.Fatalf("%q: the no-precedence pipeline refuses it before the constraint", src)
	}
	v1, ref := V1Forest(tab, []byte(src), Scope{})
	if v1 == nil {
		t.Fatalf("%q: V1 refuses it: %+v", src, ref)
	}
	all, _ := reachableAlts(built)
	kept, _ := reachableAlts(v1)
	out := map[NodeKey][]altKey{}
	for k, alts := range all {
		for _, a := range alts {
			if !slices.Contains(kept[k], a) {
				out[k] = append(out[k], a)
			}
		}
	}
	return out
}

// TestParseWithNeverConstrains: on every fixture, Parse's forest reaches none of the alternatives V1's Constrain drops
// from the no-precedence forest, and Parse never enters Constrain: the judged chart's Build already lacks them.
func TestParseWithNeverConstrains(t *testing.T) {
	tab := committedTables(t)
	for _, src := range constrainFixtures {
		dropped := droppedByV1(t, tab, src)
		n0 := constrainEntries.Load()
		f, ref := Parse(tab, []byte(src), Scope{})
		entered := constrainEntries.Load() - n0
		if f == nil {
			t.Fatalf("%q: Parse refuses it: %+v", src, ref)
		}
		got, _ := reachableAlts(f)
		kept := 0
		for k, alts := range dropped {
			for _, a := range alts {
				if slices.Contains(got[k], a) {
					kept++
				}
			}
		}
		t.Logf("%q: V1 drops alternatives on %d nodes; Parse keeps %d of them and entered Constrain %d times", src, len(dropped), kept, entered)
		if len(dropped) == 0 || kept != 0 || entered != 0 {
			t.Errorf("%q: V1 drops alternatives on %d nodes, Parse keeps %d, Constrain entered %d times; want some dropped, none kept, no entry", src, len(dropped), kept, entered)
		}
	}
}

// TestConstrainNeverEnteredByParse: Constrain's entry counter reads 0 across Parse on the fixtures, which hold
// forbidden shifts and reductions, and at least 1 when the test calls Constrain itself, so the counter counts.
func TestConstrainNeverEnteredByParse(t *testing.T) {
	tab := committedTables(t)
	prec, err := prectab.ParsePrecedence(tab)
	if err != nil {
		t.Fatal(err)
	}
	n0 := constrainEntries.Load()
	for _, src := range constrainFixtures {
		if f, ref := Parse(tab, []byte(src), Scope{}); f == nil {
			t.Fatalf("%q: Parse refuses it: %+v", src, ref)
		}
	}
	byParse := constrainEntries.Load() - n0
	n1 := constrainEntries.Load()
	builtBy(t, recognize.NewNoPrecedence, constrainFixtures[0]).Constrain(prec)
	direct := constrainEntries.Load() - n1
	t.Logf("Constrain entered %d times by Parse over %d fixtures, %d by a direct call", byParse, len(constrainFixtures), direct)
	if byParse != 0 || direct < 1 {
		t.Errorf("Constrain entered %d times by Parse and %d by a direct call; want 0 and at least 1", byParse, direct)
	}
}
