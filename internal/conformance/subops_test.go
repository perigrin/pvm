// ABOUTME: Tests that opsOf sees inside subroutine bodies, not the main program alone.
// ABOUTME: A sub body is a separate CV, so a lint blind to it cannot claim the ops that live there.
package conformance

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// TestOpsOfSeesInsideNamedSub is the measurement issue 01a0c547-516b was
// filed against.
//
// `perl -MO=Concise,-exec file.pl` with no sub named prints THE MAIN
// PROGRAM ALONE. A signature compiles `argcheck`, `argdefelem` and
// `argelem` and a body ends in `leavesub`, and every one of those lives
// in the sub's CV -- so a file carrying the whole of tier 07's subject
// compiles a visible optree of `anoncode` and `entersub` and nothing
// else. Tier 07 dropped from seven claimed ops to two on that, and tier
// 11 lost `methstart`, which opens every `method` body.
func TestOpsOfSeesInsideNamedSub(t *testing.T) {
	src := "use v5.42;\nsub add($x, $y = 10) { return $x + $y }\nsay add(1);\n"

	ops, err := opsOf(t, src)
	if err != nil {
		t.Fatalf("collecting ops: %v", err)
	}

	// Measured from `perl -MO=Concise,-exec,add` on 5.42.0: the
	// signature's three ops and the body's terminator, none of which the
	// main optree carries.
	for _, want := range []string{"argcheck", "argdefelem", "argelem", "leavesub"} {
		if !slices.Contains(ops, want) {
			t.Errorf("opsOf did not see %q, which `sub add($x, $y = 10)` compiles into its CV\n"+
				"\tgot: %s", want, strings.Join(ops, " "))
		}
	}
}

// TestOpsOfSeesInsideMethod is tier 11's half of the same bug.
//
// `methstart` opens every `method` body under the `class` feature, and a
// method body is a CV like any other. No file could be written that put
// it where the lint looked.
func TestOpsOfSeesInsideMethod(t *testing.T) {
	src := "use v5.42;\nuse experimental 'class';\n" +
		"class Point { field $x :param = 0; method show { return $x } }\n" +
		"say Point->new(x => 3)->show;\n"

	ops, err := opsOf(t, src)
	if err != nil {
		t.Fatalf("collecting ops: %v", err)
	}
	if !slices.Contains(ops, "methstart") {
		t.Errorf("opsOf did not see methstart, which opens every method body\n\tgot: %s",
			strings.Join(ops, " "))
	}
}

// TestOpsOfSeesSubsInOtherPackages pins the part a sub-name enumeration
// gets wrong by default.
//
// B::Concise takes sub names, and the name of a sub declared after
// `package Foo;` is `Foo::bar`, not `bar`. An enumeration that assumed
// `main::` would find nothing in a file whose subs all live in a package
// -- which is most of tiers 11 and 12.
func TestOpsOfSeesSubsInOtherPackages(t *testing.T) {
	src := "use v5.42;\npackage Foo;\nsub bar { my $s = shift; return $s }\n" +
		"package main;\nsay Foo::bar(1);\n"

	ops, err := opsOf(t, src)
	if err != nil {
		t.Fatalf("collecting ops: %v", err)
	}
	if !slices.Contains(ops, "leavesub") {
		t.Errorf("opsOf did not see inside Foo::bar\n\tgot: %s", strings.Join(ops, " "))
	}
}

// TestOpsOfStopsAtTheFileBoundary is the check that keeps the widening
// from swallowing the standard library.
//
// Every CV the process has loaded is reachable from the symbol table:
// `strict::import`, `Exporter::import`, the whole of `warnings`. Dumping
// those would put their ops into the measured set for EVERY corpus file,
// so a tier would appear to emit ops no line of it wrote and the lint
// would report them as unclaimed forever. The enumeration is therefore
// filtered to CVs whose `FILE` is the file under test.
func TestOpsOfStopsAtTheFileBoundary(t *testing.T) {
	// `use strict` alone drags in Exporter, whose `import` compiles
	// `goto` -- an op nothing in this source contains.
	src := "use strict;\nuse warnings;\nmy $x = 1;\nprint \"$x\\n\";\n"

	ops, err := opsOf(t, src)
	if err != nil {
		t.Fatalf("collecting ops: %v", err)
	}
	for _, unwanted := range []string{"goto", "require", "dofile"} {
		if slices.Contains(ops, unwanted) {
			t.Errorf("opsOf leaked %q from a module's CV into the file's op set\n\tgot: %s",
				unwanted, strings.Join(ops, " "))
		}
	}
}

// TestOpsOfStillSeesTheMainProgram is the regression the widening could
// cause silently.
//
// Thirteen tiers declared their op sets against the main optree. If
// enumerating subs replaced the main dump rather than adding to it, every
// one of those sets would become wrong at once and the failure would look
// like a corpus problem rather than an instrument one.
func TestOpsOfStillSeesTheMainProgram(t *testing.T) {
	src := "use v5.42;\nsub f { return 1 }\nmy $x = f();\nprint \"$x\\n\";\n"

	ops, err := opsOf(t, src)
	if err != nil {
		t.Fatalf("collecting ops: %v", err)
	}
	for _, want := range []string{"enter", "leave", "entersub", "padsv_store"} {
		if !slices.Contains(ops, want) {
			t.Errorf("opsOf lost %q from the main program\n\tgot: %s",
				want, strings.Join(ops, " "))
		}
	}
}

