// ABOUTME: Holds testdata/faithful_queue.txt to exactly the M3-dir files whose canon emission is unfaithful.
// ABOUTME: Each row is a file's first difference, its computed three-token cause, and the class that cause decides.
package parse_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"tamarou.com/pvm/internal/lexer"
	"tamarou.com/pvm/internal/parse"
)

const faithfulQueuePath = "testdata/faithful_queue.txt"

// m3Dirs are the perl.git t/ directories whose files must all re-emit
// faithfully when m3-oracle-at-scale closes.
var m3Dirs = []string{"op", "mro", "uni", "lib", "test_pl"}

// queueClasses are the seven classes, in the order the queue's header
// lists them. Each has one owning issue that brings it to zero.
var queueClasses = []string{
	"dropped-paren", "heredoc", "subless-special-sub", "stmt-boundary",
	"canon-adds-paren-op", "canon-adds-paren-term", "other",
}

// queueRow is one line of the queue: `path:line class src canon prev reason`.
type queueRow struct {
	Path             string
	Line             int
	Class            string
	Src, Canon, Prev string
	Reason           string
}

func (r queueRow) cause() string { return r.Src + " " + r.Canon + " " + r.Prev }

func (r queueRow) String() string {
	return fmt.Sprintf("%s:%d %s %s %s", r.Path, r.Line, r.Class, r.cause(), r.Reason)
}

type faithfulQueue struct {
	Rows   []queueRow
	Causes map[string]int
}

// Word categories, the names a Word token is encoded by.
var (
	declarators = []string{"my", "our", "local", "state"}
	wordOps     = []string{"not", "and", "or", "xor", "x", "lt", "gt", "le", "ge", "eq", "ne", "cmp", "isa"}
	specialSubs = []string{"DESTROY", "AUTOLOAD", "BEGIN", "END", "INIT", "CHECK", "UNITCHECK"}
)

// causeNames are the encodings that are not a token's own text. Any other
// cause field is punctuation.
var causeNames = func() map[string]bool {
	names := map[string]bool{"EOF": true, "BOF": true}
	for _, s := range []string{"declarator", "named-unary", "list-op", "word-op", "sub", "special-sub", "call", "bareword"} {
		names[s] = true
	}
	for k := lexer.Kind(0); k.String() != "Kind(?)"; k++ {
		names[k.String()] = true
	}
	return names
}()

// punctuation is the text of an Operator, CloseBracket or Semicolon token:
// never a letter or digit, which the lexer leaves to Words (`x=`, `-e`).
var punctuation = regexp.MustCompile(`^[^A-Za-z0-9_\s]+$`)

// validCauseField: a known name, or punctuation. It is what tells a cause's
// third field from a reason's first word.
func validCauseField(s string) bool {
	return causeNames[s] || punctuation.MatchString(s)
}

func isPunct(s string) bool { return !causeNames[s] }

// sigTok is one significant token as the cause encodes it.
type sigTok struct {
	Kind  lexer.Kind
	Text  string
	Start int
}

func sigToks(src []byte) []sigTok {
	var out []sigTok
	for _, tok := range lexer.Tokenize(src) {
		if !slices.Contains(parse.DroppedKinds, tok.Kind) {
			out = append(out, sigTok{tok.Kind, string(src[tok.Start:tok.End]), tok.Start})
		}
	}
	return out
}

// encodeToken is the cause encoding of toks[i]: punctuation by its text, a
// Word by its category, any other kind by its name, BOF before the stream
// and EOF after it.
func encodeToken(toks []sigTok, i int) string {
	if i < 0 {
		return "BOF"
	}
	if i >= len(toks) {
		return "EOF"
	}
	switch tok := toks[i]; tok.Kind {
	case lexer.Operator, lexer.CloseBracket, lexer.Semicolon:
		return tok.Text
	case lexer.Word:
		return wordCategory(tok.Text, i+1 < len(toks) && toks[i+1].Text == "(")
	default:
		return tok.Kind.String()
	}
}

