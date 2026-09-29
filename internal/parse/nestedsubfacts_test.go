// ABOUTME: A named sub declared inside a bare block of a required helper is still
// ABOUTME: the helper's: test.pl declares `watchdog` inside a closure block.

package parse_test

import (
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/parse"
)

// TestRequiredHelperNestedSub: readModule looked only at top-level
// statements, and t/test.pl declares `sub watchdog ($;$)` inside `{ #
// Closure ... }`. A named sub is package-global wherever it is declared;
// a lexical one is not. Measured on 5.42.0:
//
//	$ perl -e '{ my $x; sub watchdog ($;$) { "wd@_" } } print watchdog 3;
//	      { my sub lex { 1 } } print defined(&lex) ? "visible" : "not visible";'
//	wd3not visible
//
// perl.git t/mro/package_aliases.t:299, `watchdog 3;`.
func TestRequiredHelperNestedSub(t *testing.T) {
	dir := writeModules(t, map[string]string{
		"helper.pl": "{ # Closure\n  my $state;\n  sub watchdog ($;$) { 1 }\n  my sub lex { 1 }\n}\n1;\n",
		"main.pl":   "require './helper.pl';\nwatchdog 3;\nwatchdog 0;\n",
	})
	root, err := parse.ParseFileFrom(filepath.Join(dir, "main.pl"), dir)
	if err != nil {
		t.Fatal(err)
	}
	if containsKind(root, parse.Unknown) {
		t.Errorf("watchdog from the helper's closure block must resolve; got %s", shape(root))
	}
}
