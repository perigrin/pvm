// ABOUTME: Tests for parser-state snapshots: a parse resumed at any re-parse point equals the full parse,
// ABOUTME: every parser field is classified, snapshots are not aliased, and none sits inside BEGIN, use or format.
package parse

import (
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"tamarou.com/pvm/internal/types"
)

// incrementalSeedSources are the sources of chapter 7 §7.7.3's edit rows,
// before the edit: each holds the construct its row's edit breaks.
var incrementalSeedSources = map[string]string{
	"quote":      "my $s = 'a';\nsub f { my $x = 1; return $x }\nprint f(), \"\\n\";\n",
	"heredoc":    "sub f {\n    my $t = <<EOF;\nbody\nEOF\n    return $t;\n}\nprint f();\n",
	"pod":        "sub f { 1 }\n\n=pod\n\ntext\n\n=cut\n\nsub g { f() }\nprint g();\n",
	"end":        "sub f { 1 }\nprint f();\n__END__\ndata\n",
	"prototype":  "sub f() { 1 }\nsub g { f + 1 }\nmy @a = (f, 2);\nprint g();\n",
	"brace":      "sub f { if ($_[0]) { return 1 } return 0 }\nsub g { { my $x = f(1); } }\nprint g();\n",
	"splittoken": "sub f { my $foo = 1; $foo + 2 }\nprint f();\n",
}

// unfinished parses src with checkpoints and without finish's EOF passes,
// so a prefix of its children is what a resumed parse is spliced after.
func unfinished(src []byte, res *resolver) (*Node, []parserSnapshot) {
	p := newParser(src, res, false)
	p.snaps = &[]parserSnapshot{}
	return p.parseFile(), *p.snaps
}

// spliceTop is a SourceFile over src holding prefix's top-level children
// that start before at, then tail, with the EOF passes run by p.
func spliceTop(p *parser, prefix *Node, at int, tail []*Node) *Node {
	root := &Node{Kind: SourceFile, Start: 0, End: len(p.src)}
	for _, c := range prefix.Children {
		if c.Start < at {
			root.Children = append(root.Children, c)
		}
	}
	root.Children = append(root.Children, tail...)
	p.finish(root)
	return root
}

func TestParserStateResumeEqualsFullParse(t *testing.T) {
	sources := maps.Clone(incrementalSeedSources)
	edit5k, err := os.ReadFile(filepath.Join("..", "psc", "testdata", "latency", "edit5k.pl"))
	if err != nil {
		t.Fatal(err)
	}
	sources["edit5k.pl"] = string(edit5k)
	for name, s := range sources {
		src := []byte(s)
		full := Parse(src)
		prefix, snaps := unfinished(src, nil)
		var top []parserSnapshot
		for _, snap := range snaps {
			if !snap.nested {
				top = append(top, snap)
			}
		}
		if len(top) != len(full.Children) {
			t.Fatalf("%s: %d top-level snapshots for %d top-level children", name, len(top), len(full.Children))
		}
		// Resuming at k gives the full parse's children from k on when one
		// statement read from k equals child k and leaves the parser in
		// snapshot k+1's state; by induction over k that is every boundary,
		// in linear time. A resume to EOF is quadratic over edit5k.pl, so it
		// runs at three boundaries and, for the small seeds, at every one.
		for k, snap := range top {
			// The snapshot's own tokens, which are src's: restore would
			// re-lex the file at every boundary.
			p := new(parser)
			*p = snap.p.clone()
			if d := TreeDiff(full.Children[k], p.statement()); d != "" {
				t.Fatalf("%s: statement resumed at byte %d: %s", name, snap.at, d)
			}
			if k+1 < len(top) && !sameState(p, &top[k+1].p) {
				t.Fatalf("%s: the statement at byte %d leaves a state unlike the snapshot after it", name, snap.at)
			}
			if len(src) > 1000 && k != 0 && k != len(top)/2 && k != len(top)-1 {
				continue
			}
			p, tail := snap.resume(src, nil)
			if d := TreeDiff(full, spliceTop(p, prefix, snap.at, tail)); d != "" {
				t.Fatalf("%s: resumed at byte %d: %s", name, snap.at, d)
			}
		}
	}
}

// sameState reports whether a and b agree on every field parserFields
// does not mark shared.
func sameState(a, b *parser) bool {
	ca, cb := a.clone(), b.clone()
	for _, c := range []*parser{&ca, &cb} {
		c.src, c.toks, c.bodies, c.res = nil, nil, nil, nil
	}
	return reflect.DeepEqual(ca, cb)
}

