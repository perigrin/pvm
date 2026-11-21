// ABOUTME: Pins Map's whole output over the T1 corpus to hashes taken before Map was reworked for speed.
// ABOUTME: Under the committed table and under one-class tables drawn from the tdt absence claims, which decide glob and x alone.

package lex

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/chart/grammar"
	"tamarou.com/pvm/internal/lexer"
)

// mapHashes are Map's streams over T1 at dfd7f9a34, by table: "committed", or an absence claim's terminal, kind and text.
var mapHashes = map[string]string{
	"ASSIGNOP Word \"x=\"":            "b443d786dd8e6caf",
	"BAREWORD Number":                 "1d1d2b3ffd8b4516",
	"BAREWORD Quote":                  "1d1d2b3ffd8b4516",
	"BAREWORD Word":                   "139feef997d4ed6e",
	"BAREWORD Word \"ADJUST\"":        "1d1d2b3ffd8b4516",
	"BAREWORD Word \"BEGIN\"":         "50e8ad6e3319b538",
	"BAREWORD Word \"END\"":           "809c0150a9ee3536",
	"BAREWORD Word \"UNITCHECK\"":     "1d1d2b3ffd8b4516",
	"BAREWORD Word \"__CLASS__\"":     "1d1d2b3ffd8b4516",
	"BAREWORD Word \"__SUB__\"":       "55cc3d94172a9c87",
	"BAREWORD Word \"any\"":           "459134fdda54ce92",
	"BAREWORD Word \"catch\"":         "0a57426b2f2207cf",
	"BAREWORD Word \"default\"":       "027c790e0b6501b4",
	"BAREWORD Word \"fc\"":            "5d5ec9aa2ee4cb79",
	"BAREWORD Word \"field\"":         "3adea03a71473120",
	"BAREWORD Word \"finally\"":       "1d1d2b3ffd8b4516",
	"BAREWORD Word \"m\"":             "1d1d2b3ffd8b4516",
	"BAREWORD Word \"method\"":        "e3b079edaf93cec2",
	"BAREWORD Word \"qq\"":            "1d1d2b3ffd8b4516",
	"BAREWORD Word \"qw\"":            "1d1d2b3ffd8b4516",
	"BAREWORD Word \"s\"":             "1d1d2b3ffd8b4516",
	"BAREWORD Word \"tr\"":            "1d1d2b3ffd8b4516",
	"BAREWORD Word \"try\"":           "7aa092b9a5df1f24",
	"BAREWORD Word \"y\"":             "1d1d2b3ffd8b4516",
	"CHRELOP Word \"lt\"":             "15661817afab2987",
	"FUNC Word \"gethostbyaddr\"":     "1d1d2b3ffd8b4516",
	"FUNC Word \"semctl\"":            "1d1d2b3ffd8b4516",
	"FUNC0 Word \"break\"":            "1d1d2b3ffd8b4516",
	"FUNC0 Word \"getprotoent\"":      "1d1d2b3ffd8b4516",
	"FUNC0 Word \"setgrent\"":         "1d1d2b3ffd8b4516",
	"FUNC0SUB Word \"given\"":         "1d1d2b3ffd8b4516",
	"FUNC0SUB Word \"y\"":             "1d1d2b3ffd8b4516",
	"FUNC1 Word \"readdir\"":          "1d1d2b3ffd8b4516",
	"KW_DEFER Word \"defer\"":         "b7ee06fe4ff3c399",
	"KW_MY Word \"our\"":              "f5e7b01acf2986f5",
	"KW_SUB_named Word \"sub\"":       "964713355f2a31df",
	"LSTOP Word \"bind\"":             "1d1d2b3ffd8b4516",
	"LSTOP Word \"getprotobynumber\"": "1d1d2b3ffd8b4516",
	"LSTOP Word \"msgget\"":           "1d1d2b3ffd8b4516",
	"LSTOP Word \"msgrcv\"":           "1d1d2b3ffd8b4516",
	"LSTOP Word \"seek\"":             "7b53bffc2b903517",
	"LSTOP Word \"shmwrite\"":         "1d1d2b3ffd8b4516",
	"LSTOP Word \"socketpair\"":       "e7df4209adbb473e",
	"LSTOP Word \"unpack\"":           "c1a86180d1ef7412",
	"LSTOPSUB Word \"CHECK\"":         "1d1d2b3ffd8b4516",
	"LSTOPSUB Word \"default\"":       "959cd7c4273b9205",
	"LSTOPSUB Word \"s\"":             "1d1d2b3ffd8b4516",
	"LSTOPSUB Word \"try\"":           "7aa092b9a5df1f24",
	"LSTOPSUB Word \"x\"":             "1d1d2b3ffd8b4516",
	"METHCALL Word \"state\"":         "1d1d2b3ffd8b4516",
	"METHCALL Word \"x\"":             "1d1d2b3ffd8b4516",
	"METHCALL0 Word \"AUTOLOAD\"":     "1d1d2b3ffd8b4516",
	"METHCALL0 Word \"UNITCHECK\"":    "1d1d2b3ffd8b4516",
	"METHCALL0 Word \"all\"":          "1d1d2b3ffd8b4516",
	"METHCALL0 Word \"default\"":      "1d1d2b3ffd8b4516",
	"METHCALL0 Word \"finally\"":      "1d1d2b3ffd8b4516",
	"METHCALL0 Word \"qr\"":           "1d1d2b3ffd8b4516",
	"METHCALL0 Word \"qx\"":           "1d1d2b3ffd8b4516",
	"METHCALL0 Word \"when\"":         "1d1d2b3ffd8b4516",
	"METHCALL0 Word \"x\"":            "1d1d2b3ffd8b4516",
	"MULOP Word \"x\"":                "edb921415f2d9f83",
	"NOAMP Word \"break\"":            "1d1d2b3ffd8b4516",
	"NOAMP Word \"qw\"":               "1d1d2b3ffd8b4516",
	"PERLY_BRACE_OPEN Operator \"{\"": "d9b0829192f28c43",
	"PRIVATEREF Word \"finally\"":     "1d1d2b3ffd8b4516",
	"PRIVATEREF Word \"q\"":           "1d1d2b3ffd8b4516",
	"THING Quote":                     "81f06dbae34019b4",
	"UNIOP Operator \"-b\"":           "1d1d2b3ffd8b4516",
	"UNIOP Word \"chroot\"":           "1d1d2b3ffd8b4516",
	"UNIOP Word \"getc\"":             "1d1d2b3ffd8b4516",
	"UNIOP Word \"getgrnam\"":         "1d1d2b3ffd8b4516",
	"UNIOP Word \"lcfirst\"":          "1d1d2b3ffd8b4516",
	"UNIOP Word \"readpipe\"":         "1d1d2b3ffd8b4516",
	"UNIOP Word \"rewinddir\"":        "1d1d2b3ffd8b4516",
	"UNIOP Word \"study\"":            "1d1d2b3ffd8b4516",
	"UNIOPSUB Word \"CHECK\"":         "1d1d2b3ffd8b4516",
	"committed":                       "3101c22e150eabfb",
}

