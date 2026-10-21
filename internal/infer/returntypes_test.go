// ABOUTME: Holds each .tpm's declared return types to perl: perl calls each sub with its oracle call cases.
// ABOUTME: A value perl returns that the declared return type excludes fails the test.

package infer_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tamarou.com/pvm/internal/infer"
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
	"POSIX::sigprocmask": "alters the process's real signal mask for the rest of the run",
	"POSIX::sigpending":  "reads live process signal state that only the exempted signal-mutating functions above establish",
	"POSIX::tcsetpgrp":   "mutates the terminal's foreground process group",
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
	// sockaddr_un dispatches on wantarray alone: both its pack and unpack
	// forms take exactly one Str argument, so there is no arity to tell
	// them apart the way sockaddr_in's differing arities do; a path given
	// in list context dies the same way ("Bad address family"), measured.
	"Socket::sockaddr_un": "wantarray dispatch dies on the other direction's shape (no arity tells the two forms apart); held to its prototype and exports only",

	"Test::More::BAIL_OUT": "calls CORE::exit directly from Test::Builder's bail_out, bypassing the driver's exit-catching override, and would end the oracle process early",
	// plan dies "You tried to plan twice" once a plan already exists --
	// including one done_testing sets, which the oracle's alphabetically
	// earlier call to done_testing always leaves in place -- so the
	// oracle cannot call it in the same run; held in its own
	// (TestTestMorePlanReturnType), where nothing else sets a plan first.
	"Test::More::plan": "collides with done_testing, called earlier in the same oracle run, which already sets a plan; held in TestTestMorePlanReturnType's own oracleRun instead",

	// import is a declared signature to import() (perigrin's ruling,
	// 2026-10-20), read at `use` time (use.go): it states what a `use`
	// list is, not a sub the oracle calls for its return type.
	"Test::More::import":               "declares the signature to import(), read at `use` time; never called directly",
	"Locale::Maketext::Simple::import": "declares the signature to import(), read at `use` time; never called directly",

	"Getopt::Long::HelpMessage":    "calls Pod::Usage's pod2usage, which exits the process by default",
	"Getopt::Long::VersionMessage": "exits the process by default (measured: exit($pa->{-exitval}) unless -exitval is NOEXIT)",

	"ExtUtils::MakeMaker::WriteMakefile": "builds a real Makefile from a distribution's own metadata, which the oracle's call cases cannot supply",
	"ExtUtils::MakeMaker::prompt":        "reads an answer from STDIN, blocking when none is piped in",

	// Exporter's own export, export_to_level, export_tags and
	// export_ok_tags install symbols into an arbitrary caller's package
	// or mutate its @EXPORT/@EXPORT_OK from a dynamic $pkg/$callpkg pair;
	// a direct call from the oracle could collide with the driver's own
	// symbols or a later call in the same run. import, the one name
	// Exporter itself exports, is measured directly instead.
	"Exporter::export":          "installs symbols into an arbitrary caller's package from a dynamic $pkg/$callpkg pair, which could collide with the driver's own or a later call in the same run",
	"Exporter::export_to_level": "installs symbols into an arbitrary caller's package from a dynamic $pkg/$callpkg pair, which could collide with the driver's own or a later call in the same run",
	"Exporter::export_tags":     "mutates the caller package's @EXPORT from its own %EXPORT_TAGS, which could collide with a later call in the same run",
	"Exporter::export_ok_tags":  "mutates the caller package's @EXPORT_OK from its own %EXPORT_TAGS, which could collide with a later call in the same run",

	// XSLoader::load and ::bootstrap_inherit each end by loading a real
	// shared object: a name that resolves to a real extension would
	// bootstrap it a second time, and one that does not dies with "Can't
	// locate loadable object", which is not evidence of what a real
	// bootstrap returns.
	"XSLoader::load":              "loads a real compiled extension's bootstrap function or dies; the oracle's call cases cannot safely supply one",
	"XSLoader::bootstrap_inherit": "delegates to DynaLoader::bootstrap_inherit, which needs a real compiled extension the oracle's call cases cannot safely supply",

	// DynaLoader::bootstrap and ::bootstrap_inherit end the same way
	// XSLoader's do: a name that resolves to a real extension would
	// bootstrap it a second time, and one that does not dies with "Can't
	// locate loadable object", which is not evidence of what a real
	// bootstrap returns.
	"DynaLoader::bootstrap":         "loads a real compiled extension's bootstrap function or dies; the oracle's call cases cannot safely supply one",
	"DynaLoader::bootstrap_inherit": "delegates to bootstrap, which needs a real compiled extension the oracle's call cases cannot safely supply",

	// DynaLoader's four low-level XS primitives each need a real loaded
	// shared object or library reference the oracle's call cases cannot
	// safely supply (ruling 4, 2026-10-16, DynaLoader's negative scenario).
	"DynaLoader::dl_load_file":    "dynamically loads a real shared object into the process; the oracle's call cases cannot supply one safely",
	"DynaLoader::dl_find_symbol":  "looks up a symbol in a real loaded library reference, which the oracle's call cases cannot supply",
	"DynaLoader::dl_install_xsub": "installs a real XSUB from a symbol reference the oracle's call cases cannot supply",
	"DynaLoader::dl_unload_file":  "would unload a real shared object out from under the running process",

	// Net::Ping and IO::Socket: every network-touching sub the batch
	// declares, held to its prototype and exports only.
	"Net::Ping::pingecho":      "opens a real tcp connection to the given host",
	"Net::Ping::wakeonlan":     "opens a real udp socket and sends a packet onto the network",
	"IO::Socket::sockaddr_in":  "wantarray/arity dispatch dies on the other direction's shape, as Socket::sockaddr_in does; held to its prototype and exports only",
	"IO::Socket::sockaddr_in6": "wantarray/arity dispatch dies on the other direction's shape, as Socket::sockaddr_in6 does; held to its prototype and exports only",
	"IO::Socket::sockaddr_un":  "wantarray dispatch dies on the other direction's shape, as Socket::sockaddr_un does; held to its prototype and exports only",
	"IO::Socket::socket":       "opens a real socket file descriptor",
	"IO::Socket::connect":      "connects a real socket onto the network",
	"IO::Socket::bind":         "binds a real socket to a local address",
	"IO::Socket::listen":       "marks a real socket as listening for connections",
	"IO::Socket::accept":       "blocks accepting a real connection on a socket",
	"IO::Socket::send":         "writes onto a real socket",
	"IO::Socket::recv":         "blocks reading from a real socket",
	"IO::Socket::close":        "dereferences the invocant as a glob to close its socket, dying \"Can't use string ... as a symbol ref\" when the invocant is the bare class name the oracle calls with",

	// configure's own sanity check requires ref($sock) eq "IO::Socket"
	// exactly (IO/Socket.pm's configure), not a subclass: return_oracle.pl
	// resolves the method once via $module->can(name), so it always calls
	// the base IO::Socket::configure directly, never a subclass's
	// override, even against INSTANCE's object -- and that instance is
	// already blessed into IO::Socket::INET by the time configure runs
	// (new's own call to configure, measured), so a second direct call
	// dies "Cannot configure socket in domain". A separate, still-bare
	// IO::Socket instance would work (measured) but the oracle's cases
	// file supports only one shared INSTANCE per module.
	"IO::Socket::configure": "needs an instance still blessed exactly IO::Socket, which INSTANCE no longer is by the time configure runs, and the oracle supports only one shared instance",

	// IO::File::open dereferences its invocant as a glob immediately
	// (`open($fh, ...)`), dying "Can't use string ... as a symbol ref"
	// when the invocant is the bare class name the oracle calls with,
	// measured on 5.42.0; held to its prototype and exports only.
	"IO::File::open": "dereferences the invocant as a glob; dies when called with the class name itself, as the oracle calls it",

	// IO::Handle and FileHandle: format_write, print, printf and say
	// each dereference the invocant as a glob before doing anything
	// else, dying against the oracle's bare class name; see the
	// matching comments in IO/Handle.tpm and FileHandle.tpm.
	"IO::Handle::format_write": "dereferences the invocant as a glob to find its format name; dies \"Undefined format\" when called with the class name itself",
	"IO::Handle::print":        "dereferences the invocant as a glob; dies when called with the class name itself, as the oracle calls it",
	"IO::Handle::printf":       "dereferences the invocant as a glob; dies when called with the class name itself, as the oracle calls it",
	"IO::Handle::say":          "dereferences the invocant as a glob; dies when called with the class name itself, as the oracle calls it",
	"FileHandle::print":        "dereferences the invocant as a glob; dies when called with the class name itself, as the oracle calls it",
	"FileHandle::printf":       "dereferences the invocant as a glob; dies when called with the class name itself, as the oracle calls it",

	// FileHandle also aliases IO::File::open under its own name, by the
	// same code ref, so it dies the same way IO::File::open does.
	"FileHandle::open": "dereferences the invocant as a glob; dies when called with the class name itself, as the oracle calls it",
	// getpos and setpos are IO::Seekable's, aliased into FileHandle by
	// the same code ref; each dereferences the invocant as a glob (to
	// call C's fgetpos()/fsetpos()) and dies "Bad filehandle" when
	// called with the class name itself, measured on 5.42.0.
	"FileHandle::getpos": "dereferences the invocant as a glob to call fgetpos(); dies \"Bad filehandle\" when called with the class name itself, as the oracle calls it",
	"FileHandle::setpos": "dereferences the invocant as a glob to call fsetpos(); dies \"Bad filehandle\" when called with the class name itself, as the oracle calls it",
}

// pendingMisfit is one oracle misfit deferred to the issue auditing its
// module's .tpm: a parameter the declaration rejects though perl
// accepts it (Kind "parameter"), or, the DualVar knock-on, a return the
// declaration no longer admits now that the observer tells DualVar
// apart from Str and Num (Kind "return"). The fixing issue widens the
// .tpm and deletes its own entries here; nothing in this package
// widens a .tpm to absorb a misfit.
type pendingMisfit struct {
	Module, Sub, Kind, Detail, Reason string
}

// pendingAudit holds every misfit holdReturnTypes' DualVar knock-on or
// the parameter oracle finds in a module whose audit fix has not
// landed, keyed by the issue that will fix it (perigrin, 2026-10-21:
// 01a15ac0-6c4a POSIX, -6c78 sockets, -6ca2 IO and OO modules, -6ccc XS
// modules, -6cfa pure-Perl modules). POSIX's own entries were deleted
// 2026-10-22 once feat/core-module-types (01a15ac0-6c4a's fix) widened
// getpayload to Num|NaN|Inf and setpayload/setpayloadsig's $nv to Any:
// re-measured, neither misfits the parameter oracle any longer. Sockets'
// (01a15ac0-6c78) were deleted the same day once that same merge
// declared getaddrinfo's and getnameinfo's error DualVar directly:
// re-measured, the return-type oracle no longer sees a knock-on there
// either. TestPendingAuditEntriesStillMisfit holds every entry that
// remains to the same rule, so a fix that lands without deleting its
// own entries fails here instead of going unnoticed.
var pendingAudit = map[string][]pendingMisfit{
	"01a15ac0-6ca2": {
		{Module: "FileHandle", Sub: "autoflush", Kind: "parameter",
			Detail: "oracle/FileHandle.pl's own case passes 1 for $on; its declared Boolean excludes the Int perl's $| actually integerises it to (Audit C, measured: \"yes\" gives 0, 0.5 gives 0)",
			Reason: "a parameter oracle misfit: $| integerises its argument, so the declaration is Int, not a Boolean idiom"},
		{Module: "IO::Handle", Sub: "autoflush", Kind: "parameter",
			Detail: "oracle/IO-Handle.pl's own case passes 1 for $on; its declared Boolean excludes the Int perl's $| actually integerises it to (Audit C, measured: \"yes\" gives 0, 0.5 gives 0)",
			Reason: "a parameter oracle misfit: $| integerises its argument, so the declaration is Int, not a Boolean idiom"},
	},
	"01a15ac0-6ccc": {
		{Module: "Scalar::Util", Sub: "dualvar", Kind: "return",
			Detail: "the dualvar it constructs fits neither candidate's declared Str",
			Reason: "DualVar knock-on: observe_types.pl used to read dualvar()'s own result by its string half before it told DualVar apart"},
	},
}

