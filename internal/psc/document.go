// ABOUTME: Document, an open buffer's text plus its line table (spec §6.2).
// ABOUTME: Converts byte offsets to and from LSP positions in UTF-16 code units (§6.6).

package psc

import (
	"bytes"
	"sort"
	"unicode/utf8"
)

// Document is the text of one open buffer and the byte offset of each line
// start: Lines[0] == 0, and only '\n' starts a line. Offsets are bytes
// inside; LSP positions are UTF-16 code units, converted only here.
type Document struct {
	Text  []byte
	Lines []int32
}

// NewDocument indexes text's lines. The Document aliases text.
func NewDocument(text []byte) *Document {
	d := &Document{Text: text, Lines: []int32{0}}
	d.reindexLinesFrom(0)
	return d
}

// ApplyEdit replaces Text[start:oldEnd] with newText, reindexes the lines
// from start on, and returns the new end offset.
func (d *Document) ApplyEdit(start, oldEnd int, newText []byte) int {
	newEnd := start + len(newText)
	tail := d.Text[oldEnd:]
	need := newEnd + len(tail)
	if cap(d.Text) >= need {
		// In place: copy handles the overlap of a tail shifted either way.
		d.Text = d.Text[:need]
		copy(d.Text[newEnd:], d.Text[oldEnd:oldEnd+len(tail)])
		copy(d.Text[start:], newText)
	} else {
		buf := make([]byte, 0, need+need/4)
		buf = append(buf, d.Text[:start]...)
		buf = append(buf, newText...)
		d.Text = append(buf, tail...)
	}
	d.reindexLinesFrom(start)
	return newEnd
}

// reindexLinesFrom rebuilds the line starts after from. Those at or before
// it are unchanged by an edit there, and kept in place.
func (d *Document) reindexLinesFrom(from int) {
	keep := sort.Search(len(d.Lines), func(i int) bool { return d.Lines[i] > int32(from) })
	d.Lines = d.Lines[:keep]
	for i := int(d.Lines[keep-1]); ; {
		nl := bytes.IndexByte(d.Text[i:], '\n')
		if nl < 0 {
			return
		}
		i += nl + 1
		d.Lines = append(d.Lines, int32(i))
	}
}

// lineEnd is the offset where line's content ends: before its '\n', and
// before a '\r' ahead of that or at the end of the text.
func (d *Document) lineEnd(line int) int {
	end := len(d.Text)
	if line+1 < len(d.Lines) {
		end = int(d.Lines[line+1]) - 1
	}
	if end > int(d.Lines[line]) && d.Text[end-1] == '\r' {
		end--
	}
	return end
}

// utf16Width is the UTF-16 code units of r: two for a surrogate pair.
func utf16Width(r rune) int {
	if r > 0xFFFF {
		return 2
	}
	return 1
}

// ByteToPosition is the LSP position of byte offset off. An offset past the
// end is the end, one inside a character is that character's start, and one
// in a line ending is the end of the line's content.
func (d *Document) ByteToPosition(off uint32) lspPosition {
	o := min(int(off), len(d.Text))
	line := sort.Search(len(d.Lines), func(i int) bool { return int(d.Lines[i]) > o }) - 1
	o = min(o, d.lineEnd(line))
	col := 0
	for i := int(d.Lines[line]); i < o; {
		r, size := utf8.DecodeRune(d.Text[i:])
		if i+size > o {
			break
		}
		col += utf16Width(r)
		i += size
	}
	return lspPosition{Line: line, Character: col}
}

// PositionToByte is the byte offset of LSP position p. Out-of-range input is
// clamped, never rejected, since editors send stale positions: a negative
// line is offset 0, a line past the last is the end, a negative column is
// the line's start, a column past its end is its end, and a column between
// the halves of a surrogate pair is that character's start.
func (d *Document) PositionToByte(p lspPosition) uint32 {
	if p.Line < 0 {
		return 0
	}
	if p.Line >= len(d.Lines) {
		return uint32(len(d.Text))
	}
	end := d.lineEnd(p.Line)
	i, units := int(d.Lines[p.Line]), 0
	for i < end {
		r, size := utf8.DecodeRune(d.Text[i:end])
		if units+utf16Width(r) > p.Character {
			break
		}
		units += utf16Width(r)
		i += size
	}
	return uint32(i)
}
