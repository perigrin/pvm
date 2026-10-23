// ABOUTME: Guards the fidelity gate by executing its steps rather than reading them: the sweep must produce a receipt and must fail on a regression.
// ABOUTME: What execution cannot observe -- continue-on-error, if:, on: -- is read from the parsed YAML, and the header says so.

package parseoracle_test

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"tamarou.com/pvm/internal/parseoracle"
	"tamarou.com/pvm/internal/perltest"
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
//
// This is a GitHub semantic, not a shell one, so executing the step cannot
// observe it; it is the one guard here that has to read the YAML.
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

// TestCIGateRunsTheSweep executes the gate instead of reading it.
//
// Fourteen ways were found to make this workflow pass while measuring
// nothing, and every one of them got past a guard of the form "does the YAML
// not contain X": -run naming another test, a second -run that Go takes and
// a field scan does not, -skip, -list, -count=0, -exec /bin/true,
// -parseoracle.corpus=false, GOFLAGS in the step's env, the require-corpus
// switch set to '0', exports inside the script, `| tee` without pipefail,
// `|| echo`, `|| true` hidden behind a `#` inside a parameter expansion, and
// a shell without -e. The space of ways not to run a Go test is not
// enumerable by substring, so this test stops enumerating.
//
// Instead it takes the sweep step exactly as committed -- its script, its
// shell, and every env: block that applies to it -- points the corpus
// variables at a five-file shim built from the fixture corpus, and runs it
// three times:
//
//  1. against the matching baselines, per-file and per-site: it must exit 0
//     AND write a receipt that says five files were swept against that
//     baseline, carrying the sites line TestSiteRatchetCorpus adds. Every bypass above
//     produces either no receipt or the wrong one, because the only thing
//     that writes a receipt is TestRatchetCorpus after Check has passed.
//  2. against a baseline that records one file differently: it must exit
//     non-zero and write no receipt. This is the property the gate exists
//     for, and every exit-status launderer fails it.
//  3. with the corpus absent, in the shape a cache miss leaves -- an empty
//     shim directory and no perl5 -- it must exit non-zero. A sweep that
//     skips here would go green on a runner whose checkout never arrived.
//
// Then the receipt step, also as committed, must reject the five-file
// receipt against the committed 620-row baseline, reject a missing receipt,
// and accept a receipt that describes the committed baseline. Together
// those are the whole gate as the runner would execute it, short of the
// runner itself.
func TestCIGateRunsTheSweep(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the gate runs on ubuntu-latest; modelling its shell on Windows proves nothing about the runner")
	}

	p := perltest.PinnedPerl(t)
	wf := parseWorkflow(t)
	verify := stepRunning(t, wf, "cmd/receipt")
	sweep, env, good := runGateSweep(t, p, wf)
	receipt := env[receiptEnv]
	shim := env["PARSEORACLE_SHIM"]

	r, err := parseoracle.ReadReceipt(receipt)
	if err != nil {
		t.Fatalf("the sweep step exited 0 and wrote no receipt at %s: whatever it ran, "+
			"it was not TestRatchetCorpus to completion (%v)", receipt, err)
	}
	if err := r.Verify(good, pinPath, 1); err != nil {
		t.Errorf("the receipt does not describe the sweep this test asked for: %v", err)
	}
	if err := r.VerifySites(gateSitesPath()); err != nil {
		t.Errorf("the receipt does not carry the site ratchet's line: %v", err)
	}
	if want := countFixtureFiles(t); r.FilesSwept != want {
		t.Errorf("the receipt reports %d files swept, want the fixture's %d", r.FilesSwept, want)
	}

	// The receipt step, against the receipt just written: five rows is not
	// the committed corpus, and the step has to say so with a non-zero exit.
	if out, err := runStep(t, verify, env); err == nil {
		t.Errorf("the receipt step accepted a five-file receipt as the corpus gate:\n%s", out)
	}
	// ... against no receipt at all, which is what every bypass leaves.
	if err := os.Remove(receipt); err != nil {
		t.Fatal(err)
	}
	if out, err := runStep(t, verify, env); err == nil {
		t.Errorf("the receipt step passed with no receipt present:\n%s", out)
	}
	// ... and against a receipt that does describe the committed baseline,
	// so a step that fails on everything cannot pass this test either.
	if err := corpusReceipt(t).Write(receipt); err != nil {
		t.Fatal(err)
	}
	if out, err := runStep(t, verify, env); err != nil {
		t.Errorf("the receipt step rejected a receipt for the committed baseline:\n%v\n%s", err, out)
	}
	// ... but not one whose site ratchet never added its line.
	unsited := corpusReceipt(t)
	unsited.Sites = ""
	if err := unsited.Write(receipt); err != nil {
		t.Fatal(err)
	}
	if out, err := runStep(t, verify, env); err == nil {
		t.Errorf("the receipt step accepted a receipt with no sites line:\n%s", out)
	}
	if err := os.Remove(receipt); err != nil {
		t.Fatal(err)
	}

	// 2. The same step, a baseline that disagrees about one file.
	env[baselineEnv] = wrongBaseline(t, filepath.Dir(good), good)
	if out, err := runStep(t, sweep, env); err == nil {
		t.Fatalf("the sweep step exited 0 against a baseline that records a regression: "+
			"the gate cannot fail, so it is decoration\n%s", out)
	}
	if _, err := os.Stat(receipt); err == nil {
		t.Errorf("the sweep step wrote a receipt for a ratchet that failed")
	}

	// 3. The same step, the corpus gone the way a cache miss leaves it.
	env[baselineEnv] = good
	if err := os.RemoveAll(shim); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(shim, 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := runStep(t, sweep, env); err == nil {
		t.Fatalf("the sweep step exited 0 with no corpus present: on a runner whose "+
			"checkout never arrived this job would go green having measured nothing\n%s", out)
	}
}

