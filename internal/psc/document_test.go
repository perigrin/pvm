// ABOUTME: Tests for Document, the open buffer's text and line table.
// ABOUTME: Covers UTF-16 positions, clamping, CR/LF, suffix reindexing, edit ranges and snapshots.

package psc

import (
	"slices"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	// A '\n' planted ahead of the edit's line, behind the table's back: a
	// rescan of the prefix would index it, a suffix-only one cannot see it.
	d.Text[0] = '\n'
	// Replace "c" with "x\ny" on line 2.
	mustEdit(t, d, 4, 5, "x\ny")
	assert.Equal(t, "\n\nb\nx\ny\nd\n", string(d.Text))
	// The prefix was kept in place, not rebuilt into a new table.
	assert.Same(t, before, &d.Lines[0])
	assert.Equal(t, []int32{0, 2, 4, 6, 8, 10}, d.Lines)
	d.Text[0] = 'a'
	assert.Equal(t, NewDocument(d.Text).Lines, d.Lines)

	// An edit spanning lines deletes their starts.
	mustEdit(t, d, 1, 7, "")
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
	mustEdit(t, d, 0, 0, "a\n")
	check("a\nb\nc")
	mustEdit(t, d, len(d.Text), len(d.Text), "\n")
	check("a\nb\nc\n")
	mustEdit(t, d, len(d.Text)-1, len(d.Text), "")
	check("a\nb\nc")
	mustEdit(t, d, 0, len(d.Text), "")
	check("")
	mustEdit(t, d, 0, 0, "x\n\ny")
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

// mustEdit applies an edit that must be accepted.
func mustEdit(t *testing.T, d *Document, start, oldEnd int, newText string) {
	t.Helper()
	_, err := d.ApplyEdit(start, oldEnd, []byte(newText))
	require.NoError(t, err)
}

func TestDocumentApplyEditSnapshotIsolation(t *testing.T) {
	s := NewLSPServer()
	const uri = "untitled:snapshot.pl"
	src := []byte("my $x = 1;\nmy $y = 2;\n")
	require.NoError(t, s.OpenDocument(uri, src))
	// The open buffer is the server's own, not the caller's.
	src[0] = 'X'

	snap := s.Text(uri)
	want, wantLines := "my $x = 1;\nmy $y = 2;\n", slices.Clone(snap.Lines)
	require.Equal(t, want, string(snap.Text))

	editing, done := make(chan struct{}), make(chan struct{})
	o := s.owner(uri)
	go func() {
		defer close(done)
		// Edits run on the owner's goroutine, as incremental didChange will.
		o.call(func() {
			close(editing)
			doc := o.snap.text
			// Same-length edits, which fit the buffer in place, and one of
			// them adds a line.
			for range 50 {
				_, err := doc.ApplyEdit(0, 1, []byte("\n"))
				assert.NoError(t, err)
				_, err = doc.ApplyEdit(0, 1, []byte("m"))
				assert.NoError(t, err)
			}
			_, err := doc.ApplyEdit(0, 2, []byte("our"))
			assert.NoError(t, err)
		})
	}()
	<-editing
	for range 1000 {
		snap.ByteToPosition(uint32(len(snap.Text)))
		_ = string(snap.Text)
	}
	<-done

	assert.Equal(t, want, string(snap.Text))
	assert.Equal(t, wantLines, snap.Lines)
	assert.Equal(t, "our $x = 1;\nmy $y = 2;\n", string(s.Text(uri).Text))
}

func TestDocumentApplyEditRejectsBadRange(t *testing.T) {
	const text = "ab\ncd\n"
	for _, r := range [][2]int{{3, 1}, {-1, 2}, {2, 99}, {99, 99}, {-2, -1}} {
		d := NewDocument([]byte(text))
		lines := slices.Clone(d.Lines)
		_, err := d.ApplyEdit(r[0], r[1], []byte("X\n"))
		assert.Error(t, err, "range %v", r)
		assert.Equal(t, text, string(d.Text), "range %v", r)
		assert.Equal(t, lines, d.Lines, "range %v", r)
	}
}

func TestDocumentLoneCRAtEOFIsOrdinaryByte(t *testing.T) {
	d := NewDocument([]byte("ab\r"))
	assert.Equal(t, lspPosition{Line: 0, Character: 2}, d.ByteToPosition(2))
	assert.Equal(t, lspPosition{Line: 0, Character: 3}, d.ByteToPosition(3))
	assert.Equal(t, uint32(3), d.PositionToByte(lspPosition{Line: 0, Character: 99}))
	// A CRLF line still ends before its '\r'.
	d = NewDocument([]byte("ab\r\n"))
	assert.Equal(t, lspPosition{Line: 0, Character: 2}, d.ByteToPosition(3))
}
