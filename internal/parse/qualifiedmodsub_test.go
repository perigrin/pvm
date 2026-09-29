// ABOUTME: A loaded module's declared subs are known by their qualified names:
// ABOUTME: after `use overload;`, `overload::constant 'integer' => sub {...}` is a call.

package parse_test

import (
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestQualifiedModuleSub: a module's subs are callable as MODULE::name
// whether or not they are exported, and the qualified name was unknown, so
// `overload::constant 'integer' => sub {...}` refused. Measured on 5.42.0,
// `use overload; BEGIN { overload::constant "integer" => sub { return shift } }`
// compiles, and lib/overload.pm declares `sub constant` in package overload.
//
// perl.git t/op/overload_integer.t:20 and t/re/overload.t:65, 247. The second
// loads overload only with a computed import list, `use overload '""' =>
// sub {...}`, which imports nothing statically readable and still loads
// the module.
func TestQualifiedModuleSub(t *testing.T) {
	dir := writeModules(t, map[string]string{
		"Ovl.pm":      "package Ovl;\nsub constant { 1 }\n1;\n",
		"main.pl":     "use Ovl;\nOvl::constant 'integer' => sub { return shift };\n",
		"computed.pl": "use Ovl '\"\"' => sub { 1 };\nOvl::constant qr => sub { 1 };\n",
	})
	for _, file := range []string{"main.pl", "computed.pl"} {
		root, err := parse.ParseFile(filepath.Join(dir, file))
		if err != nil {
			t.Fatal(err)
		}
		if containsKind(root, parse.Unknown) {
			t.Errorf("%s: Ovl::constant is the loaded module's sub; got %s", file, shape(root))
		}
	}
}