// blockAt is the Block under n with a direct child starting at at.
func blockAt(n *Node, at int) *Node {
	for _, c := range n.Children {
		if c.Kind == Block {
			for _, g := range c.Children {
				if g.Start == at {
					return c
				}
			}
		}
		if b := blockAt(c, at); b != nil {
			return b
		}
	}
	return nil
}

// resumeNested resumes src at the nested snapshot before the statement
// starting with stmt, splices the result into a checkpointed parse's
// block, runs the EOF passes and returns that tree.
func resumeNested(t *testing.T, src []byte, stmt string) *Node {
	t.Helper()
	at := strings.Index(string(src), stmt)
	root, snaps := unfinished(src, nil)
	for _, snap := range snaps {
		if snap.at != at {
			continue
		}
		if !snap.nested {
			t.Fatalf("snapshot at %q is not nested", stmt)
		}
		p, tail := snap.resume(src, nil)
		blk := blockAt(root, at)
		if blk == nil {
			t.Fatalf("no block holds %q", stmt)
		}
		var kept []*Node
		for _, c := range blk.Children {
			if c.Start < at {
				kept = append(kept, c)
			}
		}
		kept = append(kept, tail...)
		blk.Children = kept
		p.finish(root)
		return root
	}
	t.Fatalf("no snapshot at %q", stmt)
	return nil
}

func TestParserStateResumeAtNestedSubAnchor(t *testing.T) {
	src := []byte("package Foo {\n    sub a($) { 1 }\n    sub b { a 1, 2 }\n}\n{\n    my $x = 1;\n    sub c { $x }\n    sub d { c() + 1 }\n}\nprint Foo::b(), d(), \"\\n\";\n")
	full := Parse(src)
	for _, stmt := range []string{"sub b", "sub d", "my $x"} {
		if d := TreeDiff(full, resumeNested(t, src, stmt)); d != "" {
			t.Errorf("resumed at %q: %s", stmt, d)
		}
	}
	// A sub body is not an anchor: nothing is snapshotted inside one.
	_, snaps := unfinished(src, nil)
	if in := strings.Index(string(src), "1 }"); snapshotAt(snaps, in) {
		t.Errorf("a snapshot sits inside a sub body")
	}
}

func snapshotAt(snaps []parserSnapshot, at int) bool {
	for _, s := range snaps {
		if s.at == at {
			return true
		}
	}
	return false
}

// TestParserStateResumeKeepsHasError: a parse resumed at any top-level
// boundary of a buffer holding Error and Missing nodes, spliced and
// finished, equals the full parse, HasError included -- finish flags the
// spliced tree, and a clean prefix spliced to a clean tail is not flagged.
func TestParserStateResumeKeepsHasError(t *testing.T) {
	for _, s := range []string{
		"my $a = 1;\nmy $x = ;\nsub f { $a $b; }\nprint f(\n",
		"my $a = 1;\nmy $x = 1\nmy $y = 2;\nprint $\n",
		"my $a = 1;\nprint 2;\n",
	} {
		src := []byte(s)
		full := Parse(src)
		prefix, snaps := unfinished(src, nil)
		for _, snap := range snaps {
			if snap.nested {
				continue
			}
			p, tail := snap.resume(src, nil)
			if d := TreeDiff(full, spliceTop(p, prefix, snap.at, tail)); d != "" {
				t.Errorf("%q: resumed at byte %d: %s", s, snap.at, d)
			}
		}
	}
}

