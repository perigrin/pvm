// ABOUTME: A module's XS source names its C-implemented subs and their prototypes,
// ABOUTME: read by xsubpp's rules; a loaded XS module's subs are then known.

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// sampleXS covers what xsubpp reads: the C preamble before the first MODULE
// line, PREFIX stripping, PROTOTYPES toggling, a PROTOTYPE line that
// overrides, defaults after a `;`, and `...` as `;@`. Shapes taken from
// ext/mro/mro.xs, ext/Devel-Peek/Peek.xs and dist/Time-HiRes/HiRes.xs.
const sampleXS = `#include "EXTERN.h"
static int helper(int x) { return x; }

MODULE = Foo::Bar		PACKAGE = Foo::Bar

void
quiet(sv)
    SV * sv

PROTOTYPES: ENABLE

NV
nap(useconds)
    NV useconds
    CODE:
        RETVAL = 0;
    OUTPUT:
        RETVAL

void
dump_it(sv,lim=4)
SV *	sv
I32	lim
PPCODE:
{
    helper(1);
}

void
many(a, ...)
    SV * a

void
fixed(a, b)
    PROTOTYPE: \%;$

MODULE = Foo::Bar		PACKAGE = Foo::Other		PREFIX = fo_

void
fo_isarev(...)
  PROTOTYPE: $

void
fo_gen()
  PROTOTYPE:
`

// TestXSSubs: measured against xsubpp (ExtUtils::ParseXS, Node.pm's
// proto_string): prototypes are off unless PROTOTYPES: ENABLE, each argument
// is `$`, a `;` goes before the first defaulted one and `...` adds `;@`, and
// a PROTOTYPE line wins. A sub with prototypes off has none -- a list
// operator, as `quiet` is here.
func TestXSSubs(t *testing.T) {
	want := map[string]string{
		"Foo::Bar::quiet":    "",
		"Foo::Bar::nap":      "($)",
		"Foo::Bar::dump_it":  "($;$)",
		"Foo::Bar::many":     "($;@)",
		"Foo::Bar::fixed":    `(\%;$)`,
		"Foo::Other::isarev": "($)",
		"Foo::Other::gen":    "()",
	}
	got := parse.XSSubs([]byte(sampleXS))
	for name, proto := range want {
		if p, ok := got[name]; !ok || p != proto {
			t.Errorf("%s: got %q (found %v), want %q", name, p, ok, proto)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d subs %v, want %d", len(got), got, len(want))
	}
}

// TestXSModuleLoads: a module laid out as its source tree has it --
// `Foo-Bar/Bar.pm` beside `Foo-Bar/Bar.xs`, where MakeMaker finds them --
// resolves, and its XS subs are known by qualified name and through its
// exports. perl.git t/io/eintr_print.t:61 (`Time::HiRes::usleep $big_delay;`),
// t/op/hash-clear-placeholders.t:26 (`Dump \%hash;`), t/mro/isarev.t:19
// (`mro::get_isarev $args[0]`).
func TestXSModuleLoads(t *testing.T) {
	root := t.TempDir()
	dist := filepath.Join(root, "Foo-Bar")
	if err := os.MkdirAll(dist, 0o755); err != nil {
		t.Fatal(err)
	}
	pm := "package Foo::Bar;\nuse Exporter 'import';\n@EXPORT = qw(dump_it);\nrequire XSLoader;\nXSLoader::load();\n1;\n"
	for name, body := range map[string]string{"Bar.pm": pm, "Bar.xs": sampleXS} {
		if err := os.WriteFile(filepath.Join(dist, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	load := parse.DirLoader(root)
	for _, src := range []string{
		`use Foo::Bar; my %hash; dump_it \%hash;`,
		`use Foo::Bar; my $d; Foo::Bar::nap $d;`,
		`use Foo::Bar; my @args; my @x = @{Foo::Other::isarev $args[0]};`,
	} {
		n := parse.ParseWithLoader([]byte(src), load)
		if got := countUnknown(n); got != 0 {
			t.Errorf("%q: %d Unknown, want 0; got %s", src, got, shape(n))
		}
	}
	if imp, ok := parse.Imports(parse.ParseWithLoader([]byte(`use Foo::Bar;`), load))["dump_it"]; !ok ||
		!imp.PrototypeKnown || imp.Prototype != "($;$)" {
		t.Errorf("dump_it imported as %+v, want prototype ($;$) known", imp)
	}
}
