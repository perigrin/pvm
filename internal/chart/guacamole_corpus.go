// ABOUTME: The perl-rejects rule over the Guacamole corpus: perl -c's verdict on each parses case against perl-rejects.txt.
// ABOUTME: Shared by the corpus unit test (hand-built cases) and the chart ratchet's rejects-disagree sweep.

// Package chart is the Earley chart-parser spike (docs/plans/2026-10-29-chart-parser-design.md).
package chart

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// RejectsDisagree runs perl -c on every *.pl case in casesDir, with stubsDir
// as PERL5LIB, and returns one message per case perl disagrees with the
// rejects list about, each starting with the case's bare file name. The list
// names case files by bare file name, one per line; a listed name with no
// case file is a disagreement too.
func RejectsDisagree(perl, casesDir, rejectsFile, stubsDir string) ([]string, error) {
	list, err := os.ReadFile(rejectsFile)
	if err != nil {
		return nil, err
	}
	listed := map[string]bool{}
	for _, name := range strings.Fields(string(list)) {
		listed[name] = true
	}
	cases, err := filepath.Glob(filepath.Join(casesDir, "*.pl"))
	if err != nil {
		return nil, err
	}
	var out []string
	for _, c := range cases {
		name := filepath.Base(c)
		cmd := exec.Command(perl, "-c", c)
		cmd.Env = append(os.Environ(), "PERL5LIB="+stubsDir)
		msg, runErr := cmd.CombinedOutput()
		if _, exited := runErr.(*exec.ExitError); runErr != nil && !exited {
			return nil, fmt.Errorf("perl -c %s: %w", c, runErr)
		}
		rejected := runErr != nil
		switch {
		case rejected && !listed[name]:
			out = append(out, fmt.Sprintf("%s: perl rejects it but %s does not list it: %s",
				name, filepath.Base(rejectsFile), firstLine(msg)))
		case !rejected && listed[name]:
			out = append(out, fmt.Sprintf("%s: %s lists it but perl accepts it", name, filepath.Base(rejectsFile)))
		}
		delete(listed, name)
	}
	for name := range listed {
		out = append(out, fmt.Sprintf("%s: %s lists it but there is no such case", name, filepath.Base(rejectsFile)))
	}
	sort.Strings(out)
	return out, nil
}

func firstLine(b []byte) string {
	s, _, _ := strings.Cut(string(b), "\n")
	return s
}
