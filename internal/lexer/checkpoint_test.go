// ABOUTME: Holds the lexer's checkpoints to the full stream: resuming from any one yields Tokenize's suffix.
// ABOUTME: Also the safety rule (no checkpoint inside a heredoc, POD, quote or format) and State's deep Equal/Clone.

package lexer

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// resumeFixtures are the constructs whose state outlives a token: a heredoc
// queue, POD, prototypes and signatures, quote-likes, formats, pragmas.
var resumeFixtures = map[string]string{
	"heredoc": "my $x = 1;\nprint <<A, <<'B';\nfirst $x\nA\nsecond\nB\nmy $y = 2;\n" +
		"print <<~EOT;\n    indented\n    EOT\nfoo();\n",
	"heredoc pending past a semicolon": "print <<A; foo; bar();\nbody;\nA\nbaz;\n",
	"heredoc in a substitution":        "$x =~ s/a/substr(<<E, 0, 0)/e; $y = 1;\nbody\nE\nz();\n",
	"pod":                              "my $a = 1;\n\n=pod\n\nsub f { ; }\n\n=cut\n\nmy $b = 2;\n=head1 X\n;\n=cut\n1;\n",
	"prototype": "sub f($$) { 1 }\nmy $x = $$;\nsub g(\\@;$) { }\nsub h :prototype($) { }\n" +
		"f(1, 2); g(@a);\n",
	"signatures": "use feature 'signatures';\nmy $x = 1;\nsub f ($y, $z = 2) { $y }\nf(1);\n",
	"quote": "s{a}{b}ge; my $q = q{x;y}; tr/a-z/A-Z/; my @w = qw(a ; b);\n" +
		"$s =~ m{;}x; my $r = qr/;/; print \"a;b\"; y/a/b/;\n",
	"format":       "my $a = 1;\nformat STDOUT =\n@<<<<< ; @>>>>\n$a, $b\n.\nwrite;\n",
	"blocks":       "if ($x) { y(); } else { z(); }\nmy %h = (a => 1); $h{b} = { c => 2 };\nfor (;;) { last; }\n",
	"lexical subs": "{ my sub s { 42 } print s(1); }\ns/a/b/;\nmy sub q { 1 } q(2);\n",
	"constants":    "use constant PI => 3; my $x = PI / 2;\nsub E () { 2 } my $y = E / 2;\n",
	"class":        "use feature 'class';\nclass C { method m (%o) { keys %o } }\nmy $c = C->new;\n",
	"labels":       "LOOP: for my $i (1..3) { next LOOP; }\nSKIP: { last SKIP; }\n",
	"print handle": "print $fh -1; print STDERR <<\"E\";\nx\nE\nsort $by @a;\n",
}

// resumeStarts are the three Tokenize* entry points, each with the State
// that starts it.
var resumeStarts = []struct {
	name     string
	start    State
	tokenize func([]byte) []Token
}{
	{"Tokenize", StartState(), Tokenize},
	{"TokenizeBarewords", StartStateBarewords([]string{"=>", "->"}), func(s []byte) []Token {
		return TokenizeBarewords(s, []string{"=>", "->"})
	}},
	{"TokenizeTyped", StartStateTyped(), TokenizeTyped},
}

// checkResumes asserts that the checkpointing lexer's stream is the full
// stream, and that resuming from every checkpoint it recorded yields the full
// stream's suffix from that byte. It returns the checkpoints.
func checkResumes(t *testing.T, src []byte, start State, full []Token, stride int) []Checkpoint {
	t.Helper()
	toks, cps := checkpoints(src, start, stride)
	if !reflect.DeepEqual(toks, full) {
		t.Fatalf("the checkpointing lexer's stream differs from the full stream")
	}
	for _, cp := range cps {
		got := Resume(src, cp)
		i := suffixAt(full, cp.Byte)
		if i < 0 {
			t.Fatalf("checkpoint at %d is not a token boundary", cp.Byte)
		}
		if want := full[i:]; !reflect.DeepEqual(got, want) && len(got)+len(want) > 0 {
			t.Fatalf("resuming at byte %d: %s", cp.Byte, firstTokenDiff(src, got, want))
		}
	}
	return cps
}

// suffixAt is the index of the token starting at b, len(toks) at the end of
// input, or -1 when b falls inside a token.
func suffixAt(toks []Token, b int) int {
	for i, tok := range toks {
		if tok.Start == b {
			return i
		}
		if tok.Start > b {
			return -1
		}
	}
	return len(toks)
}