// wordCategory is a Word's category; parenNext is whether its own stream
// follows it with `(`. The named sets come before the builtin shapes, which
// also list `not` (a list operator) and `isa` (niladic).
func wordCategory(text string, parenNext bool) string {
	shape, known := parse.CoreShapes()[text]
	switch {
	case slices.Contains(declarators, text):
		return "declarator"
	case slices.Contains(wordOps, text):
		return "word-op"
	case text == "sub":
		return "sub"
	case slices.Contains(specialSubs, text):
		return "special-sub"
	case known && shape == parse.ShapeUnary, lexer.IsFileTest(text):
		return "named-unary"
	case known && shape == parse.ShapeList:
		return "list-op"
	case parenNext:
		return "call"
	default:
		return "bareword"
	}
}

// classify is the class a cause decides, by the first matching rule of the
// ordered table, and that rule's 1-based number.
func classify(src, canon, prev string) (string, int) {
	either := func(set ...string) bool { return slices.Contains(set, src) || slices.Contains(set, canon) }
	switch {
	case src == "HeredocBody":
		return "heredoc", 1
	case src == "(" && canon != "(":
		return "dropped-paren", 2
	case canon == "sub" && src == "special-sub":
		return "subless-special-sub", 3
	case src == "{" && (prev == "DerefSigil" || prev == "FuncSigil"):
		return "stmt-boundary", 4
	case either("EOF"):
		return "stmt-boundary", 5
	case either(";", "}") && canon != "(":
		return "stmt-boundary", 6
	case canon == "(" && (isPunct(prev) && prev != ";" && prev != "{" ||
		slices.Contains([]string{"declarator", "named-unary", "word-op", "sub"}, prev)):
		return "canon-adds-paren-op", 7
	case canon == "(" && slices.Contains([]string{"call", "list-op", "bareword", ";", "{", "Variable", "Number", "Quote"}, prev):
		return "canon-adds-paren-term", 8
	default:
		return "other", 9
	}
}

// firstDifference is a parsed file's row, or ok false when it is faithful.
func firstDifference(rel string, n *parse.Node, src []byte) (queueRow, bool) {
	w, g, differs := parse.FaithfulStop(n, src)
	if !differs {
		return queueRow{}, false
	}
	stoks := sigToks(src)
	ctoks := sigToks([]byte(parse.Canon(n, src)))
	r := queueRow{Path: rel, Src: encodeToken(stoks, w), Canon: encodeToken(ctoks, g), Prev: encodeToken(ctoks, g-1)}
	r.Class, _ = classify(r.Src, r.Canon, r.Prev)

	at := len(src)
	if w < len(stoks) {
		at = stoks[w].Start
	} else if len(stoks) > 0 {
		at = stoks[len(stoks)-1].Start
	}
	r.Line = strings.Count(string(src[:at]), "\n") + 1
	lines := strings.Split(string(src), "\n")
	r.Reason = strings.Join(strings.Fields(lines[r.Line-1]), " ")
	if runes := []rune(r.Reason); len(runes) > 60 {
		r.Reason = strings.TrimSpace(string(runes[:60]))
	}
	if r.Reason == "" {
		r.Reason = "(blank line)"
	}
	return r, true
}

// m3Walk parses every M3-dir file under tDir outside perlRejects and returns
// the files it visited and a row for each unfaithful one. A missing dir, an
// empty walk and a file holding an Unknown are errors: the queue covers the
// files that parse, so one that stops parsing must not leave it unnoticed.
func m3Walk(tDir string) (visited []string, rows []queueRow, err error) {
	if tDir == "" {
		return nil, nil, errors.New("no corpus: the M3 walk needs perl.git t/")
	}
	roots := perlGitTRoots(tDir)
	s := parse.NewSession()
	for _, dir := range m3Dirs {
		if fi, err := os.Stat(filepath.Join(tDir, dir)); err != nil || !fi.IsDir() {
			return nil, nil, fmt.Errorf("M3 dir %s missing under %s", dir, tDir)
		}
		var paths []string
		filepath.WalkDir(filepath.Join(tDir, dir), func(p string, d os.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(p, ".t") {
				paths = append(paths, p)
			}
			return nil
		})
		for _, path := range paths {
			// A slash path, as perlRejects and the queue spell it: the walk
			// runs on fixture trees on every platform, Windows included.
			rel, _ := filepath.Rel(tDir, path)
			rel = filepath.ToSlash(rel)
			if _, rejected := perlRejects[rel]; rejected {
				continue
			}
			src, err := os.ReadFile(path)
			if err != nil {
				return nil, nil, err
			}
			n, err := s.ParseFileFrom(path, roots...)
			if err != nil {
				return nil, nil, fmt.Errorf("%s: %v", rel, err)
			}
			if c := countUnknown(n); c > 0 {
				return nil, nil, fmt.Errorf("%s: %d Unknown; an M3-dir file must parse", rel, c)
			}
			visited = append(visited, rel)
			if r, differs := firstDifference(rel, n, src); differs {
				rows = append(rows, r)
			}
		}
	}
	if len(visited) == 0 {
		return nil, nil, fmt.Errorf("the M3 walk visited no files under %s", tDir)
	}
	return visited, rows, nil
}

