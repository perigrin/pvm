// ABOUTME: Checks the committed precedence referee fixtures (testdata/precedence/) against bison's committed report.
// ABOUTME: Every (rule, token) resolution is covered by a sentence or reviewed; nothing here runs bison.

package chart

import (
	"bufio"
	"flag"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/chart/lex"
)

var precedenceDir = filepath.Join("testdata", "precedence")

// bisonConflict is one line of bison's report of perly.y, as
// offline/referee/sentences.sh copies it into bison-conflicts.txt.
var bisonConflict = regexp.MustCompile(`^\s*Conflict between rule (\d+) and token (\S+) resolved as (shift|reduce|an error) \(.*\)\.$`)

// The two reasons spec §8 allows a resolution to stay uncovered.
const (
	reasonTimedOut    = "reason: generation timed out"
	reasonUnreachable = "reason: the rendering cannot reach it"
)

// fixtureLines returns a file's lines less blank ones and # comments.
func fixtureLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	sc.Buffer(nil, 1<<20)
	for sc.Scan() {
		if l := sc.Text(); l != "" && !strings.HasPrefix(l, "#") {
			lines = append(lines, l)
		}
	}
	return lines, sc.Err()
}

// checkPrecedenceFixtures checks the four fixture files in dir and returns
// how many resolutions a kept sentence covers, and how many there are. A
// sentence is kept when its rendering round-trips (TestPrecedenceRenderingRoundTrips).
func checkPrecedenceFixtures(dir string) (covered, total int, err error) {
	read := func(name string) []string {
		if err != nil {
			return nil
		}
		var l []string
		l, err = fixtureLines(filepath.Join(dir, name))
		return l
	}
	conflicts, resolutions, sentences, uncovered := read("bison-conflicts.txt"), read("resolutions.txt"), read("sentences.txt"), read("uncovered-reviewed.txt")
	if err != nil {
		return 0, 0, err
	}

	action := map[string]string{} // "rule token" => shift, reduce or error
	for _, l := range conflicts {
		m := bisonConflict.FindStringSubmatch(l)
		if m == nil {
			return 0, 0, fmt.Errorf("bison-conflicts.txt: not a resolution: %q", l)
		}
		action[m[1]+" "+m[2]] = strings.TrimPrefix(m[3], "an ")
	}
	if len(action) != len(conflicts) {
		return 0, 0, fmt.Errorf("bison-conflicts.txt: %d lines, %d resolutions", len(conflicts), len(action))
	}

	verdict, kept := map[string]string{}, map[string]bool{}
	for _, l := range sentences {
		f := strings.Split(l, "\t")
		if len(f) != 5 || f[1] == "" || f[3] == "" || f[4] == "" {
			return 0, 0, fmt.Errorf("sentences.txt: want id, terminals, verdict, grouping, rendering: %q", l)
		}
		if _, dup := verdict[f[0]]; dup {
			return 0, 0, fmt.Errorf("sentences.txt: %s twice", f[0])
		}
		verdict[f[0]], kept[f[0]] = f[2], f[4] != "-"
	}

	listed := map[string]string{} // "rule token" => its reason
	reason := ""
	for _, l := range uncovered {
		switch {
		case strings.HasPrefix(l, "reason:"):
			if l != reasonTimedOut && l != reasonUnreachable {
				return 0, 0, fmt.Errorf("uncovered-reviewed.txt: %q is not a reason spec §8 allows", l)
			}
			reason = l
		case reason == "":
			return 0, 0, fmt.Errorf("uncovered-reviewed.txt: %q is under no reason", l)
		case action[l] == "":
			return 0, 0, fmt.Errorf("uncovered-reviewed.txt: %q is not a resolution in bison's report", l)
		case listed[l] != "":
			return 0, 0, fmt.Errorf("uncovered-reviewed.txt: %q listed twice", l)
		default:
			listed[l] = reason
		}
	}

	seen := map[string]bool{}
	for _, l := range resolutions {
		f := strings.Split(l, "\t")
		if len(f) != 4 {
			return 0, 0, fmt.Errorf("resolutions.txt: want rule, token, action, sentences: %q", l)
		}
		key := f[0] + " " + f[1]
		switch {
		case action[key] == "":
			return 0, 0, fmt.Errorf("resolutions.txt: %s is not in bison's report", key)
		case action[key] != f[2]:
			return 0, 0, fmt.Errorf("resolutions.txt: %s resolved as %s, bison's report says %s", key, f[2], action[key])
		case seen[key]:
			return 0, 0, fmt.Errorf("resolutions.txt: %s listed twice", key)
		}
		seen[key] = true
		anyKept, accepted := false, false
		if f[3] != "-" {
			for _, id := range strings.Split(f[3], ",") {
				v, ok := verdict[id]
				if !ok {
					return 0, 0, fmt.Errorf("resolutions.txt: %s names sentence %s, which sentences.txt lacks", key, id)
				}
				anyKept = anyKept || kept[id]
				accepted = accepted || kept[id] && v == "accept"
			}
		}
		switch {
		case anyKept && listed[key] != "":
			return 0, 0, fmt.Errorf("resolutions.txt: %s is covered and also on the uncovered list", key)
		case anyKept && f[2] == "error" && !accepted:
			return 0, 0, fmt.Errorf("resolutions.txt: %s is resolved as an error and no accepted neighbour covers it", key)
		case anyKept:
			covered++
		case listed[key] == "":
			return 0, 0, fmt.Errorf("resolutions.txt: %s is covered by no kept sentence and not reviewed", key)
		case (listed[key] == reasonTimedOut) != (f[3] == "-"):
			return 0, 0, fmt.Errorf("resolutions.txt: %s is under %q with sentences %s", key, listed[key], f[3])
		}
	}
	for key := range action {
		if !seen[key] {
			return 0, 0, fmt.Errorf("resolutions.txt: %s from bison's report is missing", key)
		}
	}
	return covered, len(action), nil
}