func firstTokenDiff(src []byte, got, want []Token) string {
	for i := 0; i < len(got) && i < len(want); i++ {
		if got[i] != want[i] {
			return "token " + itoa(i) + " got " + describe(src, got[i]) + ", want " + describe(src, want[i])
		}
	}
	return "got " + itoa(len(got)) + " tokens, want " + itoa(len(want))
}

func describe(src []byte, tok Token) string {
	s := tok.Kind.String() + "(" + string(src[tok.Start:tok.End]) + ")@" + itoa(tok.Start)
	if tok.OpensBlock {
		s += "+block"
	}
	if tok.TermFollows {
		s += "+term"
	}
	return s
}

func TestLexFromCheckpointMatchesFull(t *testing.T) {
	for name, src := range resumeFixtures {
		for _, s := range resumeStarts {
			t.Run(name+"/"+s.name, func(t *testing.T) {
				b := []byte(src)
				// Typed Perl reads some fixtures as garbage that never returns
				// to a statement boundary, which is fine: only Perl must.
				cps := checkResumes(t, b, s.start, s.tokenize(b), 0)
				if len(cps) < 2 && s.name != "TokenizeTyped" {
					t.Fatalf("only %d checkpoints: the fixture never reached a safe point", len(cps))
				}
			})
		}
	}
	t.Run("edit5k.pl", func(t *testing.T) {
		src, err := os.ReadFile(filepath.Join("..", "psc", "testdata", "latency", "edit5k.pl"))
		if err != nil {
			t.Fatal(err)
		}
		cps := checkResumes(t, src, StartState(), Tokenize(src), checkpointStride)
		// The stride's purpose: hundreds of checkpoints for a 5,000-line file.
		if len(cps) < 100 || len(cps) > len(src)/checkpointStride+1 {
			t.Fatalf("%d checkpoints for %d bytes", len(cps), len(src))
		}
	})
}

// TestLexFromCheckpointMatchesFullOnCorpus is the same property over perl5's
// t/, when the corpus is present.
func TestLexFromCheckpointMatchesFullOnCorpus(t *testing.T) {
	tDir, files := corpusFiles(t)
	for _, rel := range files {
		src, err := os.ReadFile(filepath.Join(tDir, rel))
		if err != nil {
			t.Fatal(err)
		}
		t.Run(rel, func(t *testing.T) {
			checkResumes(t, src, StartState(), Tokenize(src), checkpointStride)
		})
	}
}

func TestCheckpointNeverInUnsafeState(t *testing.T) {
	for name, src := range resumeFixtures {
		b := []byte(src)
		toks, cps := checkpoints(b, StartState(), 0)
		// A heredoc is pending from the end of its `<<TERM` to the start of
		// its body; bodies follow their markers in order.
		var pendingFrom, pendingTo []int
		var opens []int
		for _, tok := range toks {
			switch tok.Kind {
			case HeredocOpen:
				opens = append(opens, tok.End)
			case HeredocBody:
				if len(opens) > 0 {
					pendingFrom, pendingTo = append(pendingFrom, opens[0]), append(pendingTo, tok.Start)
					opens = opens[1:]
				}
			}
		}
		for _, cp := range cps[1:] {
			st := cp.State.l
			if !st.safe() {
				t.Errorf("%s: checkpoint at %d holds an unsafe state", name, cp.Byte)
			}
			for _, tok := range toks {
				switch tok.Kind {
				case HeredocBody, Pod, Quote, FormatBody, DataSection:
					if tok.Start < cp.Byte && cp.Byte < tok.End {
						t.Errorf("%s: checkpoint at %d inside %s", name, cp.Byte, describe(b, tok))
					}
				}
			}
			for i := range pendingFrom {
				if pendingFrom[i] <= cp.Byte && cp.Byte <= pendingTo[i] {
					t.Errorf("%s: checkpoint at %d with a heredoc pending", name, cp.Byte)
				}
			}
			if prev := toks[suffixAt(toks, cp.Byte)-1]; prev.Kind != Semicolon && prev.Kind != CloseBracket {
				t.Errorf("%s: checkpoint at %d follows %s", name, cp.Byte, describe(b, prev))
			}
		}
	}
	// An unsafe state is refused however it arose.
	for _, src := range []string{"print <<A;", "{ 1;", "format X =", "foo(;", "$h{a"} {
		l := StartState().l
		l.src = []byte(src)
		tokenize(&l)
		if l.safe() {
			t.Errorf("%q: the state after it is called safe", src)
		}
	}
}

