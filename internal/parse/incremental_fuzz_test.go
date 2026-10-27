// ABOUTME: Property test (§7.7): an incremental re-parse equals a fresh Session parse of the edited bytes.
// ABOUTME: FuzzIncremental fuzzes one edit, FuzzIncrementalSequence a burst; the §7.7.3 seeds are a table here.

package parse_test

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// The fuzzers parse as the LSP parses a file: URI, through a Session whose
// resolver searches the file's directory and then a root. incRoot holds
// Foo.pm, which exports `sub bar($)`, so a call to bar is a tree a plain
// Parse cannot build.
var (
	incPath = filepath.Join("testdata", "incremental", "edit.pl")
	incRoot = filepath.Join("testdata", "incremental", "lib")
)

// clampEdit turns the fuzzer's integers into a valid edit of a buffer of n
// bytes. Rejecting instead would throw away most inputs.
func clampEdit(n, offset, delLen int, insert string) parse.Edit {
	start := min(max(offset, 0), n)
	return parse.Edit{Start: start, OldEnd: start + min(max(delLen, 0), n-start), NewText: insert}
}

// applyEdit replaces src's bytes [e.Start, e.OldEnd) with e.NewText.
func applyEdit(src string, e parse.Edit) string {
	return src[:e.Start] + e.NewText + src[e.OldEnd:]
}

// decodeEdits reads a byte script as edits of a buffer that starts n bytes
// long. Each edit is 4 header bytes -- a little-endian uint16 start, a delete
// length, an insert length -- then the insert. Start and delete clamp to the
// buffer as it stands after every earlier edit, so each edit is valid where it
// lands; a trailing partial header is dropped.
func decodeEdits(script []byte, n int) []parse.Edit {
	var edits []parse.Edit
	for len(script) >= 4 {
		k := min(int(script[3]), len(script)-4)
		e := clampEdit(n, int(binary.LittleEndian.Uint16(script))%(n+1), int(script[2]), string(script[4:4+k]))
		script = script[4+k:]
		edits = append(edits, e)
		n += len(e.NewText) - (e.OldEnd - e.Start)
	}
	return edits
}

// encodeEdit is decodeEdits' inverse for one edit, for writing seeds.
func encodeEdit(e parse.Edit) []byte {
	b := binary.LittleEndian.AppendUint16(nil, uint16(e.Start))
	return append(append(b, byte(e.OldEnd-e.Start), byte(len(e.NewText))), e.NewText...)
}

// freshParse is the property's right-hand side: a Session that has read
// nothing, with the same path and roots.
func freshParse(src string) *parse.Node {
	return parse.NewSession().ParseSourceFrom(incPath, []byte(src), incRoot)
}

// checkIncremental asserts the property for one edit of src, in both forms:
// the Session form against a fresh Session parse, and the plain form against
// Parse. It returns the edited source.
func checkIncremental(t *testing.T, src string, e parse.Edit) string {
	t.Helper()
	edited := applyEdit(src, e)
	s := parse.NewSession()
	old := s.ParseSourceFrom(incPath, []byte(src), incRoot)
	inc := s.ReparseIncremental(old, e, incPath, []byte(edited), incRoot)
	if diff := parse.TreeDiff(inc, freshParse(edited)); diff != "" {
		t.Fatalf("Session: incremental != fresh after edit %+v of %q:\n%s", e, src, diff)
	}
	plain := parse.ReparseIncremental(parse.Parse([]byte(src)), e, []byte(edited))
	if diff := parse.TreeDiff(plain, parse.Parse([]byte(edited))); diff != "" {
		t.Fatalf("Parse: incremental != fresh after edit %+v of %q:\n%s", e, src, diff)
	}
	return edited
}

// incrementalSeed is one §7.7.3 row: the edit, and the row's "Defined
// invalidation" column. TestIncrementalEditSeeds asserts splice == fresh,
// TestTriageMatchesDefinedInvalidation the triage column.
type incrementalSeed struct {
	name     string
	src      string
	edit     parse.Edit
	triage   string // "StrategyFull" or "StrategyLocal" (§6.1.2)
	repairOK bool   // Repair's ok; false means ReparseIncremental parses the whole file
}

// insertAt is an edit inserting text at the first byte of src after the
// first occurrence of at.
func insertAt(src, at, text string) parse.Edit {
	i := strings.Index(src, at) + len(at)
	return parse.Edit{Start: i, OldEnd: i, NewText: text}
}

