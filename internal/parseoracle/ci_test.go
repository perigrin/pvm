// ABOUTME: Tests the parts of the CI fidelity job that can be tested without a GitHub runner.
// ABOUTME: The workflow YAML itself cannot run here, so what it delegates to is tested instead.

package parseoracle_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	pathpkg "path"
	"path/filepath"
	"regexp"
	"sort"
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

	// requireCorpusEnv turns a missing corpus from a skip into a failure.
	//
	// The skip is right locally: a machine with no perl5 checkout must still
	// be able to run the suite. It is wrong for the gate, which exists to
	// measure — a runner whose checkout did not materialise would skip the
	// sweep and report success having measured nothing. The workflow sets
	// this to assert that, in its world, the corpus is supposed to be there.
	requireCorpusEnv = "PARSEORACLE_REQUIRE_CORPUS"

	// baselineEnv points TestRatchetCorpus at a baseline other than the
	// committed one. It exists for exactly one caller: TestCIGateRunsTheSweep
	// runs the workflow's own sweep step over a five-file shim, and needs a
	// five-row baseline to gate it against. The workflow must never set it
	// -- that test fails if it does -- and the receipt step compares the
	// receipt's baseline hash against the committed file regardless.
	baselineEnv = "PARSEORACLE_BASELINE"

	// sitesBaselineEnv is baselineEnv for TestSiteRatchetCorpus: it points
	// the site ratchet at the gate's fixture sites file. The workflow must
	// never set it either.
	sitesBaselineEnv = "PARSEORACLE_SITES_BASELINE"

	// receiptEnv names where TestRatchetCorpus writes its receipt after the
	// ratchet passes. Unset, no receipt is written, which is right locally;
	// the workflow sets it and its next step fails on the file's absence.
	receiptEnv = "PARSEORACLE_RECEIPT"
)