func TestParserStateSnapshotCoversEveryField(t *testing.T) {
	typ := reflect.TypeFor[parser]()
	seen := map[string]bool{}
	for i := range typ.NumField() {
		f := typ.Field(i)
		seen[f.Name] = true
		if _, ok := parserFields[f.Name]; !ok {
			t.Errorf("parser field %s is not classified in parserFields", f.Name)
		}
	}
	for name := range parserFields {
		if !seen[name] {
			t.Errorf("parserFields classifies %s, which parser does not have", name)
		}
	}

	// Every snapshotted map and slice is copied, not aliased.
	var p parser
	v := reflect.ValueOf(&p).Elem()
	for i := range typ.NumField() {
		f := v.Field(i)
		switch f.Kind() {
		case reflect.Map:
			m := reflect.MakeMap(f.Type())
			val := reflect.Zero(f.Type().Elem())
			if f.Type().Elem().Kind() == reflect.Slice {
				// Spare capacity, which an append in place would share.
				val = reflect.MakeSlice(f.Type().Elem(), 1, 2)
			}
			m.SetMapIndex(reflect.Zero(f.Type().Key()), val)
			reflect.NewAt(f.Type(), f.Addr().UnsafePointer()).Elem().Set(m)
		case reflect.Slice:
			reflect.NewAt(f.Type(), f.Addr().UnsafePointer()).Elem().Set(reflect.MakeSlice(f.Type(), 1, 1))
		}
	}
	c := p.clone()
	cv := reflect.ValueOf(&c).Elem()
	for i := range typ.NumField() {
		name := typ.Field(i).Name
		if parserFields[name] != snapshotted {
			continue
		}
		if k := v.Field(i).Kind(); (k == reflect.Map || k == reflect.Slice) && v.Field(i).Pointer() == cv.Field(i).Pointer() {
			t.Errorf("snapshot aliases parser field %s", name)
		}
		// A map's slices are a second level: a clone sharing one's spare
		// capacity lets two resumes append into the same element.
		if f := v.Field(i); f.Kind() == reflect.Map && f.Type().Elem().Kind() == reflect.Slice {
			key := reflect.Zero(f.Type().Key())
			orig, cl := f.MapIndex(key), cv.Field(i).MapIndex(key)
			if cl.Cap() > cl.Len() && cl.Pointer() == orig.Pointer() {
				t.Errorf("snapshot shares spare capacity in parser field %s's slices", name)
			}
		}
	}

	// Scratch is zero at every re-parse point, which is what makes it scratch.
	_, snaps := unfinished([]byte(incrementalSeedSources["brace"]+"package P { sub x { 1 } }\n{ my $y; }\n"), nil)
	for _, s := range snaps {
		sv := reflect.ValueOf(&s.p).Elem()
		for i := range typ.NumField() {
			if name := typ.Field(i).Name; parserFields[name] == scratch && !sv.Field(i).IsZero() {
				t.Errorf("scratch field %s is set at the re-parse point at byte %d", name, s.at)
			}
		}
	}
}

func TestParserStateEOFPassesRerun(t *testing.T) {
	old := "sub f($) { 1 }\nsub g { f 2 }\nsub h(\\@) { 3 }\nprint g(), \"\\n\";\n"
	at := strings.Index(old, "sub g")
	edited := []byte(old[:at+len("sub g {")] + " {" + old[at+len("sub g {"):])
	res := func() *resolver { return &resolver{load: DirLoader(t.TempDir()), seen: map[string]bool{}} }

	full := parseRoot(edited, res())
	prefix, snaps := unfinished([]byte(old), res())
	for _, snap := range snaps {
		if snap.at != at {
			continue
		}
		p, tail := snap.resume(edited, res())
		spliced := spliceTop(p, prefix, at, tail)
		if !hasUnclosedOpener(full) {
			t.Fatal("the full parse of the edited source has no Unknown{UnclosedOpener}")
		}
		if d := childrenDiff(full, spliced); d != "" {
			t.Errorf("spliced tree: %s", d)
		}
		if !reflect.DeepEqual(readModule(full).protos, readModule(spliced).protos) {
			t.Errorf("prototypes: %v != %v", readModule(spliced).protos, readModule(full).protos)
		}
		if !reflect.DeepEqual(full.imports, spliced.imports) {
			t.Errorf("imports: %v != %v", spliced.imports, full.imports)
		}
		return
	}
	t.Fatal("no snapshot at sub g")
}

// childrenDiff is TreeDiff over two roots' children alone: a resumed parse's
// fresh resolver has not loaded what the prefix loaded, so the root's
// loaded list is its own (spec 06 §6.5.3).
func childrenDiff(a, b *Node) string {
	return TreeDiff(&Node{Children: a.Children}, &Node{Children: b.Children})
}

func hasUnclosedOpener(n *Node) bool {
	if n.Kind == Unknown && n.Refusal == UnclosedOpener {
		return true
	}
	for _, c := range n.Children {
		if hasUnclosedOpener(c) {
			return true
		}
	}
	return false
}

