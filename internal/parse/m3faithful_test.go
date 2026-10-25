// ABOUTME: Holds canonical re-emission at 100% on the M3 dirs (t/op mro uni lib test_pl): every file that parses is faithful.
// ABOUTME: No exception list: the work queue is never read here, so a row in it cannot excuse an unfaithful file.

package parse_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// m3FaithfulProblems reports each unfaithful M3-dir file among rows, the
// rows m3Walk computed. It takes no queue: perigrin's ruling (2026-10-23,
// chain-review round 4) is 100% faithful on the M3 dirs with no defects
// file, so nothing listed anywhere may stand in for a fix.
func m3FaithfulProblems(rows []queueRow) []string {
	var out []string
	for _, r := range rows {
		out = append(out, fmt.Sprintf("%s:%d re-emits unfaithfully: source has %q, canon has %q (after %q)",
			r.Path, r.Line, r.Src, r.Canon, r.Prev))
	}
	return out
}

// TestPerlGitTFaithfulOnM3Dirs: every M3-dir file outside perlRejects
// parses (m3Walk fails on an Unknown) and re-emits faithfully.
func TestPerlGitTFaithfulOnM3Dirs(t *testing.T) {
	_, visited, rows := m3Corpus(t)
	for _, p := range m3FaithfulProblems(rows) {
		t.Error(p)
	}
	t.Logf("%d/%d M3-dir files re-emit faithfully", len(visited)-len(rows), len(visited))
}

// TestPerlGitTFaithfulOnM3DirsIgnoresTheQueue: a fixture tree holding one
// unfaithful file, and a queue that lists it and matches the tree, still
// fails the 100% check, so the queue cannot serve as an exception list.
// The fixture is `print "a\n"; use strict` with no last `;`, which canon
// ends with one: a defect outside the M3 dirs, found by 01a16602-6030's
// review.
func TestPerlGitTFaithfulOnM3DirsIgnoresTheQueue(t *testing.T) {
	tmp := t.TempDir()
	for _, dir := range m3Dirs {
		if err := os.MkdirAll(filepath.Join(tmp, dir), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for name, src := range map[string]string{
		"op/good.t": "print 1;\n",
		"uni/bad.t": "print \"a\\n\"; use strict",
	} {
		if err := os.WriteFile(filepath.Join(tmp, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	_, rows, err := m3Walk(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Path != "uni/bad.t" {
		t.Fatalf("the fixture's walk computed rows %v, want one for uni/bad.t", rows)
	}
	q, err := parseQueue(renderQueue(rows))
	if err != nil {
		t.Fatal(err)
	}
	if errs := queueDiff(q, rows); len(errs) != 0 {
		t.Fatalf("a queue rendered from the rows does not match them: %q", errs)
	}
	problems := m3FaithfulProblems(rows)
	if len(problems) != 1 || !strings.Contains(problems[0], "uni/bad.t") {
		t.Errorf("an unfaithful file the queue lists passed the 100%% check: %q", problems)
	}
}

// TestFaithfulOtherRowsFixed: the queue's `other` class, the rows no
// triage issue's class fits, holds nothing. It held no rows when issue
// 01a16795-75e8 began (faithful_queue.txt's `# causes other = 0` at
// 99cd2e5d), and 01a16602-608d emptied the queue. So the class is held to
// zero on the walk itself, and every M3-dir shape the triage issues fixed
// is already held by their own tests.
func TestFaithfulOtherRowsFixed(t *testing.T) {
	_, _, rows := m3Corpus(t)
	for _, r := range rows {
		if r.Class == "other" {
			t.Errorf("%s:%d is an `other` row: source has %q, canon has %q", r.Path, r.Line, r.Src, r.Canon)
		}
	}
	if n := causeCounts(rows)["other"]; n != 0 {
		t.Errorf("%d `other` cause(s) on the M3 walk, want 0", n)
	}
}
