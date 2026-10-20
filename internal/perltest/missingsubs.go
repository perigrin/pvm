// ABOUTME: Asks a running perl which of a module's names it does not define as a callable sub at all.
// ABOUTME: Which names exist depends on the C library (or, on Windows, the SDK headers) perl was built against.

package perltest

import (
	"fmt"
	"strings"
	"testing"
)

// MissingSubs reports, name to reason, which of names module does not
// define as a callable sub on p's perl -- as opposed to one that exists
// but dies or returns a different value when called. A module like
// Errno exports a different constant set depending on the C library (and,
// on Windows, the Winsock and SDK headers) perl was built against: which
// names exist there is not something this repository can hand-list once
// for every platform CI runs on, so this asks the perl actually running
// instead of a per-OS list that could only ever describe the one runner
// that measured it.
//
// A name perl generates only once something asks for it by name (the
// same AUTOLOAD-backed pattern return_oracle.pl's own callable() check
// and by-name priming handle -- see its comment) is primed the same way
// here, so such a name is not wrongly reported missing.
func MissingSubs(t testing.TB, p Perl, inc, module string, names []string) map[string]string {
	t.Helper()
	if len(names) == 0 {
		return nil
	}
	script := `use B ();` +
		`my ($module, @names) = @ARGV;` +
		`(my $file = "$module.pm") =~ s{::}{/}g; require $file;` +
		`eval "package Held::MissingSubsBarePrime; use $module; 1";` +
		`sub callable { my ($name) = @_; return 0 if !exists &$name; return 1 if defined &$name; ` +
		`my $home = B::svref_2object(\&$name)->GV->STASH->NAME; return defined &{"${home}::AUTOLOAD"}; }` +
		`my $n = 0;` +
		`for my $name (@names) {` +
		`  my $ok = callable("${module}::$name") || callable("Held::MissingSubsBarePrime::$name");` +
		`  if (!$ok) {` +
		`    my $primed = "Held::MissingSubsPrimeByName" . $n++;` +
		`    eval "package $primed; use $module qw($name); 1";` +
		`    $ok = callable("${module}::$name") || callable("${primed}::$name");` +
		`  }` +
		`  print "$name\n" unless $ok;` +
		`}`
	args := []string{"-e", script, module}
	args = append(args, names...)
	if inc != "" {
		args = append([]string{"-I", inc}, args...)
	}
	var stderr strings.Builder
	cmd := p.Command(args...)
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("MissingSubs(%s): %v: %s", module, err, stderr.String())
	}
	missing := map[string]string{}
	for _, name := range strings.Split(strings.TrimRight(string(out), "\r\n"), "\n") {
		name = strings.TrimRight(name, "\r")
		if name == "" {
			continue
		}
		missing[name] = fmt.Sprintf("%s's perl defines no sub %s::%s -- which names a module like this exports depends on the C library perl was built against, measured directly rather than hand-listed per platform", p.Path, module, name)
	}
	return missing
}