func TestParserStateResumeAfterPackageUseAndLexicalSub(t *testing.T) {
	lib := t.TempDir()
	if err := os.WriteFile(filepath.Join(lib, "Mod.pm"), []byte("package Mod;\nuse Exporter 'import';\nour @EXPORT = ('m1');\nsub m1($) { 1 }\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := func() *resolver { return &resolver{load: DirLoader(lib), seen: map[string]bool{}} }
	src := []byte("package Foo;\nuse strict; use feature qw(signatures fc);\nuse Mod;\nmy sub lex($x) { $x }\nm1 1, 2;\nlex 3;\nprint fc 'A';\nsub s2($a, $b) { $a }\n")
	at := strings.Index(string(src), "m1 1")

	full := parseRoot(src, res())
	prefix, snaps := unfinished(src, res())
	for _, snap := range snaps {
		if snap.at != at {
			continue
		}
		st := snap.p
		if st.pkg != "Foo" || !st.features["fc"] {
			t.Errorf("snapshot has package %q, features %v", st.pkg, st.features)
		}
		if _, ok := st.imports["Foo::m1"]; !ok {
			t.Errorf("snapshot imports %v lack m1", st.imports)
		}
		if _, ok := st.lexical["lex"]; !ok {
			t.Errorf("snapshot lexical subs %v lack lex", st.lexical)
		}
		p, tail := snap.resume(src, res())
		spliced := spliceTop(p, prefix, at, tail)
		if d := childrenDiff(full, spliced); d != "" {
			t.Errorf("resumed: %s", d)
		}
		if !reflect.DeepEqual(full.imports, spliced.imports) {
			t.Errorf("imports: %v != %v", spliced.imports, full.imports)
		}

		// The check can fail: the same position with none of that state
		// restored reads the next statements differently.
		bare := newParser(src, res(), false)
		bare.pos = st.pos
		if childrenDiff(full, spliceTop(bare, prefix, at, bare.statementsTo(false, false))) == "" {
			t.Error("a parse with no restored state matched too: the test cannot tell")
		}
		return
	}
	t.Fatal("no snapshot before m1")
}

func TestParserStateSnapshotNotAliased(t *testing.T) {
	src := []byte("package Foo;\nuse feature 'say';\nmy sub l { 1 }\nsub f { l() }\npackage Bar;\nsub g { 2 }\nsay f(), g();\n")
	_, snaps := unfinished(src, nil)
	var snap parserSnapshot
	for _, s := range snaps {
		if s.at == strings.Index(string(src), "sub f") {
			snap = s
		}
	}
	before := snap.p.clone()
	_, first := snap.resume(src, nil)
	_, second := snap.resume(src, nil)
	if len(first) != len(second) {
		t.Fatalf("%d statements then %d", len(first), len(second))
	}
	for i := range first {
		if d := TreeDiff(first[i], second[i]); d != "" {
			t.Errorf("statement %d: %s", i, d)
		}
	}
	if !reflect.DeepEqual(before, snap.p) {
		t.Error("resuming changed the snapshot")
	}
	if _, ok := snap.p.lexical["l"]; len(first) == 0 || snap.p.pkg != "Foo" || !ok {
		t.Errorf("the snapshot at sub f resumed %d statements in package %q with lexical subs %v", len(first), snap.p.pkg, snap.p.lexical)
	}
}

// Two resumes from one snapshot each append a signature to a name that
// already has some, as parseTypedSignature does; neither may see the other's.
func TestParserStateSnapshotSignaturesNotShared(t *testing.T) {
	var p parser
	p.signatures = map[string][]types.Signature{"f": make([]types.Signature, 3, 4)}
	snap := parserSnapshot{p: p.clone()}
	a, b := snap.p.clone(), snap.p.clone()
	sigA, sigB := types.Signature{Invocant: &types.Param{}}, types.Signature{Invocant: &types.Param{}}
	a.signatures["f"] = append(a.signatures["f"], sigA)
	b.signatures["f"] = append(b.signatures["f"], sigB)
	if a.signatures["f"][3].Invocant != sigA.Invocant {
		t.Error("resume A sees resume B's signature")
	}
	if b.signatures["f"][3].Invocant != sigB.Invocant {
		t.Error("resume B sees resume A's signature")
	}
	if len(snap.p.signatures["f"]) != 3 {
		t.Errorf("the snapshot holds %d signatures for f, not 3", len(snap.p.signatures["f"]))
	}
}

func TestParserStateNoResumeInsideBEGINUseOrFormat(t *testing.T) {
	src := "BEGIN { my $x = 1; our $y = $x; }\nuse constant { A => 1, B => 2 };\nformat STDOUT =\n@<<<\n$y\n.\n{ BEGIN { my $z; } use constant C => 3; }\nprint A;\n"
	_, snaps := unfinished([]byte(src), nil)
	spans := [][2]string{{"BEGIN {", "}\n"}, {"use constant {", ";"}, {"format", "\n.\n"}, {"BEGIN { my $z", "} use"}, {"use constant C", ";"}}
	for _, sp := range spans {
		start := strings.Index(src, sp[0])
		end := start + strings.Index(src[start:], sp[1])
		for _, s := range snaps {
			if s.at > start && s.at < end {
				t.Errorf("a snapshot at byte %d is inside %q", s.at, src[start:end])
			}
		}
	}
	// The statements themselves are re-parse points.
	for _, stmt := range []string{"BEGIN {", "use constant {", "format", "print A"} {
		if !snapshotAt(snaps, strings.Index(src, stmt)) {
			t.Errorf("no snapshot at %q", stmt)
		}
	}
}
