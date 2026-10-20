// ABOUTME: Holds each .tpm's declared return types to perl: perl calls each sub with its oracle call cases.
// ABOUTME: A value perl returns that the declared return type excludes fails the test.

package infer_test

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/perltest"
	"tamarou.com/pvm/internal/types"
)

// oracleExempt are the declared subs the return-type oracle never calls,
// "Module::sub" to why: one that is unsafe in a scratch directory --
// network, fork/exec, signals, _exit, file mutation outside scratch,
// blocking on a terminal. Each is held to its prototype and exports only.
var oracleExempt = map[string]string{
	"POSIX::_exit":       "terminates the process immediately; never returns",
	"POSIX::abort":       "raises SIGABRT, terminating the process",
	"POSIX::pause":       "blocks until a signal arrives; nothing in the oracle run ever signals it",
	"POSIX::exit":        "calls CORE::exit directly, bypassing the driver's exit-catching override, and ends the oracle process early",
	"POSIX::fork":        "creates a real child process",
	"POSIX::kill":        "sends a real signal to a process",
	"POSIX::raise":       "sends a real signal to the current process",
	"POSIX::sleep":       "blocks for real wall-clock time",
	"POSIX::alarm":       "schedules a real SIGALRM that could terminate the oracle process later in the run",
	"POSIX::wait":        "may block indefinitely waiting for a child process",
	"POSIX::waitpid":     "may block indefinitely waiting for a child process",
	"POSIX::sigsuspend":  "blocks until a signal is delivered",
	"POSIX::system":      "spawns an arbitrary subprocess",
	"POSIX::setuid":      "changes the real process's user id",
	"POSIX::setgid":      "changes the real process's group id",
	"POSIX::setpgid":     "changes a real process's process group",
	"POSIX::setsid":      "changes the real process's session, detaching it from its controlling terminal",
	"POSIX::mkfifo":      "creates a named pipe node on disk",
	"POSIX::unlink":      "deletes a real file",
	"POSIX::rmdir":       "deletes a real directory",
	"POSIX::sigaction":   "installs a real signal handler for the rest of the run",
	"POSIX::signal":      "installs a real signal handler for the rest of the run",
	"POSIX::sigprocmask": "alters the process's real signal mask for the rest of the run",
	"POSIX::sigpending":  "reads live process signal state that only the exempted signal-mutating functions above establish",
	"POSIX::tcsetpgrp":   "mutates the terminal's foreground process group",
	"POSIX::tcsetattr":   "mutates real terminal attributes",
	"POSIX::tcflow":      "suspends or resumes real terminal transmission",
	"POSIX::tcsendbreak": "sends a real break condition on the terminal line",
	"POSIX::tcdrain":     "blocks until pending terminal output drains",
	"POSIX::tcflush":     "discards real pending terminal input or output",
	"POSIX::setlocale":   "mutates the process-wide locale, changing every later locale-sensitive call in the same run",
	"POSIX::nice":        "changes the real process's scheduling priority",
	"POSIX::read":        "its buffer argument is aliased in place; the oracle's literal call-case arguments cannot supply a safe mutable lvalue",
	// sockaddr_in and sockaddr_in6 dispatch on wantarray and arg count
	// together: the pack arity in list context, or the unpack arity in
	// scalar, croaks (measured), which the oracle's per-sub call cases
	// cannot give one direction without also calling the other in the
	// context that dies for it. pack_sockaddr_in(6) and
	// unpack_sockaddr_in(6) hold the same two directions without the
	// conflict.
	"Socket::sockaddr_in":  "wantarray/arity dispatch dies on the other direction's shape; held to its prototype and exports only",
	"Socket::sockaddr_in6": "wantarray/arity dispatch dies on the other direction's shape; held to its prototype and exports only",
}

// posixReplacementNames are the names in POSIX.pm 2.23's %replacement
// hash that POSIX also exports: measured on perl 5.42.0, each croaks
// "Unimplemented: POSIX::name(): ..." the moment it is called, regardless
// of its arguments.
var posixReplacementNames = []string{
	"atexit", "atof", "atoi", "atol", "bsearch", "calloc", "clearerr", "div",
	"execl", "execle", "execlp", "execv", "execve", "execvp", "fclose",
	"fdopen", "feof", "ferror", "fflush", "fgetc", "fgetpos", "fgets",
	"fopen", "fprintf", "fputc", "fputs", "fread", "free", "freopen",
	"fscanf", "fseek", "fsetpos", "fsync", "ftell", "fwrite", "labs",
	"ldiv", "longjmp", "malloc", "memchr", "memcmp", "memcpy", "memmove",
	"memset", "offsetof", "putc", "putchar", "puts", "qsort", "realloc",
	"scanf", "setbuf", "setjmp", "setvbuf", "siglongjmp", "sigsetjmp",
	"sscanf", "strcat", "strchr", "strcmp", "strcpy", "strcspn", "strlen",
	"strncat", "strncmp", "strncpy", "strpbrk", "strrchr", "strspn",
	"strtok", "tmpfile", "tmpnam", "ungetc", "vfprintf", "vprintf",
	"vsprintf",
}