// mapHash hashes everything Map returns for each file in order: the int, and each token's index, set, kind and offsets,
// each body and each feature change.
func mapHash(files [][]byte, raws [][]lexer.Token, tt *grammar.TokenTable) string {
	h := sha256.New()
	var b []byte
	for i, src := range files {
		s, first := Map(src, raws[i], tt)
		b = binary.AppendVarint(b[:0], int64(first))
		for _, t := range s.Tokens {
			r := s.Raw[t.Index]
			b = binary.AppendVarint(b, int64(t.Index))
			b = binary.AppendVarint(b, int64(t.TerminalSet))
			b = binary.AppendVarint(b, int64(r.Kind))
			b = binary.AppendVarint(b, int64(r.Start))
			b = binary.AppendVarint(b, int64(r.End))
		}
		b = append(b, 0xff)
		for _, o := range s.Bodies {
			b = binary.AppendVarint(b, int64(o.Opener))
			b = binary.AppendVarint(b, int64(o.Kind))
			b = binary.AppendVarint(b, int64(o.Start))
			b = binary.AppendVarint(b, int64(o.End))
		}
		b = append(b, 0xff)
		for _, f := range s.Features {
			b = binary.AppendVarint(b, int64(f.At))
			b = binary.AppendVarint(b, int64(f.On))
			b = binary.AppendVarint(b, int64(f.Maybe))
		}
		h.Write(b)
	}
	return fmt.Sprintf("%x", h.Sum(nil)[:8])
}

