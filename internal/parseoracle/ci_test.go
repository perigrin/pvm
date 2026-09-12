// ABOUTME: Tests the parts of the CI fidelity job that can be tested without a GitHub runner.
// ABOUTME: The workflow YAML itself cannot run here, so what it delegates to is tested instead.

package parseoracle_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"tamarou.com/pvm/internal/parseoracle"
)

// repoFile resolves a path relative to the repository root. The tests in this
// file reach outside the package, because what they are checking — a shell
// script and a workflow file — is what CI runs, and testing a copy would
// prove nothing about what ships.
func repoFile(t *testing.T, rel ...string) string {
	t.Helper()
	return filepath.Join(append([]string{"..", ".."}, rel...)...)
}

const (
	pinScript = "parseoracle-pin.sh"
	workflow  = "parse-fidelity.yml"

	// timeoutEnv overrides the per-file oracle timeout for a corpus sweep.
	//
	// It exists because DefaultTimeout is calibrated for this machine and CI
	// is not. Measured here, unloaded: op/pack.t costs 37s of CPU, against a
	// 60s default. B::Concise walks the whole optree and op/pack.t builds a
	// very large one; that cost is the file's, not the load's, so a runner
	// even slightly slower crosses the line. The resulting timeout is
	// recorded as a runner error, the ratchet sees a file move out of its
	// bucket, and it reports a regression caused by nothing.
	timeoutEnv = "PARSEORACLE_TIMEOUT"
)

// sweepTimeout is the per-file oracle timeout for a corpus sweep: $PARSEORACLE_TIMEOUT
// when set, and zero (meaning DefaultTimeout) otherwise.
func sweepTimeout(t *testing.T) time.Duration {
	t.Helper()

	raw := os.Getenv(timeoutEnv)
	if raw == "" {
		return 0
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		t.Fatalf("$%s=%q is not a duration: %v", timeoutEnv, raw, err)
	}
	return d
}

// TestSweepTimeoutReadsTheEnvironment proves the knob CI turns is connected
// to something. A workflow that exports an environment variable nothing reads
// is the same green check measuring nothing that this whole issue is about.
func TestSweepTimeoutReadsTheEnvironment(t *testing.T) {
	if got := sweepTimeout(t); got != 0 {
		t.Errorf("unset $%s must mean DefaultTimeout (zero), got %v", timeoutEnv, got)
	}

	t.Setenv(timeoutEnv, "4m")
	if got, want := sweepTimeout(t), 4*time.Minute; got != want {
		t.Errorf("$%s=4m gave %v, want %v", timeoutEnv, got, want)
	}
}

// TestPinScriptAgreesWithReadPin is the reason the script exists as a script.
//
// The workflow needs the pin's two halves before Go is even set up: the
// interpreter version selects which perl to install, and the revision selects
// which perl5 commit to check out. Reading them in shell means a second
// parser for a file that already has one, so this test pins the two together:
// whatever ReadPin says, the script must say.
func TestPinScriptAgreesWithReadPin(t *testing.T) {
	script := repoFile(t, ".github", "scripts", pinScript)
	pinPath := filepath.Join("testdata", "corpus.pin")

	pin, err := parseoracle.ReadPin(pinPath)
	if err != nil {
		t.Fatalf("ReadPin: %v", err)
	}

	abs, err := filepath.Abs(pinPath)
	if err != nil {
		t.Fatalf("resolving the pin: %v", err)
	}
	out, err := exec.Command("sh", script, abs).Output()
	if err != nil {
		t.Fatalf("%s: %v", script, err)
	}

	got := parseFields(string(out))
	for _, tc := range []struct{ key, want string }{
		{"interpreter", pin.Interpreter},
		{"revision", pin.Revision},
	} {
		if got[tc.key] != tc.want {
			t.Errorf("%s printed %s=%q, ReadPin says %q", pinScript, tc.key, got[tc.key], tc.want)
		}
	}

	// perl_version is the same interpreter in the dotted form
	// actions-setup-perl takes. 5.042000 is 5.42.0; getting this wrong
	// installs a perl the pin does not describe, and every verdict below
	// the baseline header was measured against something else.
	if want := "5.42.0"; got["perl_version"] != want {
		t.Errorf("%s printed perl_version=%q, want %q for interpreter %q",
			pinScript, got["perl_version"], want, pin.Interpreter)
	}
}