// runGateSweep runs the committed sweep step over a five-file fixture shim,
// gated against the fixture baselines, and fails unless it exits 0. It
// returns the step and the environment it ran in, whose receiptEnv names
// the receipt, and the per-file baseline it was given.
func runGateSweep(t *testing.T, p perltest.Perl, wf workflowFile) (workflowStep, map[string]string, string) {
	t.Helper()
	sweep := sweepStep(t, wf)
	temp := t.TempDir()
	env := gateEnv(t, wf, sweep, temp)

	if env[receiptEnv] == "" {
		t.Fatalf("the sweep step runs without %s: nothing would record that the sweep happened", receiptEnv)
	}
	for _, override := range []string{baselineEnv, sitesBaselineEnv} {
		if env[override] != "" {
			t.Fatalf("the workflow sets %s=%q: the gate must check the committed baselines, and the "+
				"override exists only so this test can point the committed step at a fixture",
				override, env[override])
		}
	}
	shim := env["PARSEORACLE_SHIM"]
	if shim == "" {
		t.Fatal("the sweep step runs without PARSEORACLE_SHIM: this test can only supply a corpus through it")
	}
	if env["PERL5_CORPUS"] == "" {
		// Never let a mutated workflow fall back to ~/dev/perl5 on a
		// developer's machine and start a real six-minute sweep.
		env["PERL5_CORPUS"] = filepath.Join(temp, "no-corpus")
	}

	good := gateBaseline(t, temp)
	env[baselineEnv] = good
	env[sitesBaselineEnv] = gateSitesPath()

	pin, err := parseoracle.ReadPin(pinPath)
	if err != nil {
		t.Fatal(err)
	}
	buildFixtureShim(t, p, shim, pin.Revision)
	if out, err := runStep(t, sweep, env); err != nil {
		t.Fatalf("the sweep step failed against baselines that match the fixture:\n%v\n%s", err, out)
	}
	return sweep, env, good
}