// TestMapCorpusUnchanged: Map's output over T1 is what it was before it was made faster, under the committed table and
// under a table of one class for each of every 16th absence claim in unseen-reviewed.txt (its terminal's first class of
// that kind and text, as falseAbsences compiles a claim alone).
func TestMapCorpusUnchanged(t *testing.T) {
	tt, _ := committed(t)
	g, _, err := grammar.Committed()
	if err != nil {
		t.Fatal(err)
	}
	t1 := os.Getenv("T1_CORPUS")
	if t1 == "" {
		t1 = "../../../testdata/perlonjava/unit"
	}
	paths, err := filepath.Glob(filepath.Join(t1, "*.t"))
	if err != nil || len(paths) == 0 {
		t.Fatalf("no T1 file under %s", t1)
	}
	var files [][]byte
	var raws [][]lexer.Token
	for _, p := range paths {
		src, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		files, raws = append(files, src), append(raws, lexer.Tokenize(src))
	}
	reviewed, err := os.ReadFile("../testdata/tdt/unseen-reviewed.txt")
	if err != nil {
		t.Fatal(err)
	}
	tables := map[string]*grammar.TokenTable{"committed": tt}
	absent, n := false, 0
	for _, l := range strings.Split(string(reviewed), "\n") {
		if r, ok := strings.CutPrefix(l, "reason:"); ok {
			absent = strings.Contains(r, "no corpus token matches the key")
			continue
		}
		if !absent || l == "" || strings.HasPrefix(l, "#") {
			continue
		}
		if n++; n%16 != 1 {
			continue
		}
		key, c := claimClass(t, g.Classes, l)
		if _, ok := tables[key]; ok {
			continue
		}
		c.Terminals = []int{0}
		if tables[key], err = grammar.CompileTokens([]grammar.TokenClass{c}); err != nil {
			t.Fatal(err)
		}
	}
	got := map[string]string{}
	for key, tt := range tables {
		got[key] = mapHash(files, raws, tt)
	}
	if len(got) < 20 {
		t.Fatalf("%d tables: the sample proves nothing", len(got))
	}
	if !maps.Equal(got, mapHashes) {
		var b strings.Builder
		for _, k := range slices.Sorted(maps.Keys(got)) {
			fmt.Fprintf(&b, "\t%q: %q,\n", k, got[k])
		}
		t.Errorf("Map's output over %d T1 files changed (or the claims drawn did); got:\n%s", len(files), b.String())
	}
}

// claimClass is the first class of a claim's terminal, kind and text, and those three as the claim's key.
func claimClass(t *testing.T, decls []grammar.ClassDecl, line string) (string, grammar.TokenClass) {
	t.Helper()
	term, rest, _ := strings.Cut(line, " ")
	kind, rest, _ := strings.Cut(strings.TrimPrefix(rest, "Kind="), " ")
	var text *string
	if q, ok := strings.CutPrefix(rest, "Text="); ok {
		q, err := strconv.QuotedPrefix(q)
		if err != nil {
			t.Fatalf("%s: %v", line, err)
		}
		s, _ := strconv.Unquote(q)
		text = &s
	}
	for _, d := range decls {
		if d.Terminal == term && d.Key.Kind != nil && d.Key.Kind.String() == kind && (text == nil) == (d.Key.Text == nil) &&
			(text == nil || *text == *d.Key.Text) {
			k := term + " " + kind
			if text != nil {
				k += " " + strconv.Quote(*text)
			}
			return k, d.Key
		}
	}
	t.Fatalf("no class for the claim %s", line)
	return "", grammar.TokenClass{}
}
