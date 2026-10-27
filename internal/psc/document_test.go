// ABOUTME: Tests for Document, the open buffer's text and line table.
// ABOUTME: Covers UTF-16 LSP positions, clamping, CRLF, and suffix-only reindexing.

package psc

import (
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
)

// nonASCII is the corpus of §6.11 step 1: é (2 bytes, 1 unit), → (3/1), 世
// (CJK, 3/1) and 💖 (4/2, a surrogate pair), on the first, middle and last
// line, with and without a trailing newline and with CRLF.
var nonASCII = []string{
	"my $café = 1;\nmy $x = '→';\nmy $y = '💖'; $z\n",
	"my $café = 1;\nmy $x = '世界';\nmy $y = '💖'; $z",
	"my $café = 1;\r\nmy $x = '→';\r\nmy $y = '💖'; $z\r\n",
	"💖💖\n\n世\n",
	"plain ascii\n",
}

func TestDocumentPositionsUTF16(t *testing.T) {
	for _, text := range nonASCII {
		d := NewDocument([]byte(text))
		for off := 0; off <= len(text); off++ {
			if off < len(text) && !utf8.RuneStart(text[off]) {
				continue
			}
			p := d.ByteToPosition(uint32(off))
			want := uint32(off)
			// The '\n' of a CRLF line has the column of its '\r'.
			if off > 0 && off < len(text) && text[off] == '\n' && text[off-1] == '\r' {
				want--
			}
			assert.Equal(t, want, d.PositionToByte(p), "%q offset %d -> %+v", text, off, p)
		}
	}

	d := NewDocument([]byte(nonASCII[0]))
	// The '=' after café is byte 11 but UTF-16 column 10 (§6.6.1).
	assert.Equal(t, lspPosition{Line: 0, Character: 10}, d.ByteToPosition(11))
	// $z on line 2 follows a 4-byte, 2-unit 💖: byte 16 of the line, column 14.
	line2 := uint32(len("my $café = 1;\nmy $x = '→';\n"))
	assert.Equal(t, lspPosition{Line: 2, Character: 14}, d.ByteToPosition(line2+16))
	assert.Equal(t, line2+16, d.PositionToByte(lspPosition{Line: 2, Character: 14}))

	// Ported from TestPositionConversion.
	d = NewDocument([]byte("my $x = 42;\nprint $x;\n"))
	assert.Equal(t, uint32(3), d.PositionToByte(lspPosition{Line: 0, Character: 3}))
	assert.Equal(t, lspPosition{Line: 0, Character: 3}, d.ByteToPosition(3))
	assert.Equal(t, uint32(12), d.PositionToByte(lspPosition{Line: 1, Character: 0}))
	assert.Equal(t, lspPosition{Line: 1, Character: 0}, d.ByteToPosition(12))
}

func TestDocumentSuffixReindex(t *testing.T) {
	d := NewDocument([]byte("a\nb\nc\nd\n"))
	before := &d.Lines[0]
	// Replace "c" with "x\ny" on line 2.
	d.ApplyEdit(4, 5, []byte("x\ny"))
	assert.Equal(t, "a\nb\nx\ny\nd\n", string(d.Text))
	assert.Equal(t, NewDocument(d.Text).Lines, d.Lines)
	// The prefix was kept in place, not rebuilt into a new table.
	assert.Same(t, before, &d.Lines[0])
	assert.Equal(t, []int32{0, 2, 4, 6, 8, 10}, d.Lines)

	// An edit spanning lines deletes their starts.
	d.ApplyEdit(1, 7, nil)
	assert.Equal(t, "a\nd\n", string(d.Text))
	assert.Equal(t, NewDocument(d.Text).Lines, d.Lines)
}

func TestDocumentPositionInsideSurrogatePairClamps(t *testing.T) {
	d := NewDocument([]byte("a💖b"))
	// Column 2 is between the surrogates of 💖, which starts at byte 1.
	assert.Equal(t, uint32(1), d.PositionToByte(lspPosition{Line: 0, Character: 2}))
	assert.Equal(t, uint32(5), d.PositionToByte(lspPosition{Line: 0, Character: 3}))
}

