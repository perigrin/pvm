// ABOUTME: Tests that CORE.pmt types its I/O, fixed-length data, file and directory builtins.
// ABOUTME: The batch is asked of perl 5.42's Pod::Functions, never kept by hand.
package parse_test

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"testing"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/types"
)

// ioFileBuiltins is perlfunc's "Input and output functions", "Functions for
// fixed-length data or records" and "Functions for filehandles, files, or
// directories" among CORE.pmt's prototyped lines, without the names the
// string, numeric, regex, array, list and hash batches hold: pack is a
// string builtin and unpack a list one.
func ioFileBuiltins(t *testing.T) []string {
	t.Helper()
	others := perlFunctionKinds(t, "String", "Regexp", "Math", "ARRAY", "LIST", "HASH")
	names := slices.DeleteFunc(perlFunctionKinds(t, "I/O", "Binary", "File"),
		func(name string) bool { return slices.Contains(others, name) })
	// Measured on 5.42.0: an empty or short answer is perl not being asked.
	if len(names) != 47 {
		t.Fatalf("Pod::Functions gives %d prototyped I/O, fixed-length data, file and directory builtins, %v; measured 47", len(names), names)
	}
	return names
}

// TestCoreIOFileBuiltinsTyped: every I/O, fixed-length data, file and
// directory builtin has a typed CORE.pmt line.
func TestCoreIOFileBuiltinsTyped(t *testing.T) {
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range untypedLines(src, ioFileBuiltins(t)) {
		t.Errorf("%s: CORE.pmt declares it untyped", name)
	}
}

// slotFaults holds each of names' typed CORE.pmt candidates to perl's
// prototype, slot by slot, and reports what fault says of each slot. A
// prototype-only line is reported too, since it has the coarse signature
// its prototype gives rather than types of its own.
func slotFaults(src []byte, names []string, perl map[string]string, fault func(slot string, optional bool, p types.Param) string) []string {
	sigs, err := parse.CoreSignaturesOf(src)
	if err != nil {
		return []string{err.Error()}
	}
	var bad []string
	for _, name := range names {
		if len(sigs[name]) == 0 || sigs[name][0].Returns == types.Unknown {
			bad = append(bad, name+": untyped")
			continue
		}
		for _, sig := range sigs[name] {
			i, optional, proto := 0, false, perl[name]
			for j := 0; j < len(proto); j++ {
				slot := proto[j : j+1]
				switch {
				case slot == ";":
					optional = true
					continue
				case slot == `\`:
					slot, j = proto[j:j+2], j+1
				}
				if i >= len(sig.Params) {
					bad = append(bad, fmt.Sprintf("%s: no parameter for (%s)'s %s", name, proto, slot))
				} else if f := fault(slot, optional, sig.Params[i]); f != "" {
					bad = append(bad, fmt.Sprintf("%s: (%s)'s %s %s", name, proto, slot, f))
				}
				i++
			}
		}
	}
	return bad
}

// starFault reports a `*` slot's parameter that is not `Glob *`: RFC 0001's
// table gives that slot, a bareword handle or any scalar (measured on 5.42,
// `binmode STDOUT` and `binmode $fh` both compile), and it derives `*`,
// never `\*` or `$`.
func starFault(slot string, _ bool, p types.Param) string {
	if slot == "*" && (p.Sigil != '*' || p.Alias || p.Type != types.Glob) {
		return fmt.Sprintf("is %v %c%s", p.Type, p.Sigil, p.Name)
	}
	return ""
}

// TestCoreStarSlotDerivesStar: every `*` slot of the batch is typed
// `Glob *`, so its derived prototype keeps the `*`.
func TestCoreStarSlotDerivesStar(t *testing.T) {
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range slotFaults(src, ioFileBuiltins(t), perlPrototypes(t), starFault) {
		t.Error(b)
	}
}

// optionalFault reports a parameter after perl's `;` that is required: an
// all-optional line, `getc` or `eof`'s `;*`, keeps its `;` only when its
// parameter is optional.
func optionalFault(_ string, optional bool, p types.Param) string {
	if optional && p.Required {
		return fmt.Sprintf("is required, %v %c%s", p.Type, p.Sigil, p.Name)
	}
	return ""
}

// TestCoreOptionalHandleKeepsSemicolon: every parameter the batch's
// prototypes make optional is typed optional, so its `;` is derived.
func TestCoreOptionalHandleKeepsSemicolon(t *testing.T) {
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	for _, b := range slotFaults(src, ioFileBuiltins(t), perlPrototypes(t), optionalFault) {
		t.Error(b)
	}
}

// TestCoreSlotChecksCatchMistypedLines: the slot checks are not
// tautological. A binmode whose handle is `Glob $fh` derives `$;$` and is
// reported as a star slot that is not one; a getc whose handle defaults to
// `die` is required, deriving `*`, and is reported as losing its `;`.
func TestCoreSlotChecksCatchMistypedLines(t *testing.T) {
	src, err := os.ReadFile("declarations/CORE.pmt")
	if err != nil {
		t.Fatal(err)
	}
	perl := perlPrototypes(t)
	for _, c := range []struct {
		name, line string
		fault      func(string, bool, types.Param) string
		want       string
	}{
		{"binmode", `sub binmode (Glob $fh, Str $layer = ":raw") Boolean|Undef;`, starFault, "binmode: (*;$)'s * is Glob $fh"},
		{"getc", `sub getc (Glob *fh = die) Str|Undef;`, optionalFault, "getc: (;*)'s * is required, Glob *fh"},
	} {
		line := regexp.MustCompile(`(?m)^sub ` + c.name + ` .*;$`)
		if len(line.FindAll(src, -1)) != 1 {
			t.Fatalf("CORE.pmt has no one %s line to mistype", c.name)
		}
		mistyped := line.ReplaceAllLiteral(src, []byte(c.line))
		if bad := slotFaults(mistyped, []string{c.name}, perl, c.fault); !slices.Equal(bad, []string{c.want}) {
			t.Errorf("%s: reported %q; want %q", c.line, bad, c.want)
		}
	}
}
