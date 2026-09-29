// ABOUTME: The string-bitwise operators `&. |. ^. ~.`, their assignments, and logical xor
// ABOUTME: `^^` / `^^=` are single tokens; a `.` before a digit is still a number.

package lexer

import "testing"

// TestStringBitwiseAndLogicalXor holds the operators the parser's precedence
// table already knew and the lexer never produced. Measured on 5.42.0 with
// -MO=Deparse under `use feature "bitwise"`: `$a |. $b`, `~.$s` and
// `$a ^^ $b` (5.40) are kept as written. The lexer split each into two
// tokens, so `22 &. 66` read as `22 & .` and a stray `66`.
//
// `&.`, `|.`, `^.` and `~.` exist only under the `bitwise` feature, and
// without it `$a |.5` is `$a | 0.5`. The lexer does not track that feature,
// so a `.` IMMEDIATELY followed by a digit stays a number's.
func TestStringBitwiseAndLogicalXor(t *testing.T) {
	for _, c := range []struct{ src, op string }{
		{"$a &. $b", "&."}, {"$a |. $b", "|."}, {"$a ^. $b", "^."},
		{"$a &.= $b", "&.="}, {"$a |.= $b", "|.="}, {"$a ^.= $b", "^.="},
		{"$a ^^ $b", "^^"}, {"$a ^^= $b", "^^="},
	} {
		toks := significant(Tokenize([]byte(c.src)))
		if len(toks) != 3 || c.src[toks[1].Start:toks[1].End] != c.op {
			t.Errorf("%q: want one %q operator between the operands, got %v", c.src, c.op, toks)
		}
	}
	src := "~.$s"
	if toks := significant(Tokenize([]byte(src))); src[toks[0].Start:toks[0].End] != "~." {
		t.Errorf("%q: want the prefix `~.`", src)
	}
	// A `.` right before a digit belongs to the number.
	src = "$a |.5"
	toks := significant(Tokenize([]byte(src)))
	if len(toks) != 3 || src[toks[1].Start:toks[1].End] != "|" {
		t.Errorf("%q: want `|` then the number .5, got %v", src, toks)
	}
}