var (
	m3Once    sync.Once
	m3Visited []string
	m3Rows    []queueRow
	m3Err     error
)

// m3Corpus is m3Walk over the perl.git corpus, once per test binary.
func m3Corpus(t *testing.T) (string, []string, []queueRow) {
	t.Helper()
	tDir, _ := perlGitTFiles(t)
	m3Once.Do(func() { m3Visited, m3Rows, m3Err = m3Walk(tDir) })
	if m3Err != nil {
		t.Fatal(m3Err)
	}
	return tDir, m3Visited, m3Rows
}

// causeCounts is the number of distinct causes in each class's rows.
func causeCounts(rows []queueRow) map[string]int {
	seen := map[string]map[string]bool{}
	for _, r := range rows {
		if seen[r.Class] == nil {
			seen[r.Class] = map[string]bool{}
		}
		seen[r.Class][r.cause()] = true
	}
	counts := map[string]int{}
	for _, c := range queueClasses {
		counts[c] = len(seen[c])
	}
	return counts
}

var causesLine = regexp.MustCompile(`^# causes (\S+) = (\d+)$`)

// parseQueue reads a queue: `# causes <class> = N` for each class exactly
// once, other `#` lines as comments, and rows of single-space fields.
func parseQueue(data string) (faithfulQueue, error) {
	q := faithfulQueue{Causes: map[string]int{}}
	paths := map[string]bool{}
	for i, line := range strings.Split(data, "\n") {
		at := fmt.Sprintf("line %d", i+1)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "# causes") {
			m := causesLine.FindStringSubmatch(line)
			if m == nil || !slices.Contains(queueClasses, m[1]) {
				return q, fmt.Errorf("%s: malformed causes line %q", at, line)
			}
			if _, dup := q.Causes[m[1]]; dup {
				return q, fmt.Errorf("%s: a second causes line for %s", at, m[1])
			}
			q.Causes[m[1]], _ = strconv.Atoi(m[2])
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, " ")
		if strings.ContainsAny(line, "\t\r") || slices.Contains(f, "") {
			return q, fmt.Errorf("%s: fields must be separated by exactly one space: %q", at, line)
		}
		if len(f) < 6 {
			return q, fmt.Errorf("%s: want path:line class src canon prev reason, got %q", at, line)
		}
		path, num, ok := strings.Cut(f[0], ":")
		n, err := strconv.Atoi(num)
		if !ok || path == "" || err != nil || n < 1 {
			return q, fmt.Errorf("%s: bad path:line %q", at, f[0])
		}
		if !slices.Contains(queueClasses, f[1]) {
			return q, fmt.Errorf("%s: unknown class %q", at, f[1])
		}
		for _, c := range f[2:5] {
			if !validCauseField(c) {
				return q, fmt.Errorf("%s: %q is not a cause field; a cause has three", at, c)
			}
		}
		if paths[path] {
			return q, fmt.Errorf("%s: a second row for %s", at, path)
		}
		paths[path] = true
		q.Rows = append(q.Rows, queueRow{path, n, f[1], f[2], f[3], f[4], strings.Join(f[5:], " ")})
	}
	for _, c := range queueClasses {
		if _, ok := q.Causes[c]; !ok {
			return q, fmt.Errorf("no causes line for %s", c)
		}
	}
	return q, nil
}