// replaceFirst is an edit replacing the first occurrence of old in src.
func replaceFirst(src, old, text string) parse.Edit {
	i := strings.Index(src, old)
	return parse.Edit{Start: i, OldEnd: i + len(old), NewText: text}
}

const (
	seedSub    = "sub f {\n  my $foo = 1;\n  return $foo;\n}\nprint f();\n"
	heredocSub = "sub f {\n  print <<EOF;\nbody\nEOF\n  return 1;\n}\nprint f();\n"
	protoSrc   = "sub f() { 1 }\nmy @a = (1, 2);\nprint f @a;\n"
)

var incrementalSeeds = []incrementalSeed{
	{"insert-quote", seedSub, insertAt(seedSub, "= ", `"`), "StrategyFull", false},
	{"insert-heredoc", seedSub, insertAt(seedSub, "1;\n", "  print <<EOF;\n"), "StrategyFull", false},
	{"delete-heredoc-terminator", heredocSub, replaceFirst(heredocSub, "EOF\n", ""), "StrategyLocal", false},
	{"pod", seedSub, insertAt(seedSub, "1;\n", "=pod\n"), "StrategyFull", false},
	{"end", seedSub, insertAt(seedSub, "1;\n", "__END__\n"), "StrategyFull", false},
	{"prototype", protoSrc, insertAt(protoSrc, "sub f(", `\@`), "StrategyFull", false},
	{"insert-brace", seedSub, insertAt(seedSub, "1;\n", "{\n"), "StrategyFull", false},
	{"split-token", seedSub, insertAt(seedSub, "my $f", "|"), "StrategyLocal", true},
}

// resolvedCallSrc calls a module-prototype sub: under a Session with incRoot,
// bar is a Resolved call with Prototype "($)"; under Parse it is not.
const resolvedCallSrc = "use Foo; my @x = (bar 1, 2);\n"

// resolvedCallEdit changes the first argument, inside the call's arguments.
var resolvedCallEdit = replaceFirst(resolvedCallSrc, "1", "42")

// fuzzIncrementalCorpus is the go test fuzz v1 encoding of one edit, as
// FuzzIncremental's arguments (src, offset, delLen, insert).
func fuzzIncrementalCorpus(src string, e parse.Edit) string {
	return fmt.Sprintf("go test fuzz v1\nstring(%q)\nint(%d)\nint(%d)\nstring(%q)\n", src, e.Start, e.OldEnd-e.Start, e.NewText)
}

// fuzzSequenceCorpus is the same edit as FuzzIncrementalSequence's (src,
// script).
func fuzzSequenceCorpus(src string, e parse.Edit) string {
	return fmt.Sprintf("go test fuzz v1\nstring(%q)\n[]byte(%q)\n", src, encodeEdit(e))
}

// TestIncrementalEditSeeds holds each §7.7.3 row: splice == fresh, and the
// committed seeds under testdata/fuzz/ are this table's edits, so the table
// and the corpus `go test` replays cannot drift apart.
func TestIncrementalEditSeeds(t *testing.T) {
	for _, s := range incrementalSeeds {
		t.Run(s.name, func(t *testing.T) {
			checkIncremental(t, s.src, s.edit)
			for dir, want := range map[string]string{
				"FuzzIncremental":         fuzzIncrementalCorpus(s.src, s.edit),
				"FuzzIncrementalSequence": fuzzSequenceCorpus(s.src, s.edit),
			} {
				got, err := os.ReadFile(filepath.Join("testdata", "fuzz", dir, s.name))
				if err != nil || string(got) != want {
					t.Errorf("%s/%s: want\n%s\ngot (err %v)\n%s", dir, s.name, want, err, got)
				}
			}
		})
	}
}

// TestTriageMatchesDefinedInvalidation: triage of each §7.7.3 row, on the
// tree a Session parsed, is the row's defined invalidation, and a
// StrategyLocal row is anchored at its enclosing sub, `sub f`.
func TestTriageMatchesDefinedInvalidation(t *testing.T) {
	for _, s := range incrementalSeeds {
		t.Run(s.name, func(t *testing.T) {
			old := parse.NewSession().ParseSourceFrom(incPath, []byte(s.src), incRoot)
			if got := parse.Triage(old, []byte(s.src), s.edit); got != s.triage {
				t.Fatalf("triage %s, want %s", got, s.triage)
			}
			if s.triage != "StrategyLocal" {
				return
			}
			start, end := parse.WidenAnchor(old, []byte(s.src), s.edit)
			if anchor := s.src[start:end]; !strings.HasPrefix(anchor, "sub f {") || !strings.HasSuffix(anchor, "}") {
				t.Errorf("anchor %q, want the enclosing sub", anchor)
			}
		})
	}
}