// TestOpsOfMutations proves the five checks above are load-bearing.
//
// A test that passes under the broken instrument is not a test, and this
// widening has two independent ways to be wrong: not looking inside subs
// at all, which is the bug, and looking too far -- past the file into
// every CV the process loaded, which would put Exporter's ops into every
// corpus file's measured set and is the failure mode hardest to notice,
// since it only WIDENS.
//
// Run as perl invocations against mutated copies of the backend rather
// than by editing opsOf, so the real function stays the one the suite
// uses.
func TestOpsOfMutations(t *testing.T) {
	perl, err := perlPath()
	if err != nil {
		t.Fatalf("%v", err)
	}

	const withSub = "use v5.42;\nsub add($x, $y = 10) { return $x + $y }\nsay add(1);\n"
	const withUse = "use strict;\nuse warnings;\nmy $x = 1;\nprint \"$x\\n\";\n"

	// run compiles source through a backend written beside it and returns
	// the ops, exactly as opsOf does.
	run := func(t *testing.T, source, backend string, args ...string) []string {
		t.Helper()
		dir := t.TempDir()
		path := filepath.Join(dir, "case.pl")
		if err := os.WriteFile(path, []byte(source), 0o600); err != nil {
			t.Fatalf("writing the case: %v", err)
		}
		if backend != "" {
			bp := filepath.Join(dir, "B", "Mutant.pm")
			if err := os.MkdirAll(filepath.Dir(bp), 0o750); err != nil {
				t.Fatalf("making room: %v", err)
			}
			if err := os.WriteFile(bp, []byte(backend), 0o600); err != nil {
				t.Fatalf("writing the backend: %v", err)
			}
		}
		argv := append([]string{"-I" + dir}, args...)
		argv = append(argv, path)
		out, err := exec.Command(perl, argv...).Output()
		if err != nil {
			t.Fatalf("perl refused the mutant: %v", err)
		}
		var ops []string
		for _, line := range strings.Split(string(out), "\n") {
			if m := reConciseOp.FindStringSubmatch(line); m != nil {
				ops = append(ops, m[1])
			}
		}
		return ops
	}

	// MUTANT 1: the instrument as it was -- no sub named, main program
	// alone. Must LOSE the four signature-and-body ops, or
	// TestOpsOfSeesInsideNamedSub is testing nothing.
	t.Run("narrow loses the sub body", func(t *testing.T) {
		ops := run(t, withSub, "", "-MO=Concise,-exec")
		for _, op := range []string{"argcheck", "argdefelem", "argelem", "leavesub"} {
			if slices.Contains(ops, op) {
				t.Errorf("the bare -MO=Concise,-exec already sees %q, so the widening "+
					"proves nothing\n\tgot: %s", op, strings.Join(ops, " "))
			}
		}
	})

	// MUTANT 2: the FILE filter dropped, so every CV in the process is
	// dumped. Must LEAK a module's ops, or TestOpsOfStopsAtTheFileBoundary
	// is asserting a property nothing threatens.
	t.Run("unfiltered leaks the standard library", func(t *testing.T) {
		unfiltered := strings.Replace(fileOpsBackend,
			"                next unless (eval { $cv->FILE } // '') eq $file;\n", "", 1)
		if unfiltered == fileOpsBackend {
			t.Fatal("the FILE filter line was not found, so this mutant is not a mutant")
		}
		unfiltered = strings.Replace(unfiltered, "package B::FileOps;", "package B::Mutant;", 1)

		ops := run(t, withUse, unfiltered, "-MO=Mutant")
		var leaked []string
		for _, op := range []string{"goto", "require", "dofile"} {
			if slices.Contains(ops, op) {
				leaked = append(leaked, op)
			}
		}
		if len(leaked) == 0 {
			t.Errorf("dropping the FILE filter leaked nothing, so the filter is not "+
				"what keeps the standard library out\n\tgot: %s", strings.Join(ops, " "))
		}
	})

	// MUTANT 3: `-main` dropped, so the sub dump REPLACES the main
	// program. Must lose the main program's ops, or
	// TestOpsOfStillSeesTheMainProgram is asserting nothing.
	t.Run("without -main the main program vanishes", func(t *testing.T) {
		noMain := strings.Replace(fileOpsBackend,
			"B::Concise::compile('-exec', '-main', @names)->();",
			"B::Concise::compile('-exec', @names)->();", 1)
		if noMain == fileOpsBackend {
			t.Fatal("the -main argument was not found, so this mutant is not a mutant")
		}
		noMain = strings.Replace(noMain, "package B::FileOps;", "package B::Mutant;", 1)

		ops := run(t, withSub, noMain, "-MO=Mutant")
		if slices.Contains(ops, "enter") {
			t.Errorf("dropping -main still dumped the main program, so `-main` is not "+
				"what keeps it\n\tgot: %s", strings.Join(ops, " "))
		}
	})
}