func TestLexStateEqualAndCloneAreDeep(t *testing.T) {
	base := StartState()
	if !base.Equal(base.Clone()) {
		t.Fatal("a state is unequal to its clone")
	}
	empty := base.Clone()
	empty.l.brackets, empty.l.pending, empty.l.lexSubs = []bracket{}, []pendingHeredoc{}, []lexSub{}
	empty.l.constSubs = map[string]bool{}
	if !base.Equal(empty) {
		t.Fatal("empty and nil collections compare unequal")
	}
	mutations := map[string]func(*lexer){
		"heredoc queue":    func(l *lexer) { l.pending = append(l.pending, pendingHeredoc{term: []byte("A")}) },
		"heredoc term":     func(l *lexer) { l.pending[0].term[0] = 'B' },
		"bracket stack":    func(l *lexer) { l.brackets = append(l.brackets, paren) },
		"bracket kind":     func(l *lexer) { l.brackets[0] = braceBlock },
		"lexical subs":     func(l *lexer) { l.lexSubs = append(l.lexSubs, lexSub{name: "s"}) },
		"constant subs":    func(l *lexer) { l.constSubs = map[string]bool{"PI": true} },
		"expect":           func(l *lexer) { l.expect = XTerm },
		"utf8":             func(l *lexer) { l.utf8Pragma = true },
		"pending format":   func(l *lexer) { l.pendingFormat = true },
		"signature parens": func(l *lexer) { l.sigDepth = 1 },
	}
	// A state with one of everything, so each mutation has something to change.
	full := base.Clone()
	full.l.pending = []pendingHeredoc{{term: []byte("A")}}
	full.l.brackets = []bracket{square}
	full.l.lexSubs = []lexSub{{name: "t", depth: 1}}
	full.l.constSubs = map[string]bool{"E": true}
	for name, mutate := range mutations {
		c := full.Clone()
		mutate(&c.l)
		if full.Equal(c) {
			t.Errorf("%s: states differing only in it compare equal", name)
		}
		if !reflect.DeepEqual(full.l.pending[0].term, []byte("A")) || full.l.brackets[0] != square ||
			full.l.lexSubs[0].name != "t" || !full.l.constSubs["E"] || full.l.expect != XState {
			t.Fatalf("%s: mutating a clone changed the original", name)
		}
	}
	// Resuming must not write into the checkpoint it resumed from.
	src := []byte("my $x = 1; { print <<A, [1, (2)]; my sub f { } }\nA\nuse constant K => 1;\n")
	_, cps := checkpoints(src, StartState(), 0)
	saved := make([]State, len(cps))
	for i, cp := range cps {
		saved[i] = cp.State.Clone()
	}
	for _, cp := range cps {
		Resume(src, cp)
	}
	for i, cp := range cps {
		if !cp.State.Equal(saved[i]) {
			t.Errorf("resuming at %d changed the checkpoint's state", cp.Byte)
		}
	}
}

// TestLexStateClonesEveryReference keeps Clone honest as lexer grows: a slice
// or map field it does not deep-copy would alias between a checkpoint and the
// lexer that goes on mutating it.
func TestLexStateClonesEveryReference(t *testing.T) {
	cloned := map[string]bool{
		"src": true, "toks": true, // cleared, not cloned: the state holds neither
		"barewordOps": true, "brackets": true, "lexSubs": true, "pending": true, "constSubs": true,
	}
	typ := reflect.TypeOf(lexer{})
	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)
		switch f.Type.Kind() {
		case reflect.Slice, reflect.Map, reflect.Pointer, reflect.Interface, reflect.Func, reflect.Chan:
			if !cloned[f.Name] {
				t.Errorf("lexer.%s is a reference that State.Clone does not copy", f.Name)
			}
		}
	}
}

func TestLexResumeAtOffsetZeroAndEOF(t *testing.T) {
	for _, src := range []string{"", "my $x = 1;", "print <<A;\nbody\nA\n", "=pod\n=cut\n"} {
		b := []byte(src)
		full := Tokenize(b)
		if got := Resume(b, Checkpoint{Byte: 0, State: StartState()}); !reflect.DeepEqual(got, full) {
			t.Errorf("%q: resuming at 0: %s", src, firstTokenDiff(b, got, full))
		}
		_, cps := checkpoints(b, StartState(), 0)
		if cps[0].Byte != 0 || !cps[0].State.Equal(StartState()) {
			t.Errorf("%q: the first checkpoint is not the start of input", src)
		}
		eof := Checkpoint{Byte: len(b), State: StartState()}
		if got := Resume(b, eof); len(got) != 0 {
			t.Errorf("%q: resuming at EOF yields %d tokens", src, len(got))
		}
	}
	_, cps := checkpoints([]byte("my $x = 1;"), StartState(), 0)
	if last := cps[len(cps)-1]; last.Byte != len("my $x = 1;") {
		t.Errorf("no checkpoint at a statement ending at EOF")
	}
}

