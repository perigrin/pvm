// ABOUTME: Exporter's %EXPORT_TAGS: an export list built from a tag, and `:tag` in an
// ABOUTME: import list, name the subs the tag lists.

package parse_test

import (
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestExportTags: perl 5.42.0's Hash/Util/FieldHash.pm declares
//
//	our %EXPORT_TAGS = ( 'all' => [ qw( fieldhash ... ) ] );
//	our @EXPORT_OK = ( @{ $EXPORT_TAGS{'all'} } );
//
// and `use Hash::Util::FieldHash qw(fieldhash id)` imports both, with
// fieldhash's `(\%)` prototype, so `fieldhash my %fields;` is
// `fieldhash(my %fields)`. The export list was read as computed and nothing
// was imported. `:tag` in an import list names the tag's subs, as Exporter
// defines it. PerlOnJava unit/hash_util_fieldhash_identity.t:7.
func TestExportTags(t *testing.T) {
	lib := t.TempDir()
	if err := os.MkdirAll(filepath.Join(lib, "Foo"), 0o755); err != nil {
		t.Fatal(err)
	}
	pm := "package Foo::Tags;\nuse Exporter 'import';\n" +
		"our %EXPORT_TAGS = ( 'all' => [ qw( fieldhash id ) ], extra => ['more'] );\n" +
		"our @EXPORT_OK = ( @{ $EXPORT_TAGS{'all'} }, @{ $EXPORT_TAGS{extra} } );\n" +
		"sub fieldhash (\\%) { } sub id ($) { } sub more { }\n1;\n"
	if err := os.WriteFile(filepath.Join(lib, "Foo", "Tags.pm"), []byte(pm), 0o644); err != nil {
		t.Fatal(err)
	}
	load := parse.DirLoader(lib)
	for src, names := range map[string][]string{
		"use Foo::Tags qw(fieldhash id); fieldhash my %fields;": {"fieldhash", "id"},
		"use Foo::Tags ':all';":                                 {"fieldhash", "id"},
		"use Foo::Tags qw(:extra);":                             {"more"},
	} {
		root := parse.ParseWithLoader([]byte(src), load)
		if countUnknown(root) != 0 {
			t.Errorf("%q: %d Unknown; got %s", src, countUnknown(root), shape(root))
		}
		imps := parse.Imports(root)
		for _, name := range names {
			if imp, ok := imps[name]; !ok || !imp.PrototypeKnown {
				t.Errorf("%q: %s imported as %+v (found %v)", src, name, imp, ok)
			}
		}
	}
}

// TestPragmaTagIsNotAnExport: `use open qw( :utf8 :std )` names layers, not
// Exporter tags, and leaves the sub table as complete as it was --
// perl.git t/uni/gv.t:15 still reads `s2 $f` as `$f->s2` after it.
func TestPragmaTagIsNotAnExport(t *testing.T) {
	lib := t.TempDir()
	if err := os.WriteFile(filepath.Join(lib, "open.pm"),
		[]byte("package open;\nsub import { }\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := "use open qw( :utf8 :std ); my $f; s2 $f;"
	if indirectCall(parse.ParseWithLoader([]byte(src), parse.DirLoader(lib)), "s2") == nil {
		t.Errorf("%q: `s2 $f` is `$f->s2` with the table complete", src)
	}
}