// TestCISweepRunsOracleOnce: the sweep step runs both ratchets in one go
// test, and they share one walk of the corpus. The receipt records how many
// files the process walked, which must equal the files swept, and carries
// the sites line only the site ratchet adds after passing.
func TestCISweepRunsOracleOnce(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the gate runs on ubuntu-latest; modelling its shell on Windows proves nothing about the runner")
	}
	_, env, _ := runGateSweep(t, perltest.PinnedPerl(t), parseWorkflow(t))

	r, err := parseoracle.ReadReceipt(env[receiptEnv])
	if err != nil {
		t.Fatalf("the sweep step wrote no receipt: %v", err)
	}
	if r.FilesSwept == 0 || r.FilesWalked != r.FilesSwept {
		t.Errorf("the receipt records %d files walked for %d swept, want equal: both ratchets "+
			"must share one walk of each file", r.FilesWalked, r.FilesSwept)
	}
	if err := r.VerifySites(gateSitesPath()); err != nil {
		t.Errorf("the receipt carries no sites line for the fixture: %v", err)
	}
}

// TestCIGateCannotBeSkipped reads the switches execution cannot see.
//
// `if: false` on the job, on either gate step, or on any step that sets
// PVM_REQUIRE_PINNED_PERL, and triggers that never name pu, each leave a
// workflow that parses, passes every executable check here, and never
// runs. This is partial on purpose: nothing in this repository can make a
// skipped or absent check block a merge. That takes a ruleset on pu with
// required_status_checks naming the ratchet job, which the workflow header
// spells out and only the repository owner can apply.
func TestCIGateCannotBeSkipped(t *testing.T) {
	wf := parseWorkflow(t)

	for _, branches := range []struct {
		trigger string
		list    []string
	}{{"push", wf.On.Push.Branches}, {"pull_request", wf.On.PullRequest.Branches}} {
		named := false
		for _, b := range branches.list {
			if b == "pu" {
				named = true
			}
		}
		if !named {
			t.Errorf("on.%s.branches is %v and does not name pu: the gate would never "+
				"run for the branch it protects", branches.trigger, branches.list)
		}
	}

	// Both gate steps must exist for the check below to mean anything.
	sweepStep(t, wf)
	stepRunning(t, wf, "cmd/receipt")
	for _, problem := range skipProblems(wf) {
		t.Error(problem)
	}
}

// TestCIGateCannotBeSkippedCoversRequireSteps: a step that requires the
// pinned perl is a gate too. With `if: false` on it the held tests never
// run and the step reports success, so the 5.44-held step is flagged the
// same as the sweep.
func TestCIGateCannotBeSkippedCoversRequireSteps(t *testing.T) {
	var wf workflowFile
	if err := yaml.Unmarshal([]byte(`
jobs:
  ratchet:
    steps:
    - name: Run the 5.44-held tests
      if: false
      env:
        PVM_REQUIRE_PINNED_PERL: '1'
      run: go test ./internal/parseoracle/ -run '^(TestRatchet)$'
    - name: Run the fidelity ratchet
      env:
        PVM_REQUIRE_PINNED_PERL: '1'
      run: go test ./internal/parseoracle/ -run '^(TestRatchetCorpus)$' -parseoracle.corpus
`), &wf); err != nil {
		t.Fatal(err)
	}
	problems := skipProblems(wf)
	if len(problems) != 1 || !strings.Contains(problems[0], "Run the 5.44-held tests") {
		t.Errorf("an if: on the 5.44-held step must be flagged, and only it; got %q", problems)
	}
}

