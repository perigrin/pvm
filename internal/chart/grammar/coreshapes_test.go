// ABOUTME: Holds the builtin classes' differences from parse.CoreShapes() equal to the committed, reviewed list.
// ABOUTME: Also shows the comparison can fail, and that every listed entry names both sides and a reason.

package grammar

import (
	"os"
	"regexp"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

const coreShapeList = "testdata/core-shape-differences.txt"

// listEntries are the committed list's entries without their reasons; blank lines and comment lines are skipped.
func listEntries(t *testing.T, path string) (entries, raw []string) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range strings.Split(string(b), "\n") {
		if l = strings.TrimRight(l, " "); l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		raw = append(raw, l)
		entry, _, _ := strings.Cut(l, "\t# ")
		entries = append(entries, entry)
	}
	return entries, raw
}

// diffAgainstList describes how the computed differences and a list disagree, "" when they are equal.
func diffAgainstList(got, list []string) string {
	var out []string
	for i, l := range list {
		if slices.Contains(list[:i], l) {
			out = append(out, "listed twice: "+l)
		}
	}
	for _, g := range got {
		if !slices.Contains(list, g) {
			out = append(out, "not in the list: "+g)
		}
	}
	for _, l := range list {
		if !slices.Contains(got, l) {
			out = append(out, "listed but not a difference: "+l)
		}
	}
	return strings.Join(out, "\n")
}

func TestCoreShapeDifferences(t *testing.T) {
	list, _ := listEntries(t, coreShapeList)
	if d := diffAgainstList(CoreShapeDifferences(parse.CoreShapes(), perlyClasses), list); d != "" {
		t.Errorf("the builtin classes differ from parse.CoreShapes() other than as %s lists; review each and edit the list:\n%s", coreShapeList, d)
	}
}

func TestCoreShapeDifferencesDetects(t *testing.T) {
	cls := func(terminal, word string) ClassDecl { return ClassDecl{terminal, TokenClass{Text: ptr(word)}} }
	shapes := map[string]parse.Shape{
		"coreonly": parse.ShapeUnary,
		"uni":      parse.ShapeUnary, "fun": parse.ShapeUnary, "both": parse.ShapeUnary,
		"unil": parse.ShapeUnary, "mix": parse.ShapeUnary, "nilu": parse.ShapeNiladic,
		"nil": parse.ShapeNiladic, "nilop": parse.ShapeNiladic, "lst": parse.ShapeList,
	}
	classes := []ClassDecl{
		cls("UNIOP", "uni"), cls("FUNC1", "fun"), cls("UNIOP", "both"), cls("FUNC1", "both"),
		cls("LSTOP", "unil"), cls("UNIOP", "mix"), cls("LSTOP", "mix"), cls("UNIOP", "nilu"), cls("FUNC0", "nil"), cls("FUNC0OP", "nilop"), cls("LSTOP", "lst"),
		cls("LSTOP", "portonly"), cls("KW_DO", "other"),
	}
	want := []string{
		"coreonly\tcore: unary\tport: none",
		"mix\tcore: unary\tport: LSTOP/UNIOP",
		"nilu\tcore: niladic\tport: UNIOP",
		"portonly\tcore: none\tport: LSTOP",
		"unil\tcore: unary\tport: LSTOP",
	}
	got := CoreShapeDifferences(shapes, classes)
	if !slices.Equal(got, want) {
		t.Errorf("synthetic differences:\n got %q\nwant %q", got, want)
	}
	if d := diffAgainstList(got, want); d != "" {
		t.Errorf("the equal list was refused:\n%s", d)
	}
	if diffAgainstList(got, want[1:]) == "" {
		t.Error("a list omitting an entry passed")
	}
	if diffAgainstList(got, append(slices.Clone(want), want[0])) == "" {
		t.Error("a list repeating an entry passed")
	}
	if diffAgainstList(got, append(slices.Clone(want), "extra\tcore: list\tport: none")) == "" {
		t.Error("a list adding an entry passed")
	}
}

var listForm = regexp.MustCompile(`^\S+\tcore: (niladic|unary|list|block|none)\tport: (none|[A-Z0-9]+(/[A-Z0-9]+)*)\t# \S.*$`)

func TestCoreShapeDifferencesListForm(t *testing.T) {
	_, raw := listEntries(t, coreShapeList)
	for _, l := range raw {
		if !listForm.MatchString(l) {
			t.Errorf("entry does not name both sides and a reason (`name<TAB>core: <shape|none><TAB>port: <terminals|none><TAB># reason`): %q", l)
		}
	}
	for _, bad := range []string{"abs", "abs\tcore: unary", "abs\tcore: unary\tport: LSTOP", "abs\tcore: unary\tport: LSTOP\t# "} {
		if listForm.MatchString(bad) {
			t.Errorf("the form accepts %q", bad)
		}
	}
	if !listForm.MatchString("abs\tcore: unary\tport: LSTOP\t# why") {
		t.Error("the form refuses a whole entry")
	}
}
