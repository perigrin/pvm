// ABOUTME: After the nine UNIDOR builtins `//` is defined-or, not an empty pattern,
// ABOUTME: as toke.c's XTERMORDORDOR state decides it.

package lexer

import "testing"

// TestDefinedOrAfterUnidor holds toke.c's UNIDOR rule. Nine named unaries
// leave PL_expect = XTERMORDORDOR, and yyl_slash takes `//` as DORDOR there:
//
//	#define UNIDOR(f) UNI3(f,XTERMORDORDOR,1)
//	if ((PL_expect == XOPERATOR || PL_expect == XTERMORDORDOR) && s[1] == '/')
//	    ... AOPERATOR(DORDOR);
//
// Measured on 5.42.0, `my $y = shift // 7` deparses as `shift() // 7`, and
// the same holds for getc, pop, pos, readline, readpipe, readlink, undef and
// umask. Any other named unary is plain XTERM, where `//` opens a pattern:
// `my $y = lc // 7;` reports "Number found where operator expected".
func TestDefinedOrAfterUnidor(t *testing.T) {
	for _, w := range []string{
		"getc", "pop", "pos", "readline", "readpipe", "readlink",
		"shift", "undef", "umask", "CORE::shift",
	} {
		src := "my $y = " + w + " // 7;"
		toks := significant(Tokenize([]byte(src)))
		if len(toks) < 6 || src[toks[4].Start:toks[4].End] != "//" || toks[4].Kind != Operator {
			t.Errorf("%q: want `//` as the defined-or operator after %s", src, w)
		}
	}
	src := "my $y = lc // 7;"
	toks := significant(Tokenize([]byte(src)))
	if len(toks) < 5 || toks[4].Kind != Quote {
		t.Errorf("%q: after a plain named unary `//` is an empty pattern", src)
	}
}

// TestUnidorSubNameStillOpensABlock: a UNIDOR word used as a sub's NAME is a
// name, so its `{` is the body. op/exec.t has `sub readpipe { pop }`, which
// split into a forward declaration and a bare block when the UNIDOR state
// was tested before the declared-name one.
func TestUnidorSubNameStillOpensABlock(t *testing.T) {
	for _, src := range []string{"sub readpipe { pop }", "sub shift { 1 }"} {
		opens, found := lastBraceOpensBlock(src)
		if !found || !opens {
			t.Errorf("%q: the `{` after a sub name opens its body", src)
		}
	}
}