// posixUndefinedMacroNames are exported POSIX names whose AUTOLOAD falls
// through to constant(), which does not recognize or define them on this
// platform: measured on perl 5.42.0, each dies every time it is called.
var posixUndefinedMacroNames = []string{
	"LC_SYNTAX", "LC_TOD", "NDEBUG", "cfgetispeed", "cfgetospeed",
	"cfsetispeed", "cfsetospeed", "mbstowcs", "wcstombs", "stdin", "stdout",
	"stderr", "tcgetattr",
}

// holdReturnTypes holds each module's shipped .tpm to perl: every sub it
// declares, called with the cases in testdata/oracle/<Module>.pl (File::Spec
// at File-Spec.pl), returns
// only values its declared return type admits. A batch of declarations
// calls it once with its modules.
func holdReturnTypes(t *testing.T, modules ...string) {
	t.Helper()
	p := perltest.Perl542(t)
	for _, module := range modules {
		rel := strings.ReplaceAll(module, "::", "/")
		tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", filepath.FromSlash(rel)+".tpm"))
		if err != nil {
			t.Errorf("%s: no declaration: %v", module, err)
			continue
		}
		cases, err := filepath.Abs(filepath.Join("testdata", "oracle", strings.ReplaceAll(module, "::", "-")+".pl"))
		require.NoError(t, err)
		for _, f := range returnTypeFailures(t, p, oracleRun{Module: module, TPM: tpm, Cases: cases, Exempt: oracleExempt}) {
			t.Error(f)
		}
	}
}

// oracleRun is one module the return-type oracle holds to perl.
type oracleRun struct {
	// Module is the package perl requires, and TPM its declaration source.
	Module string
	TPM    []byte
	// Cases is the call-case file: sub name => [ [args], ... ].
	Cases string
	// Inc is a directory perl searches first, or "".
	Inc string
	// Exempt are the subs never called, "Module::sub" to why.
	Exempt map[string]string
}

// returnTypeFailures has perl call each sub r.TPM declares, except those
// exempt, with its call cases, and returns one line per returned value
// the declared return type excludes, or why the oracle could not run.
// Perl runs with its working directory, HOME and TMPDIR in a temp dir.
func returnTypeFailures(t *testing.T, p perltest.Perl, r oracleRun) []string {
	t.Helper()
	sigs, errs := parse.DeclaredSignatures(r.TPM)
	if len(errs) > 0 {
		return []string{fmt.Sprintf("%s: declaration errors: %v", r.Module, errs)}
	}
	observer, err := filepath.Abs(filepath.Join("testdata", "observe_types.pl"))
	require.NoError(t, err)
	driver, err := filepath.Abs(filepath.Join("testdata", "return_oracle.pl"))
	require.NoError(t, err)

	args := []string{driver, observer, r.Cases, r.Module}
	if r.Inc != "" {
		args = append([]string{"-I", r.Inc}, args...)
	}
	for _, name := range slices.Sorted(maps.Keys(sigs)) {
		if _, exempt := r.Exempt[r.Module+"::"+name]; exempt {
			continue
		}
		for _, call := range calls(sigs[name]) {
			args = append(args, name+":"+call)
		}
	}

	scratch := t.TempDir()
	cmd := p.Command(args...)
	cmd.Dir = scratch
	cmd.Env = append(cmd.Env, "HOME="+scratch, "TMPDIR="+scratch)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return []string{fmt.Sprintf("%s: the oracle did not run: %v: %s", r.Module, err, stderr.String())}
	}
	// The driver says it is complete once every sub is called, so a sub
	// that ends perl early fails here rather than passing the rest.
	report, complete := strings.CutPrefix(string(out), "complete\n")
	if !complete {
		return []string{fmt.Sprintf("%s: the oracle ended before calling every sub: %s", r.Module, stderr.String())}
	}
	var obs []observation
	if err := json.Unmarshal([]byte(report), &obs); err != nil {
		return []string{fmt.Sprintf("%s: the oracle emitted %q: %v", r.Module, out, err)}
	}

	var failures []string
	// failed are the calls a returned value already failed, so the call's
	// Returned marker, which follows its values, adds no second failure.
	failed := map[string]bool{}
	for _, o := range obs {
		f := strings.Split(o.Name, ":")
		name, ctx, form := f[0], f[1], f[2]
		n, _ := strconv.Atoi(f[3])
		call := fmt.Sprintf("%s:%d", o.Name, o.Line)
		if o.Type == "Returned" {
			// A call returning the empty list records no value; it still
			// returned, which only a candidate that returns may take.
			returns := slices.ContainsFunc(sigs[name], func(s types.Signature) bool {
				return s.Returns != types.None && takes(s, form, n) && callContext(s) == ctx
			})
			if !returns && !failed[call] {
				failures = append(failures, fmt.Sprintf("%s::%s call case %d in %s context returned, which no candidate taking %d arguments in that context admits",
					r.Module, name, o.Line, ctx, n))
			}
			continue
		}
		if o.Type == "Died" {
			if !dies(sigs[name], form, n) {
				failures = append(failures, fmt.Sprintf("%s::%s call case %d in %s context died, which no candidate declared None taking %d arguments admits",
					r.Module, name, o.Line, ctx, n))
			}
			continue
		}
		got, known := observedToLattice[o.Type]
		admitted := known && slices.ContainsFunc(sigs[name], func(s types.Signature) bool {
			if s.Returns == types.None || !takes(s, form, n) || callContext(s) != ctx {
				return false
			}
			// In list context each value is one element of the list.
			if ctx == "list" && s.ReturnElement != types.Unknown {
				return types.IsSubtype(got, s.ReturnElement)
			}
			return types.IsSubtype(got, s.Returns)
		})
		if !admitted {
			failed[call] = true
			failures = append(failures, fmt.Sprintf("%s::%s call case %d in %s context returned %s, which its declared return type excludes",
				r.Module, name, o.Line, ctx, o.Type))
		}
	}
	sort.Strings(failures)
	return failures
}