// TestPinScriptConvertsInterpreterVersions covers the arithmetic on its own,
// because the committed pin exercises exactly one value and a conversion that
// is right once may be right by accident.
func TestPinScriptConvertsInterpreterVersions(t *testing.T) {
	script := repoFile(t, ".github", "scripts", pinScript)

	for _, tc := range []struct{ interpreter, want string }{
		{"5.042000", "5.42.0"},
		{"5.036002", "5.36.2"},
		{"5.008009", "5.8.9"},
		{"5.040000", "5.40.0"},
	} {
		t.Run(tc.interpreter, func(t *testing.T) {
			pin := filepath.Join(t.TempDir(), "corpus.pin")
			body := "interpreter = " + tc.interpreter + "\nrevision    = deadbeef\n"
			if err := os.WriteFile(pin, []byte(body), 0o644); err != nil {
				t.Fatalf("writing a pin: %v", err)
			}

			out, err := exec.Command("sh", script, pin).Output()
			if err != nil {
				t.Fatalf("%s: %v", script, err)
			}
			if got := parseFields(string(out))["perl_version"]; got != tc.want {
				t.Errorf("interpreter %s -> perl_version %q, want %q", tc.interpreter, got, tc.want)
			}
		})
	}
}

// TestPinScriptFailsOnAnIncompletePin: a pin missing either half must abort
// rather than emit an empty value. An empty revision checks out blead's HEAD,
// which is the drift the pin exists to prevent, and it would do so quietly.
func TestPinScriptFailsOnAnIncompletePin(t *testing.T) {
	script := repoFile(t, ".github", "scripts", pinScript)

	for name, body := range map[string]string{
		"no revision":    "interpreter = 5.042000\n",
		"no interpreter": "revision = deadbeef\n",
		"empty":          "# nothing here\n",
	} {
		t.Run(name, func(t *testing.T) {
			pin := filepath.Join(t.TempDir(), "corpus.pin")
			if err := os.WriteFile(pin, []byte(body), 0o644); err != nil {
				t.Fatalf("writing a pin: %v", err)
			}
			if err := exec.Command("sh", script, pin).Run(); err == nil {
				t.Errorf("%s accepted a pin with %s: CI would run against an "+
					"unpinned world and report the skew as a parser change", pinScript, name)
			}
		})
	}
}

// TestCINeverRebaselines is the non-negotiable, asserted where a test can see
// it rather than left as a comment.
//
// A workflow that passes -parseoracle.update rewrites the baseline from its
// own run, so the check always passes and proves nothing. That is how
// perl-lsp arrived at a parser ratchet whose baseline file contains the
// single line `0`. The update is a deliberate local act that lands in the
// same commit as the change that moved it.
func TestCINeverRebaselines(t *testing.T) {
	// Comments are stripped first, because the workflow explains at length
	// WHY it does not re-baseline and a whole-file substring match cannot
	// tell that prose from a command. Matching only what the runner executes
	// is also the stricter test: a flag smuggled in after a `#` is inert.
	body := uncommented(readWorkflow(t))

	if strings.Contains(body, "-parseoracle.update") {
		t.Error("the fidelity workflow passes -parseoracle.update: a job that " +
			"re-baselines launders regressions into the baseline and can never fail")
	}
	// The converse: without -parseoracle.corpus the corpus ratchet skips,
	// and a skipped test is a green check that measured nothing.
	if !strings.Contains(body, "-parseoracle.corpus") {
		t.Error("the fidelity workflow does not pass -parseoracle.corpus: " +
			"TestRatchetCorpus skips without it, so the job would be decoration")
	}
	if !strings.Contains(body, "TestRatchetCorpus") {
		t.Error("the fidelity workflow does not run TestRatchetCorpus")
	}
}

// TestCIPinsTheWorldItMeasuresIn: the workflow must name the pin's two halves
// from the pin file, not from literals of its own. A hardcoded revision in
// YAML is a third copy of the pin that nothing keeps in step, and it drifts
// silently — the failure mode is a job that measures a different corpus while
// reporting on this one.
func TestCIPinsTheWorldItMeasuresIn(t *testing.T) {
	body := readWorkflow(t)

	if !strings.Contains(uncommented(body), pinScript) {
		t.Errorf("the fidelity workflow does not call %s: it must read the pin "+
			"rather than carry its own copy of the version and revision", pinScript)
	}

	pin, err := parseoracle.ReadPin(filepath.Join("testdata", "corpus.pin"))
	if err != nil {
		t.Fatalf("ReadPin: %v", err)
	}
	if strings.Contains(body, pin.Revision) {
		t.Errorf("the fidelity workflow hardcodes the corpus revision %s: "+
			"that is a second copy of the pin, and it will drift", pin.Revision)
	}
}