func renderQueue(rows []queueRow) string {
	rows = slices.Clone(rows)
	sort.Slice(rows, func(i, j int) bool { return rows[i].Path < rows[j].Path })
	var b strings.Builder
	b.WriteString("# The re-emission work queue: every M3-dir file (t/op mro uni lib test_pl) whose\n" +
		"# canon emission is unfaithful, at its first difference. A work list, not a register:\n" +
		"# it must be empty when m3-oracle-at-scale closes. Rows are\n" +
		"#   path:line class src canon prev reason\n" +
		"# with the cause (src canon prev) and class computed by TestFaithfulQueueMatchesTree.\n" +
		"# Regenerate with -parse.update-ratchet; never merge it by hand.\n")
	counts := causeCounts(rows)
	for _, c := range queueClasses {
		fmt.Fprintf(&b, "# causes %s = %d\n", c, counts[c])
	}
	for _, r := range rows {
		b.WriteString(r.String() + "\n")
	}
	return b.String()
}

// queueDiff compares a queue with the rows computed from the tree, both ways,
// and its causes lines with the computed counts.
func queueDiff(q faithfulQueue, computed []queueRow) []string {
	var errs []string
	have := map[string]queueRow{}
	for _, r := range q.Rows {
		have[r.Path] = r
	}
	for _, c := range computed {
		r, ok := have[c.Path]
		delete(have, c.Path)
		switch {
		case !ok:
			errs = append(errs, fmt.Sprintf("%s is unfaithful and has no row; want %s", c.Path, c))
		case r.Line != c.Line:
			errs = append(errs, fmt.Sprintf("%s: row line %d is not the file's first difference, line %d", c.Path, r.Line, c.Line))
		case r.Class != c.Class || r.cause() != c.cause():
			errs = append(errs, fmt.Sprintf("%s: row says %s %s, computed %s %s", c.Path, r.Class, r.cause(), c.Class, c.cause()))
		}
	}
	for path := range have {
		errs = append(errs, fmt.Sprintf("%s has a row but re-emits faithfully; remove it", path))
	}
	counts := causeCounts(computed)
	for _, c := range queueClasses {
		if q.Causes[c] != counts[c] {
			errs = append(errs, fmt.Sprintf("# causes %s = %d, computed %d", c, q.Causes[c], counts[c]))
		}
	}
	sort.Strings(errs)
	return errs
}

// rowPathsErr: every row names an existing file in an M3 dir.
func rowPathsErr(rows []queueRow, tDir string) error {
	for _, r := range rows {
		dir, _, _ := strings.Cut(r.Path, "/")
		if !slices.Contains(m3Dirs, dir) {
			return fmt.Errorf("%s is outside the M3 dirs %v", r.Path, m3Dirs)
		}
		if _, err := os.Stat(filepath.Join(tDir, r.Path)); err != nil {
			return fmt.Errorf("%s is not in the corpus: %v", r.Path, err)
		}
	}
	return nil
}

// causeZeroLine is the split-sibling template's cause-zero criterion: the
// queue exists and holds no row with any of the class's given causes. A
// fixed-string grep, because a cause holds `(` and an ERE with an unmatched
// one exits 2, which `!` turns into a pass.
func causeZeroLine(queue, class string, causes []string) string {
	quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }
	parts := []string{"test -f " + quote(queue)}
	for _, c := range causes {
		parts = append(parts, "! grep -Fq "+quote(" "+class+" "+c+" ")+" "+quote(queue))
	}
	return strings.Join(parts, " && ")
}

func readQueue(t *testing.T) faithfulQueue {
	t.Helper()
	data, err := os.ReadFile(faithfulQueuePath)
	if err != nil {
		t.Fatal(err)
	}
	q, err := parseQueue(string(data))
	if err != nil {
		t.Fatalf("%s: %v", faithfulQueuePath, err)
	}
	return q
}

const goodQueue = "# a comment\n" +
	"# causes dropped-paren = 1\n# causes heredoc = 0\n# causes subless-special-sub = 0\n" +
	"# causes stmt-boundary = 0\n# causes canon-adds-paren-op = 0\n# causes canon-adds-paren-term = 0\n" +
	"# causes other = 0\n" +
	"op/a.t:3 dropped-paren ( Variable ( is(($x), 1);\n"

