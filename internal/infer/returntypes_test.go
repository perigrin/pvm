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
	"runtime"
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
	"IO::Socket::socket":       "opens a real socket file descriptor",
	"IO::Socket::connect":      "connects a real socket onto the network",
	"IO::Socket::bind":         "binds a real socket to a local address",
	"IO::Socket::listen":       "marks a real socket as listening for connections",
	"IO::Socket::accept":       "blocks accepting a real connection on a socket",
	"IO::Socket::send":         "writes onto a real socket",
	"IO::Socket::recv":         "blocks reading from a real socket",
	"IO::Socket::close":        "dereferences the invocant as a glob to close its socket, dying \"Can't use string ... as a symbol ref\" when the invocant is the bare class name the oracle calls with",

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

// platformExempt holds, per runtime.GOOS, the additional subs the
// return-type oracle exempts on that platform alone: a name whose call
// dies, or whose returned type the .tpm (written against Linux perl
// 5.42.0) does not declare, there and nowhere else. Each reason cites the
// CI job that measured it; withPlatformExempt merges the current OS's
// entries into a module's base exemption map, so every other name still
// runs there. A name genuinely absent from the running perl (as opposed
// to one that exists and dies or differs) belongs in neither map --
// perltest.MissingSubs measures that directly instead, since which names
// a module like Errno exports depends on the C library perl was built
// against, not something this repository can hand-list once for every
// platform CI runs on.
var platformExempt = map[string]map[string]string{
	"darwin": {
		"Fcntl::DN_ACCESS":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::DN_ATTRIB":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::DN_CREATE":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::DN_DELETE":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::DN_MODIFY":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::DN_MULTISHOT":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::DN_RENAME":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_ADD_SEALS":                           "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_EXLCK":                               "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_GETLK64":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_GETPIPE_SZ":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_GETSIG":                              "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_GET_SEALS":                           "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_NOTIFY":                              "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_SEAL_FUTURE_WRITE":                   "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_SEAL_GROW":                           "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_SEAL_SEAL":                           "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_SEAL_SHRINK":                         "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_SEAL_WRITE":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_SETLK64":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_SETLKW64":                            "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_SETPIPE_SZ":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_SETSIG":                              "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::F_SHLCK":                               "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::LOCK_MAND":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::LOCK_READ":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::LOCK_RW":                               "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::LOCK_WRITE":                            "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::O_DIRECT":                              "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::O_EVTONLY":                             "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Fcntl::O_EXLOCK":                              "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Fcntl::O_LARGEFILE":                           "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::O_NOATIME":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::O_RSYNC":                               "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::O_SHLOCK":                              "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Fcntl::O_SYMLINK":                             "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Fcntl::O_TMPFILE":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Fcntl::S_IFWHT":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Fcntl::S_ISTXT":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Fcntl::S_ISWHT":                               "returns Boolean on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_CCITT":                             "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_CHAOS":                             "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_DATAKIT":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_DLI":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_ECMA":                              "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_HYLINK":                            "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_IMPLINK":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_ISO":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_KEY":                               "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::AF_LAT":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_LINK":                              "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_NS":                                "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_OSI":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_PUP":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::AF_X25":                               "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::AI_CANONIDN":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::AI_IDN":                               "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::EAI_BADHINTS":                         "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::EAI_PROTOCOL":                         "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::IPV6_ADD_MEMBERSHIP":                  "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IPV6_DROP_MEMBERSHIP":                 "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IPV6_MTU":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IPV6_MTU_DISCOVER":                    "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IPV6_RECVERR":                         "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IPV6_ROUTER_ALERT":                    "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IP_BIND_ADDRESS_NO_PORT":              "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IP_FREEBIND":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IP_MTU":                               "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IP_MTU_DISCOVER":                      "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IP_MULTICAST_ALL":                     "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IP_NODEFRAG":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IP_PMTUDISC_DO":                       "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IP_PMTUDISC_DONT":                     "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IP_PMTUDISC_PROBE":                    "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IP_PMTUDISC_WANT":                     "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IP_RECVERR":                           "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::IP_TRANSPARENT":                       "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::MSG_EOF":                              "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::MSG_ERRQUEUE":                         "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::MSG_FASTOPEN":                         "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::MSG_FIN":                              "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::MSG_PROXY":                            "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::MSG_RST":                              "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::MSG_SYN":                              "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::NI_IDN":                               "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::PF_CCITT":                             "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::PF_CHAOS":                             "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::PF_DATAKIT":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::PF_DLI":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::PF_ECMA":                              "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::PF_HYLINK":                            "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::PF_IMPLINK":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::PF_ISO":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::PF_LAT":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::PF_LINK":                              "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::PF_NS":                                "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::PF_OSI":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::PF_PUP":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::PF_X25":                               "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SCM_CREDENTIALS":                      "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SCM_CREDS":                            "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::SOCK_CLOEXEC":                         "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SOCK_NONBLOCK":                        "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_ATTACH_FILTER":                     "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_DETACH_FILTER":                     "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_DOMAIN":                            "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_INCOMING_CPU":                      "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_INCOMING_NAPI_ID":                  "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_LOCK_FILTER":                       "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_PASSCRED":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_PEERCRED":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_PROTOCOL":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_RCVBUFFORCE":                       "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_SECURITY_AUTHENTICATION":           "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_SECURITY_ENCRYPTION_NETWORK":       "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_SECURITY_ENCRYPTION_TRANSPORT":     "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_SNDBUFFORCE":                       "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::SO_USELOOPBACK":                       "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::TCP_CC_INFO":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_CONGESTION":                       "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_CONNECTIONTIMEOUT":                "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::TCP_CORK":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_DEFER_ACCEPT":                     "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_FASTOPEN_CONNECT":                 "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_FASTOPEN_KEY":                     "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_FASTOPEN_NO_COOKIE":               "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_INFO":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_KEEPALIVE":                        "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::TCP_KEEPIDLE":                         "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_LINGER2":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_MD5SIG":                           "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_MD5SIG_EXT":                       "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_NOOPT":                            "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::TCP_NOPUSH":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Socket::TCP_QUICKACK":                         "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_SYNCNT":                           "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_TIMESTAMP":                        "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_TX_DELAY":                         "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_USER_TIMEOUT":                     "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::TCP_WINDOW_CLAMP":                     "dies on macOS's perl -- measured in CI job 39489951290",
		"Socket::UIO_MAXIOV":                           "dies on macOS's perl -- measured in CI job 39489951290",
		"FileHandle::F_EXLCK":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"FileHandle::F_GETLK64":                        "dies on macOS's perl -- measured in CI job 39489951290",
		"FileHandle::F_SETLK64":                        "dies on macOS's perl -- measured in CI job 39489951290",
		"FileHandle::F_SETLKW64":                       "dies on macOS's perl -- measured in CI job 39489951290",
		"FileHandle::F_SHLCK":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"FileHandle::O_DIRECT":                         "dies on macOS's perl -- measured in CI job 39489951290",
		"FileHandle::O_EXLOCK":                         "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"FileHandle::O_LARGEFILE":                      "dies on macOS's perl -- measured in CI job 39489951290",
		"FileHandle::O_RSYNC":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"FileHandle::O_SHLOCK":                         "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::File::O_DIRECT":                           "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::File::O_EXLOCK":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::File::O_LARGEFILE":                        "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::File::O_RSYNC":                            "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::File::O_SHLOCK":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_CCITT":                         "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_CHAOS":                         "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_DATAKIT":                       "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_DLI":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_ECMA":                          "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_HYLINK":                        "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_IMPLINK":                       "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_ISO":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_KEY":                           "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::AF_LAT":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_LINK":                          "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_NS":                            "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_OSI":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_PUP":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::AF_X25":                           "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::MSG_EOF":                          "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::MSG_ERRQUEUE":                     "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::MSG_FASTOPEN":                     "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::MSG_FIN":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::MSG_PROXY":                        "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::MSG_RST":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::MSG_SYN":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::PF_CCITT":                         "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::PF_CHAOS":                         "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::PF_DATAKIT":                       "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::PF_DLI":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::PF_ECMA":                          "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::PF_HYLINK":                        "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::PF_IMPLINK":                       "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::PF_ISO":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::PF_LAT":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::PF_LINK":                          "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::PF_NS":                            "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::PF_OSI":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::PF_PUP":                           "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::PF_X25":                           "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::SCM_CREDENTIALS":                  "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::SCM_CREDS":                        "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::SO_ATTACH_FILTER":                 "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::SO_DETACH_FILTER":                 "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::SO_DOMAIN":                        "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::SO_PASSCRED":                      "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::SO_PEERCRED":                      "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::SO_PROTOCOL":                      "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::SO_SECURITY_AUTHENTICATION":       "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::SO_SECURITY_ENCRYPTION_NETWORK":   "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::SO_SECURITY_ENCRYPTION_TRANSPORT": "dies on macOS's perl -- measured in CI job 39489951290",
		"IO::Socket::SO_USELOOPBACK":                   "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"IO::Socket::UIO_MAXIOV":                       "dies on macOS's perl -- measured in CI job 39489951290",
		"POSIX::ARG_MAX":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"POSIX::CHILD_MAX":                             "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"POSIX::CLK_TCK":                               "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"POSIX::EPROCLIM":                              "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"POSIX::ERESTART":                              "dies on macOS's perl -- measured in CI job 39489951290",
		"POSIX::LC_ADDRESS":                            "dies on macOS's perl -- measured in CI job 39489951290",
		"POSIX::LC_IDENTIFICATION":                     "dies on macOS's perl -- measured in CI job 39489951290",
		"POSIX::LC_MEASUREMENT":                        "dies on macOS's perl -- measured in CI job 39489951290",
		"POSIX::LC_NAME":                               "dies on macOS's perl -- measured in CI job 39489951290",
		"POSIX::LC_PAPER":                              "dies on macOS's perl -- measured in CI job 39489951290",
		"POSIX::LC_TELEPHONE":                          "dies on macOS's perl -- measured in CI job 39489951290",
		"POSIX::LDBL_EPSILON":                          "returns Num on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"POSIX::LDBL_MAX":                              "returns Num on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"POSIX::LDBL_MIN":                              "returns Num on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"POSIX::LINK_MAX":                              "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"POSIX::L_cuserid":                             "dies on macOS's perl -- measured in CI job 39489951290",
		"POSIX::OPEN_MAX":                              "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"POSIX::SIGPOLL":                               "dies on macOS's perl -- measured in CI job 39489951290",
		"POSIX::SIGRTMAX":                              "dies on macOS's perl -- measured in CI job 39489951290",
		"POSIX::SIGRTMIN":                              "dies on macOS's perl -- measured in CI job 39489951290",
		"POSIX::cuserid":                               "dies on macOS's perl -- measured in CI job 39489951290",
		"POSIX::getgrnam":                              "returns Undef on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
		"Time::HiRes::CLOCK_BOOTTIME":                  "dies on macOS's perl -- measured in CI job 39489951290",
		"Time::HiRes::CLOCK_MONOTONIC_COARSE":          "dies on macOS's perl -- measured in CI job 39489951290",
		"Time::HiRes::CLOCK_REALTIME_COARSE":           "dies on macOS's perl -- measured in CI job 39489951290",
		"Time::HiRes::CLOCK_UPTIME_RAW":                "returns Int on macOS's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39489951290",
	},
	"windows": {
		"Time::HiRes::clock":                           `unimplemented on Windows's perl -- measured in CI job 39489951290: "Time::HiRes::clock(): unimplemented in this platform"`,
		"Fcntl::DN_ACCESS":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::DN_ATTRIB":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::DN_CREATE":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::DN_DELETE":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::DN_MODIFY":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::DN_MULTISHOT":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::DN_RENAME":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::FAPPEND":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::FASYNC":                                "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::FD_CLOEXEC":                            "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::FNDELAY":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::FNONBLOCK":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_ADD_SEALS":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_DUPFD":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_EXLCK":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_GETFD":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_GETFL":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_GETLEASE":                            "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_GETLK":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_GETLK64":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_GETOWN":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_GETPIPE_SZ":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_GETSIG":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_GET_SEALS":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_NOTIFY":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_RDLCK":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SEAL_FUTURE_WRITE":                   "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SEAL_GROW":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SEAL_SEAL":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SEAL_SHRINK":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SEAL_WRITE":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SETFD":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SETFL":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SETLEASE":                            "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SETLK":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SETLK64":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SETLKW":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SETLKW64":                            "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SETOWN":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SETPIPE_SZ":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SETSIG":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_SHLCK":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_UNLCK":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::F_WRLCK":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::LOCK_MAND":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::LOCK_READ":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::LOCK_RW":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::LOCK_WRITE":                            "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::O_ASYNC":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::O_DIRECT":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::O_DIRECTORY":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::O_DSYNC":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::O_LARGEFILE":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::O_NDELAY":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::O_NOATIME":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::O_NOCTTY":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::O_NOFOLLOW":                            "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::O_NOINHERIT":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Fcntl::O_NONBLOCK":                            "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::O_RANDOM":                              "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Fcntl::O_RAW":                                 "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Fcntl::O_RSYNC":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::O_SEQUENTIAL":                          "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Fcntl::O_SYNC":                                "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::O_TEMPORARY":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Fcntl::O_TMPFILE":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::S_IFLNK":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::S_IFSOCK":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::S_ISLNK":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::S_ISSOCK":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Fcntl::S_ISVTX":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::AF_CCITT":                             "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::AF_CHAOS":                             "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::AF_DATAKIT":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::AF_DLI":                               "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::AF_ECMA":                              "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::AF_HYLINK":                            "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::AF_IMPLINK":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::AF_ISO":                               "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::AF_KEY":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::AF_LAT":                               "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::AF_NS":                                "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::AF_OSI":                               "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::AF_PUP":                               "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::AF_ROUTE":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::AF_X25":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::AI_CANONIDN":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::AI_IDN":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::EAI_ADDRFAMILY":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::EAI_SYSTEM":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IOV_MAX":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IPV6_ROUTER_ALERT":                    "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IP_BIND_ADDRESS_NO_PORT":              "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IP_FREEBIND":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IP_MULTICAST_ALL":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IP_NODEFRAG":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IP_PMTUDISC_DO":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IP_PMTUDISC_DONT":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IP_PMTUDISC_PROBE":                    "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IP_PMTUDISC_WANT":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IP_RECVOPTS":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IP_RECVRETOPTS":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IP_RETOPTS":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::IP_TRANSPARENT":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::MSG_CTRUNC":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::MSG_DONTWAIT":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::MSG_EOR":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::MSG_ERRQUEUE":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::MSG_FASTOPEN":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::MSG_FIN":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::MSG_MAXIOVLEN":                        "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::MSG_NOSIGNAL":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::MSG_PROXY":                            "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::MSG_RST":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::MSG_SYN":                              "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::MSG_TRUNC":                            "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::NI_IDN":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::PF_CCITT":                             "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::PF_CHAOS":                             "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::PF_DATAKIT":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::PF_DLI":                               "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::PF_ECMA":                              "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::PF_HYLINK":                            "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::PF_IMPLINK":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::PF_ISO":                               "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::PF_KEY":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::PF_LAT":                               "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::PF_NS":                                "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::PF_OSI":                               "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::PF_PUP":                               "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::PF_ROUTE":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::PF_X25":                               "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SCM_CREDENTIALS":                      "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SCM_RIGHTS":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SCM_TIMESTAMP":                        "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SOCK_CLOEXEC":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SOCK_NONBLOCK":                        "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_ATTACH_FILTER":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_DETACH_FILTER":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_DOMAIN":                            "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_DONTLINGER":                        "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::SO_INCOMING_CPU":                      "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_INCOMING_NAPI_ID":                  "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_LOCK_FILTER":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_PASSCRED":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_PEERCRED":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_PROTOCOL":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_RCVBUFFORCE":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_REUSEPORT":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_SECURITY_AUTHENTICATION":           "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_SECURITY_ENCRYPTION_NETWORK":       "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_SECURITY_ENCRYPTION_TRANSPORT":     "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_SNDBUFFORCE":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::SO_USELOOPBACK":                       "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Socket::TCP_CC_INFO":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_CONGESTION":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_CORK":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_DEFER_ACCEPT":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_FASTOPEN":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_FASTOPEN_CONNECT":                 "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_FASTOPEN_KEY":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_FASTOPEN_NO_COOKIE":               "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_INFO":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_KEEPCNT":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_KEEPIDLE":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_KEEPINTVL":                        "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_LINGER2":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_MAXSEG":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_MD5SIG":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_MD5SIG_EXT":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_NOTSENT_LOWAT":                    "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_QUICKACK":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_SYNCNT":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_TIMESTAMP":                        "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_TX_DELAY":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_USER_TIMEOUT":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::TCP_WINDOW_CLAMP":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"Socket::UIO_MAXIOV":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"File::Spec::Functions::file_name_is_absolute": "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"FileHandle::FD_CLOEXEC":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_DUPFD":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_EXLCK":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_GETFD":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_GETFL":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_GETLK":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_GETLK64":                        "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_GETOWN":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_RDLCK":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_SETFD":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_SETFL":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_SETLK":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_SETLK64":                        "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_SETLKW":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_SETLKW64":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_SETOWN":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_SHLCK":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_UNLCK":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::F_WRLCK":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::O_ASYNC":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::O_DIRECT":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::O_DIRECTORY":                      "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::O_DSYNC":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::O_LARGEFILE":                      "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::O_NDELAY":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::O_NOCTTY":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::O_NOFOLLOW":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::O_NOINHERIT":                      "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"FileHandle::O_NONBLOCK":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::O_RANDOM":                         "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"FileHandle::O_RAW":                            "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"FileHandle::O_RSYNC":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::O_SEQUENTIAL":                     "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"FileHandle::O_SYNC":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"FileHandle::O_TEMPORARY":                      "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::File::O_ASYNC":                            "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::File::O_DIRECT":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::File::O_DIRECTORY":                        "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::File::O_DSYNC":                            "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::File::O_LARGEFILE":                        "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::File::O_NDELAY":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::File::O_NOCTTY":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::File::O_NOFOLLOW":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::File::O_NOINHERIT":                        "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::File::O_NONBLOCK":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::File::O_RANDOM":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::File::O_RAW":                              "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::File::O_RSYNC":                            "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::File::O_SEQUENTIAL":                       "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::File::O_SYNC":                             "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::File::O_TEMPORARY":                        "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::AF_CCITT":                         "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::AF_CHAOS":                         "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::AF_DATAKIT":                       "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::AF_DLI":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::AF_ECMA":                          "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::AF_HYLINK":                        "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::AF_IMPLINK":                       "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::AF_ISO":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::AF_KEY":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::AF_LAT":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::AF_NS":                            "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::AF_OSI":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::AF_PUP":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::AF_ROUTE":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::AF_X25":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::IOV_MAX":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::IP_RECVOPTS":                      "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::IP_RECVRETOPTS":                   "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::IP_RETOPTS":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::MSG_CTRUNC":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::MSG_DONTWAIT":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::MSG_EOR":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::MSG_ERRQUEUE":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::MSG_FASTOPEN":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::MSG_FIN":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::MSG_MAXIOVLEN":                    "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::MSG_NOSIGNAL":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::MSG_PROXY":                        "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::MSG_RST":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::MSG_SYN":                          "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::MSG_TRUNC":                        "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::PF_CCITT":                         "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::PF_CHAOS":                         "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::PF_DATAKIT":                       "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::PF_DLI":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::PF_ECMA":                          "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::PF_HYLINK":                        "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::PF_IMPLINK":                       "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::PF_ISO":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::PF_KEY":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::PF_LAT":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::PF_NS":                            "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::PF_OSI":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::PF_PUP":                           "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::PF_ROUTE":                         "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::PF_X25":                           "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SCM_CREDENTIALS":                  "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SCM_RIGHTS":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SCM_TIMESTAMP":                    "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SO_ATTACH_FILTER":                 "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SO_DETACH_FILTER":                 "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SO_DOMAIN":                        "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SO_DONTLINGER":                    "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::SO_PASSCRED":                      "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SO_PEERCRED":                      "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SO_PROTOCOL":                      "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SO_REUSEPORT":                     "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SO_SECURITY_AUTHENTICATION":       "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SO_SECURITY_ENCRYPTION_NETWORK":   "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SO_SECURITY_ENCRYPTION_TRANSPORT": "dies on Windows's perl -- measured in CI job 39498377022",
		"IO::Socket::SO_USELOOPBACK":                   "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::UIO_MAXIOV":                       "dies on Windows's perl -- measured in CI job 39498377022",
		"File::Spec::file_name_is_absolute":            "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"IO::Socket::IP::SO_REUSEPORT":                 "dies on Windows's perl -- measured in CI job 39498377022",
		"Term::ReadLine::MinLine":                      "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
		"Term::ReadLine::addhistory":                   "returns Int on Windows's perl, which this .tpm (describing Linux perl 5.42.0) does not declare -- measured in CI job 39498377022",
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
		for _, f := range returnTypeFailures(t, p, oracleRun{Module: module, TPM: tpm, Cases: cases, Exempt: exemptForModule(t, p, module, tpm)}) {
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
func exemptForModule(t *testing.T, p perltest.Perl, module string, tpm []byte) map[string]string {
	t.Helper()
	exempt := withPlatformExempt(oracleExempt)
	if module == "Errno" {
		sigs, errs := parse.DeclaredSignatures(tpm)
		if len(errs) == 0 {
			for name, reason := range perltest.MissingSubs(t, p, "", module, slices.Sorted(maps.Keys(sigs))) {
				exempt[module+"::"+name] = reason
			}
		}
	}
	return exempt
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
		m["IO::Socket::IP::"+name] = "lives in IO::Socket's own stash, from its file-scope `use Socket;`, not IO::Socket::IP's; the sub-form oracle cannot reach it there (measured), and Socket.tpm's own oracle already holds its return type"
	}
	for _, name := range []string{"connect", "sockhost", "sockport", "peerhost", "peerport"} {
		m["IO::Socket::IP::"+name] = "opens or reads a real network connection"
	}
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
	exempt := ooExempt()
	for _, module := range perltest.OOModules {
		rel := strings.ReplaceAll(module, "::", "/")
		tpm, err := os.ReadFile(filepath.Join("..", "parse", "declarations", filepath.FromSlash(rel)+".tpm"))
		if err != nil {
			t.Errorf("%s: no declaration: %v", module, err)
			continue
		}
		cases, err := filepath.Abs(filepath.Join("testdata", "oracle", strings.ReplaceAll(module, "::", "-")+".pl"))
		require.NoError(t, err)
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
		"IO::Socket::IP": {"connect", "sockhost", "sockport", "peerhost", "peerport"},
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
// touch the network (or, for IO::Socket's sockaddr_in and
// sockaddr_in6, dispatch on context the way Socket's already-exempt
// ones do); every one must be in oracleExempt.
var ioExportersNetworkSubs = []string{
	"Net::Ping::pingecho", "Net::Ping::wakeonlan",
	"IO::Socket::socket", "IO::Socket::connect", "IO::Socket::bind", "IO::Socket::listen",
	"IO::Socket::accept", "IO::Socket::send", "IO::Socket::recv", "IO::Socket::close",
	"IO::Socket::sockaddr_in", "IO::Socket::sockaddr_in6",
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