// TestIncrementalSessionResolverMatchesFresh: an edit inside the arguments
// of a call whose prototype a module installed. The Session form must still
// resolve it, as a fresh Session does; Parse alone would not.
func TestIncrementalSessionResolverMatchesFresh(t *testing.T) {
	edited := checkIncremental(t, resolvedCallSrc, resolvedCallEdit)
	s := parse.NewSession()
	old := s.ParseSourceFrom(incPath, []byte(resolvedCallSrc), incRoot)
	inc := s.ReparseIncremental(old, resolvedCallEdit, incPath, []byte(edited), incRoot)
	if c := findCall(inc, "bar"); c == nil || !c.Resolved || c.Prototype != "($)" {
		t.Fatalf("want bar a Resolved call with Prototype ($); got %s (%+v)", shape(inc), c)
	}
	if diff := parse.TreeDiff(inc, parse.Parse([]byte(edited))); diff == "" {
		t.Fatal("Session tree equals Parse's: the resolver is not in play, so this test checks nothing")
	}
	got, err := os.ReadFile(filepath.Join("testdata", "fuzz", "FuzzIncremental", "resolved-call"))
	if want := fuzzIncrementalCorpus(resolvedCallSrc, resolvedCallEdit); err != nil || string(got) != want {
		t.Errorf("resolved-call seed: want\n%s\ngot (err %v)\n%s", want, err, got)
	}
}

// TestSessionReparseIncrementalUsesFreshSession: a Session's second parse
// reports none of the modules, declaration errors or declaration warnings its
// first already read, so a re-parse on the receiver would lose them.
// ReparseIncremental must re-parse fresh. Odd's declaration, overlaid from
// incRoot, carries one warning and one error, so each list is non-empty.
func TestSessionReparseIncrementalUsesFreshSession(t *testing.T) {
	t.Cleanup(parse.OverlayDeclarations(os.DirFS(incRoot)))
	src := "use Odd;\n" + resolvedCallSrc
	e := replaceFirst(src, "1", "42")
	s := parse.NewSession()
	old := s.ParseSourceFrom(incPath, []byte(src), incRoot)
	edited := applyEdit(src, e)
	inc := s.ReparseIncremental(old, e, incPath, []byte(edited), incRoot)
	fresh := freshParse(edited)
	reused := s.ParseSourceFrom(incPath, []byte(edited), incRoot)
	for _, c := range []struct {
		name string
		of   func(*parse.Node) []string
	}{
		{"LoadedModules", parse.LoadedModules},
		{"DeclarationErrors", func(n *parse.Node) []string { return errorStrings(parse.DeclarationErrors(n)) }},
		{"DeclarationWarnings", func(n *parse.Node) []string { return errorStrings(parse.DeclarationWarnings(n)) }},
	} {
		if len(c.of(fresh)) == 0 {
			t.Fatalf("%s: fresh parse has none; the fixture checks nothing", c.name)
		}
		if got := c.of(reused); len(got) != 0 {
			t.Fatalf("%s: a reused Session's second parse has %v; this test assumes none", c.name, got)
		}
		if got, want := c.of(inc), c.of(fresh); !slices.Equal(got, want) {
			t.Errorf("%s: got %v, want %v", c.name, got, want)
		}
	}
	if diff := parse.TreeDiff(inc, fresh); diff != "" {
		t.Errorf("incremental != fresh:\n%s", diff)
	}
}

// errorStrings is errs' texts, which is how TreeDiff compares them.
func errorStrings(errs []error) []string {
	var out []string
	for _, err := range errs {
		out = append(out, err.Error())
	}
	return out
}