func TestDocumentPositionOutOfRangeClamps(t *testing.T) {
	// Ported from TestPositionConversionEdgeCases.
	source := []byte("abc\ndef\n")
	d := NewDocument(source)
	n := uint32(len(source))
	// A column past end-of-line clamps to the line's end.
	assert.Equal(t, uint32(3), d.PositionToByte(lspPosition{Line: 0, Character: 100}))
	// A line past the last maps to len(source).
	assert.Equal(t, n, d.PositionToByte(lspPosition{Line: 99, Character: 0}))
	// Line 2 is the empty line after the final newline.
	assert.Equal(t, n, d.PositionToByte(lspPosition{Line: 2, Character: 0}))
	// A negative column is column 0, a negative line offset 0.
	assert.Equal(t, uint32(0), d.PositionToByte(lspPosition{Line: 0, Character: -1}))
	assert.Equal(t, uint32(4), d.PositionToByte(lspPosition{Line: 1, Character: -5}))
	assert.Equal(t, uint32(0), d.PositionToByte(lspPosition{Line: -1, Character: 2}))
	// An offset past the end is the end.
	assert.Equal(t, lspPosition{Line: 2, Character: 0}, d.ByteToPosition(1000))
}

func TestDocumentByteOffsetMidRuneClamps(t *testing.T) {
	d := NewDocument([]byte("é💖x"))
	assert.Equal(t, lspPosition{Line: 0, Character: 0}, d.ByteToPosition(1))
	for off := uint32(3); off < 6; off++ {
		assert.Equal(t, lspPosition{Line: 0, Character: 1}, d.ByteToPosition(off), "offset %d", off)
	}
	assert.Equal(t, lspPosition{Line: 0, Character: 3}, d.ByteToPosition(6))
}

func TestDocumentCRLFColumnsExcludeCR(t *testing.T) {
	d := NewDocument([]byte("ab\r\ncé\r\nx\ry"))
	// Only '\n' starts a line.
	assert.Equal(t, []int32{0, 4, 9}, d.Lines)
	// End of line 0 is column 2, at the '\r' and at the '\n' alike.
	assert.Equal(t, lspPosition{Line: 0, Character: 2}, d.ByteToPosition(2))
	assert.Equal(t, lspPosition{Line: 0, Character: 2}, d.ByteToPosition(3))
	assert.Equal(t, uint32(2), d.PositionToByte(lspPosition{Line: 0, Character: 99}))
	assert.Equal(t, uint32(7), d.PositionToByte(lspPosition{Line: 1, Character: 99}))
	// A lone '\r' is an ordinary byte, one column wide.
	assert.Equal(t, lspPosition{Line: 2, Character: 2}, d.ByteToPosition(11))
	assert.Equal(t, uint32(12), d.PositionToByte(lspPosition{Line: 2, Character: 99}))
}

func TestDocumentEmptyText(t *testing.T) {
	d := NewDocument(nil)
	assert.Equal(t, []int32{0}, d.Lines)
	for _, p := range []lspPosition{{}, {Line: 0, Character: 5}, {Line: 3, Character: 1}, {Line: -1, Character: -1}} {
		assert.Equal(t, uint32(0), d.PositionToByte(p), "%+v", p)
	}
	assert.Equal(t, lspPosition{}, d.ByteToPosition(0))
}

func TestDocumentEditAtOffsetZeroAndEOF(t *testing.T) {
	d := NewDocument([]byte("b\nc"))
	check := func(want string) {
		t.Helper()
		assert.Equal(t, want, string(d.Text))
		assert.Equal(t, NewDocument([]byte(want)).Lines, d.Lines)
	}
	d.ApplyEdit(0, 0, []byte("a\n"))
	check("a\nb\nc")
	d.ApplyEdit(len(d.Text), len(d.Text), []byte("\n"))
	check("a\nb\nc\n")
	d.ApplyEdit(len(d.Text)-1, len(d.Text), nil)
	check("a\nb\nc")
	d.ApplyEdit(0, len(d.Text), nil)
	check("")
	d.ApplyEdit(0, 0, []byte("x\n\ny"))
	check("x\n\ny")
}

func TestDocumentInvalidUTF8NoPanic(t *testing.T) {
	text := []byte("a\x80\xff\xe2\x86b\n\xf0\x9f\x92")
	d := NewDocument(text)
	for off := 0; off <= len(text); off++ {
		p := d.ByteToPosition(uint32(off))
		got := d.PositionToByte(p)
		assert.LessOrEqual(t, got, uint32(len(text)), "offset %d", off)
		assert.LessOrEqual(t, got, uint32(off), "offset %d", off)
	}
	for c := -1; c < 10; c++ {
		assert.LessOrEqual(t, d.PositionToByte(lspPosition{Line: 1, Character: c}), uint32(len(text)))
	}
}