// pendingAuditMatch reports whether module::sub already has a kind
// entry in pendingAudit, so holdReturnTypes and the parameter oracle
// suppress exactly the misfits already deferred to a fixing issue, and
// nothing else.
func pendingAuditMatch(module, sub, kind string) bool {
	for _, entries := range pendingAudit {
		for _, m := range entries {
			if m.Module == module && m.Sub == sub && m.Kind == kind {
				return true
			}
		}
	}
	return false
}

// platformExempt holds, per runtime.GOOS, the subs that still need a
// hand-kept exemption on that platform alone: ones that take a real
// parameter, so exemptForModule's own live measurement
// (perltest.ConstantDifferences) does not cover them, or whose return
// type is not a plain scalar-or-None (a union, say). Each reason cites
// the CI job that measured it. A zero-arg constant whose declared
// return is a scalar type or None -- Fcntl's, Socket's, POSIX's and the
// rest's TCP_*, DN_*, F_*, SO_* and B0-style families among them -- is
// measured directly instead (see exemptForModule), since such a family
// can run into the thousands and differs per C library and per build,
// not something this repository could ever hand-list once for every
// platform CI runs on.
var platformExempt = map[string]map[string]string{
	"darwin": {
		"Fcntl::S_ISWHT":                      "returns Boolean on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"POSIX::getgrnam":                     "returns Undef on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Time::HiRes::CLOCK_BOOTTIME":         "dies on macOS's perl -- measured in CI job 39489951290",
		"Time::HiRes::CLOCK_MONOTONIC_COARSE": "dies on macOS's perl -- measured in CI job 39489951290",
		"Time::HiRes::CLOCK_REALTIME_COARSE":  "dies on macOS's perl -- measured in CI job 39489951290",
		"Time::HiRes::CLOCK_UPTIME_RAW":       "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
	},
	"windows": {
		"POSIX::fpclassify":                            `unimplemented on Windows's perl -- measured in CI job 39508730008: "POSIX::fpclassify not implemented on this architecture"`,
		"File::Copy::copy":                             "returns Boolean on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39508730008",
		"Time::HiRes::clock_nanosleep":                 `unimplemented on Windows's perl -- measured in CI job 39508730008: "Time::HiRes::clock_nanosleep(): unimplemented in this platform"`,
		"Time::HiRes::nanosleep":                       `unimplemented on Windows's perl -- measured in CI job 39489951290 (heldtoperl_test.go's own export probe): "Time::HiRes::nanosleep(): unimplemented in this platform"`,
		"Time::HiRes::ualarm":                          `unimplemented on Windows's perl -- measured in CI job 39489951290 (heldtoperl_test.go's own export probe): "Time::HiRes::ualarm(): unimplemented in this platform"`,
		"Time::HiRes::utime":                           `unimplemented on Windows's perl -- measured in CI job 39489951290 (heldtoperl_test.go's own export probe): "Time::HiRes::utime(): unimplemented in this platform"`,
		"Fcntl::S_ISLNK":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::S_ISSOCK":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"File::Spec::Functions::file_name_is_absolute": "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"File::Spec::file_name_is_absolute":            "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Term::ReadLine::MinLine":                      "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Term::ReadLine::addhistory":                   "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::blocking":                         "returns Boolean on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39514360942",
		"POSIX::S_ISLNK":                               "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::S_ISSOCK":                              "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::WEXITSTATUS":                           "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::WIFEXITED":                             "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::WIFSIGNALED":                           "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::WIFSTOPPED":                            "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::WSTOPSIG":                              "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::WTERMSIG":                              "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::fcntl":                                 "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::fesetround":                            "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::fpathconf":                             "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::lchown":                                "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::pathconf":                              "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::sysconf":                               "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::tcgetpgrp":                             "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::ttyname":                               "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::tzset":                                 "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::stat":                                  "returns Str in list context on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39514360942",
		"POSIX::fstat":                                 "returns Str in list context on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39514360942",
		"POSIX::strtold":                               "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::getpwuid":                              "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::getpwnam":                              "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::getgrnam":                              "dies on Windows's perl -- measured in CI job 39514360942",
		"POSIX::getgrgid":                              "dies on Windows's perl -- measured in CI job 39514360942",
		"Time::HiRes::CLOCK_BOOTTIME":                  "dies on Windows's perl -- measured in CI job 39514360942",
		"Time::HiRes::CLOCK_MONOTONIC_COARSE":          "dies on Windows's perl -- measured in CI job 39514360942",
		"Time::HiRes::CLOCK_MONOTONIC_RAW":             "dies on Windows's perl -- measured in CI job 39514360942",
		"Time::HiRes::ITIMER_PROF":                     "dies on Windows's perl -- measured in CI job 39514360942",
		"Time::HiRes::ITIMER_REAL":                     "dies on Windows's perl -- measured in CI job 39514360942",
		"Time::HiRes::ITIMER_VIRTUAL":                  "dies on Windows's perl -- measured in CI job 39514360942",
		"Time::HiRes::alarm":                           "dies on Windows's perl -- measured in CI job 39514360942",
		"Time::HiRes::getitimer":                       "dies on Windows's perl -- measured in CI job 39514360942",
		"Time::HiRes::setitimer":                       "dies on Windows's perl -- measured in CI job 39514360942",
	},
}

// withPlatformExempt clones base and adds this OS's platformExempt
// entries.
func withPlatformExempt(base map[string]string) map[string]string {
	m := maps.Clone(base)
	for name, reason := range platformExempt[runtime.GOOS] {
		m[name] = reason
	}
	return m
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
	"cfsetispeed", "cfsetospeed", "mbstowcs", "wcstombs", "signal", "stdin",
	"stdout", "stderr", "tcgetattr", "tcsetattr",
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
		for _, f := range returnTypeFailures(t, p, oracleRun{Module: module, TPM: tpm, Cases: cases, Exempt: exemptForModule(t, p, module, tpm, withPlatformExempt(oracleExempt))}) {
			t.Error(f)
		}
	}
}

// exemptForModule is oracleExempt extended with this OS's platformExempt,
// and, for a module whose exported set depends on the platform's C
// library (Errno's), the names perltest.MissingSubs measures absent on
// p's perl right now -- a name genuinely missing there kills the whole
// oracle run on the first one return_oracle.pl is asked to call (its
// callable() check dies rather than recording a per-name failure), so it
// must never reach the call list, and hand-listing which names a given
// libc or Windows's SDK headers omit is not something this repository
// can do once for every platform CI runs on.
func exemptForModule(t *testing.T, p perltest.Perl, module string, tpm []byte, base map[string]string) map[string]string {
	t.Helper()
	exempt := maps.Clone(base)
	if runtime.GOOS == "linux" {
		return exempt
	}
	sigs, errs := parse.DeclaredSignatures(tpm)
	if len(errs) != 0 {
		return exempt
	}
	candidates := constantWantsValue(sigs)
	for name := range candidates {
		// Already exempt for its own (usually behavioral, not
		// platform) reason -- oracleExempt's _exit, abort, pause,
		// fork, wait and the rest are each zero-arg and return-type
		// shaped exactly like a constant, and actually calling one to
		// measure it would do the very thing it is exempt for (abort
		// the process, block waiting for a child that never arrives).
		if _, already := base[module+"::"+name]; already {
			delete(candidates, name)
		}
	}
	for name, reason := range perltest.ConstantDifferences(t, p, "", module, candidates) {
		exempt[module+"::"+name] = reason
	}
	return exempt
}

// constantWantsValue selects, from sigs, the names a module's zero-arg
// constant family is made of: every declared candidate takes no
// parameters and returns a concrete scalar type or None, with every
// candidate agreeing on which. The map's value is true for a name
// declared to return a value, false for one declared None (expected to
// die whenever it is called).
func constantWantsValue(sigs map[string][]types.Signature) map[string]bool {
	out := map[string]bool{}
	for name, cands := range sigs {
		if len(cands) == 0 {
			continue
		}
		qualifies := true
		wants := cands[0].Returns != types.None
		for _, s := range cands {
			// A method (one with an invocant) called bare, with
			// literally no arguments, is not called the way the real
			// oracle calls it (the module name as invocant) -- a
			// plain, truly argument-less sub is the only shape this
			// measures.
			if s.Invocant != nil || len(s.Params) != 0 || !isScalarConstant(s.Returns) {
				qualifies = false
				break
			}
			if (s.Returns != types.None) != wants {
				qualifies = false
				break
			}
		}
		if qualifies {
			out[name] = wants
		}
	}
	return out
}