// calls is how the oracle calls a sub's cases, "CONTEXT:FORM", once for
// each context and form one of its candidates answers. A candidate
// declared never to return answers no context: its cases die in whatever
// context the others are called in, or in scalar context when no other
// candidate takes its form.
func calls(cands []types.Signature) []string {
	var out []string
	for _, s := range cands {
		if call := callContext(s) + ":" + callForm(s); s.Returns != types.None && !slices.Contains(out, call) {
			out = append(out, call)
		}
	}
	for _, s := range cands {
		form := callForm(s)
		if s.Returns == types.None && !slices.ContainsFunc(out, func(c string) bool { return strings.HasSuffix(c, ":"+form) }) {
			out = append(out, "scalar:"+form)
		}
	}
	return out
}

// callContext is the context the oracle calls a candidate's cases in,
// and the one its returns are held in: list for one answering list
// context, RFC 0001 "Context selects by return type", or declared Void,
// where a bare return is the empty list; scalar for any other.
func callContext(s types.Signature) string {
	if a := s.Answers(); a == types.ListCtx || a == types.VoidCtx {
		return "list"
	}
	return "scalar"
}

// callForm is how the oracle calls a candidate: as a method, with the
// module as its invocant, when it declares one; as a sub otherwise.
func callForm(s types.Signature) string {
	if s.Invocant != nil {
		return "method"
	}
	return "sub"
}

// dies is whether a call of form with n arguments may die: some
// candidate declared None takes it.
func dies(cands []types.Signature, form string, n int) bool {
	return slices.ContainsFunc(cands, func(s types.Signature) bool {
		return s.Returns == types.None && takes(s, form, n)
	})
}

// takes is whether candidate s takes a call of form with n arguments
// besides any invocant, so a value the call returns, or its death, is
// held to s.
func takes(s types.Signature, form string, n int) bool {
	if form == "method" {
		n++
	}
	return callForm(s) == form && types.Accepts([]types.Signature{s}, n)
}

// oracleFixture writes module's perl source under a lib directory in a
// temp dir, and its call cases beside it, returning the lib directory and
// the case file.
func oracleFixture(t *testing.T, module, pm, cases string) (lib, caseFile string) {
	t.Helper()
	dir := t.TempDir()
	lib = filepath.Join(dir, "lib")
	pmPath := filepath.Join(lib, filepath.FromSlash(strings.ReplaceAll(module, "::", "/"))+".pm")
	require.NoError(t, os.MkdirAll(filepath.Dir(pmPath), 0o755))
	require.NoError(t, os.WriteFile(pmPath, []byte(pm), 0o644))
	caseFile = filepath.Join(dir, "cases.pl")
	require.NoError(t, os.WriteFile(caseFile, []byte(cases), 0o644))
	return lib, caseFile
}