// TestIncrementalEditClamping: out-of-range integers clamp to a valid edit,
// and the boundary edits satisfy the property.
func TestIncrementalEditClamping(t *testing.T) {
	const src = "my $x = 1;\nprint $x;\n"
	n := len(src)
	for _, c := range []struct {
		name               string
		src                string
		offset, del        int
		insert             string
		wantStart, wantEnd int
	}{
		{"negative offset", src, -5, 2, "a", 0, 2},
		{"negative delete", src, 3, -7, "a", 3, 3},
		{"oversized offset", src, n + 100, 4, "a", n, n},
		{"oversized delete", src, 4, n * 3, "a", 4, n},
		{"offset 0", src, 0, 0, "#", 0, 0},
		{"at EOF", src, n, 0, "1;", n, n},
		{"delete everything", src, 0, n, "", 0, n},
		{"empty source", "", 7, 7, "my $y;", 0, 0},
		{"empty insert", src, 3, 2, "", 3, 5},
	} {
		t.Run(c.name, func(t *testing.T) {
			e := clampEdit(len(c.src), c.offset, c.del, c.insert)
			if e.Start != c.wantStart || e.OldEnd != c.wantEnd {
				t.Fatalf("clampEdit(%d, %d, %d) = [%d, %d), want [%d, %d)", len(c.src), c.offset, c.del, e.Start, e.OldEnd, c.wantStart, c.wantEnd)
			}
			checkIncremental(t, c.src, e)
		})
	}
}

// TestIncrementalEditSplitsUTF8Rune: edit bounds land inside a multi-byte
// character, leaving invalid UTF-8 on one or both sides.
func TestIncrementalEditSplitsUTF8Rune(t *testing.T) {
	const src = "my $s = \"é€\";\nprint $s;\n"
	i := strings.Index(src, "é")
	for _, e := range []parse.Edit{
		{Start: i + 1, OldEnd: i + 1, NewText: "x"},
		{Start: i + 1, OldEnd: i + 3, NewText: ""},
		{Start: i, OldEnd: i + 1, NewText: "\xe2"},
	} {
		checkIncremental(t, src, e)
	}
}

// TestDecodeEditsStaysInBounds: for any script, each decoded edit is valid
// against the buffer as the earlier edits left it.
func TestDecodeEditsStaysInBounds(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for range 2000 {
		script := make([]byte, r.Intn(200))
		r.Read(script)
		cur := strings.Repeat("x", r.Intn(300))
		for _, e := range decodeEdits(script, len(cur)) {
			if e.Start < 0 || e.Start > e.OldEnd || e.OldEnd > len(cur) {
				t.Fatalf("edit %+v outside buffer of %d bytes (script %x)", e, len(cur), script)
			}
			cur = applyEdit(cur, e)
		}
	}
}

// FuzzIncremental is §7.7.2's property for one edit. Run it with
//
//	go test ./internal/parse/ -run '^$' -fuzz '^FuzzIncremental$' -fuzztime 10000000x
//
// or `make fuzz-incremental`. Plain `go test` replays the committed seeds
// under testdata/fuzz/FuzzIncremental/, the §7.7.3 edits among them.
func FuzzIncremental(f *testing.F) {
	f.Add("my $x = 1;\nprint $x;\n", 5, 1, "42")
	f.Add("print <<'EOF';\nbody\nEOF\n", 6, 3, "\"\"")
	f.Add("s{a}{b}g;", 2, 1, "{")
	f.Fuzz(func(t *testing.T, src string, offset, delLen int, insert string) {
		if len(src) > 64*1024 || len(insert) > 1024 {
			t.Skip()
		}
		checkIncremental(t, src, clampEdit(len(src), offset, delLen, insert))
	})
}

// FuzzIncrementalSequence is §7.7.4's: a burst of edits on one tree, checked
// after every edit, since corruption compounds and checking only at the end
// would not say which edit caused it.
func FuzzIncrementalSequence(f *testing.F) {
	f.Fuzz(func(t *testing.T, src string, script []byte) {
		if len(src) > 64*1024 || len(script) > 4096 {
			t.Skip()
		}
		s := parse.NewSession()
		tree := s.ParseSourceFrom(incPath, []byte(src), incRoot)
		cur := src
		for _, e := range decodeEdits(script, len(cur)) {
			cur = applyEdit(cur, e)
			tree = s.ReparseIncremental(tree, e, incPath, []byte(cur), incRoot)
			if diff := parse.TreeDiff(tree, freshParse(cur)); diff != "" {
				t.Fatalf("diverged after edit %+v:\n%s", e, diff)
			}
		}
	})
}