// isScalarConstant reports whether ty is exactly one of the concrete
// scalar leaf types a plain constant sub returns, or None -- never a
// union (Int|Undef is not one of these), since a union is not the shape
// ConstantDifferences' die-or-return question is about.
func isScalarConstant(ty types.Type) bool {
	switch ty {
	case types.Int, types.Num, types.Str, types.Boolean, types.None:
		return true
	default:
		return false
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
	// IgnorePending turns off pendingAudit's and parameterExempt's own
	// filtering, so TestPendingAuditEntriesStillMisfit can measure
	// whether an entry's misfit is still real.
	IgnorePending bool
}

// oracleCallArgs is the NAME:CONTEXT:FORM arguments return_oracle.pl and
// parameter_oracle.pl both take: one per context and form a non-exempt
// declared sub's candidates answer, so each driver calls (or, for the
// parameter oracle, observes the arguments of) the same cases the other
// does.
func oracleCallArgs(sigs map[string][]types.Signature, module string, exempt map[string]string) []string {
	var args []string
	for _, name := range slices.Sorted(maps.Keys(sigs)) {
		if _, skip := exempt[module+"::"+name]; skip {
			continue
		}
		for _, call := range calls(sigs[name]) {
			args = append(args, name+":"+call)
		}
	}
	return args
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
	args = append(args, oracleCallArgs(sigs, r.Module, r.Exempt)...)

	// Resolved, not t.TempDir()'s own path: macOS's /tmp is a symlink to
	// /private/tmp, so perl's getcwd() (which resolves symlinks) would
	// otherwise report a path that never string-equals $ENV{HOME} or
	// $ENV{TMPDIR} set from the unresolved one, even though all three
	// name the same directory (measured: TestReturnOracleRunsInScratchDir).
	scratch, err := filepath.EvalSymlinks(t.TempDir())
	require.NoError(t, err)
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
			if o.Type == "DualVar" && !r.IgnorePending && pendingAuditMatch(r.Module, name, "return") {
				continue
			}
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
	perltest.WriteModule(t, lib, module, pm)
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
    my ( $cwd, $home, $tmpdir ) = ( Cwd::getcwd(), $ENV{HOME}, $ENV{TMPDIR} );
    if ( $^O eq 'MSWin32' ) {
        # getcwd(), $ENV{HOME} and $ENV{TMPDIR} can each spell the same
        # directory differently here -- an 8.3 short name, a
        # differently-cased drive letter, or backslashes versus forward
        # slashes (measured in CI) -- so canonicalize all three to the
        # same long-name, same-case, same-separator form before comparing.
        require Win32;
        $_ = lc( Win32::GetLongPathName($_) =~ tr{\\}{/}r ) for $cwd, $home, $tmpdir;
    }
    return $cwd eq $home && $home eq $tmpdir ? 1 : 'elsewhere';
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

// TestXSUtilitiesReturnTypes holds the XS utility modules' declared return
// types to perl (perltest.XSUtilities, the same list parse's
// heldtoperl_test.go holds to perl's prototypes and exports), with the
// call cases in testdata/oracle/.
func TestXSUtilitiesReturnTypes(t *testing.T) {
	holdReturnTypes(t, perltest.XSUtilities...)
}

// TestScalarUtilBlessedUnblessedReturnIsDeclared: blessed on an unblessed
// reference returns undef, which Scalar/Util.tpm's declared Str|Undef
// admits; testdata/oracle/Scalar-Util.pl's blessed case includes one,
// {}. Measured on 5.42.0, `Scalar::Util::blessed({})` is undef.
func TestScalarUtilBlessedUnblessedReturnIsDeclared(t *testing.T) {
	holdReturnTypes(t, "Scalar::Util")
}

// TestMIMEBase64InvalidInputReturnIsDeclared: decode_base64 given invalid
// base64 returns the garbage bytes its decoder produces, a Str,
// MIME/Base64.tpm's declared return type for it; testdata/oracle/
// MIME-Base64.pl's decode_base64 case includes one, 'not valid base64
// !!!'. Measured on 5.42.0, `MIME::Base64::decode_base64('not valid
// base64 !!!')` is defined, not undef, and not a death.
func TestMIMEBase64InvalidInputReturnIsDeclared(t *testing.T) {
	holdReturnTypes(t, "MIME::Base64")
}

// TestXSUtilitiesOracleStaysInScratch: Storable's oracle cases (fixture.store,
// store.out, and the rest) write under the oracle's temp directory, never
// the package directory `go test` runs from -- running Storable's return-type
// oracle leaves that directory's entries unchanged.
func TestXSUtilitiesOracleStaysInScratch(t *testing.T) {
	before := dirEntries(t, ".")
	holdReturnTypes(t, "Storable")
	after := dirEntries(t, ".")
	assert.Equal(t, before, after, "Storable's oracle cases wrote outside the test's scratch directory")
}

// dirEntries is the sorted names dir holds, for before/after comparisons
// that must see no new file appear.
func dirEntries(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	names := make([]string, len(entries))
	for i, e := range entries {
		names[i] = e.Name()
	}
	sort.Strings(names)
	return names
}

// TestPurePerlExportersReturnTypes holds each pure-Perl exporter's
// declared return types to perl (perltest.PureExporters; see its
// comment for Symbol's history), called with its oracle's cases.
func TestPurePerlExportersReturnTypes(t *testing.T) {
	holdReturnTypes(t, perltest.PureExporters...)
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
	assert.Equal(t, "2:1", strings.TrimRight(string(out), "\r\n"), "strtol returns two values in list context and one in scalar context")

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

// posixFailureReturnsScript measures, in one perl process, every POSIX
// sub whose documented "Returns undef on failure" (the T_SYSRET typemap,
// or a name lookup that finds nothing) a one-shot process can safely
// induce: a bad fd, a nonexistent path, an unknown name or id, or a
// removed cwd. Each case prints "name\tUNDEF" or "name\tDEF:value".
// setuid and setgid are skipped when run as root, where they succeed.
const posixFailureReturnsScript = `
use v5.42; no warnings;
use POSIX;
use File::Temp qw(tempfile);

my $buf;
my @cases = (
  ['access', sub { POSIX::access("/definitely-does-not-exist-xyz-pvm",0) }],
  ['close', sub { POSIX::close(987654) }],
  ['dup', sub { POSIX::dup(987654) }],
  ['dup2', sub { POSIX::dup2(987654,987655) }],
  ['open', sub { POSIX::open("/definitely-does-not-exist-xyz-pvm/f", &POSIX::O_RDONLY) }],
  ['creat', sub { POSIX::creat("/definitely-does-not-exist-xyz-pvm/f",0) }],
  ['fcntl', sub {
      my ($fh, $fn) = tempfile(); my $fd = fileno($fh); close $fh;
      POSIX::fcntl($fd, &POSIX::F_GETFD, 0);
  }],
  ['lchown', sub { POSIX::lchown(0,0,"/definitely-does-not-exist-xyz-pvm") }],
  ['mkfifo', sub { POSIX::mkfifo("/definitely-does-not-exist-xyz-pvm/f",0600) }],
  ['read', sub { POSIX::read(987654,$buf,1) }],
  ['write', sub { POSIX::write(987654,"x",1) }],
  ['fpathconf', sub { POSIX::fpathconf(987654,0) }],
  ['pathconf', sub { POSIX::pathconf("/definitely-does-not-exist-xyz-pvm",0) }],
  ['sysconf', sub { POSIX::sysconf(987654) }],
  ['setpgid', sub { POSIX::setpgid(999999999,0) }],
  ['tcdrain', sub { POSIX::tcdrain(987654) }],
  ['tcflow', sub { POSIX::tcflow(987654,0) }],
  ['tcflush', sub { POSIX::tcflush(987654,0) }],
  ['tcsendbreak', sub { POSIX::tcsendbreak(987654,0) }],
  ['tcsetpgrp', sub { POSIX::tcsetpgrp(987654,0) }],
  ['getgrgid', sub { POSIX::getgrgid(999999999) }],
  ['getgrnam', sub { POSIX::getgrnam("definitely-nonexistent-group-xyz") }],
  ['getpwnam', sub { POSIX::getpwnam("definitely-nonexistent-user-xyz") }],
  ['getpwuid', sub { POSIX::getpwuid(999999999) }],
);
push @cases, ['setuid', sub { POSIX::setuid(0) }], ['setgid', sub { POSIX::setgid(0) }] unless $> == 0;

for my $c (@cases) {
    my ($name, $code) = @$c;
    my $v = eval { $code->() };
    print "$name\t", (defined $v ? "DEF:$v" : "UNDEF"), ($@ ? " DIED:$@" : ''), "\n";
}

my $base = POSIX::getcwd();
my $tmp = "$base/pvm-oracle-removed-$$";
mkdir $tmp or die "mkdir: $!";
chdir $tmp or die "chdir: $!";
rmdir $tmp or die "rmdir: $!";
my $v = eval { POSIX::getcwd() };
print "getcwd\t", (defined $v ? "DEF:$v" : "UNDEF"), ($@ ? " DIED:$@" : ''), "\n";
chdir $base;
`

// posixScriptLines runs script (in dir, unless dir is "") and returns
// its "name\tresult" lines, split and freed of perl's own \r\n line
// ending so a trailing \r never sticks to the last field of a line --
// measured: Windows's perl prints \r\n, and trimming only the whole
// output's own trailing \n (not \r) leaves one on every line but the
// last. Perl's eval cannot catch a C-level crash, which one of several
// bogus arguments these scripts feed a POSIX sub induces on some
// platforms (measured: Windows's MSVCRT aborts the process outright for
// some of them, rather than setting errno the way glibc does) -- a
// nonzero exit here still means whatever lines printed before that are
// real measurements, so it is logged, not failed, and parsed the same
// way a clean exit's output is. A name the crash cut off before it
// could print is simply not among the lines returned, so this run does
// not check it, same as a name genuinely exempt.
func posixScriptLines(t *testing.T, p perltest.Perl, dir, script string) []string {
	t.Helper()
	cmd := p.Command("-e", script)
	if dir != "" {
		cmd.Dir = dir
	}
	out, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		require.True(t, errors.As(err, &exitErr), "could not even run the script: %v", err)
		t.Logf("script exited %v before finishing; using what it printed up to there", err)
	}
	text := strings.TrimRight(string(out), "\r\n")
	if text == "" {
		return nil
	}
	var lines []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimRight(line, "\r")
		// A crash mid-print, or a stray blank line a Windows abort
		// handler's own output can leave behind, is not a record to
		// parse -- only a real "name\tresult" line is.
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}

// TestPOSIXFailureReturnsAdmitUndef: every sub the T_SYSRET typemap, or
// a name lookup finding nothing, documents as "Returns undef on
// failure" (scratchpad/audit-A.md F1) really returns undef when induced
// safely in a one-shot process, and POSIX.tpm's declared return type
// admits it.
func TestPOSIXFailureReturnsAdmitUndef(t *testing.T) {
	p := perltest.Perl542(t)
	sigs := posixSignatures(t)
	lines := posixScriptLines(t, p, "", posixFailureReturnsScript)
	require.NotEmpty(t, lines)
	for _, line := range lines {
		name, result, ok := strings.Cut(line, "\t")
		require.True(t, ok, "unparseable line: %q", line)
		if !assert.True(t, strings.HasPrefix(result, "UNDEF"), "POSIX::%s: measured %s, not undef", name, result) {
			continue
		}
		cands, ok := sigs[name]
		if !assert.True(t, ok, "POSIX::%s is not declared", name) {
			continue
		}
		admits := slices.ContainsFunc(cands, func(s types.Signature) bool { return types.IsSubtype(types.Undef, s.Returns) })
		assert.True(t, admits, "POSIX::%s measured undef on failure; its declared return type does not admit Undef", name)
	}
}

// posixMathReturnsScript measures, for each POSIX math sub CORE.tpm's
// own `exp` precedent already admits Inf for, a finite argument that
// overflows to Inf or is outside the real domain and so is NaN
// (scratchpad/audit-A.md F2). Each case prints "name\tINF", "name\tNAN"
// or "name\tNUM".
const posixMathReturnsScript = `
use v5.42; no warnings;
use POSIX;

sub classify {
    my $v = shift;
    return 'NAN' if $v != $v;
    return 'INF' if $v == 9**9**9 || $v == -9**9**9;
    return 'NUM';
}

my @cases = (
  ['exp', sub { POSIX::exp(1000) }],
  ['exp2', sub { POSIX::exp2(2000) }],
  ['expm1', sub { POSIX::expm1(1000) }],
  ['cosh', sub { POSIX::cosh(1000) }],
  ['sinh', sub { POSIX::sinh(1000) }],
  ['ldexp', sub { POSIX::ldexp(1,2000) }],
  ['scalbn', sub { POSIX::scalbn(1,5000) }],
  ['hypot', sub { POSIX::hypot(1.7e308,1.7e308) }],
  ['fdim', sub { POSIX::fdim(1.7e308,-1.7e308) }],
  ['fma', sub { POSIX::fma(1e308,10,0) }],
  ['atanh', sub { POSIX::atanh(1) }],
  ['log10', sub { POSIX::log10(0) }],
  ['log10', sub { POSIX::log10(-1) }],
  ['log2', sub { POSIX::log2(0) }],
  ['log1p', sub { POSIX::log1p(-1) }],
  ['logb', sub { POSIX::logb(0) }],
  ['y0', sub { POSIX::y0(0) }],
  ['y1', sub { POSIX::y1(0) }],
  ['yn', sub { POSIX::yn(1,0) }],
  ['lgamma', sub { POSIX::lgamma(0) }],
  ['tgamma', sub { POSIX::tgamma(0) }],
  ['tgamma', sub { POSIX::tgamma(-1) }],
  ['pow', sub { POSIX::pow(10,400) }],
  ['pow', sub { POSIX::pow(-1,.5) }],
  ['acos', sub { POSIX::acos(2) }],
  ['asin', sub { POSIX::asin(2) }],
  ['acosh', sub { POSIX::acosh(0) }],
  ['fmod', sub { POSIX::fmod(1,0) }],
  ['remainder', sub { POSIX::remainder(1,0) }],
  ['strtod', sub { scalar POSIX::strtod("inf") }],
  ['strtod', sub { scalar POSIX::strtod("nan") }],
  ['strtold', sub { scalar POSIX::strtold("inf") }],
  ['strtold', sub { scalar POSIX::strtold("nan") }],
  ['remquo', sub { my ($q) = POSIX::remquo(1,0); $q }],
);

for my $c (@cases) {
    my ($name, $code) = @$c;
    my $v = $code->();
    print "$name\t", classify($v), "\n";
}
`

// TestPOSIXMathReturnsAdmitInfNaN: every math sub scratchpad/audit-A.md's
// F2 names overflows to Inf or leaves its real domain to NaN for a
// finite argument, measured directly, and POSIX.tpm's declared return
// type admits it -- CORE.tpm's own `exp` precedent (Num|Inf) already
// does for the language's builtin.
func TestPOSIXMathReturnsAdmitInfNaN(t *testing.T) {
	p := perltest.Perl542(t)
	wants := map[string]types.Type{"INF": types.Inf, "NAN": types.NaN}
	sigs := posixSignatures(t)
	lines := posixScriptLines(t, p, "", posixMathReturnsScript)
	require.NotEmpty(t, lines)
	for _, line := range lines {
		name, class, ok := strings.Cut(line, "\t")
		require.True(t, ok, "unparseable line: %q", line)
		want, known := wants[class]
		if !assert.True(t, known, "POSIX::%s: unexpected class %q", name, class) {
			continue
		}
		cands, ok := sigs[name]
		if !assert.True(t, ok, "POSIX::%s is not declared", name) {
			continue
		}
		admits := slices.ContainsFunc(cands, func(s types.Signature) bool {
			if s.Returns == types.None {
				return false
			}
			r := s.Returns
			if s.ReturnElement != types.Unknown {
				r = s.ReturnElement
			}
			return types.IsSubtype(want, r)
		})
		assert.True(t, admits, "POSIX::%s measured %s; its declared return type does not admit it", name, class)
	}
}

// posixClassifierScript calls each classification and NaN-payload sub
// scratchpad/audit-A.md's F3 names with a NaN or an Inf, the values
// these subs exist to classify, and prints "name\tvalue" or
// "name\tvalue DIED:...".
const posixClassifierScript = `
use v5.42; no warnings;
use POSIX;
my $nan = POSIX::nan();
my $inf = 9**9**9;
my @cases = (
  ['fpclassify', sub { POSIX::fpclassify($inf) }],
  ['ilogb', sub { POSIX::ilogb($inf) }],
  ['isfinite', sub { POSIX::isfinite($nan) }],
  ['isgreater', sub { POSIX::isgreater($nan, 1) }],
  ['isgreaterequal', sub { POSIX::isgreaterequal($nan, 1) }],
  ['isinf', sub { POSIX::isinf($inf) }],
  ['isless', sub { POSIX::isless($nan, 1) }],
  ['islessequal', sub { POSIX::islessequal($nan, 1) }],
  ['islessgreater', sub { POSIX::islessgreater($nan, 1) }],
  ['isnan', sub { POSIX::isnan($nan) }],
  ['isnormal', sub { POSIX::isnormal($inf) }],
  ['issignaling', sub { POSIX::issignaling($nan) }],
  ['isunordered', sub { POSIX::isunordered($nan, 1) }],
  ['signbit', sub { POSIX::signbit(-$inf) }],
  ['getpayload', sub { POSIX::getpayload(POSIX::nan(7)) }],
);
for my $c (@cases) {
    my ($name, $code) = @$c;
    my $v = eval { $code->() };
    print "$name\t", (defined $v ? $v : 'undef'), ($@ ? " DIED:$@" : ''), "\n";
}
`

// TestPOSIXClassifiersTakeInfNaN: fpclassify, ilogb, isfinite, isinf,
// isnan, isnormal, issignaling, signbit, getpayload and the paired
// is-comparison subs exist to classify NaN and Inf, measured taking
// them without dying, and POSIX.tpm's declared parameter type admits
// both (scratchpad/audit-A.md F3) -- Num alone excludes them, so a
// coercion-mismatch check would see a lossy coercion although nothing
// is lost: these subs coerce their argument to a C double, which
// preserves NaN and Inf exactly.
func TestPOSIXClassifiersTakeInfNaN(t *testing.T) {
	p := perltest.Perl542(t)
	sigs := posixSignatures(t)
	lines := posixScriptLines(t, p, "", posixClassifierScript)
	require.NotEmpty(t, lines)
	for _, line := range lines {
		name, result, ok := strings.Cut(line, "\t")
		require.True(t, ok, "unparseable line: %q", line)
		if reason, exempt := platformExempt[runtime.GOOS]["POSIX::"+name]; exempt {
			assert.Contains(t, result, "DIED", "POSIX::%s is exempt (%s) because it dies taking NaN/Inf, but it did not -- stale exemption", name, reason)
			continue
		}
		if !assert.NotContains(t, result, "DIED", "POSIX::%s died taking NaN/Inf: %s", name, result) {
			continue
		}
		cands, ok := sigs[name]
		if !assert.True(t, ok, "POSIX::%s is not declared", name) {
			continue
		}
		for _, s := range cands {
			for _, param := range s.Params {
				assert.True(t, types.IsSubtype(types.NaN, param.Type), "POSIX::%s's $%s does not admit NaN", name, param.Name)
				assert.True(t, types.IsSubtype(types.Inf, param.Type), "POSIX::%s's $%s does not admit Inf", name, param.Name)
			}
		}
	}
}

// posixArityScript measures, for nan, open, strtol, strtoul, mblen,
// mbtowc, asctime, mktime and strftime, an arity one below their
// minimum, each arity usage.txt's optional parameters admit, and one
// above their maximum, printing "name:n\tOK" or "name:n\tDIE".
const posixArityScript = `
use v5.42; no warnings;
use POSIX;

sub tryn {
    my ($label, $code) = @_;
    my $ok = eval { $code->(); 1 };
    print "$label\t", ($ok ? "OK" : "DIE"), "\n";
}

tryn('nan:0', sub { POSIX::nan() });
tryn('nan:1', sub { POSIX::nan(5) });
tryn('nan:2', sub { POSIX::nan(5,6) });

tryn('open:0', sub { POSIX::open() });
tryn('open:1', sub { POSIX::open("f") });
tryn('open:2', sub { POSIX::open("f", &POSIX::O_RDONLY) });
tryn('open:3', sub { POSIX::open("f", &POSIX::O_RDONLY, 0666) });
tryn('open:4', sub { POSIX::open("f", &POSIX::O_RDONLY, 0666, 1) });

tryn('strtol:0', sub { POSIX::strtol() });
tryn('strtol:1', sub { POSIX::strtol("42") });
tryn('strtol:2', sub { POSIX::strtol("42", 10) });
tryn('strtol:3', sub { POSIX::strtol("42", 10, 1) });

tryn('strtoul:0', sub { POSIX::strtoul() });
tryn('strtoul:1', sub { POSIX::strtoul("42") });
tryn('strtoul:2', sub { POSIX::strtoul("42", 10) });
tryn('strtoul:3', sub { POSIX::strtoul("42", 10, 1) });

tryn('mblen:0', sub { POSIX::mblen() });
tryn('mblen:1', sub { POSIX::mblen("a") });
tryn('mblen:2', sub { POSIX::mblen("a", 1) });
tryn('mblen:3', sub { POSIX::mblen("a", 1, 2) });
tryn('mblen:undef', sub { POSIX::mblen(undef) });

tryn('mbtowc:1', sub { my $x; POSIX::mbtowc($x) });
tryn('mbtowc:2', sub { my $x; POSIX::mbtowc($x, "a") });
tryn('mbtowc:3', sub { my $x; POSIX::mbtowc($x, "a", 1) });
tryn('mbtowc:4', sub { my $x; POSIX::mbtowc($x, "a", 1, 2) });
tryn('mbtowc:undef', sub { my $x; POSIX::mbtowc($x, undef) });

tryn('perror:0', sub { POSIX::perror() });
tryn('perror:1', sub { POSIX::perror("a") });
tryn('perror:2', sub { POSIX::perror("a", "b") });

tryn('asctime:5', sub { POSIX::asctime(0,0,0,1,0) });
tryn('asctime:6', sub { POSIX::asctime(0,0,0,1,0,70) });
tryn('asctime:9', sub { POSIX::asctime(0,0,0,1,0,70,0,0,-1) });
tryn('asctime:10', sub { POSIX::asctime(0,0,0,1,0,70,0,0,-1,1) });

tryn('mktime:5', sub { POSIX::mktime(0,0,0,1,0) });
tryn('mktime:6', sub { POSIX::mktime(0,0,0,1,0,70) });
tryn('mktime:9', sub { POSIX::mktime(0,0,0,1,0,70,0,0,-1) });
tryn('mktime:10', sub { POSIX::mktime(0,0,0,1,0,70,0,0,-1,1) });

tryn('strftime:6', sub { POSIX::strftime("%Y",0,0,0,1,0) });
tryn('strftime:7', sub { POSIX::strftime("%Y",0,0,0,1,0,70) });
tryn('strftime:10', sub { POSIX::strftime("%Y",0,0,0,1,0,70,0,0,-1) });
tryn('strftime:11', sub { POSIX::strftime("%Y",0,0,0,1,0,70,0,0,-1,1) });
`

// TestPOSIXOptionalParametersMatchPerl: nan, open, strtol, strtoul,
// mblen, mbtowc, asctime, mktime and strftime each take an arity range
// perl's own "Usage: ..." croak states (usage.txt), not the single
// required arity POSIX.tpm declared. Each arity perl accepts is
// measured directly and admitted by the declared signature; the one
// just past each end, perl's own negative scenario, still dies in perl
// and is still refused by it (scratchpad/audit-A.md F4).
func TestPOSIXOptionalParametersMatchPerl(t *testing.T) {
	p := perltest.Perl542(t)
	sigs := posixSignatures(t)
	lines := posixScriptLines(t, p, "", posixArityScript)
	require.NotEmpty(t, lines)
	for _, line := range lines {
		label, result, ok := strings.Cut(line, "\t")
		require.True(t, ok, "unparseable line: %q", line)
		name, narg, ok := strings.Cut(label, ":")
		require.True(t, ok, "unparseable label: %q", label)

		cands, ok := sigs[name]
		if !assert.True(t, ok, "POSIX::%s is not declared", name) {
			continue
		}
		if narg == "undef" {
			// mblen and mbtowc: undef in their Str slot is a documented
			// distinct meaning, not an arity case; perl must not die.
			assert.Equal(t, "OK", result, "POSIX::%s(undef, ...) died", name)
			continue
		}
		n, convErr := strconv.Atoi(narg)
		require.NoError(t, convErr)

		dies := result == "DIE"
		accepts := types.Accepts(cands, n)
		assert.NotEqual(t, dies, accepts,
			"POSIX::%s with %d argument(s): perl %s, declared signature accepts=%v", name, n, result, accepts)
	}

	// asctime, mktime and strftime also take their 6 (7 with strftime's
	// format) through 9 (10) required arguments flattened from a single
	// localtime call -- the idiom strftime($fmt, localtime) must keep
	// working -- and still refuse too few. A loader, even one resolving
	// nothing, is what has the parser track a name `use POSIX qw(...)`
	// imports bare, as checkFile's real one does; loweredDiags's
	// loader-less parse.Parse does not, and so checks nothing here.
	assert.Empty(t, posixDeclaredDiags(t, "use POSIX qw(asctime mktime strftime);\nmy $a = asctime(localtime);\nmy $m = mktime(localtime);\nmy $s = strftime(\"%Y\", localtime);\n"))
	wantArity := func(src string) {
		diags := posixDeclaredDiags(t, src)
		require.Len(t, diags, 1, "%q: diagnostics %v", src, diags)
		assert.Equal(t, infer.CodeArityMismatch, diags[0].Code, diags[0].Message)
	}
	wantArity("use POSIX qw(asctime);\nmy $a = asctime(0,0,0,1,0);\n")
	wantArity("use POSIX qw(strftime);\nmy $s = strftime(\"%Y\",0,0,0,1,0);\n")
}

// posixDeclaredDiags analyses src, resolving `use POSIX` and its sibling
// declared libraries through a loader (even one resolving nothing), the
// way checkFile's real one does; loweredDiags's loader-less parse.Parse
// does not track a bare imported name the same way and so checks nothing
// for it.
func posixDeclaredDiags(t *testing.T, src string) []infer.Diagnostic {
	t.Helper()
	load := func(string) ([]byte, bool) { return nil, false }
	root := parse.ParseWithLoader([]byte(src), load)
	return infer.AnalyzeParsed(root, nil, infer.Options{}).Diagnostics
}

// posixNarrowParamsScript measures, for isatty, chdir, setpayload and
// utime, an argument form scratchpad/audit-A.md's F5 says POSIX.tpm
// declared too narrow for: a glob, a bareword filehandle name and an fd
// for isatty; a dirhandle for chdir; a fresh non-Num lexical for
// setpayload, which the call writes a NaN into, so it is a true output
// lvalue, not a coerced argument; undef, meaning "now", for utime.
const posixNarrowParamsScript = `
use v5.42; no warnings;
use POSIX;

sub tryn {
    my ($label, $code) = @_;
    my $ok = eval { $code->(); 1 };
    print "$label\t", ($ok ? "OK" : "DIE:$@"), "\n";
}

tryn('isatty:glob', sub { POSIX::isatty(\*STDIN) });
tryn('isatty:fd', sub { POSIX::isatty(0) });
tryn('isatty:str', sub { POSIX::isatty("STDIN") });

my $dh = POSIX::opendir('.');
tryn('chdir:handle', sub { POSIX::chdir($dh) });

my $x;
my @list = POSIX::setpayload($x, 7);
print "setpayload:list\t", (scalar(@list) == 0 ? "OK" : "DIE:not empty"), "\n";
print "setpayload:wrote\t", (POSIX::isnan($x) ? "OK" : "DIE:did not write NaN"), "\n";

open my $seed, '>', 'posix-f5-scratch.txt' or die; close $seed;
tryn('utime:undef', sub { POSIX::utime("posix-f5-scratch.txt", undef, undef) });
`

// posixNarrowParamsExempt holds, per runtime.GOOS, posixNarrowParamsScript
// labels that die there instead of measuring OK -- a real platform
// difference in the underlying call, not a declaration bug.
var posixNarrowParamsExempt = map[string]map[string]string{
	"windows": {
		"chdir:handle": "Windows's perl has no fchdir -- measured in CI job 39508730008: \"The fchdir function is unimplemented\"",
	},
}

// TestPOSIXParametersAdmitWhatPerlAccepts: isatty, chdir, setpayload,
// setpayloadsig and utime each declared a parameter type narrower than
// what perl and the POD accept (scratchpad/audit-A.md F5). isatty and
// chdir coerce to or preserve a filehandle, not its stringification;
// setpayload's and setpayloadsig's $nv is an output lvalue the call
// writes, not an argument it coerces; utime's undef means "now" and is
// preserved, not coerced to 0. Each case is measured directly, and
// POSIX.tpm's declared parameter type admits it.
func TestPOSIXParametersAdmitWhatPerlAccepts(t *testing.T) {
	p := perltest.Perl542(t)
	scratch := t.TempDir()
	lines := posixScriptLines(t, p, scratch, posixNarrowParamsScript)
	require.NotEmpty(t, lines)
	for _, line := range lines {
		label, result, ok := strings.Cut(line, "\t")
		require.True(t, ok, "unparseable line: %q", line)
		if reason, exempt := posixNarrowParamsExempt[runtime.GOOS][label]; exempt {
			assert.NotEqual(t, "OK", result, "%s: exempt (%s) because it dies, but it did not -- stale exemption", label, reason)
			continue
		}
		assert.Equal(t, "OK", result, "%s: %s", label, result)
	}

	sigs := posixSignatures(t)
	isatty := sigs["isatty"]
	require.NotEmpty(t, isatty)
	for _, s := range isatty {
		assert.True(t, types.IsSubtype(types.FileHandle, s.Params[0].Type), "isatty's $%s does not admit a filehandle", s.Params[0].Name)
		assert.True(t, types.IsSubtype(types.Str, s.Params[0].Type), "isatty's $%s does not admit a bareword name", s.Params[0].Name)
	}

	chdir := sigs["chdir"]
	require.NotEmpty(t, chdir)
	for _, s := range chdir {
		assert.True(t, types.IsSubtype(types.FileHandle, s.Params[0].Type), "chdir's $%s does not admit a dirhandle", s.Params[0].Name)
	}

	for _, name := range []string{"setpayload", "setpayloadsig"} {
		cands, ok := sigs[name]
		if !assert.True(t, ok, "POSIX::%s is not declared", name) {
			continue
		}
		for _, s := range cands {
			assert.Equal(t, types.Any, s.Params[0].Type, "%s's $%s is not Any, an output lvalue the call writes rather than an argument it coerces", name, s.Params[0].Name)
			assert.Equal(t, types.Void, s.Returns, "%s does not declare Void: the XS returns the empty list", name)
		}
	}

	utime := sigs["utime"]
	require.NotEmpty(t, utime)
	for _, s := range utime {
		for _, param := range s.Params[1:] {
			assert.True(t, types.IsSubtype(types.Undef, param.Type), "utime's $%s does not admit undef, which means \"now\"", param.Name)
		}
	}
}

// TestXSLibrariesReturnTypes holds each XS library's declared return
// types to perl (perltest.XSLibraries), called with its oracle's cases.
func TestXSLibrariesReturnTypes(t *testing.T) {
	holdReturnTypes(t, perltest.XSLibraries...)
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

// ioSocketIPSocketReexports are 161 of the 170 Socket default names
// IO/Socket/IP.tpm marks :export (measured on 5.42.0): they live only in
// IO::Socket's own stash, from its file-scope `use Socket;`, not
// IO::Socket::IP's own, so the return-type oracle's sub-form call --
// which looks up Module::name directly -- cannot reach any of them
// here, though IO::Socket::IP->can(name) finds them by inheritance.
// Socket.tpm's own oracle (TestConstantExportersReturnTypes) already
// holds their return types. The other nine -- AF_INET, sockaddr_family,
// SO_REUSEADDR, SO_REUSEPORT, SO_BROADCAST, SO_ERROR, SOCK_DGRAM,
// SOCK_STREAM and SOL_SOCKET -- are also imported directly into
// IO::Socket::IP's own package, by its own file-scope `use Socket ...
// qw(...)` (measured), so the oracle holds them directly instead
// (oracle/IO-Socket-IP.pl), not here.
var ioSocketIPSocketReexports = []string{
	"AF_802", "AF_AAL", "AF_APPLETALK", "AF_CCITT", "AF_CHAOS", "AF_CTF", "AF_DATAKIT", "AF_DECnet",
	"AF_DLI", "AF_ECMA", "AF_GOSIP", "AF_HYLINK", "AF_IMPLINK", "AF_INET6", "AF_ISO",
	"AF_KEY", "AF_LAST", "AF_LAT", "AF_LINK", "AF_MAX", "AF_NBS", "AF_NIT", "AF_NS",
	"AF_OSI", "AF_OSINET", "AF_PUP", "AF_ROUTE", "AF_SNA", "AF_UNIX", "AF_UNSPEC", "AF_USER",
	"AF_WAN", "AF_X25", "INADDR_ANY", "INADDR_BROADCAST", "INADDR_LOOPBACK", "INADDR_NONE", "IOV_MAX", "IP_HDRINCL",
	"IP_OPTIONS", "IP_RECVOPTS", "IP_RECVRETOPTS", "IP_RETOPTS", "IP_TOS", "IP_TTL", "MSG_BCAST", "MSG_BTAG",
	"MSG_CTLFLAGS", "MSG_CTLIGNORE", "MSG_CTRUNC", "MSG_DONTROUTE", "MSG_DONTWAIT", "MSG_EOF", "MSG_EOR", "MSG_ERRQUEUE",
	"MSG_ETAG", "MSG_FASTOPEN", "MSG_FIN", "MSG_MAXIOVLEN", "MSG_MCAST", "MSG_NOSIGNAL", "MSG_OOB", "MSG_PEEK",
	"MSG_PROXY", "MSG_RST", "MSG_SYN", "MSG_TRUNC", "MSG_URG", "MSG_WAITALL", "MSG_WIRE", "PF_802",
	"PF_AAL", "PF_APPLETALK", "PF_CCITT", "PF_CHAOS", "PF_CTF", "PF_DATAKIT", "PF_DECnet", "PF_DLI",
	"PF_ECMA", "PF_GOSIP", "PF_HYLINK", "PF_IMPLINK", "PF_INET", "PF_INET6", "PF_ISO", "PF_KEY",
	"PF_LAST", "PF_LAT", "PF_LINK", "PF_MAX", "PF_NBS", "PF_NIT", "PF_NS", "PF_OSI",
	"PF_OSINET", "PF_PUP", "PF_ROUTE", "PF_SNA", "PF_UNIX", "PF_UNSPEC", "PF_USER", "PF_WAN",
	"PF_X25", "SCM_CONNECT", "SCM_CREDENTIALS", "SCM_CREDS", "SCM_RIGHTS", "SCM_TIMESTAMP", "SHUT_RD", "SHUT_RDWR",
	"SHUT_WR", "SOCK_RAW", "SOCK_RDM", "SOCK_SEQPACKET", "SOMAXCONN",
	"SO_ACCEPTCONN", "SO_ATTACH_FILTER", "SO_BACKLOG", "SO_CHAMELEON", "SO_DEBUG", "SO_DETACH_FILTER", "SO_DGRAM_ERRIND",
	"SO_DOMAIN", "SO_DONTLINGER", "SO_DONTROUTE", "SO_FAMILY", "SO_KEEPALIVE", "SO_LINGER", "SO_OOBINLINE",
	"SO_PASSCRED", "SO_PASSIFNAME", "SO_PEERCRED", "SO_PROTOCOL", "SO_PROTOTYPE", "SO_RCVBUF", "SO_RCVLOWAT", "SO_RCVTIMEO",
	"SO_SECURITY_AUTHENTICATION", "SO_SECURITY_ENCRYPTION_NETWORK", "SO_SECURITY_ENCRYPTION_TRANSPORT", "SO_SNDBUF", "SO_SNDLOWAT", "SO_SNDTIMEO",
	"SO_STATE", "SO_TYPE", "SO_USELOOPBACK", "SO_XOPEN", "SO_XSE", "UIO_MAXIOV", "inet_aton", "inet_ntoa",
	"pack_sockaddr_in", "pack_sockaddr_in6", "pack_sockaddr_un", "sockaddr_in", "sockaddr_in6", "sockaddr_un", "unpack_sockaddr_in",
	"unpack_sockaddr_in6", "unpack_sockaddr_un",

	// The :export_ok names IO/Socket/IP.tpm mirrors from Socket.tpm, the
	// same way and for the same reason (IO/Socket.tpm's own comment):
	// each lives only in IO::Socket's or Socket's own stash, not
	// IO::Socket::IP's, so the sub-form oracle cannot reach it here
	// either, though IO::Socket::IP->can(name) finds it by inheritance.
	"AI_ADDRCONFIG", "AI_ALL", "AI_CANONIDN", "AI_CANONNAME", "AI_IDN", "AI_IDN_ALLOW_UNASSIGNED",
	"AI_IDN_USE_STD3_ASCII_RULES", "AI_NUMERICHOST", "AI_NUMERICSERV", "AI_PASSIVE", "AI_V4MAPPED",
	"CR", "CRLF", "LF", "EAI_ADDRFAMILY", "EAI_AGAIN", "EAI_BADFLAGS", "EAI_BADHINTS", "EAI_FAIL",
	"EAI_FAMILY", "EAI_NODATA", "EAI_NONAME", "EAI_PROTOCOL", "EAI_SERVICE", "EAI_SOCKTYPE", "EAI_SYSTEM",
	"IN6ADDR_ANY", "IN6ADDR_LOOPBACK", "IPPROTO_AH", "IPPROTO_ESP", "IPPROTO_GRE", "IPPROTO_ICMP",
	"IPPROTO_ICMPV6", "IPPROTO_IGMP", "IPPROTO_IP", "IPPROTO_IPV6", "IPPROTO_RAW", "IPPROTO_SCTP",
	"IPPROTO_TCP", "IPPROTO_UDP", "IPTOS_LOWDELAY", "IPTOS_MINCOST", "IPTOS_RELIABILITY", "IPTOS_THROUGHPUT",
	"IPV6_ADDRFROM", "IPV6_ADD_MEMBERSHIP", "IPV6_DROP_MEMBERSHIP", "IPV6_JOIN_GROUP", "IPV6_LEAVE_GROUP",
	"IPV6_MTU", "IPV6_MTU_DISCOVER", "IPV6_MULTICAST_HOPS", "IPV6_MULTICAST_IF", "IPV6_MULTICAST_LOOP",
	"IPV6_RECVERR", "IPV6_ROUTER_ALERT", "IPV6_TCLASS", "IPV6_UNICAST_HOPS", "IPV6_V6ONLY",
	"IP_ADD_MEMBERSHIP", "IP_ADD_SOURCE_MEMBERSHIP", "IP_BIND_ADDRESS_NO_PORT", "IP_DROP_MEMBERSHIP",
	"IP_DROP_SOURCE_MEMBERSHIP", "IP_FREEBIND", "IP_MTU", "IP_MTU_DISCOVER", "IP_MULTICAST_ALL",
	"IP_MULTICAST_IF", "IP_MULTICAST_LOOP", "IP_MULTICAST_TTL", "IP_NODEFRAG", "IP_PMTUDISC_DO",
	"IP_PMTUDISC_DONT", "IP_PMTUDISC_PROBE", "IP_PMTUDISC_WANT", "IP_RECVERR", "IP_TRANSPARENT",
	"NI_DGRAM", "NI_IDN", "NI_IDN_ALLOW_UNASSIGNED", "NI_IDN_USE_STD3_ASCII_RULES", "NI_NAMEREQD",
	"NI_NOFQDN", "NI_NUMERICHOST", "NI_NUMERICSERV", "NIx_NOHOST", "NIx_NOSERV",
	"SOCK_CLOEXEC", "SOCK_NONBLOCK", "SO_INCOMING_CPU", "SO_INCOMING_NAPI_ID", "SO_LOCK_FILTER",
	"SO_RCVBUFFORCE", "SO_SNDBUFFORCE",
	"TCP_CC_INFO", "TCP_CONGESTION", "TCP_CONNECTIONTIMEOUT", "TCP_CORK", "TCP_DEFER_ACCEPT",
	"TCP_FASTOPEN", "TCP_FASTOPEN_CONNECT", "TCP_FASTOPEN_KEY", "TCP_FASTOPEN_NO_COOKIE", "TCP_INFO",
	"TCP_INIT_CWND", "TCP_KEEPALIVE", "TCP_KEEPCNT", "TCP_KEEPIDLE", "TCP_KEEPINTVL", "TCP_LINGER2",
	"TCP_MAXRT", "TCP_MAXSEG", "TCP_MD5SIG", "TCP_MD5SIG_EXT", "TCP_NODELAY", "TCP_NOOPT", "TCP_NOPUSH",
	"TCP_NOTSENT_LOWAT", "TCP_QUICKACK", "TCP_SACK_ENABLE", "TCP_STDURG", "TCP_SYNCNT", "TCP_TIMESTAMP",
	"TCP_TX_DELAY", "TCP_USER_TIMEOUT", "TCP_WINDOW_CLAMP",
	"getaddrinfo", "getnameinfo", "inet_ntop", "inet_pton",
	"pack_ip_mreq", "pack_ip_mreq_source", "pack_ipv6_mreq",
	"unpack_ip_mreq", "unpack_ip_mreq_source", "unpack_ipv6_mreq",
}

// ooExempt extends oracleExempt with the OO batch's own exemptions:
// Socket's re-exports under IO::Socket::IP that its own stash cannot
// reach, and the subs that touch the network or a real terminal.
// IO::Select's and HTTP::Tiny's own instance methods are not here: the
// return-type oracle calls each of those on a constructed instance
// (INSTANCE in oracle/IO-Select.pl and oracle/HTTP-Tiny.pl), not the
// bare class name, so they need no exemption.
func ooExempt() map[string]string {
	m := withPlatformExempt(oracleExempt)
	for _, name := range ioSocketIPSocketReexports {
		m["IO::Socket::IP::"+name] = "lives in IO::Socket's or Socket's own stash, not IO::Socket::IP's; the sub-form oracle cannot reach it there (measured), and that module's own oracle already holds its return type"
	}
	for _, name := range []string{"connect", "sockhost", "sockport", "peerhost", "peerport"} {
		m["IO::Socket::IP::"+name] = "opens or reads a real network connection"
	}
	m["IO::Socket::IP::accept"] = "blocks accepting a real connection on a socket, as IO::Socket's own accept does"
	m["IO::Socket::IP::peeraddr"] = "unpacks peername's address directly, dying \"Bad arg length\" when peername is undef (measured, unconnected)"
	for _, name := range []string{"get", "head", "put", "post", "patch", "delete", "post_form", "mirror", "request"} {
		m["HTTP::Tiny::"+name] = "performs a real HTTP request over the network"
	}
	m["Term::ReadLine::readline"] = "blocks reading a line from a real terminal"
	return m
}

// TestOOModulesReturnTypes holds each OO module's declared return types
// to perl, called with its oracle's cases, under ooExempt's extended
// exemptions.
func TestOOModulesReturnTypes(t *testing.T) {
	p := perltest.Perl542(t)
	for _, module := range perltest.OOModules {
		rel := strings.ReplaceAll(module, "::", "/")
		tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", filepath.FromSlash(rel)+".tpm"))
		if err != nil {
			t.Errorf("%s: no declaration: %v", module, err)
			continue
		}
		cases, err := filepath.Abs(filepath.Join("testdata", "oracle", strings.ReplaceAll(module, "::", "-")+".pl"))
		require.NoError(t, err)
		exempt := exemptForModule(t, p, module, tpm, ooExempt())
		for _, f := range returnTypeFailures(t, p, oracleRun{Module: module, TPM: tpm, Cases: cases, Exempt: exempt}) {
			t.Error(f)
		}
	}
}

// TestOOModulesNetworkSubsExempt: HTTP::Tiny's and IO::Socket::IP's
// network-touching subs are each exempt, with a reason, and have no
// call case in their oracle file, so TestOOModulesReturnTypes never
// calls them.
func TestOOModulesNetworkSubsExempt(t *testing.T) {
	exempt := ooExempt()
	networked := map[string][]string{
		"HTTP::Tiny":     {"get", "head", "put", "post", "patch", "delete", "post_form", "mirror", "request"},
		"IO::Socket::IP": {"connect", "sockhost", "sockport", "peerhost", "peerport", "accept", "peeraddr"},
	}
	for module, names := range networked {
		cases, err := os.ReadFile(filepath.Join("testdata", "oracle", strings.ReplaceAll(module, "::", "-")+".pl"))
		require.NoError(t, err)
		for _, name := range names {
			reason, ok := exempt[module+"::"+name]
			assert.True(t, ok, "%s::%s is not exempt", module, name)
			assert.NotEmpty(t, reason, "%s::%s is exempt with no reason", module, name)
			assert.NotContains(t, string(cases), name+" =>", "%s::%s has a call case despite being exempt", module, name)
		}
	}
}

// TestTermReadLineNeverCalled: Term::ReadLine's readline is exempt, with
// a reason, and has no call case in Term-ReadLine.pl, since it blocks
// reading a line from a real terminal.
func TestTermReadLineNeverCalled(t *testing.T) {
	exempt := ooExempt()
	reason, ok := exempt["Term::ReadLine::readline"]
	assert.True(t, ok, "Term::ReadLine::readline is not exempt")
	assert.NotEmpty(t, reason)

	cases, err := os.ReadFile(filepath.Join("testdata", "oracle", "Term-ReadLine.pl"))
	require.NoError(t, err)
	assert.NotContains(t, string(cases), "readline =>", "Term::ReadLine::readline has a call case despite being exempt")
}

// TestVariableExportersReturnTypes holds each variable exporter's
// declared return types to perl (perltest.VariableExporters), called
// with its oracle's cases.
func TestVariableExportersReturnTypes(t *testing.T) {
	holdReturnTypes(t, perltest.VariableExporters...)
}

// TestTestMorePlanReturnType holds Test::More::plan's declared return
// type to perl in its own oracleRun (Review Findings, finding 2): plan
// collides with done_testing in the batch's combined run (oracleExempt),
// so it is called alone here, where nothing else has set a plan first,
// with the one case oracle/Test-More-plan.pl gives it.
func TestTestMorePlanReturnType(t *testing.T) {
	p := perltest.Perl542(t)
	tpm := []byte("package Test::More;\nsub plan (List @args) Int;\n")
	cases, err := filepath.Abs(filepath.Join("testdata", "oracle", "Test-More-plan.pl"))
	require.NoError(t, err)
	for _, f := range returnTypeFailures(t, p, oracleRun{Module: "Test::More", TPM: tpm, Cases: cases}) {
		t.Error(f)
	}
}

// declaredVariables reads module's .tpm variables, failing t on any
// declaration error.
func declaredVariables(t *testing.T, module string) map[string]types.Param {
	t.Helper()
	rel := strings.ReplaceAll(module, "::", "/")
	tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", filepath.FromSlash(rel)+".tpm"))
	require.NoError(t, err)
	vars, errs := parse.DeclaredVariables(tpm)
	require.Empty(t, errs)
	return vars
}

// TestConfigValuesFitDeclaredType: ruling 3 (perigrin, 2026-10-16).
// %Config is declared Hash[Str|Undef], and every value perl's %Config
// holds on 5.42.0 is a member of that element type -- some of them
// undef, which Hash[Str] alone would exclude. Exactly how many are
// undef depends on which build of 5.42.0 is running (measured: pvm's
// own reference build has 172, CI's actions-setup-perl build has 174,
// job 39489951290's Linux log), so that exact count is pinned only
// against internal/perltest.Perl.IsReferenceBuild.
func TestConfigValuesFitDeclaredType(t *testing.T) {
	want := declaredVariables(t, "Config")["%Config"]
	require.Equal(t, types.Hash, want.Type, "%%Config is not declared Hash")

	p := perltest.Perl542(t)
	out, err := p.Command("-MConfig", "-e",
		`for my $k (sort keys %Config::Config) { my $v = $Config::Config{$k}; print "$k\t", (defined $v ? $v : "\x00UNDEF\x00"), "\n"; }`,
	).Output()
	require.NoError(t, err)

	var undef int
	for _, line := range strings.Split(strings.TrimRight(string(out), "\r\n"), "\n") {
		_, v, ok := strings.Cut(strings.TrimRight(line, "\r"), "\t")
		require.True(t, ok, "unparsed line %q", line)
		got := types.Str
		if v == "\x00UNDEF\x00" {
			got, undef = types.Undef, undef+1
		}
		assert.True(t, types.IsSubtype(got, want.Element),
			"a %%Config value is %v, which the declared element type %v excludes", got, want.Element)
	}
	assert.Greater(t, undef, 0, "no %%Config value is undef on this perl; Hash[Str] would have served as well as Hash[Str|Undef]")
	assert.False(t, types.IsSubtype(types.Undef, types.Str),
		"Hash[Str] would exclude the undef values Hash[Str|Undef] admits")

	if !p.IsReferenceBuild(t) {
		t.Logf("not the reference build (perl -V's useshrplib is not false): skipping the exact undef count, pinned only against it")
		return
	}
	assert.Equal(t, 172, undef, "measured on perl 5.42.0's reference build: 172 of %%Config's values are undef")
}

// TestTestMoreBailOutNeverCalled: BAIL_OUT is declared None and is in
// the return-type oracle's exemption map, since it exits rather than
// dies (measured: it calls CORE::exit directly from Test::Builder's
// bail_out, bypassing the driver's exit-catching override).
func TestTestMoreBailOutNeverCalled(t *testing.T) {
	reason, ok := oracleExempt["Test::More::BAIL_OUT"]
	if !assert.True(t, ok, "Test::More::BAIL_OUT is not in the return-type oracle's exemption map") {
		return
	}
	assert.NotEmpty(t, strings.TrimSpace(reason), "Test::More::BAIL_OUT is exempt without a reason")

	sigs, errs := parse.DeclaredSignatures(func() []byte {
		tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", "Test", "More.tpm"))
		require.NoError(t, err)
		return tpm
	}())
	require.Empty(t, errs)
	for _, s := range sigs["BAIL_OUT"] {
		assert.Equal(t, types.None, s.Returns, "Test::More::BAIL_OUT is not declared None")
	}
}

// TestMakeMakerInteractiveSubsExempt: prompt and WriteMakefile are
// exempt with a reason and never called (prompt blocks on STDIN;
// WriteMakefile needs a distribution's own metadata the oracle cannot
// supply).
func TestMakeMakerInteractiveSubsExempt(t *testing.T) {
	for _, name := range []string{"prompt", "WriteMakefile"} {
		reason, ok := oracleExempt["ExtUtils::MakeMaker::"+name]
		if !assert.True(t, ok, "ExtUtils::MakeMaker::%s is not in the return-type oracle's exemption map", name) {
			continue
		}
		assert.NotEmpty(t, strings.TrimSpace(reason), "ExtUtils::MakeMaker::%s is exempt without a reason", name)
	}
	holdReturnTypes(t, "ExtUtils::MakeMaker")
}

// Exporter, XSLoader and DynaLoader are the infrastructure-module batch
// (01a1464e-dd6c, -ddb0, -dde3), each held to perl separately so its own
// test name matches its issue's criteria command.

// TestExporterReturnTypes holds Exporter.tpm's declared return types to
// perl, called with its oracle's cases.
func TestExporterReturnTypes(t *testing.T) {
	holdReturnTypes(t, "Exporter")
}

// TestXSLoaderReturnTypes holds XSLoader.tpm's declared return types to
// perl, called with its oracle's cases.
func TestXSLoaderReturnTypes(t *testing.T) {
	holdReturnTypes(t, "XSLoader")
}

// TestXSLoaderSubsExempt: load and bootstrap_inherit are exempt with a
// reason and never called -- each needs a real compiled extension the
// oracle cannot safely supply.
func TestXSLoaderSubsExempt(t *testing.T) {
	for _, name := range []string{"load", "bootstrap_inherit"} {
		reason, ok := oracleExempt["XSLoader::"+name]
		if !assert.True(t, ok, "XSLoader::%s is not in the return-type oracle's exemption map", name) {
			continue
		}
		assert.NotEmpty(t, strings.TrimSpace(reason), "XSLoader::%s is exempt without a reason", name)
	}
}

// TestDynaLoaderReturnTypes holds DynaLoader.tpm's declared return types
// to perl, called with its oracle's cases.
func TestDynaLoaderReturnTypes(t *testing.T) {
	holdReturnTypes(t, "DynaLoader")
}

// TestDynaLoaderLowLevelSubsExempt: DynaLoader's low-level subs
// dl_load_file, dl_find_symbol, dl_install_xsub and dl_unload_file are
// exempt with a reason and never called (DynaLoader's negative scenario,
// ruling 4, 2026-10-16): each needs a real loaded shared object or
// library reference the oracle's call cases cannot supply.
func TestDynaLoaderLowLevelSubsExempt(t *testing.T) {
	for _, name := range []string{"dl_load_file", "dl_find_symbol", "dl_install_xsub", "dl_unload_file"} {
		reason, ok := oracleExempt["DynaLoader::"+name]
		if !assert.True(t, ok, "DynaLoader::%s is not in the return-type oracle's exemption map", name) {
			continue
		}
		assert.NotEmpty(t, strings.TrimSpace(reason), "DynaLoader::%s is exempt without a reason", name)
	}
}

// TestFileStatMissingFileReturnIsDeclared: File::stat::stat on a path
// that does not exist returns the empty list in list context and undef
// in scalar (measured on 5.42.0), both of which File::stat.tpm's
// declared Object|Undef admits. oracle/File-stat.pl includes this case
// among stat's and lstat's own.
func TestFileStatMissingFileReturnIsDeclared(t *testing.T) {
	assert.Empty(t, moduleFailures(t, "File::stat"))
}

// TestIOExportersReturnTypes holds each module's declared return types
// to perl (perltest.IOExporters), called with its oracle's cases.
func TestIOExportersReturnTypes(t *testing.T) {
	holdReturnTypes(t, perltest.IOExporters...)
}

// ioExportersNetworkSubs are the Net::Ping and IO::Socket subs that
// touch the network (or, for IO::Socket's sockaddr_in, sockaddr_in6
// and sockaddr_un, dispatch on context the way Socket's already-exempt
// ones do); every one must be in oracleExempt.
var ioExportersNetworkSubs = []string{
	"Net::Ping::pingecho", "Net::Ping::wakeonlan",
	"IO::Socket::socket", "IO::Socket::connect", "IO::Socket::bind", "IO::Socket::listen",
	"IO::Socket::accept", "IO::Socket::send", "IO::Socket::recv", "IO::Socket::close",
	"IO::Socket::sockaddr_in", "IO::Socket::sockaddr_in6", "IO::Socket::sockaddr_un",
}

// TestIOExportersNetworkSubsExempt: every Net::Ping and IO::Socket sub
// that touches the network is in oracleExempt with a reason, and the
// return-type oracle never calls it -- holdReturnTypes skips any name
// in oracleExempt when it builds the subs it asks perl to call, so a
// name here is simply never part of that list.
func TestIOExportersNetworkSubsExempt(t *testing.T) {
	for _, name := range ioExportersNetworkSubs {
		reason, ok := oracleExempt[name]
		assert.True(t, ok, "%s touches the network but is not in oracleExempt", name)
		assert.NotEmpty(t, reason, "%s's oracleExempt reason is empty", name)
	}
}

// TestIOExportersOracleStaysInScratch: File::Temp and Archive::Tar
// create files and (for File::Temp) directories; their oracle cases
// name only relative paths -- an absolute one would reach outside the
// scratch directory returnTypeFailures confines perl's cwd, HOME and
// TMPDIR to.
func TestIOExportersOracleStaysInScratch(t *testing.T) {
	absolute := regexp.MustCompile(`['"]/\S`)
	for _, module := range []string{"File-Temp", "Archive-Tar"} {
		cases, err := os.ReadFile(filepath.Join("testdata", "oracle", module+".pl"))
		require.NoError(t, err)
		assert.False(t, absolute.Match(cases), "%s's oracle cases name an absolute path", module)
	}
}

// TestIOSocketConstructorTakesArgs: IO::Socket::new's own perl, measured
// on 5.42.0, is `my($class,%arg) = @_`, and the POD says new "optionally
// takes arguments" and "If the constructor fails it will return undef"
// (measured: IO::Socket->new(Domain=>AF_INET, ..., PeerPort=>1) with no
// listener gives undef, errstr "connect: Connection refused"); a prior
// pass of IO/Socket.tpm declared new taking only its class name
// (01a15ac0, IO::Socket finding 1).
func TestIOSocketConstructorTakesArgs(t *testing.T) {
	tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", "IO", "Socket.tpm"))
	require.NoError(t, err)
	sigs, errs := parse.DeclaredSignatures(tpm)
	require.Empty(t, errs)
	cands, ok := sigs["new"]
	require.True(t, ok, "IO::Socket.tpm does not declare new")
	require.Len(t, cands, 1)
	s := cands[0]
	require.NotEmpty(t, s.Params, "IO::Socket::new takes no parameters besides its class name, but perl's new(%%arg) accepts a Domain=>... argument list (measured)")
	assert.Equal(t, byte('@'), s.Params[0].Sigil, "IO::Socket::new's argument list should be a slurpy, matching perl's my($class,%%arg) = @_")
	assert.True(t, types.IsSubtype(types.Undef, s.Returns), "IO::Socket::new does not admit Undef, but perl returns undef when the constructor fails (measured)")
}

// TestIOSocketMethodsReturnSocketOrUndef: IO::Socket's socket, connect,
// bind and listen each return $sock (the invocant) on success or undef
// on failure (measured on 5.42.0, IO/Socket.pm:91,165,223,233), never
// the core Boolean a prior pass of IO/Socket.tpm declared instead
// (01a15ac0, IO::Socket finding 2).
func TestIOSocketMethodsReturnSocketOrUndef(t *testing.T) {
	tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", "IO", "Socket.tpm"))
	require.NoError(t, err)
	sigs, errs := parse.DeclaredSignatures(tpm)
	require.Empty(t, errs)
	for _, name := range []string{"socket", "connect", "bind", "listen"} {
		cands, ok := sigs[name]
		if !assert.True(t, ok, "IO::Socket::%s is not declared", name) {
			continue
		}
		require.Len(t, cands, 1)
		s := cands[0]
		assert.True(t, types.IsSubtype(types.Object, s.Returns), "IO::Socket::%s should admit Object: perl returns the invocant itself on success (measured)", name)
		assert.True(t, types.IsSubtype(types.Undef, s.Returns), "IO::Socket::%s should admit Undef: perl returns undef on failure (measured)", name)
	}
}

// TestIOSocketSendFlagsOptional: IO::Socket::send's own perl, measured
// on 5.42.0, is `@_ >= 2 && @_ <= 4`, and the POD shows
// `send(BUF, [FLAGS, [TO]])` -- flags is optional (measured:
// $sock->send("hi") returns 2), not the required parameter a prior pass
// of IO/Socket.tpm declared (01a15ac0, IO::Socket finding 5).
func TestIOSocketSendFlagsOptional(t *testing.T) {
	tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", "IO", "Socket.tpm"))
	require.NoError(t, err)
	sigs, errs := parse.DeclaredSignatures(tpm)
	require.Empty(t, errs)
	cands, ok := sigs["send"]
	require.True(t, ok, "IO::Socket::send is not declared")
	require.Len(t, cands, 1)
	s := cands[0]
	require.GreaterOrEqual(t, len(s.Params), 2, "IO::Socket::send should declare buf and flags parameters")
	assert.False(t, s.Params[1].Required, `IO::Socket::send's flags parameter should be optional: perl accepts $sock->send("hi") with no flags (measured)`)
}

// TestSocketListContextUnpackers: Socket::sockaddr_un dispatches on
// wantarray alone, the same way sockaddr_in and sockaddr_in6 do
// (measured on 5.42.0, Socket.pm:876-883: `unpack_sockaddr_un` runs
// under `if(wantarray)`; `my @l = sockaddr_un(pack_sockaddr_un($path))`
// gives `($path)`), so it needs a List[Str] candidate for list context
// beside its scalar (pack) one -- left as a single scalar-only
// candidate by a prior pass of Socket.tpm (01a15ac0, Socket finding 4).
func TestSocketListContextUnpackers(t *testing.T) {
	tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", "Socket.tpm"))
	require.NoError(t, err)
	sigs, errs := parse.DeclaredSignatures(tpm)
	require.Empty(t, errs)
	cands, ok := sigs["sockaddr_un"]
	require.True(t, ok, "Socket::sockaddr_un is not declared")
	require.Len(t, cands, 2, "sockaddr_un should be a multi, pack and unpack, the same way sockaddr_in and sockaddr_in6 are")
	hasListCandidate := slices.ContainsFunc(cands, func(s types.Signature) bool {
		return s.Answers() == types.ListCtx && s.ReturnElement == types.Str
	})
	assert.True(t, hasListCandidate, "sockaddr_un has no List[Str] candidate for its list-context (unpack) form")
}

// TestIOFileNewMissingFileReturnIsDeclared: IO::File->new on a file
// that does not exist returns undef rather than dying, measured on
// 5.42.0, which IO::File.tpm's declared Object|Undef admits.
// oracle/IO-File.pl's own `new` case (no filename) already exercises
// new's Object branch; this is the Undef one, checked directly rather
// than added to the module-wide oracle (new's declared signature takes
// an optional filename, and giving one here would otherwise need a
// second call form the oracle driver has no way to select between).
func TestIOFileNewMissingFileReturnIsDeclared(t *testing.T) {
	p := perltest.Perl542(t)
	scratch := t.TempDir()
	cmd := p.Command("-MIO::File", "-e",
		`my $f = IO::File->new("does-not-exist.txt", "r"); print defined($f) ? "defined" : "undef";`)
	cmd.Dir = scratch
	out, err := cmd.Output()
	require.NoError(t, err)
	require.Equal(t, "undef", string(out), "IO::File->new on a missing file should return undef, not die or succeed")

	tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", "IO", "File.tpm"))
	require.NoError(t, err)
	sigs, errs := parse.DeclaredSignatures(tpm)
	require.Empty(t, errs)
	cands, ok := sigs["new"]
	require.True(t, ok, "IO::File.tpm does not declare new")
	admits := false
	for _, s := range cands {
		if types.IsSubtype(types.Undef, s.Returns) {
			admits = true
		}
	}
	assert.True(t, admits, "IO::File.tpm's declared new does not admit Undef, but perl returns undef on a missing file")
}

// allExemptAllowlist names the modules the return-type oracle may
// legitimately call nothing from: XSLoader declares exactly load and
// bootstrap_inherit, and Net::Ping exactly pingecho and wakeonlan, and
// every one of the four needs a real compiled extension or a real
// network connection (see oracleExempt). Any other module whose every
// declared sub ends up exempt passes TestOOModulesReturnTypes or
// TestIOExportersReturnTypes vacuously -- never actually calling perl --
// which is the defect this guards against.
var allExemptAllowlist = map[string]bool{
	"XSLoader":  true,
	"Net::Ping": true,
}

// exemptMapFor is the exemption map holdReturnTypes or
// TestOOModulesReturnTypes actually calls module with: ooExempt's
// extended map for the OO batch, oracleExempt for everyone else.
func exemptMapFor(module string) map[string]string {
	if slices.Contains(perltest.OOModules, module) {
		return ooExempt()
	}
	return withPlatformExempt(oracleExempt)
}

// TestNoModuleIsAllExempt: for every module with a .tpm declaration, at
// least one of its declared subs is not in the exemption map its own
// return-type oracle run uses, unless the module is in
// allExemptAllowlist. Exempting every sub a module declares makes its
// return-type oracle test pass without ever calling perl, which the
// allowlist proves this test would otherwise miss: XSLoader stays green
// only because it is named here.
func TestNoModuleIsAllExempt(t *testing.T) {
	root := filepath.Join("..", "parse", "declarations")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() || !strings.HasSuffix(path, ".tpm") {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		module := strings.ReplaceAll(strings.TrimSuffix(filepath.ToSlash(rel), ".tpm"), "/", "::")
		if module == "CORE" {
			// CORE's builtins are not modules the return-type oracle
			// calls perl's `require` on; a different mechanism holds them.
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sigs, errs := parse.DeclaredSignatures(src)
		if len(errs) > 0 || len(sigs) == 0 {
			// A declaration error is a different test's job to catch; a
			// module with no declared subs has nothing to exempt.
			return nil
		}
		exempt := exemptMapFor(module)
		allExempt := true
		for name := range sigs {
			if _, ok := exempt[module+"::"+name]; !ok {
				allExempt = false
				break
			}
		}
		if allExempt && !allExemptAllowlist[module] {
			t.Errorf("%s: every declared sub is exempt, so its return-type oracle test never calls perl; "+
				"add a real call case for at least one, or add %q to allExemptAllowlist with why", module, module)
		}
		return nil
	})
	require.NoError(t, err)
}

// caseArities runs perl to read a cases file's own hash literal and
// reports each sub name's call cases by their argument count. The cases
// file is perl data, some of it built from the module's own constants
// (Socket::pack_sockaddr_in and the like), so the module is required
// first, exactly as return_oracle.pl itself requires it before reading
// the same file (ruling: perl is the only safe reader of this data, not
// a Go-side parse of the literal).
func caseArities(t *testing.T, p perltest.Perl, module, casesPath string) map[string][]int {
	t.Helper()
	script := `my ($module, $cases_file) = @ARGV;` +
		`open my $out, '>&', \*STDOUT or die $!; open STDOUT, '>&', \*STDERR or die $!;` +
		`(my $file = "$module.pm") =~ s{::}{/}g; require $file;` +
		`my %cases = do $cases_file; die "cannot read $cases_file: ", $@ || $!, "\n" if $@;` +
		`delete $cases{INSTANCE};` +
		`open STDOUT, '>&', $out or die $!;` +
		`for my $n (sort keys %cases) { print "$n ", join(",", map { scalar @$_ } @{$cases{$n}}), "\n" }`
	scratch := t.TempDir()
	cmd := p.Command("-e", script, module, casesPath)
	cmd.Dir = scratch
	cmd.Env = append(cmd.Env, "HOME="+scratch, "TMPDIR="+scratch)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	require.NoError(t, err, "caseArities(%s): %s", module, stderr.String())
	arities := map[string][]int{}
	for _, line := range strings.Split(strings.TrimRight(string(out), "\r\n"), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		name, counts, ok := strings.Cut(line, " ")
		require.True(t, ok, "caseArities: unparseable line %q from %s", line, casesPath)
		for _, s := range strings.Split(counts, ",") {
			n, err := strconv.Atoi(s)
			require.NoError(t, err)
			arities[name] = append(arities[name], n)
		}
	}
	return arities
}

// TestEveryMultiCandidateIsExercised: for every module with a .tpm
// declaration, each candidate of a `multi` sub is reachable by at least
// one oracle call case of a matching arity and form, or the whole name
// is in the module's exemption map. RFC 0001 ("Context selects by
// return type") lets calls() collapse two candidates sharing a context
// into one request, told apart only by arity; a candidate calls() never
// asks for, and no case ever exercises, can pass unnoticed, since
// returnTypeFailures only requires SOME candidate to admit an observed
// value, never every one.
func TestEveryMultiCandidateIsExercised(t *testing.T) {
	p := perltest.Perl542(t)
	root := filepath.Join("..", "parse", "declarations")
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() || !strings.HasSuffix(path, ".tpm") {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		module := strings.ReplaceAll(strings.TrimSuffix(filepath.ToSlash(rel), ".tpm"), "/", "::")
		if module == "CORE" {
			// CORE's multi candidates are told apart by parse-time
			// syntax, not a runtime oracle call; a different mechanism
			// holds them.
			return nil
		}
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sigs, errs := parse.DeclaredSignatures(src)
		if len(errs) > 0 {
			return nil
		}
		hasMulti := false
		for _, cands := range sigs {
			if len(cands) > 1 {
				hasMulti = true
				break
			}
		}
		if !hasMulti {
			return nil
		}
		casesPath, err := filepath.Abs(filepath.Join("testdata", "oracle", strings.ReplaceAll(module, "::", "-")+".pl"))
		if err != nil {
			return err
		}
		if _, statErr := os.Stat(casesPath); statErr != nil {
			// No oracle file for this module (a non-core .tpm outside
			// the core-module-types milestone): nothing to check here.
			return nil
		}
		arities := caseArities(t, p, module, casesPath)
		exempt := exemptMapFor(module)
		for name, cands := range sigs {
			if len(cands) < 2 {
				continue
			}
			if _, ok := exempt[module+"::"+name]; ok {
				continue
			}
			for _, s := range cands {
				form := callForm(s)
				if !slices.ContainsFunc(arities[name], func(n int) bool { return takes(s, form, n) }) {
					t.Errorf("%s::%s: a multi candidate (%s form, returns %s) has no oracle call case of a matching arity, and %s::%s is not exempt",
						module, name, form, s.Returns, module, name)
				}
			}
		}
		return nil
	})
	require.NoError(t, err)
}