// TestReturnOracleRejectsAnExcludedValue: a sub declared Int whose perl
// returns a string fails the oracle, and one returning an integer passes.
func TestReturnOracleRejectsAnExcludedValue(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Oracle;\nsub count () Int;\n")
	cases := "(count => [ [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Oracle", "package My::Oracle;\nsub count { 'many' }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Oracle", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a Str where Int is declared")
	assert.Contains(t, failures[0], "count")

	lib, caseFile = oracleFixture(t, "My::Oracle", "package My::Oracle;\nsub count { 3 }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Oracle", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleCallsInTheDeclaredContext: a sub declared to return a
// scalar is called in scalar context, where perl's wantarray is false.
func TestReturnOracleCallsInTheDeclaredContext(t *testing.T) {
	p := perltest.Perl542(t)
	lib, caseFile := oracleFixture(t, "My::Wants", "package My::Wants;\nsub n { wantarray ? 'list' : 1 }\n1;\n", "(n => [ [] ]);\n")
	tpm := []byte("package My::Wants;\nsub n () Int;\n")

	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Wants", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleNeverCallsAnExemptSub: an exempt sub that would write a
// marker file is not called, so no marker is left.
func TestReturnOracleNeverCallsAnExemptSub(t *testing.T) {
	p := perltest.Perl542(t)
	marker := filepath.Join(t.TempDir(), "called")
	pm := fmt.Sprintf("package My::Exempt;\nsub touch { open my $fh, '>', %q or die; 1 }\nsub safe { 1 }\n1;\n", marker)
	lib, caseFile := oracleFixture(t, "My::Exempt", pm, "(touch => [ [] ], safe => [ [] ]);\n")
	tpm := []byte("package My::Exempt;\nsub touch () Int;\nsub safe () Int;\n")
	exempt := map[string]string{"My::Exempt::touch": "writes a file outside scratch"}

	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Exempt", TPM: tpm, Cases: caseFile, Inc: lib, Exempt: exempt}))
	assert.NoFileExists(t, marker, "the exempt sub was called")
}

// TestReturnOracleExemptionsNameAReason: every exemption says why.
func TestReturnOracleExemptionsNameAReason(t *testing.T) {
	for sub, reason := range oracleExempt {
		assert.NotEmpty(t, strings.TrimSpace(reason), "%s is exempt from the return-type oracle without a reason", sub)
	}
}

// TestReturnOracleRejectsAReturningNoneSub: a sub declared None that
// returns fails the oracle, and one that dies passes it.
func TestReturnOracleRejectsAReturningNoneSub(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Never;\nsub stop () None;\n")
	cases := "(stop => [ [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Never", "package My::Never;\nsub stop { 1 }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Never", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a sub declared None returned")
	assert.Contains(t, failures[0], "stop")

	lib, caseFile = oracleFixture(t, "My::Never", "package My::Never;\nsub stop { die 'stopped' }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Never", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleRejectsAnUncoveredSub: a declared sub with neither a
// call case nor an exemption fails the oracle by name.
func TestReturnOracleRejectsAnUncoveredSub(t *testing.T) {
	p := perltest.Perl542(t)
	lib, caseFile := oracleFixture(t, "My::Uncovered",
		"package My::Uncovered;\nsub one { 1 }\nsub two { 2 }\n1;\n", "(one => [ [] ]);\n")
	tpm := []byte("package My::Uncovered;\nsub one () Int;\nsub two () Int;\n")

	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Uncovered", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "two has no call case")
	assert.Contains(t, failures[0], "two has no call case")
}

// TestReturnOracleRunsInScratchDir: perl's working directory, HOME and
// TMPDIR are the oracle's temp dir, so a sub writing under $ENV{HOME}
// writes there, as one writing to a relative path or under TMPDIR does.
func TestReturnOracleRunsInScratchDir(t *testing.T) {
	p := perltest.Perl542(t)
	pm := `package My::Scratch;
use Cwd ();
sub where {
    open my $fh, '>', "$ENV{HOME}/home-marker" or die;
    open $fh, '>', "$ENV{TMPDIR}/tmp-marker" or die;
    open $fh, '>', 'cwd-marker' or die;
    return Cwd::getcwd() eq $ENV{HOME} && $ENV{HOME} eq $ENV{TMPDIR} ? 1 : 'elsewhere';
}
1;
`
	lib, caseFile := oracleFixture(t, "My::Scratch", pm, "(where => [ [] ]);\n")
	tpm := []byte("package My::Scratch;\nsub where () Int;\n")
	home, _ := os.UserHomeDir()

	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Scratch", TPM: tpm, Cases: caseFile, Inc: lib}),
		"perl's cwd, HOME and TMPDIR are one scratch directory")
	assert.NoFileExists(t, filepath.Join(home, "home-marker"), "the sub wrote under the real HOME")
}

// TestCarpReturnTypes holds Carp.tpm's return types to perl.
func TestCarpReturnTypes(t *testing.T) {
	holdReturnTypes(t, "Carp")
}

// TestConstantExportersReturnTypes holds Fcntl.tpm's and Socket.tpm's
// return types to perl.
func TestConstantExportersReturnTypes(t *testing.T) {
	holdReturnTypes(t, "Fcntl", "Socket")
}

// TestSocketOracleNeverResolvesAName: the oracle's Socket call cases
// give inet_aton a numeric address only, and give getaddrinfo and
// getnameinfo a numeric address along with AI_NUMERICHOST or
// NI_NUMERICHOST, so none of the three ever asks libc to resolve a
// hostname over the network.
func TestSocketOracleNeverResolvesAName(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("testdata", "oracle", "Socket.pl"))
	require.NoError(t, err)
	text := string(src)

	ipLike := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
	for _, m := range regexp.MustCompile(`inet_aton\('([^']*)'\)`).FindAllStringSubmatch(text, -1) {
		assert.Regexp(t, ipLike, m[1], "inet_aton called with %q, which is not a numeric address", m[1])
	}

	for _, name := range []string{"getaddrinfo", "getnameinfo"} {
		block := oracleCaseBlock(t, text, name)
		assert.Contains(t, block, "NUMERICHOST", "%s's call cases never set *NUMERICHOST, so a bad address could resolve over the network", name)
	}
}

// oracleCaseBlock is the bracketed value `name => [ ... ],` gives in an
// oracle call-case file's source, found by counting brackets rather than
// a regexp, since the value nests arrayrefs and hashrefs.
func oracleCaseBlock(t *testing.T, src, name string) string {
	t.Helper()
	key := name + " => ["
	start := strings.Index(src, key)
	require.NotEqual(t, -1, start, "%s has no call case in the oracle file", name)
	depth := 0
	for i := start + len(key) - 1; i < len(src); i++ {
		switch src[i] {
		case '[':
			depth++
		case ']':
			depth--
			if depth == 0 {
				return src[start : i+1]
			}
		}
	}
	t.Fatalf("%s: unbalanced brackets", name)
	return ""
}

// TestReturnOracleChecksEachListElement: a sub declared List[Int] is held
// element by element, so a list holding a Str, a reference or undef fails.
func TestReturnOracleChecksEachListElement(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Nums;\nsub nums () List[Int];\n")
	cases := "(nums => [ [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Nums", "package My::Nums;\nsub nums { ('a', [1], undef) }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Nums", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 3, "a Str, an ArrayRef and an Undef where List[Int] is declared")
	assert.Contains(t, strings.Join(failures, "\n"), "returned Str")

	lib, caseFile = oracleFixture(t, "My::Nums", "package My::Nums;\nsub nums { (1, 2) }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Nums", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleCallsEachCandidateInItsContext: a multi sub whose
// candidates answer scalar and list context is called in both, each
// return held to the candidate answering that context.
func TestReturnOracleCallsEachCandidateInItsContext(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Both;\nmulti sub n () Int;\nmulti sub n () List[Str];\n")
	cases := "(n => [ [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Both", "package My::Both;\nsub n { wantarray ? ('a', 'b') : 'one' }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Both", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a Str in scalar context, where the scalar candidate returns Int")
	assert.Contains(t, failures[0], "in scalar context returned Str")

	lib, caseFile = oracleFixture(t, "My::Both", "package My::Both;\nsub n { wantarray ? ('a', 'b') : 2 }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Both", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleFailsOnASubThatExits: a sub calling exit ends perl
// before the subs after it are called, which fails the oracle rather than
// passing it with those subs unchecked.
func TestReturnOracleFailsOnASubThatExits(t *testing.T) {
	p := perltest.Perl542(t)
	// exit dies, which a sub declared Int may not; CORE::exit ends perl
	// before the driver says it is complete.
	want := map[string]string{
		"exit 0":       "My::Exits::a_exits call case 0 in scalar context died",
		"CORE::exit 0": "My::Exits: the oracle ended before calling every sub",
	}
	for _, exit := range []string{"exit 0", "CORE::exit 0"} {
		t.Run(exit, func(t *testing.T) {
			pm := fmt.Sprintf("package My::Exits;\nsub a_exits { %s }\nsub b_str { 'str' }\n1;\n", exit)
			lib, caseFile := oracleFixture(t, "My::Exits", pm, "(a_exits => [ [] ], b_str => [ [] ]);\n")
			tpm := []byte("package My::Exits;\nsub a_exits () Int;\nsub b_str () Int;\n")

			failures := returnTypeFailures(t, p, oracleRun{Module: "My::Exits", TPM: tpm, Cases: caseFile, Inc: lib})
			assert.Contains(t, strings.Join(failures, "\n"), want[exit])
		})
	}
}

// TestReturnOracleTurnsExitIntoADeath: exit dies under the oracle, so a
// sub declared None that calls it is held as one that dies.
func TestReturnOracleTurnsExitIntoADeath(t *testing.T) {
	p := perltest.Perl542(t)
	lib, caseFile := oracleFixture(t, "My::Quits", "package My::Quits;\nsub quit { exit 0 }\n1;\n", "(quit => [ [] ]);\n")
	tpm := []byte("package My::Quits;\nsub quit () None;\n")

	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Quits", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleKeepsAModulesLoadTimeOutputOutOfItsReport: a module that
// dups STDOUT as it loads, as Test::Builder does, writes to STDERR through
// its dup, and its END setting perl's exit status, as Test::Builder's does
// when no plan was declared, leaves a complete run successful.
func TestReturnOracleKeepsAModulesLoadTimeOutputOutOfItsReport(t *testing.T) {
	p := perltest.Perl542(t)
	pm := "package My::Builder;\nuse Test::More ();\nsub pass_one { Test::More::ok(1, 'name') }\n1;\n"
	lib, caseFile := oracleFixture(t, "My::Builder", pm, "(pass_one => [ [] ]);\n")
	tpm := []byte("package My::Builder;\nsub pass_one () Int;\n")

	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Builder", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleCallsAVoidSubInListContext: a sub declared Void is
// called in list context, where a bare return is the empty list and any
// value it returns fails.
func TestReturnOracleCallsAVoidSubInListContext(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Void;\nsub v () Void;\n")
	cases := "(v => [ [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Void", "package My::Void;\nsub v { return }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Void", TPM: tpm, Cases: caseFile, Inc: lib}))

	lib, caseFile = oracleFixture(t, "My::Void", "package My::Void;\nsub v { 1 }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Void", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a value where Void is declared")
	assert.Contains(t, failures[0], "in list context returned Int")
}

// TestReturnOracleAdmitsADeathOnlyThroughTheNoneCandidate: a multi sub
// with an Int candidate taking one argument and a None candidate taking
// none may die when called with none, and must return an Int when called
// with one.
func TestReturnOracleAdmitsADeathOnlyThroughTheNoneCandidate(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Maybe;\nmulti sub f (Int $x) Int;\nmulti sub f () None;\n")
	cases := "(f => [ [1], [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Maybe", "package My::Maybe;\nsub f { @_ ? $_[0] : die 'none' }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Maybe", TPM: tpm, Cases: caseFile, Inc: lib}))

	lib, caseFile = oracleFixture(t, "My::Maybe", "package My::Maybe;\nsub f { die 'always' }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Maybe", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a death with one argument, which only the Int candidate takes")
	assert.Contains(t, failures[0], "My::Maybe::f call case 0 in scalar context died")

	lib, caseFile = oracleFixture(t, "My::Maybe", "package My::Maybe;\nsub f { @_ ? $_[0] : 0 }\n1;\n", cases)
	failures = returnTypeFailures(t, p, oracleRun{Module: "My::Maybe", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a return with no argument, which only the None candidate takes")
	assert.Contains(t, failures[0], "My::Maybe::f call case 1 in scalar context returned Int")
}

// TestReturnOracleRejectsAnEmptyReturnOnlyTheNoneCandidateTakes: a call
// that returns the empty list records no value, yet it returned, so a
// call only the None candidate takes fails even when it returns nothing.
func TestReturnOracleRejectsAnEmptyReturnOnlyTheNoneCandidateTakes(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package My::Empty;\nmulti sub f (Int $x) List[Int];\nmulti sub f () None;\n")
	cases := "(f => [ [1], [] ]);\n"

	lib, caseFile := oracleFixture(t, "My::Empty", "package My::Empty;\nsub f { @_ ? ($_[0]) : die 'none' }\n1;\n", cases)
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Empty", TPM: tpm, Cases: caseFile, Inc: lib}))

	lib, caseFile = oracleFixture(t, "My::Empty", "package My::Empty;\nsub f { @_ ? ($_[0]) : () }\n1;\n", cases)
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Empty", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "an empty return with no argument, which only the None candidate takes")
	assert.Contains(t, failures[0], "My::Empty::f call case 1 in list context returned")
}

// TestReturnOracleCallsAnInheritedClassMethod: a declaration with an
// invocant is a method, found through can as perl's method call finds an
// inherited one, and called with the class as its invocant.
func TestReturnOracleCallsAnInheritedClassMethod(t *testing.T) {
	p := perltest.Perl542(t)
	pm := `package My::Kid;
our @ISA = ('My::Base');
package My::Base;
sub cat { my ($class, @parts) = @_; $class eq 'My::Kid' ? join('/', @parts) : [] }
1;
`
	lib, caseFile := oracleFixture(t, "My::Kid", pm, "(cat => [ [ 'a', 'b' ] ]);\n")

	tpm := []byte("package My::Kid;\nsub cat (Str $class: List[Str] @parts) Str;\n")
	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Kid", TPM: tpm, Cases: caseFile, Inc: lib}))

	tpm = []byte("package My::Kid;\nsub cat (Str $class: List[Str] @parts) Int;\n")
	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Kid", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1, "a Str where Int is declared")
	assert.Contains(t, failures[0], "returned Str")
}

// TestReturnOracleRejectsASubWithNoCases: a call-case entry listing no
// argument lists calls nothing, so it fails as a missing entry does.
func TestReturnOracleRejectsASubWithNoCases(t *testing.T) {
	p := perltest.Perl542(t)
	lib, caseFile := oracleFixture(t, "My::Empty", "package My::Empty;\nsub one { 'one' }\n1;\n", "(one => []);\n")
	tpm := []byte("package My::Empty;\nsub one () Int;\n")

	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Empty", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1)
	assert.Contains(t, failures[0], "one has no call case")
}

// TestReturnOracleRejectsAnUndefinedNoneSub: a sub declared None that
// perl does not define dies when called, "Undefined subroutine", which is
// no evidence it never returns; it fails the oracle.
func TestReturnOracleRejectsAnUndefinedNoneSub(t *testing.T) {
	p := perltest.Perl542(t)
	lib, caseFile := oracleFixture(t, "My::Gone", "package My::Gone;\n1;\n", "(stop => [ [] ]);\n")
	tpm := []byte("package My::Gone;\nsub stop () None;\n")

	failures := returnTypeFailures(t, p, oracleRun{Module: "My::Gone", TPM: tpm, Cases: caseFile, Inc: lib})
	require.Len(t, failures, 1)
	assert.Contains(t, failures[0], "perl defines no sub stop")
}

// TestReturnOracleKeepsASubsOutputOutOfItsReport: a sub printing to
// STDOUT, as Test::More's ok does, leaves the oracle's report readable.
func TestReturnOracleKeepsASubsOutputOutOfItsReport(t *testing.T) {
	p := perltest.Perl542(t)
	lib, caseFile := oracleFixture(t, "My::Prints", "package My::Prints;\nsub ok { print \"ok 1\\n\"; 1 }\n1;\n", "(ok => [ [] ]);\n")
	tpm := []byte("package My::Prints;\nsub ok () Int;\n")

	assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "My::Prints", TPM: tpm, Cases: caseFile, Inc: lib}))
}

// TestReturnOracleIgnoresTheCallersPerlLib: perl finds the module under
// test as an installed perl does, whatever PERL5LIB, PERLLIB or PERL5OPT
// the test runs with: a Carp there whose croak returns is not measured.
func TestReturnOracleIgnoresTheCallersPerlLib(t *testing.T) {
	p := perltest.Perl542(t)
	fake, _ := oracleFixture(t, "Carp", "package Carp;\nsub croak { 1 }\nsub confess { 1 }\nsub carp { 1 }\nsub cluck { 1 }\nsub longmess { 'l' }\nsub shortmess { 's' }\n1;\n", "")
	tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", "Carp.tpm"))
	require.NoError(t, err)
	cases, err := filepath.Abs(filepath.Join("testdata", "oracle", "Carp.pl"))
	require.NoError(t, err)
	for _, env := range [][2]string{{"PERL5LIB", fake}, {"PERLLIB", fake}, {"PERL5OPT", "-I" + fake}} {
		t.Run(env[0], func(t *testing.T) {
			t.Setenv(env[0], env[1])
			assert.Empty(t, returnTypeFailures(t, p, oracleRun{Module: "Carp", TPM: tpm, Cases: cases}))
		})
	}
}

// pureExporters is the pure-Perl exporters batch (01a1464e-db4c): every
// module whose .tpm declares the names a `use` of it installs. Symbol is
// held back; see the matching comment in internal/parse/heldtoperl_test.go.
var pureExporters = []string{
	"File::Basename", "File::Find", "File::Path", "Errno",
	"Text::ParseWords", "Class::Struct", "File::Copy", "Time::Local", "Getopt::Std",
}

// TestPurePerlExportersReturnTypes holds each pure-Perl exporter's
// declared return types to perl, called with its oracle's cases.
func TestPurePerlExportersReturnTypes(t *testing.T) {
	holdReturnTypes(t, pureExporters...)
}

// moduleFailures holds module's shipped .tpm return types to perl,
// called with its shipped oracle/<Module>.pl cases.
func moduleFailures(t *testing.T, module string) []string {
	t.Helper()
	p := perltest.Perl542(t)
	rel := strings.ReplaceAll(module, "::", "/")
	tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", filepath.FromSlash(rel)+".tpm"))
	require.NoError(t, err)
	cases, err := filepath.Abs(filepath.Join("testdata", "oracle", strings.ReplaceAll(module, "::", "-")+".pl"))
	require.NoError(t, err)
	return returnTypeFailures(t, p, oracleRun{Module: module, TPM: tpm, Cases: cases})
}

// TestGetoptStdUnknownOptionReturnIsDeclared: getopts given a switch its
// spec does not list still returns a Boolean (measured on 5.42.0, a
// false one, after warning "Unknown option"), which Getopt::Std.tpm's
// declared return type admits. oracle/Getopt-Std.pl stages @ARGV so
// getopts' one call sees exactly this: a known switch and an unknown one.
func TestGetoptStdUnknownOptionReturnIsDeclared(t *testing.T) {
	assert.Empty(t, moduleFailures(t, "Getopt::Std"))
}

// TestTextParseWordsUnbalancedQuoteReturnIsDeclared: shellwords given an
// unbalanced quote returns the empty list in list context and undef in
// scalar context (measured on 5.42.0), both of which
// Text::ParseWords.tpm's declared return type admits. oracle/Text-ParseWords.pl
// includes this case among shellwords' own.
func TestTextParseWordsUnbalancedQuoteReturnIsDeclared(t *testing.T) {
	assert.Empty(t, moduleFailures(t, "Text::ParseWords"))
}

// TestFileBasenameWindowsPathReturnIsDeclared: fileparse, basename and
// dirname given a Windows-style path still return a plain Str (measured
// on 5.42.0: this platform's fstype is Unix, so none of the three
// splits the path on a backslash), which File::Basename.tpm's declared
// return types admit. oracle/File-Basename.pl includes a Windows path
// among each of the three's own cases.
func TestFileBasenameWindowsPathReturnIsDeclared(t *testing.T) {
	assert.Empty(t, moduleFailures(t, "File::Basename"))
}

// TestPurePerlExportersOracleStaysInScratch: File::Path, File::Copy and
// File::Find touch the filesystem, so their oracle cases name only
// relative paths -- an absolute one would reach outside the scratch
// directory returnTypeFailures confines perl's cwd, HOME and TMPDIR to.
func TestPurePerlExportersOracleStaysInScratch(t *testing.T) {
	absolute := regexp.MustCompile(`['"]/\S`)
	for _, module := range []string{"File::Path", "File::Copy", "File::Find"} {
		cases, err := os.ReadFile(filepath.Join("testdata", "oracle", strings.ReplaceAll(module, "::", "-")+".pl"))
		require.NoError(t, err)
		assert.False(t, absolute.Match(cases), "%s's oracle cases name an absolute path", module)
	}
}

// posixSignatures parses POSIX.tpm, failing t on any declaration error.
func posixSignatures(t *testing.T) map[string][]types.Signature {
	t.Helper()
	tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", "POSIX.tpm"))
	require.NoError(t, err)
	sigs, errs := parse.DeclaredSignatures(tpm)
	require.Empty(t, errs)
	return sigs
}

// TestPOSIXReturnTypes holds POSIX.tpm's return types to perl.
func TestPOSIXReturnTypes(t *testing.T) {
	holdReturnTypes(t, "POSIX")
}

// TestPOSIXAutoloadStubsDie: every name in POSIX.pm's %replacement, and
// every exported name whose AUTOLOAD falls through to a constant()
// perl's build does not recognize or define, is declared None, and perl
// dies calling each with no arguments -- Ruling 5 (perigrin, 2026-10-16).
func TestPOSIXAutoloadStubsDie(t *testing.T) {
	p := perltest.Perl542(t)
	sigs := posixSignatures(t)
	names := append(append([]string{}, posixReplacementNames...), posixUndefinedMacroNames...)
	for _, name := range names {
		cands, ok := sigs[name]
		if !assert.True(t, ok, "POSIX::%s is not declared", name) {
			continue
		}
		for _, s := range cands {
			assert.Equal(t, types.None, s.Returns, "POSIX::%s is not declared None", name)
		}

		out, err := p.Command("-MPOSIX", "-e", fmt.Sprintf(`no strict 'refs'; &{"POSIX::%s"}()`, name)).CombinedOutput()
		assert.Error(t, err, "POSIX::%s did not die: %s", name, out)
	}
}

// TestPOSIXOracleNeverCallsExit: _exit, abort and pause are in the
// return-type oracle's exemption map with a reason, so holdReturnTypes
// (TestReturnOracleNeverCallsAnExemptSub proves the mechanism generally)
// skips them rather than calling them.
func TestPOSIXOracleNeverCallsExit(t *testing.T) {
	for _, name := range []string{"_exit", "abort", "pause"} {
		reason, ok := oracleExempt["POSIX::"+name]
		if !assert.True(t, ok, "POSIX::%s is not in the return-type oracle's exemption map", name) {
			continue
		}
		assert.NotEmpty(t, strings.TrimSpace(reason), "POSIX::%s is exempt without a reason", name)
	}
}

// TestPOSIXStrtolContextReturnIsDeclared: POSIX::strtol returns two
// values in list context (the parsed number and how much of the string
// it consumed) and one in scalar context, measured directly, and
// POSIX.tpm declares a candidate for each.
func TestPOSIXStrtolContextReturnIsDeclared(t *testing.T) {
	p := perltest.Perl542(t)
	out, err := p.Command("-MPOSIX", "-e",
		`my @l = POSIX::strtol("42abc", 10); my $s = POSIX::strtol("42abc", 10); print scalar(@l), ":", (defined $s ? "1" : "0"), "\n";`,
	).Output()
	require.NoError(t, err)
	assert.Equal(t, "2:1\n", string(out), "strtol returns two values in list context and one in scalar context")

	cands := posixSignatures(t)["strtol"]
	require.Len(t, cands, 2, "strtol should have a list candidate and a scalar candidate")
	var sawList, sawScalar bool
	for _, s := range cands {
		switch callContext(s) {
		case "list":
			sawList = true
			assert.True(t, types.IsSubtype(types.Int, s.ReturnElement) || types.IsSubtype(types.Int, s.Returns),
				"strtol's list candidate does not admit the Int values perl returns")
		case "scalar":
			sawScalar = true
			assert.True(t, types.IsSubtype(types.Int, s.Returns), "strtol's scalar candidate does not admit the Int perl returns")
		}
	}
	assert.True(t, sawList, "no candidate answers list context")
	assert.True(t, sawScalar, "no candidate answers scalar context")
}

// TestPOSIXReimplementedSubsAreNotNone: a %reimpl name that returns a
// real value, such as getpid ($$), is not declared None -- Ruling 5
// applies only to the 107 names that die.
func TestPOSIXReimplementedSubsAreNotNone(t *testing.T) {
	cands, ok := posixSignatures(t)["getpid"]
	require.True(t, ok, "POSIX::getpid is not declared")
	for _, s := range cands {
		assert.NotEqual(t, types.None, s.Returns, "POSIX::getpid is declared None, but perl returns $$ when it is called")
	}
}

// xsLibraries is the XS-library batch (01a1464e-dbd3): Encode,
// Time::HiRes, Compress::Raw::Zlib and SDBM_File.
var xsLibraries = []string{"Encode", "Time::HiRes", "Compress::Raw::Zlib", "SDBM_File"}

// TestXSLibrariesReturnTypes holds each XS library's declared return
// types to perl, called with its oracle's cases.
func TestXSLibrariesReturnTypes(t *testing.T) {
	holdReturnTypes(t, xsLibraries...)
}

// TestEncodeUndefInputReturnIsDeclared: encode_utf8(undef) returns undef
// (measured on 5.42.0, with a "Use of uninitialized value" warning, never
// a die), which Encode.tpm's declared Str|Undef admits. oracle/Encode.pl
// includes this case among encode_utf8's own.
func TestEncodeUndefInputReturnIsDeclared(t *testing.T) {
	assert.Empty(t, moduleFailures(t, "Encode"))
}

// TestSDBMFileOracleStaysInScratch: SDBM_File ties a database file, so
// its oracle cases name only a relative path -- an absolute one would
// reach outside the scratch directory returnTypeFailures confines perl's
// cwd, HOME and TMPDIR to.
func TestSDBMFileOracleStaysInScratch(t *testing.T) {
	absolute := regexp.MustCompile(`['"]/\S`)
	cases, err := os.ReadFile(filepath.Join("testdata", "oracle", "SDBM_File.pl"))
	require.NoError(t, err)
	assert.False(t, absolute.Match(cases), "SDBM_File's oracle cases name an absolute path")
}