func TestPrecedenceFixturesWellFormed(t *testing.T) {
	covered, total, err := checkPrecedenceFixtures(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("precedence fixtures: %d of %d bison resolutions covered", covered, total)
}

func TestPrecedenceFixturesRejectMalformed(t *testing.T) {
	good := map[string]string{
		"bison-conflicts.txt": "    Conflict between rule 1 and token ADDOP resolved as reduce (%left ADDOP).\n" +
			"    Conflict between rule 2 and token CHRELOP resolved as an error (%nonassoc CHRELOP).\n" +
			"    Conflict between rule 3 and token MULOP resolved as shift (ADDOP < MULOP).\n",
		"resolutions.txt":        "1\tADDOP\treduce\ts0001\n2\tCHRELOP\terror\ts0002,s0001\n3\tMULOP\tshift\t-\n",
		"sentences.txt":          "s0001\tGRAMPROG THING ADDOP THING ADDOP THING\taccept\t1:1-4\tx + x + x\ns0002\tGRAMPROG THING\trefuse@4\t-\tx\n",
		"uncovered-reviewed.txt": "# reviewed\nreason: generation timed out\n3 MULOP\n",
	}
	write := func(t *testing.T, edits map[string]string) string {
		dir := t.TempDir()
		for name, body := range good {
			if e, ok := edits[name]; ok {
				body = e
			}
			if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}
	if _, _, err := checkPrecedenceFixtures(write(t, nil)); err != nil {
		t.Fatalf("the well-formed fixture is refused: %v", err)
	}
	for name, c := range map[string]struct{ file, body, want string }{
		"a resolution missing":                        {"resolutions.txt", "1\tADDOP\treduce\ts0001\n3\tMULOP\tshift\t-\n", "is missing"},
		"an error with no accepted neighbour":         {"resolutions.txt", "1\tADDOP\treduce\ts0001\n2\tCHRELOP\terror\ts0002\n3\tMULOP\tshift\t-\n", "no accepted neighbour"},
		"a dangling sentence id":                      {"resolutions.txt", "1\tADDOP\treduce\ts0009\n2\tCHRELOP\terror\ts0002,s0001\n3\tMULOP\tshift\t-\n", "sentences.txt lacks"},
		"a resolution listed twice":                   {"resolutions.txt", "1\tADDOP\treduce\ts0001\n1\tADDOP\treduce\ts0001\n2\tCHRELOP\terror\ts0002,s0001\n3\tMULOP\tshift\t-\n", "listed twice"},
		"covered and on the uncovered list":           {"uncovered-reviewed.txt", "reason: generation timed out\n3 MULOP\n1 ADDOP\n", "also on the uncovered list"},
		"uncovered and unreviewed":                    {"uncovered-reviewed.txt", "# none\n", "not reviewed"},
		"a reason section 8 does not allow":           {"uncovered-reviewed.txt", "reason: too hard\n3 MULOP\n", "not a reason"},
		"an entry under no reason":                    {"uncovered-reviewed.txt", "3 MULOP\nreason: generation timed out\n", "under no reason"},
		"a sentence with no rendering column":         {"sentences.txt", "s0001\tGRAMPROG THING\taccept\t1:1-4\n", "want id, terminals"},
		"a resolution only an unkept sentence covers": {"sentences.txt", "s0001\tGRAMPROG THING ADDOP THING ADDOP THING\taccept\t1:1-4\t-\ns0002\tGRAMPROG THING\trefuse@4\t-\tx\n", "no kept sentence"},
		"unreachable with no generated sentence":      {"uncovered-reviewed.txt", "reason: the rendering cannot reach it\n3 MULOP\n", "is under"},
		"timed out with a generated sentence":         {"uncovered-reviewed.txt", "reason: generation timed out\n3 MULOP\n1 ADDOP\n", "also on the uncovered list"},
		"an action bison's report contradicts":        {"resolutions.txt", "1\tADDOP\tshift\ts0001\n2\tCHRELOP\terror\ts0002,s0001\n3\tMULOP\tshift\t-\n", "bison's report says"},
		"a resolution bison does not report":          {"resolutions.txt", "1\tADDOP\treduce\ts0001\n2\tCHRELOP\terror\ts0002,s0001\n3\tMULOP\tshift\t-\n4\tPOWOP\tshift\ts0001\n", "not in bison's report"},
	} {
		t.Run(name, func(t *testing.T) {
			_, _, err := checkPrecedenceFixtures(write(t, map[string]string{c.file: c.body}))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("got %v, want an error containing %q", err, c.want)
			}
		})
	}
}

// The rendering (spec §8): each sentence's terminals joined by spaces into source, kept only if
// internal/lexer and chart/lex read that source back as the same terminals.

var updateRendering = flag.Bool("chart.update-rendering", false,
	"rewrite testdata/precedence/sentences.txt's rendering column and nonroundtrip-reviewed.txt from rendering.txt")

// startSelector is the terminal that selects the grammar's start rule and is no token (perly.go: "GRAMPROG is a start
// selector"), so it renders to nothing.
const startSelector = "GRAMPROG"

// precedenceSentence is one row of sentences.txt: its id, terminals, and rendering ("-" when not kept).
type precedenceSentence struct {
	id, rendering string
	terms         []string
}

func readSentences(dir string) ([]precedenceSentence, error) {
	lines, err := fixtureLines(filepath.Join(dir, "sentences.txt"))
	if err != nil {
		return nil, err
	}
	var out []precedenceSentence
	for _, l := range lines {
		f := strings.Split(l, "\t")
		if len(f) < 4 {
			return nil, fmt.Errorf("sentences.txt: %q", l)
		}
		s := precedenceSentence{id: f[0], terms: strings.Fields(f[1]), rendering: "-"}
		if len(f) > 4 {
			s.rendering = f[4]
		}
		out = append(out, s)
	}
	return out, nil
}

// readLexemes reads rendering.txt: per line a terminal and its one lexeme, tab separated.
func readLexemes(dir string) (map[string]string, error) {
	lines, err := fixtureLines(filepath.Join(dir, "rendering.txt"))
	if err != nil {
		return nil, err
	}
	lex := map[string]string{}
	for _, l := range lines {
		t, x, ok := strings.Cut(l, "\t")
		if !ok || x == "" || strings.ContainsAny(x, " \t") {
			return nil, fmt.Errorf("rendering.txt: want a terminal and one lexeme: %q", l)
		}
		if _, dup := lex[t]; dup {
			return nil, fmt.Errorf("rendering.txt: %s has two lexemes", t)
		}
		lex[t] = x
	}
	return lex, nil
}

// readNonRoundTrip reads nonroundtrip-reviewed.txt: sentence ids, each under a reason: line. It returns id to reason.
func readNonRoundTrip(dir string) (map[string]string, error) {
	lines, err := fixtureLines(filepath.Join(dir, "nonroundtrip-reviewed.txt"))
	if err != nil {
		return nil, err
	}
	listed, reason := map[string]string{}, ""
	for _, l := range lines {
		switch {
		case strings.HasPrefix(l, "reason: "):
			reason = l
		case reason == "":
			return nil, fmt.Errorf("nonroundtrip-reviewed.txt: %q is under no reason", l)
		case listed[l] != "":
			return nil, fmt.Errorf("nonroundtrip-reviewed.txt: %q listed twice", l)
		default:
			listed[l] = reason
		}
	}
	return listed, nil
}

// renderTerms joins the lexemes of terms by spaces; ok is false when a terminal has none.
func renderTerms(lexemes map[string]string, terms []string) (src, missing string, ok bool) {
	var words []string
	for _, t := range terms {
		if t == startSelector {
			continue
		}
		x, found := lexemes[t]
		if !found {
			return "", t, false
		}
		words = append(words, x)
	}
	return strings.Join(words, " "), "", true
}

// roundTrip reads src through internal/lexer and chart/lex and returns why it does not lex back to terms, or "" when each
// token's terminal set holds its terminal and there is one token per terminal.
func roundTrip(tab *grammar.Tables, src string, terms []string) string {
	id := map[string]int{}
	for i, s := range tab.Symbols {
		if s.Terminal {
			id[s.Name] = i
		}
	}
	var want []string
	for _, t := range terms {
		if t != startSelector {
			want = append(want, t)
		}
	}
	s, _ := lex.Lex([]byte(src), tab.Tokens)
	if len(s.Tokens) != len(want) {
		return "the lexer makes a different number of tokens than there are terminals"
	}
	for i, t := range s.Tokens {
		if !slices.Contains(tab.TerminalSets[t.TerminalSet], id[want[i]]) {
			return want[i] + " is not among the terminals its token admits"
		}
	}
	return ""
}

// renderingCause is why a sentence is not kept, or "" when it is.
func renderingCause(tab *grammar.Tables, lexemes map[string]string, s precedenceSentence) string {
	src, missing, ok := renderTerms(lexemes, s.terms)
	if !ok {
		if strings.HasPrefix(missing, "PLUGIN_") {
			return "no rendering for " + missing + " until a library operator is in scope"
		}
		return "no single-lexeme rendering for " + missing
	}
	return roundTrip(tab, src, s.terms)
}

func TestPrecedenceRenderingRoundTrips(t *testing.T) {
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	lexemes, err := readLexemes(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	sentences, err := readSentences(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	listed, err := readNonRoundTrip(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	if *updateRendering {
		writeRendering(t, tab, lexemes, sentences)
		return
	}
	kept := 0
	for _, s := range sentences {
		cause := renderingCause(tab, lexemes, s)
		src, _, _ := renderTerms(lexemes, s.terms)
		switch {
		case cause == "" && listed[s.id] != "":
			t.Errorf("%s round-trips and is on nonroundtrip-reviewed.txt", s.id)
		case cause == "" && s.rendering != src:
			t.Errorf("%s: stored rendering %q, the table renders %q", s.id, s.rendering, src)
		case cause == "":
			kept++
		case s.rendering != "-":
			t.Errorf("%s does not round-trip (%s) and keeps a rendering", s.id, cause)
		case listed[s.id] != "reason: "+cause:
			t.Errorf("%s does not round-trip (%s) and is listed under %q", s.id, cause, listed[s.id])
		}
	}
	t.Logf("%d of %d sentences round-trip; %d are on nonroundtrip-reviewed.txt", kept, len(sentences), len(listed))
}

// writeRendering rewrites sentences.txt's rendering column and nonroundtrip-reviewed.txt.
func writeRendering(t *testing.T, tab *grammar.Tables, lexemes map[string]string, sentences []precedenceSentence) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(precedenceDir, "sentences.txt"))
	if err != nil {
		t.Fatal(err)
	}
	byCause, keptNow := map[string][]string{}, map[string]bool{}
	var out strings.Builder
	n := 0
	for _, l := range strings.SplitAfter(string(raw), "\n") {
		if l == "" || strings.HasPrefix(l, "#") {
			out.WriteString(l)
			continue
		}
		f := strings.Split(strings.TrimSuffix(l, "\n"), "\t")[:4]
		s := sentences[n]
		n++
		r := "-"
		if cause := renderingCause(tab, lexemes, s); cause != "" {
			byCause[cause] = append(byCause[cause], s.id)
		} else {
			r, _, _ = renderTerms(lexemes, s.terms)
			keptNow[s.id] = true
		}
		out.WriteString(strings.Join(append(f, r), "\t") + "\n")
	}
	list := "# Sentences whose rendering does not lex back to their terminals, each under the reason it fails (spec section 8).\n" +
		"# Written by `go test ./internal/chart/ -run TestPrecedenceRenderingRoundTrips -chart.update-rendering`; the test\n" +
		"# recomputes every reason, so a sentence moves here or off it only by a change to rendering.txt or the lexer.\n"
	for _, cause := range slices.Sorted(maps.Keys(byCause)) {
		list += "reason: " + cause + "\n" + strings.Join(byCause[cause], "\n") + "\n"
	}
	// A resolution with sentences and none of them kept is unreachable; one with none is generation's (sentences.sh).
	timedOut, unreachable := "", ""
	resolutions, err := fixtureLines(filepath.Join(precedenceDir, "resolutions.txt"))
	if err != nil {
		t.Fatal(err)
	}
	for _, l := range resolutions {
		f := strings.Split(l, "\t")
		switch ids := strings.Split(f[3], ","); {
		case f[3] == "-":
			timedOut += f[0] + " " + f[1] + "\n"
		case !slices.ContainsFunc(ids, func(id string) bool { return keptNow[id] }):
			unreachable += f[0] + " " + f[1] + "\n"
		}
	}
	uncovered := "# Resolutions no kept sentence covers, each under its reason (spec section 8): `reason: generation timed out`\n" +
		"# (no sentence was generated; none at 94e50866) or `reason: the rendering cannot reach it` (every sentence generated\n" +
		"# for it is on nonroundtrip-reviewed.txt). Written by the -chart.update-rendering run that writes\n" +
		"# sentences.txt's rendering column.\n" +
		reasonTimedOut + "\n" + timedOut + reasonUnreachable + "\n" + unreachable
	for name, body := range map[string]string{"sentences.txt": out.String(), "nonroundtrip-reviewed.txt": list, "uncovered-reviewed.txt": uncovered} {
		if err := os.WriteFile(filepath.Join(precedenceDir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPrecedenceRenderingRejectsNonRoundTrip(t *testing.T) {
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	for name, c := range map[string]struct {
		src   string
		terms []string
	}{
		"an inserted ;":                       {"qw(a) ;", []string{"QWLIST"}},
		"UNIOP ( read as FUNC1":               {"ref (", []string{"UNIOP", "PERLY_PAREN_OPEN"}},
		"two tokens for one terminal":         {"qw(a) qw(b)", []string{"QWLIST"}},
		"one token where two terminals stand": {"qw(a)", []string{"QWLIST", "QWLIST"}},
	} {
		t.Run(name, func(t *testing.T) {
			if why := roundTrip(tab, c.src, c.terms); why == "" {
				t.Fatalf("%q round-trips as %v", c.src, c.terms)
			}
		})
	}
	if why := roundTrip(tab, "qw(a) + qw(b) ;", []string{startSelector, "QWLIST", "ADDOP", "QWLIST", "PERLY_SEMICOLON"}); why != "" {
		t.Fatalf("a plain sum does not round-trip: %s", why)
	}
}

func TestPrecedenceRenderingAccountsForEverySentence(t *testing.T) {
	tab, err := committed()
	if err != nil {
		t.Fatal(err)
	}
	lexemes, err := readLexemes(precedenceDir) // refuses a terminal given two lexemes
	if err != nil {
		t.Fatal(err)
	}
	terminal := map[string]bool{}
	for _, s := range tab.Symbols {
		terminal[s.Name] = s.Terminal
	}
	for term := range lexemes {
		if !terminal[term] {
			t.Errorf("rendering.txt renders %s, which is no perly.y terminal", term)
		}
	}
	sentences, err := readSentences(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	listed, err := readNonRoundTrip(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	kept := 0
	for _, s := range sentences {
		switch {
		case s.rendering != "-" && listed[s.id] != "":
			t.Errorf("%s is both kept and listed", s.id)
		case s.rendering != "-":
			kept++
		case listed[s.id] == "":
			t.Errorf("%s is neither kept nor listed", s.id)
		}
	}
	if kept+len(listed) != len(sentences) {
		t.Errorf("%d kept + %d listed != %d sentences (the list names a sentence sentences.txt lacks)", kept, len(listed), len(sentences))
	}
}

// checkUnreachable checks the resolutions under `reason: the rendering cannot reach it`: each has a generated sentence
// and every one of them is on nonroundtrip-reviewed.txt.
func checkUnreachable(dir string) (int, error) {
	uncovered, err := fixtureLines(filepath.Join(dir, "uncovered-reviewed.txt"))
	if err != nil {
		return 0, err
	}
	resolutions, err := fixtureLines(filepath.Join(dir, "resolutions.txt"))
	if err != nil {
		return 0, err
	}
	listed, err := readNonRoundTrip(dir)
	if err != nil {
		return 0, err
	}
	ids := map[string]string{}
	for _, l := range resolutions {
		if f := strings.Split(l, "\t"); len(f) == 4 {
			ids[f[0]+" "+f[1]] = f[3]
		}
	}
	n, under := 0, false
	for _, l := range uncovered {
		if strings.HasPrefix(l, "reason:") {
			under = l == "reason: the rendering cannot reach it"
			continue
		}
		if !under {
			continue
		}
		n++
		if ids[l] == "" || ids[l] == "-" {
			return n, fmt.Errorf("%s is unreachable by rendering and has no generated sentence", l)
		}
		for _, id := range strings.Split(ids[l], ",") {
			if listed[id] == "" {
				return n, fmt.Errorf("%s is unreachable by rendering but %s is not on nonroundtrip-reviewed.txt", l, id)
			}
		}
	}
	return n, nil
}

func TestPrecedenceUnreachableIsNonRoundTrip(t *testing.T) {
	n, err := checkUnreachable(precedenceDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%d resolutions the rendering cannot reach", n)
	write := func(uncovered, reviewed string) string {
		dir := t.TempDir()
		for name, body := range map[string]string{
			"uncovered-reviewed.txt":    uncovered,
			"nonroundtrip-reviewed.txt": reviewed,
			"resolutions.txt":           "1\tADDOP\treduce\ts0001,s0002\n2\tMULOP\tshift\t-\n",
		} {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
				t.Fatal(err)
			}
		}
		return dir
	}
	const reason = "reason: the rendering cannot reach it\n"
	if _, err := checkUnreachable(write(reason+"1 ADDOP\n", "reason: x\ns0001\ns0002\n")); err != nil {
		t.Fatalf("a true use is refused: %v", err)
	}
	for name, c := range map[string]struct{ uncovered, reviewed, want string }{
		"a sentence that is kept":            {reason + "1 ADDOP\n", "reason: x\ns0001\n", "is not on nonroundtrip-reviewed.txt"},
		"a resolution with no sentence":      {reason + "2 MULOP\n", "reason: x\ns0001\ns0002\n", "no generated sentence"},
		"a resolution resolutions.txt lacks": {reason + "9 POWOP\n", "reason: x\ns0001\n", "no generated sentence"},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := checkUnreachable(write(c.uncovered, c.reviewed)); err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("got %v, want an error containing %q", err, c.want)
			}
		})
	}
}
