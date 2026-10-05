// ABOUTME: Tests for the lattice's IO and FileHandle types and the lookup from a type name.
// ABOUTME: A .pmt declaration names types by string; this is the table it resolves against.
package types

import "testing"

// TestIOLeafAndFileHandleUnion: RFC 0001 "Type names". IO joins the lattice
// as a top-level leaf beside Code and Glob -- the paper's hierarchy puts it
// there, and measured on 5.42 `*STDOUT{IO}` is an IO object, not a scalar --
// and FileHandle is the named union Glob|GlobRef|IO a handle may be: a
// bareword handle is a Glob, `open my $fh` gives a GLOB reference.
func TestIOLeafAndFileHandleUnion(t *testing.T) {
	if IO&(List|Code|Glob) != 0 || IO&None != 0 {
		t.Errorf("IO overlaps an existing type: %032b", IO)
	}
	if FileHandle != Glob|GlobRef|IO {
		t.Errorf("FileHandle = %v, want Glob|GlobRef|IO", FileHandle)
	}
	if got := FileHandle.String(); got != "FileHandle" {
		t.Errorf("FileHandle.String() = %q", got)
	}
	if got := IO.String(); got != "IO" {
		t.Errorf("IO.String() = %q", got)
	}
	if !IsSubtype(IO, FileHandle) || !IsSubtype(IO, Any) {
		t.Errorf("IO is not under FileHandle and Any")
	}
}
