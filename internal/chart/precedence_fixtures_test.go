// ABOUTME: Checks the committed precedence referee fixtures (testdata/precedence/) against bison's committed report.
// ABOUTME: Every (rule, token) resolution is covered by a sentence or reviewed; nothing here runs bison.

package chart

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var precedenceDir = filepath.Join("testdata", "precedence")

// bisonConflict is one line of bison's report of perly.y, as
// offline/referee/sentences.sh copies it into bison-conflicts.txt.
var bisonConflict = regexp.MustCompile(`^\s*Conflict between rule (\d+) and token (\S+) resolved as (shift|reduce|an error) \(.*\)\.$`)

// uncoveredReasons are the two reasons spec §8 allows a resolution to stay uncovered.
var uncoveredReasons = map[string]bool{
	"reason: generation timed out":          true,
	"reason: the rendering cannot reach it": true,
}

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
// how many resolutions a sentence covers, and how many there are.
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

	verdict := map[string]string{}
	for _, l := range sentences {
		f := strings.Split(l, "\t")
		if len(f) != 4 || f[1] == "" || f[3] == "" {
			return 0, 0, fmt.Errorf("sentences.txt: want id, terminals, verdict, grouping: %q", l)
		}
		if _, dup := verdict[f[0]]; dup {
			return 0, 0, fmt.Errorf("sentences.txt: %s twice", f[0])
		}
		verdict[f[0]] = f[2]
	}

	listed := map[string]bool{}
	reason := ""
	for _, l := range uncovered {
		switch {
		case strings.HasPrefix(l, "reason:"):
			if !uncoveredReasons[l] {
				return 0, 0, fmt.Errorf("uncovered-reviewed.txt: %q is not a reason spec §8 allows", l)
			}
			reason = l
		case reason == "":
			return 0, 0, fmt.Errorf("uncovered-reviewed.txt: %q is under no reason", l)
		case action[l] == "":
			return 0, 0, fmt.Errorf("uncovered-reviewed.txt: %q is not a resolution in bison's report", l)
		case listed[l]:
			return 0, 0, fmt.Errorf("uncovered-reviewed.txt: %q listed twice", l)
		default:
			listed[l] = true
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
		if f[3] == "-" {
			if !listed[key] {
				return 0, 0, fmt.Errorf("resolutions.txt: %s is covered by no sentence and not reviewed", key)
			}
			continue
		}
		if listed[key] {
			return 0, 0, fmt.Errorf("resolutions.txt: %s is covered and also on the uncovered list", key)
		}
		accepted := false
		for _, id := range strings.Split(f[3], ",") {
			v, ok := verdict[id]
			if !ok {
				return 0, 0, fmt.Errorf("resolutions.txt: %s names sentence %s, which sentences.txt lacks", key, id)
			}
			accepted = accepted || v == "accept"
		}
		if f[2] == "error" && !accepted {
			return 0, 0, fmt.Errorf("resolutions.txt: %s is resolved as an error and no accepted neighbour covers it", key)
		}
		covered++
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
		"sentences.txt":          "s0001\tGRAMPROG THING ADDOP THING ADDOP THING\taccept\t1:1-4\ns0002\tGRAMPROG THING\trefuse@4\t-\n",
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
		"a resolution missing":                 {"resolutions.txt", "1\tADDOP\treduce\ts0001\n3\tMULOP\tshift\t-\n", "is missing"},
		"an error with no accepted neighbour":  {"resolutions.txt", "1\tADDOP\treduce\ts0001\n2\tCHRELOP\terror\ts0002\n3\tMULOP\tshift\t-\n", "no accepted neighbour"},
		"a dangling sentence id":               {"resolutions.txt", "1\tADDOP\treduce\ts0009\n2\tCHRELOP\terror\ts0002,s0001\n3\tMULOP\tshift\t-\n", "sentences.txt lacks"},
		"a resolution listed twice":            {"resolutions.txt", "1\tADDOP\treduce\ts0001\n1\tADDOP\treduce\ts0001\n2\tCHRELOP\terror\ts0002,s0001\n3\tMULOP\tshift\t-\n", "listed twice"},
		"covered and on the uncovered list":    {"uncovered-reviewed.txt", "reason: generation timed out\n3 MULOP\n1 ADDOP\n", "also on the uncovered list"},
		"uncovered and unreviewed":             {"uncovered-reviewed.txt", "# none\n", "not reviewed"},
		"a reason section 8 does not allow":    {"uncovered-reviewed.txt", "reason: too hard\n3 MULOP\n", "not a reason"},
		"an entry under no reason":             {"uncovered-reviewed.txt", "3 MULOP\nreason: generation timed out\n", "under no reason"},
		"an action bison's report contradicts": {"resolutions.txt", "1\tADDOP\tshift\ts0001\n2\tCHRELOP\terror\ts0002,s0001\n3\tMULOP\tshift\t-\n", "bison's report says"},
		"a resolution bison does not report":   {"resolutions.txt", "1\tADDOP\treduce\ts0001\n2\tCHRELOP\terror\ts0002,s0001\n3\tMULOP\tshift\t-\n4\tPOWOP\tshift\ts0001\n", "not in bison's report"},
	} {
		t.Run(name, func(t *testing.T) {
			_, _, err := checkPrecedenceFixtures(write(t, map[string]string{c.file: c.body}))
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("got %v, want an error containing %q", err, c.want)
			}
		})
	}
}