// skipProblems names every `if:` that could turn the gate off: on the
// ratchet job, on the sweep, on the receipt check, and on every step that
// sets PVM_REQUIRE_PINNED_PERL.
func skipProblems(wf workflowFile) []string {
	var problems []string
	job := wf.Jobs["ratchet"]
	if job.If != nil {
		problems = append(problems, fmt.Sprintf(
			"the ratchet job carries if: %v -- a skipped job reports success to a required check", job.If))
	}
	for _, step := range job.Steps {
		gate := isSweep(step) || strings.Contains(step.Run, "cmd/receipt") || step.Env[requirePinnedEnv] != ""
		if gate && step.If != nil {
			problems = append(problems, fmt.Sprintf(
				"step %q carries if: %v -- a skipped step leaves nothing to fail", step.Name, step.If))
		}
	}
	return problems
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

// runnerTemp is the one expression the gate's env uses. Anything else is
// refused rather than guessed: a value this test cannot model is a value it
// cannot claim to have executed.
var runnerTemp = regexp.MustCompile(`\$\{\{\s*runner\.temp\s*\}\}`)

// gateEnv builds the environment the runner would give a step: the process
// environment, then the workflow's env, the job's, and the step's, later
// levels winning. Reading every level is what makes a GOFLAGS smuggled into
// any of them reach the executed command, exactly as it would on GitHub.
func gateEnv(t *testing.T, wf workflowFile, step workflowStep, temp string) map[string]string {
	t.Helper()

	env := make(map[string]string)
	for _, kv := range os.Environ() {
		if k, v, ok := strings.Cut(kv, "="); ok {
			env[k] = v
		}
	}
	env["RUNNER_TEMP"] = temp
	env["RUNNER_OS"] = "Linux"
	env["RUNNER_ARCH"] = "X64"
	env["GITHUB_WORKSPACE"] = repoRoot(t)

	for _, level := range []map[string]string{wf.Env, wf.Jobs["ratchet"].Env, step.Env} {
		for k, v := range level {
			v = runnerTemp.ReplaceAllLiteralString(v, temp)
			if strings.Contains(v, "${{") {
				t.Fatalf("env %s=%q uses an expression this test does not model; it cannot "+
					"claim to have executed a step whose environment it had to guess", k, v)
			}
			env[k] = v
		}
	}
	return env
}

// runStep runs one step's script the way the runner would: written to a
// file and handed to the step's shell, which defaults to `bash -e` -- no
// pipefail, which is why `| tee` was a bypass and why the default is
// modelled rather than improved on here.
func runStep(t *testing.T, step workflowStep, env map[string]string) (string, error) {
	t.Helper()

	script := filepath.Join(t.TempDir(), "step.sh")
	if err := os.WriteFile(script, []byte(step.Run), 0o644); err != nil {
		t.Fatal(err)
	}

	argv := shellArgv(t, step.Shell)
	for i, a := range argv {
		argv[i] = strings.ReplaceAll(a, "{0}", script)
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Dir = repoRoot(t)
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// shellArgv is GitHub's documented resolution of a step's shell: absent means
// `bash -e {0}`, the named shells get their documented defaults, and anything
// else is a template that must place the script with {0}.
func shellArgv(t *testing.T, shell string) []string {
	t.Helper()
	switch shell {
	case "":
		return []string{"bash", "-e", "{0}"}
	case "bash":
		return []string{"bash", "--noprofile", "--norc", "-eo", "pipefail", "{0}"}
	case "sh":
		return []string{"sh", "-e", "{0}"}
	}
	if !strings.Contains(shell, "{0}") {
		t.Fatalf("shell %q is not a shell this test models and not a {0} template", shell)
	}
	return strings.Fields(shell)
}

// stepRunning finds the step whose script mentions needle. The receipt
// check is the step that invokes cmd/receipt; it is executed, so a step
// that merely mentions the name in an echo is found here and then fails.
func stepRunning(t *testing.T, wf workflowFile, needle string) workflowStep {
	t.Helper()
	for _, step := range wf.Jobs["ratchet"].Steps {
		if strings.Contains(step.Run, needle) {
			return step
		}
	}
	t.Fatalf("no step in the ratchet job runs %s", needle)
	return workflowStep{}
}

// sweepName is the sweep step's name in parse-fidelity.yml.
const sweepName = "Run the fidelity ratchet"

// isSweep reports whether step is the corpus sweep: the step with the
// sweep's name, or whose script passes -parseoracle.corpus. Never a test
// name: the 5.44-held step may name the same tests the sweep runs.
func isSweep(step workflowStep) bool {
	return step.Name == sweepName || strings.Contains(step.Run, "-parseoracle.corpus")
}

// sweepStep finds the corpus sweep step.
func sweepStep(t *testing.T, wf workflowFile) workflowStep {
	t.Helper()
	for _, step := range wf.Jobs["ratchet"].Steps {
		if isSweep(step) {
			return step
		}
	}
	t.Fatalf("no step in the ratchet job is named %q or passes -parseoracle.corpus", sweepName)
	return workflowStep{}
}

// TestGateFindsTheSweepStep: a REQUIRE step naming TestRatchetCorpus that
// comes before the sweep must not be mistaken for it, or the gate would
// execute the wrong step and prove nothing about the sweep.
func TestGateFindsTheSweepStep(t *testing.T) {
	var wf workflowFile
	if err := yaml.Unmarshal([]byte(`
jobs:
  ratchet:
    steps:
    - name: Run the 5.44-held tests
      env:
        PVM_REQUIRE_PINNED_PERL: '1'
      run: go test ./internal/parseoracle/ -run '^(TestRatchetCorpus)$'
    - name: Run the fidelity ratchet
      env:
        PVM_REQUIRE_PINNED_PERL: '1'
      run: go test ./internal/parseoracle/ -run '^(TestRatchetCorpus)$' -parseoracle.corpus
`), &wf); err != nil {
		t.Fatal(err)
	}
	if got := sweepStep(t, wf); got.Name != sweepName {
		t.Errorf("the gate picked step %q as the sweep, want %q", got.Name, sweepName)
	}
}

// buildFixtureShim lays the five fixture files out the way corpusReport
// expects a built shim: a t/ with test.pl in it and .t files to walk,
// stamped with p's $] and threads and with revision. The sweep sees a shim that is already
// built for the perl it runs, which is the cache-hit path; unstamped, it
// would try to rebuild the shim from the corpus this gate withholds.
func buildFixtureShim(t *testing.T, p perltest.Perl, shim, revision string) {
	t.Helper()

	built, err := parseoracle.ObservePin(context.Background(), p.Path, revision)
	if err != nil {
		t.Fatalf("asking %s what it is: %v", p.Path, err)
	}

	shimT := filepath.Join(shim, "t")
	if err := os.MkdirAll(shimT, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(shimT, "test.pl"), []byte("1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, src := range fixtureCorpus(t) {
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}
		dst := filepath.Join(shimT, gateName(src))
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := parseoracle.WriteShimStamp(shim, built); err != nil {
		t.Fatal(err)
	}
}

// gateName is where buildFixtureShim puts a fixture file: its base name,
// .pl swapped for .t, at the top of the shim's t/.
func gateName(fixture string) string {
	return strings.TrimSuffix(filepath.Base(fixture), ".pl") + ".t"
}

// gateBaseline is the committed fixture baseline with its paths rewritten to
// where buildFixtureShim puts the files. Deriving it keeps this test in step
// with fixture.txt when TestRatchet re-baselines it.
func gateBaseline(t *testing.T, temp string) string {
	t.Helper()

	base := loadFixtureBaseline(t)
	for i := range base.Rows {
		base.Rows[i].Path = gateName(base.Rows[i].Path)
	}
	path := filepath.Join(temp, "gate-baseline.txt")
	if err := parseoracle.WriteBaseline(path, base); err != nil {
		t.Fatal(err)
	}
	return path
}

// wrongBaseline is the gate baseline with one verdict moved, so a sweep of
// the unchanged fixture reads as a regression.
func wrongBaseline(t *testing.T, temp, good string) string {
	t.Helper()

	base, err := parseoracle.LoadBaseline(good)
	if err != nil {
		t.Fatal(err)
	}
	moved := parseoracle.BucketWrong.String()
	if base.Rows[0].Status == moved {
		moved = parseoracle.BucketExact.String()
	}
	base.Rows[0].Status = moved

	path := filepath.Join(temp, "wrong-baseline.txt")
	if err := parseoracle.WriteBaseline(path, base); err != nil {
		t.Fatal(err)
	}
	return path
}

// corpusReceipt is the receipt a sweep of the committed corpus would write,
// built from the committed baseline's own rows and sites.txt's digest. It is what the receipt step
// must accept; a step that rejects everything is not a check either.
func corpusReceipt(t *testing.T) parseoracle.Receipt {
	t.Helper()

	base, err := parseoracle.LoadBaseline(corpusBaselinePath())
	if err != nil {
		t.Fatal(err)
	}
	var report parseoracle.Report
	for _, row := range base.Rows {
		report.Files = append(report.Files, reportWithStatus(t, row.Path, row.Status).Files...)
	}
	r, err := parseoracle.NewReceipt("TestRatchetCorpus", corpusBaselinePath(), report)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.RecordSites(corpusSitesPath()); err != nil {
		t.Fatal(err)
	}
	return r
}
