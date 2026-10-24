// ABOUTME: Tests that the committed rv2hv wider count fell below the counts m3start.txt recorded
// ABOUTME: once op/magic.t was answered and once arrow hash reads were decided.

package parseoracle_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// sitesPath is the committed per-site ratchet.
var sitesPath = filepath.Join("testdata", "ratchet", "sites.txt")

// sitesTotal reads one `total SCOPE MARKER BUCKET COUNT` row of sites.txt.
func sitesTotal(path, scope, marker, bucket string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	for _, line := range strings.Split(string(data), "\n") {
		f := strings.Fields(line)
		if len(f) == 5 && f[0] == "total" && f[1] == scope && f[2] == marker && f[3] == bucket {
			return strconv.Atoi(f[4])
		}
	}
	return 0, fmt.Errorf("%s has no `total %s %s %s` row", path, scope, marker, bucket)
}

// rv2hvFellProblems says why sites.txt's rv2hv wider total is not below
// m3start.txt's key, or that either is missing.
func rv2hvFellProblems(sites, m3start, key string) []string {
	now, err := sitesTotal(sites, "all", "rv2hv", "wider")
	if err != nil {
		return []string{err.Error()}
	}
	keys, err := readM3Start(m3start)
	if err != nil {
		return []string{err.Error()}
	}
	v, ok := keys[key]
	if !ok {
		return []string{m3start + " records no " + key}
	}
	before, err := strconv.Atoi(v)
	if err != nil {
		return []string{fmt.Sprintf("%s: %s: %v", m3start, key, err)}
	}
	if now >= before {
		return []string{fmt.Sprintf("rv2hv wider is %d, not below %s = %d", now, key, before)}
	}
	return nil
}

// TestRv2hvWiderFellFromM3Start: deciding arrow hash reads lowered the
// corpus's wider rv2hv sites from the count recorded once op/magic.t was
// answered.
func TestRv2hvWiderFellFromM3Start(t *testing.T) {
	for _, p := range rv2hvFellProblems(sitesPath, m3startPath, "rv2hv_wider_after_magic") {
		t.Error(p)
	}
}

// TestRv2hvWiderFellBelowArrowFix: deciding a named hash's rv2hv from its
// declaration lowered the wider count below the one recorded once arrow
// hash reads were decided.
func TestRv2hvWiderFellBelowArrowFix(t *testing.T) {
	for _, p := range rv2hvFellProblems(sitesPath, m3startPath, "rv2hv_wider_after_arrow") {
		t.Error(p)
	}
}

// TestRv2hvWiderFellRejectsNoFallAndMissingValues: the check fails when
// the count did not fall, and when either value is absent.
func TestRv2hvWiderFellRejectsNoFallAndMissingValues(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	sites := write("sites.txt", "total all rv2hv wider 416\n")
	noSites := write("nosites.txt", "total all rv2hv exact 3\n")
	m3 := write("m3.txt", "rv2hv_wider_after_magic = 416\n")
	noKey := write("nokey.txt", "rv2hv_wider = 396\n")
	for name, c := range map[string][2]string{
		"no fall":      {sites, m3},
		"no sites row": {noSites, m3},
		"no magic key": {sites, noKey},
	} {
		if len(rv2hvFellProblems(c[0], c[1], "rv2hv_wider_after_magic")) == 0 {
			t.Errorf("%s: the check passed", name)
		}
	}
	fell := write("fell.txt", "total all rv2hv wider 415\n")
	if p := rv2hvFellProblems(fell, m3, "rv2hv_wider_after_magic"); len(p) != 0 {
		t.Errorf("a fall from 416 to 415 failed: %v", p)
	}
}