func TestLexResumeUnterminatedConstructs(t *testing.T) {
	for _, src := range []string{
		"my $a = 1; foo();\nprint <<EOF;\nnever ends\n",
		"my $a = 1; foo();\nmy $s = \"never ends;\n",
		"my $a = 1; foo();\n$x =~ m/never ends;\n",
		"my $a = 1; foo();\n$x =~ s{a}{never ends;\n",
		"my $a = 1; foo();\n\n=pod\n\nnever cut;\n",
		"my $a = 1; foo();\nformat STDOUT =\n@<<<\n",
	} {
		b := []byte(src)
		if cps := checkResumes(t, b, StartState(), Tokenize(b), 0); len(cps) < 3 {
			t.Errorf("%q: %d checkpoints, want one per statement", src, len(cps))
		}
	}
}

func TestLexResumeAroundEND(t *testing.T) {
	for _, src := range []string{
		"my $x = 1;\nfoo();\n__END__\nbar; baz;\n}\n",
		"my $x = 1;\nfoo();\n__DATA__\nbar; baz;\n}\n",
	} {
		b := []byte(src)
		cps := checkResumes(t, b, StartState(), Tokenize(b), 0)
		data := strings.Index(src, "__")
		for _, cp := range cps {
			if cp.Byte > data {
				t.Errorf("%q: checkpoint at %d, inside the data section from %d", src, cp.Byte, data)
			}
		}
		if cps[len(cps)-1].Byte != strings.Index(src, "\n__") {
			t.Errorf("%q: no checkpoint just before the data section", src)
		}
	}
}

func TestLexResumeCRLFHeredoc(t *testing.T) {
	src := []byte("my $x = 1;\r\nprint <<EOT; foo();\r\nline;\r\nEOT\r\nmy $y = 2;\r\nprint <<~E;\r\n  a;\r\n  E\r\nz();\r\n")
	full := Tokenize(src)
	bodies := 0
	for _, tok := range full {
		if tok.Kind == HeredocBody {
			bodies++
		}
		if tok.Kind == UnknownRest {
			t.Fatalf("a CRLF heredoc ran to EOF: %s", describe(src, tok))
		}
	}
	if bodies != 2 {
		t.Fatalf("%d heredoc bodies, want 2", bodies)
	}
	if cps := checkResumes(t, src, StartState(), full, 0); len(cps) < 4 {
		t.Fatalf("%d checkpoints", len(cps))
	}
}

func TestLexResumeCarriesUTF8AndFeatures(t *testing.T) {
	src := []byte("use utf8;\nuse feature 'signatures';\nno feature 'apostrophe_as_package_separator';\n" +
		"my $x = 1;\nmy $café = 2; sub naïve ($ü) { $ü }\nmy $s = $x'foo';\n")
	full := Tokenize(src)
	cps := checkResumes(t, src, StartState(), full, 0)
	last := cps[len(cps)-1].State.l
	if !last.utf8Pragma || !last.signatures || !last.noApostropheSeparator {
		t.Fatalf("the pragmas did not reach the last checkpoint: utf8=%v signatures=%v noApostrophe=%v",
			last.utf8Pragma, last.signatures, last.noApostropheSeparator)
	}
	// The identifier is one token only under `use utf8`, so the resumed
	// stream matching proves the flag was carried rather than re-derived.
	for _, tok := range full {
		if string(src[tok.Start:tok.End]) == "$café" {
			return
		}
	}
	t.Fatal("$café is not one token: the fixture does not exercise use utf8")
}

// FuzzLexResume is the resume property on any input: the checkpointing
// lexer's stream is Tokenize's, and resuming from each of its checkpoints --
// every safe point, no stride -- yields Tokenize's suffix.
//
//	go test ./internal/lexer/ -run '^$' -fuzz '^FuzzLexResume$' -fuzztime 50000x
func FuzzLexResume(f *testing.F) {
	for _, s := range fuzzSeeds {
		f.Add(s)
	}
	for _, s := range resumeFixtures {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, src string) {
		b := []byte(src)
		checkResumes(t, b, StartState(), Tokenize(b), 0)
	})
}
