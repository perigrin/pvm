// ABOUTME: Guards the two ways the fidelity gate can be made advisory while every other CI test stays green.
// ABOUTME: A job or step marked continue-on-error, and a sweep that skips itself when the corpus is missing.

package parseoracle_test

import (
	"os/exec"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// TestCIGateCannotBeAdvisory: `continue-on-error: true` reports a failed job or
// step as a success. Anywhere in this workflow it turns the gate into
// decoration -- the sweep runs, finds regressions, fails, and the commit gets
// a green check anyway. That is the same neutralised-guard shape as a
// re-baselining job, reached by a different route.
//
// Both levels matter. GitHub honours the key on a job and on each step
// independently, so guarding one leaves the other open, and the step level is
// the one that hides best: a `continue-on-error` on the sweep step alone is
// four words in the middle of a long file.
func TestCIGateCannotBeAdvisory(t *testing.T) {
	const key = "continue-on-error"

	// Comments are stripped first, because the header above discusses this
	// very construct in prose and an earlier assertion in this suite failed
	// against its own explanatory text. A `#`-commented key is inert on the
	// runner, so it must be inert here too.
	if !strings.Contains(uncommented(readWorkflow(t)), key) {
		return // nothing set anywhere: the common case, and the passing one
	}

	var parsed struct {
		Jobs map[string]struct {
			ContinueOnError any `yaml:"continue-on-error"`
			Steps           []struct {
				Name            string `yaml:"name"`
				ContinueOnError any    `yaml:"continue-on-error"`
			} `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal([]byte(readWorkflow(t)), &parsed); err != nil {
		t.Fatalf("%s is not valid YAML: %v", workflow, err)
	}

	for name, job := range parsed.Jobs {
		if job.ContinueOnError != nil {
			t.Errorf("job %q sets %s: %v. A failing gate that reports success is "+
				"not a gate -- the sweep would find regressions and the commit would "+
				"still go green", name, key, job.ContinueOnError)
		}
		for i, step := range job.Steps {
			if step.ContinueOnError != nil {
				t.Errorf("job %q step %d (%q) sets %s: %v. The same bypass one level "+
					"down, and harder to see in review", name, i, step.Name, key, step.ContinueOnError)
			}
		}
	}
}

// TestCISweepFailureIsNotSwallowed covers the shell-level version of the same
// bypass. `go test ... || true` makes the step succeed however the sweep
// exited, without the words continue-on-error appearing anywhere for a
// reviewer to catch.
func TestCISweepFailureIsNotSwallowed(t *testing.T) {
	for _, step := range workflowSteps(t, readWorkflow(t)) {
		if !strings.Contains(step.Run, "TestRatchetCorpus") {
			continue
		}
		run := uncommented(step.Run)
		for _, bad := range []string{"|| true", "|| :", "exit 0", "set +e"} {
			if strings.Contains(run, bad) {
				t.Errorf("the sweep step contains %q: the ratchet's exit status is "+
					"the gate's only output, and this discards it", bad)
			}
		}
	}
}

// TestCIRequiresTheCorpusToBePresent is the second bypass, in the test rather
// than the YAML.
//
// corpusReport skips when CorpusRoot finds no checkout, which is right for a
// developer and wrong for the gate: on a runner where the perl5 checkout did
// not materialise -- a cache miss, a network blip, a wrong revision -- the
// sweep skips, the job exits 0, and CI reports success having measured zero
// files. That is the perl-lsp failure exactly, a green check that means
// nothing.
//
// The fix is an assertion the workflow makes about its own world:
// $PARSEORACLE_REQUIRE_CORPUS says "a missing corpus here is a failure, not a
// local convenience", so the same test skips for a developer and fails for
// the gate.
func TestCIRequiresTheCorpusToBePresent(t *testing.T) {
	var sweep map[string]string
	for _, step := range workflowSteps(t, readWorkflow(t)) {
		if strings.Contains(step.Run, "TestRatchetCorpus") {
			sweep = step.Env
		}
	}
	if sweep == nil {
		t.Fatal("no step runs TestRatchetCorpus")
	}
	if sweep[requireCorpusEnv] == "" {
		t.Errorf("the sweep step does not set %s: without it a runner whose perl5 "+
			"checkout failed to materialise skips the sweep and reports success "+
			"having measured nothing", requireCorpusEnv)
	}
}

// TestCISweepSelectsATestThatExists is the bypass one level out from the two
// the review named, and the cheapest of the three to reach by accident.
//
// `go test -run TestRatchetCorpusX` exits 0 with "no tests to run". Every
// other guard in this suite survives that: they ask whether the string
// "TestRatchetCorpus" appears in the workflow, and "TestRatchetCorpusX"
// contains it. A one-character typo -- or a rename that misses the YAML --
// therefore produces a job that runs nothing, passes, and is indistinguishable
// on the commit from a job that swept 620 files.
//
// Substring matching cannot answer this, so the pattern is resolved against
// the package the way the runner would resolve it: `go test -list` prints the
// tests that actually exist, and at least one of them must match.
func TestCISweepSelectsATestThatExists(t *testing.T) {
	var run string
	for _, step := range workflowSteps(t, readWorkflow(t)) {
		if strings.Contains(step.Run, "TestRatchetCorpus") {
			run = uncommented(step.Run)
		}
	}
	if run == "" {
		t.Fatal("no step runs TestRatchetCorpus")
	}

	fields := strings.Fields(run)
	pattern := ""
	for i, f := range fields {
		if f == "-run" && i+1 < len(fields) {
			pattern = fields[i+1]
		}
	}
	if pattern == "" {
		t.Fatal("the sweep step passes no -run pattern, so it would run the " +
			"whole package rather than the ratchet")
	}

	// -list takes the same regexp as -run and prints matching test names
	// without running them, so this asks the toolchain the question rather
	// than reimplementing Go's matching rules.
	out, err := exec.Command("go", "test", "-list", pattern, ".").CombinedOutput()
	if err != nil {
		t.Fatalf("go test -list %s: %v\n%s", pattern, err, out)
	}
	var matched []string
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "Test") {
			matched = append(matched, line)
		}
	}
	if len(matched) == 0 {
		t.Errorf("the sweep's -run pattern %q matches no test in this package: "+
			"`go test` exits 0 with \"no tests to run\", so the job would sweep "+
			"nothing and still report success", pattern)
	}
}

// TestCorpusRequiredReadsTheEnvironment proves the knob is connected, the way
// TestSweepTimeoutReadsTheEnvironment does for the timeout. A workflow that
// exports a variable nothing reads is the green check this issue is about.
func TestCorpusRequiredReadsTheEnvironment(t *testing.T) {
	if corpusRequired() {
		t.Errorf("unset $%s must mean a missing corpus skips", requireCorpusEnv)
	}
	for _, value := range []string{"1", "true", "yes"} {
		t.Setenv(requireCorpusEnv, value)
		if !corpusRequired() {
			t.Errorf("$%s=%q must mean a missing corpus fails", requireCorpusEnv, value)
		}
	}
	// An explicit off switch, so a developer who inherits the variable from a
	// shell profile is not stuck with a failing suite.
	for _, value := range []string{"0", "false", ""} {
		t.Setenv(requireCorpusEnv, value)
		if corpusRequired() {
			t.Errorf("$%s=%q must mean a missing corpus skips", requireCorpusEnv, value)
		}
	}
}