// corpusRequired reports whether a missing corpus must fail rather than skip.
func corpusRequired() bool {
	switch os.Getenv(requireCorpusEnv) {
	case "", "0", "false":
		return false
	}
	return true
}

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
	// actions-setup-perl takes. 5.044000 is 5.44.0; getting this wrong
	// installs a perl the pin does not describe, and every verdict below
	// the baseline header was measured against something else.
	if want := "5.44.0"; got["perl_version"] != want {
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
			body := "interpreter = " + tc.interpreter + "\nrevision    = deadbeef\nthreads     = define\n"
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
		"no revision":    "interpreter = 5.042000\nthreads = define\n",
		"no interpreter": "revision = deadbeef\nthreads = define\n",
		"no threads":     "interpreter = 5.042000\nrevision = deadbeef\n",
		"bogus threads":  "interpreter = 5.042000\nrevision = deadbeef\nthreads = yes\n",
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
	//
	// This is a must-not-contain check, and uncommented() can hide text
	// from it (see its comment). It stays because it is cheap and its
	// message names the exact flag; the check that cannot be hidden from
	// is TestCIGateRunsTheSweep, where a step that re-baselines exits 0
	// against a baseline that should have failed it.
	body := uncommented(readWorkflow(t))

	if strings.Contains(body, "-parseoracle.update") {
		t.Error("the fidelity workflow passes -parseoracle.update: a job that " +
			"re-baselines launders regressions into the baseline and can never fail")
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
		if isSweep(step) {
			sweep = step.Env
		}
	}
	if sweep == nil {
		t.Fatal("no step is the sweep")
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
		if isSweep(step) {
			sweep = i
		}
	}
	if guard < 0 {
		t.Fatal("no step compares the runner's perl against the pin: a skewed " +
			"runner would spend the whole sweep before CheckPin said so")
	}
	if sweep < 0 {
		t.Fatal("no step is the sweep")
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

// workflowStep is the subset of a step these tests reason about. Run, Shell
// and Env are what TestCIGateRunsTheSweep executes; If is what it cannot.
type workflowStep struct {
	Name  string            `yaml:"name"`
	ID    string            `yaml:"id"`
	Uses  string            `yaml:"uses"`
	With  map[string]string `yaml:"with"`
	Run   string            `yaml:"run"`
	Shell string            `yaml:"shell"`
	Env   map[string]string `yaml:"env"`
	If    any               `yaml:"if"`
}

// workflowJob is the subset of a job these tests reason about.
type workflowJob struct {
	If    any               `yaml:"if"`
	Env   map[string]string `yaml:"env"`
	Steps []workflowStep    `yaml:"steps"`
}

// workflowFile is the subset of the workflow these tests reason about: the
// triggers, and every level env: can be set at, because a variable set at
// any of them reaches the executed command.
type workflowFile struct {
	On struct {
		Push struct {
			Branches []string `yaml:"branches"`
		} `yaml:"push"`
		PullRequest struct {
			Branches []string `yaml:"branches"`
		} `yaml:"pull_request"`
	} `yaml:"on"`
	Env  map[string]string      `yaml:"env"`
	Jobs map[string]workflowJob `yaml:"jobs"`
}

// parseWorkflow reads the fidelity workflow into workflowFile.
func parseWorkflow(t *testing.T) workflowFile {
	t.Helper()
	var wf workflowFile
	if err := yaml.Unmarshal([]byte(readWorkflow(t)), &wf); err != nil {
		t.Fatalf("%s is not valid YAML: %v", workflow, err)
	}
	return wf
}

// workflowSteps parses the ratchet job's steps in order.
func workflowSteps(t *testing.T, body string) []workflowStep {
	t.Helper()

	var parsed workflowFile
	if err := yaml.Unmarshal([]byte(body), &parsed); err != nil {
		t.Fatalf("%s is not valid YAML: %v", workflow, err)
	}
	return parsed.Jobs["ratchet"].Steps
}

// repoRoot is the repository root as an absolute path, which is the
// directory the runner executes every step from.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs(repoFile(t))
	if err != nil {
		t.Fatal(err)
	}
	return root
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

// uncommented drops YAML comments, leaving roughly what the runner executes.
//
// It is line-oriented and cuts at the first `#` on a line, so it does not
// understand a `#` inside a quoted string or a parameter expansion. That
// makes it fit for MUST-CONTAIN checks only, where dropping text can only
// make the check stricter. For a must-NOT-contain check it is a sieve:
// `-parallel "${#RUNNER_ARCH}" || true` is executed in full by the shell and
// truncated to `-parallel "${` here, so the `|| true` is invisible to every
// substring assertion and live on the runner. An earlier version of this
// comment claimed the opposite; that claim was disproved by exactly that
// line. The exit-status property is guarded by executing the step in
// TestCIGateRunsTheSweep, not by reading it here.
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

// requirePinnedEnv turns perltest.PinnedPerl's skip into a failure. The
// fidelity workflow sets it on each step that installed the pinned perl
// and runs tests held to it; ci.yml's 5.42 jobs never set it, so there
// those tests skip.
const requirePinnedEnv = "PVM_REQUIRE_PINNED_PERL"

// heldName is the name of parse-fidelity.yml's step that runs every test
// held to the pinned perl other than the sweep.
const heldName = "Run the 5.44-held tests"

// TestCIRequiresThePinnedPerl: both steps that run tests held to the
// pinned perl -- the sweep and the 5.44-held step -- set
// PVM_REQUIRE_PINNED_PERL, so a held test that cannot find that perl fails
// there rather than skipping. No ci.yml job sets it: those jobs install
// 5.42, and a held test must skip there, not fail.
func TestCIRequiresThePinnedPerl(t *testing.T) {
	wf := parseWorkflow(t)
	held := false
	for _, step := range wf.Jobs["ratchet"].Steps {
		if step.Name == heldName {
			held = true
			if step.Env[requirePinnedEnv] != "1" {
				t.Errorf("step %q sets %s=%q, want '1'", heldName, requirePinnedEnv, step.Env[requirePinnedEnv])
			}
		}
	}
	if !held {
		t.Errorf("%s has no step named %q", workflow, heldName)
	}
	if got := sweepStep(t, wf).Env[requirePinnedEnv]; got != "1" {
		t.Errorf("the sweep step sets %s=%q, want '1'", requirePinnedEnv, got)
	}

	ci, err := os.ReadFile(repoFile(t, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(uncommented(string(ci)), requirePinnedEnv) {
		t.Errorf("ci.yml sets %s: its jobs install perl 5.42, so every test held to "+
			"the pinned perl would fail there instead of skipping", requirePinnedEnv)
	}
}

// TestEveryPinnedPerlTestIsRequired: every test function in internal/parse
// and internal/parseoracle that calls perltest.PinnedPerl is named in the
// -run pattern of a step that sets PVM_REQUIRE_PINNED_PERL, and that
// step's package list covers the test's package -- `go test
// ./internal/parseoracle/ -run X` for an X in internal/parse runs nothing
// and exits 0. No other function may call PinnedPerl: a helper that does
// hides the held test behind it. And each REQUIRE step's fail-unless-PASS
// loop names exactly its -run names, so the two cannot drift.
func TestEveryPinnedPerlTestIsRequired(t *testing.T) {
	files := map[string][]byte{}
	for _, dir := range []string{"internal/parse", "internal/parseoracle"} {
		paths, err := filepath.Glob(filepath.Join(repoFile(t, dir), "*_test.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			// A slash path on every platform: the workflow names packages
			// with `/`, and filepath.Join wrote `internal\parse` on Windows.
			files[dir+"/"+filepath.Base(path)] = src
		}
	}
	if len(files) == 0 {
		t.Fatal("no test files found to scan")
	}
	for _, problem := range pinnedPerlProblems(t, parseWorkflow(t), files) {
		t.Error(problem)
	}
}

// TestEveryPinnedPerlTestIsRequiredRejectsHelperCaller: a test that reaches
// PinnedPerl only through a helper is invisible to the coverage check, so
// the helper is flagged.
func TestEveryPinnedPerlTestIsRequiredRejectsHelperCaller(t *testing.T) {
	problems := pinnedPerlProblems(t, requireFixtureWorkflow(t), map[string][]byte{
		"internal/parseoracle/x_test.go": []byte(`package parseoracle_test
func pinned(t *testing.T) perltest.Perl { return perltest.PinnedPerl(t) }
func TestHeld(t *testing.T) { _ = pinned(t) }
`),
	})
	if !anyContains(problems, "pinned") {
		t.Errorf("a helper that calls PinnedPerl was not flagged; got %q", problems)
	}
}

// TestEveryPinnedPerlTestIsRequiredRejectsIndirectReferences: PinnedPerl
// reached through a package-level func literal or a function value hides
// the test that uses it just as a helper does, and PinnedPerlAt outside
// givenperl_test.go is a way round the check altogether. Each is flagged.
func TestEveryPinnedPerlTestIsRequiredRejectsIndirectReferences(t *testing.T) {
	for name, src := range map[string]string{
		"func literal": `package parseoracle_test
var pinned = func(t *testing.T) perltest.Perl { return perltest.PinnedPerl(t) }
func TestHeldA(t *testing.T) { _ = pinned(t) }
`,
		"function value": `package parseoracle_test
var pp = perltest.PinnedPerl
func TestHeldC(t *testing.T) { _ = pp(t) }
`,
		"PinnedPerlAt": `package parseoracle_test
func TestHeldB(t *testing.T) { _ = perltest.PinnedPerlAt(t, "testdata/corpus.pin") }
`,
	} {
		t.Run(name, func(t *testing.T) {
			problems := pinnedPerlProblems(t, requireFixtureWorkflow(t), map[string][]byte{
				"internal/parseoracle/x_test.go": []byte(src),
			})
			if len(problems) == 0 {
				t.Errorf("a %s reaching PinnedPerl was not flagged", name)
			}
		})
	}
}

// TestEveryPinnedPerlTestIsRequiredFlagsUnlistedSweepTest: a held test
// named in neither REQUIRE step is flagged, even one whose name is a
// prefix-sharing neighbour of a listed one.
func TestEveryPinnedPerlTestIsRequiredFlagsUnlistedSweepTest(t *testing.T) {
	problems := pinnedPerlProblems(t, requireFixtureWorkflow(t), map[string][]byte{
		"internal/parseoracle/r_test.go": []byte(`package parseoracle_test
func TestRatchetCorpusSites(t *testing.T) { _ = perltest.PinnedPerl(t) }
`),
	})
	if !anyContains(problems, "TestRatchetCorpusSites") {
		t.Errorf("a held sweep test in no REQUIRE step was not flagged; got %q", problems)
	}
}

// TestEveryPinnedPerlTestIsRequiredChecksThePackage: a held test in
// internal/parse named only by a step that tests internal/parseoracle
// never runs, and is flagged.
func TestEveryPinnedPerlTestIsRequiredChecksThePackage(t *testing.T) {
	problems := pinnedPerlProblems(t, requireFixtureWorkflow(t), map[string][]byte{
		"internal/parse/p_test.go": []byte(`package parse
func TestRatchet(t *testing.T) { _ = perltest.PinnedPerl(t) }
`),
	})
	if !anyContains(problems, "TestRatchet") {
		t.Errorf("a held internal/parse test run only against internal/parseoracle was not flagged; got %q", problems)
	}
}

// TestEveryPinnedPerlTestIsRequiredMatchesTheLoop: a REQUIRE step whose
// fail-unless-PASS loop names fewer tests than its -run is flagged, since
// a test it no longer checks can skip unnoticed.
func TestEveryPinnedPerlTestIsRequiredMatchesTheLoop(t *testing.T) {
	var wf workflowFile
	if err := yaml.Unmarshal([]byte(`
jobs:
  ratchet:
    steps:
    - name: Run the 5.44-held tests
      env:
        PVM_REQUIRE_PINNED_PERL: '1'
      run: |
        go test ./internal/parseoracle/ -run '^(TestA|TestB)$' -v 2>&1 | tee held.log
        for t in TestA; do grep -q "^--- PASS: $t (" held.log; done
`), &wf); err != nil {
		t.Fatal(err)
	}
	if problems := pinnedPerlProblems(t, wf, nil); !anyContains(problems, "TestB") {
		t.Errorf("a loop that drops TestB from -run was not flagged; got %q", problems)
	}
}

// requireFixtureWorkflow is a workflow with the two REQUIRE steps' shape:
// the sweep naming TestRatchetCorpus and the held step naming TestRatchet,
// both over internal/parseoracle only.
func requireFixtureWorkflow(t *testing.T) workflowFile {
	t.Helper()
	var wf workflowFile
	if err := yaml.Unmarshal([]byte(`
jobs:
  ratchet:
    steps:
    - name: Run the 5.44-held tests
      env:
        PVM_REQUIRE_PINNED_PERL: '1'
      run: |
        go test ./internal/parseoracle/ -run '^(TestRatchet)$' -v 2>&1 | tee held.log
        for t in TestRatchet; do grep -q "^--- PASS: $t (" held.log; done
    - name: Run the fidelity ratchet
      env:
        PVM_REQUIRE_PINNED_PERL: '1'
      run: |
        go test ./internal/parseoracle/ -run '^(TestRatchetCorpus)$' -parseoracle.corpus -v 2>&1 | tee sweep.log
        for t in TestRatchetCorpus; do grep -q "^--- PASS: $t (" sweep.log; done
`), &wf); err != nil {
		t.Fatal(err)
	}
	return wf
}

func anyContains(list []string, needle string) bool {
	for _, s := range list {
		if strings.Contains(s, needle) {
			return true
		}
	}
	return false
}

var (
	// runPattern is a step's -run, written '^(TestA|TestB)$' so the names
	// can be read back exactly.
	runPattern = regexp.MustCompile(`-run[ =]'\^\(([\w|]+)\)\$'`)
	// loopPattern is a step's fail-unless-PASS loop header.
	loopPattern = regexp.MustCompile(`for \w+ in ([\w\s\\]+?); do`)
	// pkgPattern is a package path handed to go test.
	pkgPattern = regexp.MustCompile(`\./(internal/[\w/]*\w)/?(?:\s|$)`)
)

// requireStep is a step that sets PVM_REQUIRE_PINNED_PERL, read from its
// uncommented script.
type requireStep struct {
	name      string
	run, loop []string
	pkgs      map[string]bool
}

// requireSteps reads every step in wf that sets PVM_REQUIRE_PINNED_PERL.
func requireSteps(wf workflowFile) []requireStep {
	var steps []requireStep
	for _, job := range wf.Jobs {
		for _, step := range job.Steps {
			if step.Env[requirePinnedEnv] == "" {
				continue
			}
			body := uncommented(step.Run)
			s := requireStep{name: step.Name, pkgs: map[string]bool{}}
			if m := runPattern.FindStringSubmatch(body); m != nil {
				s.run = strings.Split(m[1], "|")
			}
			if m := loopPattern.FindStringSubmatch(body); m != nil {
				s.loop = strings.Fields(strings.ReplaceAll(m[1], `\`, " "))
			}
			for _, m := range pkgPattern.FindAllStringSubmatch(body, -1) {
				s.pkgs[m[1]] = true
			}
			steps = append(steps, s)
		}
	}
	return steps
}

// pinnedPerlProblems checks files (keyed by repository path) against wf:
// every Test function that calls PinnedPerl must be named in some REQUIRE
// step's -run whose package list holds its package, no other function may
// call PinnedPerl, and each REQUIRE step's loop must name exactly its -run.
func pinnedPerlProblems(t *testing.T, wf workflowFile, files map[string][]byte) []string {
	t.Helper()
	var problems []string
	steps := requireSteps(wf)
	if len(steps) == 0 {
		problems = append(problems, "no step sets "+requirePinnedEnv)
	}
	for _, s := range steps {
		if len(s.run) == 0 {
			problems = append(problems, fmt.Sprintf("step %q sets %s but has no -run '^(...)$' to read", s.name, requirePinnedEnv))
		}
		if !sameNames(s.run, s.loop) {
			problems = append(problems, fmt.Sprintf("step %q runs %v but its fail-unless-PASS loop checks %v",
				s.name, s.run, s.loop))
		}
	}

	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	fset := token.NewFileSet()
	for _, path := range paths {
		f, err := parser.ParseFile(fset, path, files[path], 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		pkg := pathpkg.Dir(path)
		if pathpkg.Base(path) != pinnedPerlAtFile && references(f, "PinnedPerlAt") {
			problems = append(problems, fmt.Sprintf("%s refers to PinnedPerlAt, which only %s may: "+
				"elsewhere it reaches a pinned perl that no REQUIRE step holds", path, pinnedPerlAtFile))
		}
		for _, decl := range f.Decls {
			if !references(decl, "PinnedPerl") {
				continue
			}
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Recv != nil || !strings.HasPrefix(fn.Name.Name, "Test") {
				problems = append(problems, fmt.Sprintf("%s: %s refers to PinnedPerl, but only a test "+
					"function may; take a perltest.Perl parameter instead so the test that uses it is "+
					"visible", path, declName(decl)))
				continue
			}
			name := fn.Name.Name
			covered := false
			for _, s := range steps {
				if s.pkgs[pkg] && containsName(s.run, name) {
					covered = true
				}
			}
			if !covered {
				problems = append(problems, fmt.Sprintf("%s: %s calls PinnedPerl, but no step that sets %s "+
					"names it in -run with ./%s/ in its package list, so CI never requires it to run",
					path, name, requirePinnedEnv, pkg))
			}
		}
	}
	return problems
}

// pinnedPerlAtFile is the one test file that may use PinnedPerlAt: its
// tests write their own pins, so none is held to the committed one.
const pinnedPerlAtFile = "givenperl_test.go"

// references reports whether node names the identifier name anywhere: a
// call, a package selector, a function value, a func literal's body.
// Matching references rather than calls is what keeps `var pp =
// perltest.PinnedPerl` from hiding the tests that call pp.
func references(node ast.Node, name string) bool {
	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if id, ok := n.(*ast.Ident); ok && id.Name == name {
			found = true
		}
		return !found
	})
	return found
}

// declName names a declaration for a problem message.
func declName(decl ast.Decl) string {
	switch d := decl.(type) {
	case *ast.FuncDecl:
		return d.Name.Name
	case *ast.GenDecl:
		for _, spec := range d.Specs {
			if v, ok := spec.(*ast.ValueSpec); ok && len(v.Names) > 0 {
				return v.Names[0].Name
			}
		}
	}
	return "a declaration"
}

func containsName(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}

// sameNames reports whether a and b hold the same names, in any order.
func sameNames(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for _, n := range a {
		if !containsName(b, n) {
			return false
		}
	}
	return true
}

// The two perls CI installs, each from its own pin. The parse oracle is
// judged by corpus.pin's perl; the .tpm libraries, the conformance corpus
// and the CORE.tpm check are judged by perl 5.42.0, from perl542.pin.
const (
	oraclePin  = "internal/parseoracle/testdata/corpus.pin"
	perl542Pin = "internal/perltest/testdata/perl542.pin"
)

// workflowPerls names every job in every workflow and the pin its perl is
// installed from, "" for a job that installs none. A job added later is
// unclassified until it is named here, which fails the build.
var workflowPerls = map[string]string{
	"build-perl-binaries.yml/prepare": "",
	"build-perl-binaries.yml/build":   "",
	"build-perl-binaries.yml/upload":  "",
	"build-perl-binaries.yml/summary": "",
	"ci.yml/test":                     perl542Pin,
	"ci.yml/ratchets":                 perl542Pin,
	"ci.yml/smoke-test":               "",
	"parse-fidelity.yml/ratchet":      oraclePin,
	"parser-fuzz.yml/fuzz-parser":     "",
	"release.yml/build-matrix":        "",
	"release.yml/create-release":      "",
}

var (
	// perlVersionOutput is the only perl-version a setup-perl step may
	// take: the output of a step in the same job that read a pin.
	perlVersionOutput = regexp.MustCompile(`^\$\{\{\s*steps\.([\w-]+)\.outputs\.perl_version\s*\}\}$`)
	// pinScriptCall is the pin script's invocation, with its optional pin
	// argument; corpus.pin is the script's default.
	pinScriptCall = regexp.MustCompile(`sh \.github/scripts/parseoracle-pin\.sh(?:[ \t]+([^\s>]+))?[ \t]*>>`)
)

// installedPin is the pin each actions-setup-perl step in job installs
// from, as repository paths, with the problems found tracing them.
func installedPin(job workflowJob) ([]string, []string) {
	var pins, problems []string
	for _, step := range job.Steps {
		if !strings.HasPrefix(step.Uses, "shogo82148/actions-setup-perl@") {
			continue
		}
		m := perlVersionOutput.FindStringSubmatch(step.With["perl-version"])
		if m == nil {
			problems = append(problems, fmt.Sprintf("step %q installs perl-version %q, not a pin's perl_version",
				step.Name, step.With["perl-version"]))
			continue
		}
		found := false
		for _, src := range job.Steps {
			if src.ID != m[1] {
				continue
			}
			found = true
			call := pinScriptCall.FindStringSubmatch(uncommented(src.Run))
			switch {
			case call == nil:
				problems = append(problems, fmt.Sprintf("step %q does not read a pin with %s", src.Name, pinScript))
			case call[1] == "":
				pins = append(pins, oraclePin)
			default:
				pins = append(pins, call[1])
			}
		}
		if !found {
			problems = append(problems, fmt.Sprintf("step %q reads steps.%s, which no step in its job is", step.Name, m[1]))
		}
	}
	return pins, problems
}

// readsPin reports whether the script run names pin, by its path or, for
// corpus.pin, by calling the pin script with no argument.
func readsPin(run, pin string) bool {
	body := uncommented(run)
	if strings.Contains(body, pin) {
		return true
	}
	if pin != oraclePin {
		return false
	}
	for _, call := range pinScriptCall.FindAllStringSubmatch(body, -1) {
		if call[1] == "" {
			return true
		}
	}
	return false
}

// workflowPinProblems checks every job in workflows (keyed by file name)
// against want: each job is classified, installs perl from exactly its
// pin, and never names the other pin.
func workflowPinProblems(workflows map[string]workflowFile, want map[string]string) []string {
	var problems []string
	seen := map[string]bool{}
	for file, wf := range workflows {
		for name, job := range wf.Jobs {
			key := file + "/" + name
			seen[key] = true
			pin, ok := want[key]
			if !ok {
				problems = append(problems, fmt.Sprintf("%s is not classified: name the perl it installs in workflowPerls", key))
				continue
			}
			pins, traced := installedPin(job)
			for _, p := range traced {
				problems = append(problems, key+": "+p)
			}
			if pin == "" && len(pins) > 0 {
				problems = append(problems, fmt.Sprintf("%s installs perl from %v, but is classified as installing none", key, pins))
			}
			if pin != "" && len(pins) == 0 && len(traced) == 0 {
				problems = append(problems, fmt.Sprintf("%s installs no perl, but is classified as installing %s's", key, pin))
			}
			for _, p := range pins {
				if p != pin {
					problems = append(problems, fmt.Sprintf("%s installs perl from %s, want %s", key, p, pin))
				}
			}
			if pin == "" {
				continue
			}
			// The full path, since t1corpus.pin ends in corpus.pin too.
			other := oraclePin
			if pin == oraclePin {
				other = perl542Pin
			}
			for _, step := range job.Steps {
				if readsPin(step.Run, other) {
					problems = append(problems, fmt.Sprintf("%s step %q reads %s, the other perl's pin", key, step.Name, other))
				}
			}
		}
	}
	for key := range want {
		if !seen[key] {
			problems = append(problems, fmt.Sprintf("workflowPerls names %s, which no workflow has", key))
		}
	}
	sort.Strings(problems)
	return problems
}

// TestCIWorkflowsPinTheirPerls: the oracle's job installs corpus.pin's
// perl, and ci.yml's jobs -- the Test matrix, whose go test ./... runs the
// conformance and CORE.tpm checks, and "Ratchets (corpus-backed)" -- install
// perl 5.42.0 from perl542.pin. Both pins measure the same perl5 corpus.
func TestCIWorkflowsPinTheirPerls(t *testing.T) {
	paths, err := filepath.Glob(repoFile(t, ".github", "workflows", "*.yml"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no workflows found: %v", err)
	}
	workflows := map[string]workflowFile{}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var wf workflowFile
		if err := yaml.Unmarshal(data, &wf); err != nil {
			t.Fatalf("%s is not valid YAML: %v", path, err)
		}
		workflows[filepath.Base(path)] = wf
	}
	for _, p := range workflowPinProblems(workflows, workflowPerls) {
		t.Error(p)
	}

	oracle, err := parseoracle.ReadPin(repoFile(t, oraclePin))
	if err != nil {
		t.Fatal(err)
	}
	p542, err := parseoracle.ReadPin(repoFile(t, perl542Pin))
	if err != nil {
		t.Fatal(err)
	}
	if p542.Interpreter != "5.042000" {
		t.Errorf("%s pins %s; the 5.42 jobs judge the .tpm libraries with 5.042000", perl542Pin, p542.Interpreter)
	}
	if p542.Revision != oracle.Revision {
		t.Errorf("%s records perl5 at %s and %s at %s: the ratchets job and the oracle measure one corpus",
			perl542Pin, p542.Revision, oraclePin, oracle.Revision)
	}
}

// TestCIWorkflowsPinTheirPerlsRejectsSharedPin: a 5.42 job that installs
// from, or reads, corpus.pin fails, and so does the oracle job reading
// perl542.pin.
func TestCIWorkflowsPinTheirPerlsRejectsSharedPin(t *testing.T) {
	as542 := map[string]string{"w.yml/test": perl542Pin}
	asOracle := map[string]string{"w.yml/test": oraclePin}
	for name, tc := range map[string]struct {
		want map[string]string
		body string
	}{
		"5.42 job installs from corpus.pin": {as542, `
jobs:
  test:
    steps:
    - id: perlpin
      run: sh .github/scripts/parseoracle-pin.sh >> "$GITHUB_OUTPUT"
    - uses: shogo82148/actions-setup-perl@v1
      with:
        perl-version: ${{ steps.perlpin.outputs.perl_version }}
`},
		"5.42 job reads corpus.pin": {as542, `
jobs:
  test:
    steps:
    - id: perlpin
      run: sh .github/scripts/parseoracle-pin.sh internal/perltest/testdata/perl542.pin >> "$GITHUB_OUTPUT"
    - uses: shogo82148/actions-setup-perl@v1
      with:
        perl-version: ${{ steps.perlpin.outputs.perl_version }}
    - run: grep revision internal/parseoracle/testdata/corpus.pin
`},
		"5.42 job reads corpus.pin through the script's default": {as542, `
jobs:
  test:
    steps:
    - id: perlpin
      run: sh .github/scripts/parseoracle-pin.sh internal/perltest/testdata/perl542.pin >> "$GITHUB_OUTPUT"
    - id: opin
      run: sh .github/scripts/parseoracle-pin.sh >> "$GITHUB_OUTPUT"
    - uses: actions/checkout@v5
      with:
        ref: ${{ steps.opin.outputs.revision }}
    - uses: shogo82148/actions-setup-perl@v1
      with:
        perl-version: ${{ steps.perlpin.outputs.perl_version }}
        multi-thread: ${{ steps.opin.outputs.threads == 'define' }}
`},
		"5.42 job names its own version": {as542, `
jobs:
  test:
    steps:
    - uses: shogo82148/actions-setup-perl@v1
      with:
        perl-version: '5.42.0'
`},
		"oracle job installs from perl542.pin": {asOracle, `
jobs:
  test:
    steps:
    - id: pin
      run: sh .github/scripts/parseoracle-pin.sh internal/perltest/testdata/perl542.pin >> "$GITHUB_OUTPUT"
    - uses: shogo82148/actions-setup-perl@v1
      with:
        perl-version: ${{ steps.pin.outputs.perl_version }}
`},
		"oracle job reads perl542.pin": {asOracle, `
jobs:
  test:
    steps:
    - id: pin
      run: sh .github/scripts/parseoracle-pin.sh >> "$GITHUB_OUTPUT"
    - uses: shogo82148/actions-setup-perl@v1
      with:
        perl-version: ${{ steps.pin.outputs.perl_version }}
    - run: cat internal/perltest/testdata/perl542.pin
`},
	} {
		t.Run(name, func(t *testing.T) {
			var wf workflowFile
			if err := yaml.Unmarshal([]byte(tc.body), &wf); err != nil {
				t.Fatal(err)
			}
			if problems := workflowPinProblems(map[string]workflowFile{"w.yml": wf}, tc.want); len(problems) == 0 {
				t.Errorf("%s passed the workflow pin check", name)
			}
		})
	}
}
