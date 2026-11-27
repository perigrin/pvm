// ABOUTME: Tests that no caller of Parse or ParseWith can turn precedence off (F3-7): no parameter or option is a
// ABOUTME: switch, and ParseWith panics on a recognizer that judges with anything but the precedence New derives.

package forest

import (
	"reflect"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
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