func TestFaithfulQueueRowsAreWellFormed(t *testing.T) {
	q := readQueue(t)
	for _, r := range q.Rows {
		if got, _ := classify(r.Src, r.Canon, r.Prev); got != r.Class {
			t.Errorf("%s: class %s, its cause decides %s", r.Path, r.Class, got)
		}
	}
	fixture, err := parseQueue(goodQueue)
	if err != nil {
		t.Fatalf("a well-formed fixture is rejected: %v", err)
	}
	want := queueRow{"op/a.t", 3, "dropped-paren", "(", "Variable", "(", "is(($x), 1);"}
	if len(fixture.Rows) != 1 || fixture.Rows[0] != want || fixture.Causes["dropped-paren"] != 1 || len(fixture.Causes) != 7 {
		t.Errorf("fixture parsed as %+v", fixture)
	}
}

func TestFaithfulQueueRowsRejectMalformed(t *testing.T) {
	row := "op/a.t:3 dropped-paren ( Variable ( is(($x), 1);\n"
	for name, bad := range map[string]string{
		"empty reason":       strings.Replace(goodQueue, row, "op/a.t:3 dropped-paren ( Variable (\n", 1),
		"trailing space":     strings.Replace(goodQueue, row, "op/a.t:3 dropped-paren ( Variable ( \n", 1),
		"zero line":          strings.Replace(goodQueue, "op/a.t:3", "op/a.t:0", 1),
		"no line":            strings.Replace(goodQueue, "op/a.t:3", "op/a.t", 1),
		"unknown class":      strings.Replace(goodQueue, "op/a.t:3 dropped-paren", "op/a.t:3 misc", 1),
		"two-field cause":    strings.Replace(goodQueue, row, "op/a.t:3 dropped-paren ( Variable is(($x), 1);\n", 1),
		"missing causes":     strings.Replace(goodQueue, "# causes other = 0\n", "", 1),
		"duplicate causes":   strings.Replace(goodQueue, "# causes other = 0\n", "# causes other = 0\n# causes other = 0\n", 1),
		"double space":       strings.Replace(goodQueue, "dropped-paren (", "dropped-paren  (", 1),
		"tab":                strings.Replace(goodQueue, "dropped-paren (", "dropped-paren\t(", 1),
		"leading space":      strings.Replace(goodQueue, row, " "+row, 1),
		"unknown causes key": strings.Replace(goodQueue, "# causes other", "# causes misc", 1),
	} {
		if bad == goodQueue {
			t.Fatalf("%s: the fixture did not change", name)
		}
		if _, err := parseQueue(bad); err == nil {
			t.Errorf("%s: accepted\n%s", name, bad)
		}
	}
}