// TestCIGivesTheSweepHeadroom guards the measured fact that makes this job
// flaky by default.
//
// op/pack.t costs 37s of CPU on the machine this was measured on, unloaded,
// against a DefaultTimeout of 60s. A shared runner is slower, so the default
// would time that file out, record a runner error, and the ratchet would
// report a regression that is nothing but load. The workflow therefore raises
// the per-file timeout; without that it is a job that cries wolf and then
// gets disabled.
func TestCIGivesTheSweepHeadroom(t *testing.T) {
	if !strings.Contains(uncommented(readWorkflow(t)), timeoutEnv) {
		t.Errorf("the fidelity workflow does not set %s: op/pack.t costs 37s "+
			"unloaded against a %s default, so a shared runner times it out and "+
			"the ratchet reports a phantom regression", timeoutEnv, parseoracle.DefaultTimeout)
	}
}

// TestCIWorkflowIsWellFormed parses the YAML.
//
// It is the cheapest half of "does this workflow work", and the only half
// that can be answered without a runner: a file that does not parse fails on
// GitHub with no job and no red X on the commit, which is indistinguishable
// from having no workflow at all. Substring assertions elsewhere in this file
// would all still pass on a file with a broken indent, so this runs first.
func TestCIWorkflowIsWellFormed(t *testing.T) {
	var parsed struct {
		Name string `yaml:"name"`
		Jobs map[string]struct {
			RunsOn string `yaml:"runs-on"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal([]byte(readWorkflow(t)), &parsed); err != nil {
		t.Fatalf("%s is not valid YAML: %v — GitHub would run no job at all, "+
			"which looks exactly like having no ratchet", workflow, err)
	}
	if _, ok := parsed.Jobs["ratchet"]; !ok {
		t.Fatalf("%s has no `ratchet` job, only %v", workflow, parsed.Jobs)
	}

	steps := workflowSteps(t, readWorkflow(t))
	if len(steps) == 0 {
		t.Fatal("the ratchet job has no steps")
	}
	for i, step := range steps {
		if step.Uses == "" && step.Run == "" {
			t.Errorf("step %d (%q) neither uses an action nor runs a command", i, step.Name)
		}
		// An action must be pinned to a major version at least. An unpinned
		// `uses:` resolves to the default branch, which is a third party's
		// HEAD running in this repo's CI.
		if step.Uses != "" && !strings.Contains(step.Uses, "@") {
			t.Errorf("step %d uses %q with no version", i, step.Uses)
		}
	}

	// The sweep's own step must carry both environment variables that point
	// the runner at the pinned world. Asserting on the parsed env rather
	// than on the file's text means a variable set in the wrong step, or at
	// the wrong indent, is caught here rather than on the runner.
	var sweep map[string]string
	for _, step := range steps {
		if strings.Contains(step.Run, "TestRatchetCorpus") {
			sweep = step.Env
		}
	}
	if sweep == nil {
		t.Fatal("no step runs TestRatchetCorpus")
	}
	for _, key := range []string{"PERL5_CORPUS", "PARSEORACLE_SHIM", timeoutEnv} {
		if sweep[key] == "" {
			t.Errorf("the sweep step does not set %s", key)
		}
	}
	if _, err := time.ParseDuration(sweep[timeoutEnv]); err != nil {
		t.Errorf("%s=%q is not a duration Go can parse, so the sweep would "+
			"fail before measuring anything: %v", timeoutEnv, sweep[timeoutEnv], err)
	}
	// The headroom has to be real. op/pack.t costs 37s unloaded here; a
	// timeout that merely matches DefaultTimeout would change nothing.
	if d, err := time.ParseDuration(sweep[timeoutEnv]); err == nil && d <= parseoracle.DefaultTimeout {
		t.Errorf("%s=%v is no more than DefaultTimeout (%v): the point is headroom "+
			"for a runner slower than the machine the default was calibrated on",
			timeoutEnv, d, parseoracle.DefaultTimeout)
	}
}

// TestCIReportsSkewBeforeSweeping covers the acceptance criterion that a
// runner whose perl does not match the pin reports "re-baseline needed"
// rather than a list of regressions.
//
// CheckPin is the backstop and TestRatchetPinMismatch proves it says the
// right thing, but reaching it costs the whole sweep. The workflow therefore
// compares perl's $] against the pin immediately after installing it, so a
// skewed runner fails in seconds with both versions named. The failure has to
// read as skew, not as the parser breaking overnight, which is why the
// message and not merely the comparison is asserted here.
func TestCIReportsSkewBeforeSweeping(t *testing.T) {
	body := readWorkflow(t)

	guard, sweep := -1, -1
	for i, step := range workflowSteps(t, body) {
		if strings.Contains(step.Run, "perl -e 'print $]'") {
			guard = i
		}
		if strings.Contains(step.Run, "TestRatchetCorpus") {
			sweep = i
		}
	}
	if guard < 0 {
		t.Fatal("no step compares the runner's perl against the pin: a skewed " +
			"runner would spend the whole sweep before CheckPin said so")
	}
	if sweep < 0 {
		t.Fatal("no step runs TestRatchetCorpus")
	}
	if guard > sweep {
		t.Errorf("the interpreter check is step %d and the sweep is step %d: "+
			"checking after measuring wastes the run it exists to avoid", guard, sweep)
	}

	// "wrong" is not an answer. The message must name the pinned version,
	// the observed one, and say the two were measured in different worlds --
	// otherwise a perl upgrade reads as the parser breaking.
	step := workflowSteps(t, body)[guard]
	for _, want := range []string{"$actual", "$PINNED", "skew"} {
		if !strings.Contains(step.Run, want) {
			t.Errorf("the interpreter check's message omits %q; it must say which "+
				"perl is present, which the baseline used, and that the difference "+
				"is skew rather than a regression", want)
		}
	}
}

// workflowStep is the subset of a step these tests reason about.
type workflowStep struct {
	Name string            `yaml:"name"`
	Uses string            `yaml:"uses"`
	Run  string            `yaml:"run"`
	Env  map[string]string `yaml:"env"`
}

// workflowSteps parses the ratchet job's steps in order.
func workflowSteps(t *testing.T, body string) []workflowStep {
	t.Helper()

	var parsed struct {
		Jobs map[string]struct {
			Steps []workflowStep `yaml:"steps"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("%s is not valid YAML: %v", workflow, err)
	}
	return parsed.Jobs["ratchet"].Steps
}

// readWorkflow reads the fidelity workflow, failing rather than skipping when
// it is absent: these tests exist to notice its removal.
func readWorkflow(t *testing.T) string {
	t.Helper()
	path := repoFile(t, ".github", "workflows", workflow)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(data)
}

// uncommented drops YAML comments, leaving what the runner actually executes.
//
// It is line-oriented and does not understand `#` inside a quoted string,
// which is fine for the question being asked: every use here is "does this
// flag appear in a command", and erring towards dropping text can only make
// these assertions stricter, never laxer.
func uncommented(body string) string {
	var kept []string
	for _, line := range strings.Split(body, "\n") {
		if i := strings.IndexByte(line, '#'); i >= 0 {
			line = line[:i]
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// TestUncommentedDropsOnlyComments guards the helper the two assertions above
// lean on. A stripper that dropped everything would make them both vacuous.
func TestUncommentedDropsOnlyComments(t *testing.T) {
	const in = "run: go test -parseoracle.corpus\n# never -parseoracle.update\nkey: value # trailing\n"

	got := uncommented(in)
	if strings.Contains(got, "-parseoracle.update") {
		t.Error("uncommented kept a commented-out flag")
	}
	for _, want := range []string{"-parseoracle.corpus", "key: value"} {
		if !strings.Contains(got, want) {
			t.Errorf("uncommented dropped %q, which is not a comment", want)
		}
	}
}

// parseFields reads `key=value` lines, which is the form GITHUB_OUTPUT takes.
func parseFields(out string) map[string]string {
	fields := make(map[string]string)
	for _, line := range strings.Split(out, "\n") {
		if key, value, ok := strings.Cut(strings.TrimSpace(line), "="); ok {
			fields[key] = value
		}
	}
	return fields
}
