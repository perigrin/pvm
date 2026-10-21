// ABOUTME: Measures, on the running perl, whether a zero-arg constant's die-or-return behavior
// ABOUTME: differs from what the Linux declaration says, so a platform-specific constant family never needs a hand-kept list.

package perltest

import (
	"fmt"
	"strings"
	"testing"
)

// ConstantDifferences reports, name to reason, which of wantsValue's
// names behave differently on p's perl than their declared shape says:
// a name declared with a value (wantsValue[name] true) that dies when
// called with no arguments in scalar context, or one declared None
// (wantsValue[name] false) that instead returns. A name the module does
// not define at all is treated the same as one that dies -- there is no
// third way to call it.
//
// A constant family like Socket's or POSIX's can number in the hundreds,
// differs per C library and per build, and keeps growing as more of a
// module gets declared: measuring this directly, the same way
// MissingSubs measures which names a module defines, is the only way
// such a family never needs a hand-kept per-platform list that goes
// stale the moment the build changes under it.
func ConstantDifferences(t testing.TB, p Perl, inc, module string, wantsValue map[string]bool) map[string]string {
	t.Helper()
	if len(wantsValue) == 0 {
		return nil
	}
	names := make([]string, 0, len(wantsValue))
	for name := range wantsValue {
		names = append(names, name)
	}
	script := `use B ();` +
		`my ($module, @names) = @ARGV;` +
		`(my $file = "$module.pm") =~ s{::}{/}g; require $file;` +
		`eval "package Held::ConstPrime; use $module; 1";` +
		`sub callable { my ($name) = @_; return 0 if !exists &$name; return 1 if defined &$name; ` +
		`my $home = B::svref_2object(\&$name)->GV->STASH->NAME; return defined &{"${home}::AUTOLOAD"}; }` +
		`my $n = 0;` +
		`for my $name (@names) {` +
		`  my $code;` +
		`  if (callable("${module}::$name")) { $code = \&{"${module}::$name"}; }` +
		`  elsif (callable("Held::ConstPrime::$name")) { $code = \&{"Held::ConstPrime::$name"}; }` +
		`  if (!$code) {` +
		`    my $primed = "Held::ConstPrimeByName" . $n++;` +
		`    eval "package $primed; use $module qw($name); 1";` +
		`    if (callable("${module}::$name")) { $code = \&{"${module}::$name"}; }` +
		`    elsif (callable("${primed}::$name")) { $code = \&{"${primed}::$name"}; }` +
		`  }` +
		`  my $died = 1;` +
		`  $died = eval { $code->(); 1 } ? 0 : 1 if $code;` +
		`  print "$name\t", ($died ? "DIED" : "OK"), "\n";` +
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
		t.Fatalf("ConstantDifferences(%s): %v: %s", module, err, stderr.String())
	}
	diffs := map[string]string{}
	for _, line := range strings.Split(strings.TrimRight(string(out), "\r\n"), "\n") {
		line = strings.TrimRight(line, "\r")
		if line == "" {
			continue
		}
		name, result, ok := strings.Cut(line, "\t")
		if !ok {
			continue
		}
		died := result == "DIED"
		if wantsValue[name] != died {
			continue
		}
		if died {
			diffs[name] = fmt.Sprintf("%s's perl dies calling %s::%s, which this .tpm (describing Linux perl 5.42.0) declares with a value -- measured directly, not hand-listed", p.Path, module, name)
		} else {
			diffs[name] = fmt.Sprintf("%s's perl returns a value calling %s::%s, which this .tpm (describing Linux perl 5.42.0) declares None -- measured directly, not hand-listed", p.Path, module, name)
		}
	}
	return diffs
}