func TestPerlGitTFaithfulOnM3DirsVisitsEveryFile(t *testing.T) {
	_, files := perlGitTFiles(t)
	_, visited, _ := m3Corpus(t)
	var want []string
	for _, rel := range files {
		dir, _, _ := strings.Cut(rel, "/")
		if _, rejected := perlRejects[rel]; slices.Contains(m3Dirs, dir) && !rejected {
			want = append(want, rel)
		}
	}
	sort.Strings(visited)
	if !slices.Equal(visited, want) {
		t.Errorf("the M3 walk visited %d files, the M3 dirs hold %d", len(visited), len(want))
	}
	if len(visited) != 347 {
		t.Errorf("the M3 walk visited %d files, want 347", len(visited))
	}

	if _, _, err := m3Walk(""); err == nil {
		t.Error("an unset corpus passed the walk")
	}
	tmp := t.TempDir()
	for _, dir := range m3Dirs[1:] {
		if err := os.MkdirAll(filepath.Join(tmp, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := m3Walk(tmp); err == nil || !strings.Contains(err.Error(), "op") {
		t.Errorf("a corpus without op/ passed the walk: %v", err)
	}
	if err := os.Mkdir(filepath.Join(tmp, "op"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, _, err := m3Walk(tmp); err == nil {
		t.Error("a walk over empty M3 dirs passed")
	}
}

func TestPerlGitTFaithfulOnM3DirsFailsOnUnparsedFile(t *testing.T) {
	tmp := t.TempDir()
	for _, dir := range m3Dirs {
		if err := os.MkdirAll(filepath.Join(tmp, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(name, src string) {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("op/good.t", "print 1;\n")
	if _, _, err := m3Walk(tmp); err != nil {
		t.Fatalf("a parsing fixture failed the walk: %v", err)
	}
	// perl rejects `&= 2;`, and the parser refuses it (bareampersand_test.go).
	write("uni/bad.t", "&= 2;\n")
	if _, _, err := m3Walk(tmp); err == nil || !strings.Contains(err.Error(), "uni/bad.t") {
		t.Errorf("a file holding an Unknown passed the walk: %v", err)
	}
}

func TestFaithfulQueueRowsNameM3Files(t *testing.T) {
	tDir, _ := perlGitTFiles(t)
	if err := rowPathsErr(readQueue(t).Rows, tDir); err != nil {
		t.Error(err)
	}
	for _, path := range []string{"comp/parser.t", "op/no-such-file.t", "opx/a.t"} {
		if err := rowPathsErr([]queueRow{{Path: path}}, tDir); err == nil {
			t.Errorf("a row naming %s passed", path)
		}
	}
}

func TestFaithfulQueueRowsAreUnique(t *testing.T) {
	dup := goodQueue + "op/a.t:9 dropped-paren ( Variable ( again\n"
	if _, err := parseQueue(dup); err == nil || !strings.Contains(err.Error(), "op/a.t") {
		t.Errorf("two rows for one path passed: %v", err)
	}
}

// queueFixture is the computed rows rendered and read back, with one row
// changed by edit; it fails the test when no row has the given class.
func queueFixture(t *testing.T, rows []queueRow, class string, edit func(*queueRow)) (faithfulQueue, string) {
	t.Helper()
	for i, r := range rows {
		if class == "" || r.Class == class {
			changed := slices.Clone(rows)
			edit(&changed[i])
			q, err := parseQueue(renderQueue(changed))
			if err != nil {
				t.Fatal(err)
			}
			return q, r.Path
		}
	}
	t.Fatalf("no computed row has class %q to build the fixture from", class)
	return faithfulQueue{}, ""
}

// fixtureQueueRows are queue rows computed without the corpus, for the
// tests of the queue's own checks: the M3 walk computes none once
// re-emission is 100% there. One is a fixture tree's unfaithful file,
// `print "a\n"; use strict` with no last `;`, which canon ends with one;
// the other is `my $y = -2 ** 2` regrouped by hand with the minus as
// `**`'s left operand, which canon writes `(-2) ** 2`, a canon-adds-paren-op
// row.
func fixtureQueueRows(t *testing.T) []queueRow {
	t.Helper()
	tmp := t.TempDir()
	for _, dir := range m3Dirs {
		if err := os.MkdirAll(filepath.Join(tmp, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(tmp, "uni/bad.t"), []byte("print \"a\\n\"; use strict"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, rows, err := m3Walk(tmp)
	if err != nil || len(rows) != 1 {
		t.Fatalf("the fixture tree's walk: rows %v, error %v", rows, err)
	}

	src := "my $y = -2 ** 2;\n"
	n := parse.Parse([]byte(src))
	decl := n.Children[0].Children[0]
	neg := decl.Children[1]
	pow := neg.Children[0]
	neg.Children = []*parse.Node{pow.Children[0]}
	neg.End = pow.Children[0].End
	pow.Children[0], pow.Start = neg, neg.Start
	decl.Children[1] = pow
	r, differs := firstDifference("op/regrouped.t", n, []byte(src))
	if !differs {
		t.Fatalf("the regrouped %q re-emits faithfully", src)
	}
	return append(rows, r)
}

func TestFaithfulQueueLineIsFirstDifference(t *testing.T) {
	rows := fixtureQueueRows(t)
	q, path := queueFixture(t, rows, "", func(r *queueRow) { r.Line++ })
	errs := queueDiff(q, rows)
	if len(errs) != 1 || !strings.Contains(errs[0], path) || !strings.Contains(errs[0], "first difference") {
		t.Errorf("a row off its first difference by one line: %q", errs)
	}
}

func TestFaithfulQueueMatchesTree(t *testing.T) {
	_, _, rows := m3Corpus(t)
	for _, r := range rows {
		t.Logf("%s:%d cause %q class %s", r.Path, r.Line, r.cause(), r.Class)
	}
	t.Logf("%d unfaithful M3-dir files; causes per class %v", len(rows), causeCounts(rows))
	if *updateRatchet {
		if err := os.WriteFile(faithfulQueuePath, []byte(renderQueue(rows)), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("rewrote %s; commit it with the change that moved it", faithfulQueuePath)
		return
	}
	for _, e := range queueDiff(readQueue(t), rows) {
		t.Error(e)
	}
}

func TestFaithfulQueueClassIsComputed(t *testing.T) {
	rows := fixtureQueueRows(t)
	for name, edit := range map[string]func(*queueRow){
		"relabelled other": func(r *queueRow) { r.Class = "other" },
		"wrong cause":      func(r *queueRow) { r.Prev = "declarator" },
	} {
		q, path := queueFixture(t, rows, "canon-adds-paren-op", edit)
		errs := queueDiff(q, rows)
		if !slices.ContainsFunc(errs, func(e string) bool { return strings.HasPrefix(e, path+": row says") }) {
			t.Errorf("%s: a %s row passed: %q", name, path, errs)
		}
	}
}

func TestFaithfulQueueCauseDeterminesClass(t *testing.T) {
	classOf := map[string]string{}
	for _, r := range readQueue(t).Rows {
		if prior, ok := classOf[r.cause()]; ok && prior != r.Class {
			t.Errorf("cause %q is both %s and %s", r.cause(), prior, r.Class)
		}
		classOf[r.cause()] = r.Class
	}

	reached := map[int]bool{}
	for _, f := range []struct{ src, canon, prev, class string }{
		{"HeredocBody", "{", ",", "heredoc"},
		{"(", "Variable", "(", "dropped-paren"},
		{"special-sub", "sub", ";", "subless-special-sub"},
		{"{", "Variable", "DerefSigil", "stmt-boundary"},
		{"EOF", ";", ")", "stmt-boundary"},
		{"call", ";", "}", "stmt-boundary"},
		{"Variable", "(", `\`, "canon-adds-paren-op"},
		{"named-unary", "(", "call", "canon-adds-paren-term"},
		{"Variable", "Number", ",", "other"},
	} {
		class, rule := classify(f.src, f.canon, f.prev)
		if class != f.class {
			t.Errorf("cause %s %s %s: class %s, want %s", f.src, f.canon, f.prev, class, f.class)
		}
		reached[rule] = true
	}
	for rule := 1; rule <= 9; rule++ {
		if !reached[rule] {
			t.Errorf("no fixture reaches table rule %d", rule)
		}
	}

	// A word's category comes from its own stream: a source word followed
	// by `(` is a call.
	for _, f := range []struct{ src, want string }{
		{"foo(1);", "call"},
		{"foo;", "bareword"},
		{"print(1);", "list-op"},
		{"defined $x;", "named-unary"},
		{"not $x;", "word-op"},
		{"my $x;", "declarator"},
		{"DESTROY {}", "special-sub"},
		{"sub f {}", "sub"},
	} {
		if got := encodeToken(sigToks([]byte(f.src)), 0); got != f.want {
			t.Errorf("%q: first token encodes %s, want %s", f.src, got, f.want)
		}
	}
	toks := sigToks([]byte("$x"))
	if got := encodeToken(toks, -1) + " " + encodeToken(toks, 0) + " " + encodeToken(toks, 1); got != "BOF Variable EOF" {
		t.Errorf("$x encodes %q, want BOF Variable EOF", got)
	}
}

func TestFaithfulQueueCauseZeroGrepIsRedWhileRowExists(t *testing.T) {
	dir := t.TempDir()
	queue := filepath.Join(dir, "queue.txt")
	line := causeZeroLine(queue, "dropped-paren", []string{"( Variable ("})
	ere := strings.ReplaceAll(line, "grep -Fq", "grep -Eq")
	run := func(content, cmd string) int {
		if err := os.WriteFile(queue, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		err := exec.Command("sh", "-c", cmd).Run()
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return exit.ExitCode()
		}
		if err != nil {
			t.Fatal(err)
		}
		return 0
	}
	with := "op/a.t:3 dropped-paren ( Variable ( r\n"
	without := "op/b.t:3 heredoc HeredocBody { , r\n"
	if run(with, line) == 0 {
		t.Errorf("the cause-zero line passed while the row exists: %s", line)
	}
	if code := run(without, line); code != 0 {
		t.Errorf("the cause-zero line exited %d without the row: %s", code, line)
	}
	if run(with, ere) != 0 || run(without, ere) != 0 {
		t.Errorf("the grep -E form is expected to pass both ways, which is why it is forbidden: %s", ere)
	}
}
