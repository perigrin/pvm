// ABOUTME: Tests for the lattice's IO and FileHandle types and the lookup from a type name.
// ABOUTME: A .pmt declaration names types by string; this is the table it resolves against.
package types

import (
	"strings"
	"testing"
)

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

// TestBooleanIsThePapersName: RFC 0001 "Type names" -- names follow the
// paper, which says Boolean, so a .pmt writes Boolean and the lattice
// prints it.
func TestBooleanIsThePapersName(t *testing.T) {
	if got := Boolean.String(); got != "Boolean" {
		t.Errorf("Boolean.String() = %q", got)
	}
}

// TestTypeFromName: a .pmt names types by string (RFC 0001 "Type names"),
// so every named type resolves, and a union is spelled A|B.
func TestTypeFromName(t *testing.T) {
	for typ, name := range typeNames {
		if typ == Unknown || typ == None {
			continue
		}
		got, err := FromName(name)
		if err != nil || got != typ {
			t.Errorf("FromName(%q) = %v, %v; want %v", name, got, err, typ)
		}
	}
	for name, want := range map[string]Type{
		"FileHandle": FileHandle,
		"Boolean":    Boolean,
		"Str|Undef":  Str | Undef,
		"Int|Num":    Num,
	} {
		if got, err := FromName(name); err != nil || got != want {
			t.Errorf("FromName(%q) = %v, %v; want %v", name, got, err, want)
		}
	}
	for _, bad := range []string{"Strng", "Str|"} {
		if _, err := FromName(bad); err == nil {
			t.Errorf("FromName(%q) succeeded; want an error", bad)
		}
	}
}

// TestTypeFromNameRejectsMalformed: a malformed or partly unknown name is
// an error naming the offending part, never a panic.
func TestTypeFromNameRejectsMalformed(t *testing.T) {
	for bad, part := range map[string]string{
		"":           `""`,
		"   ":        `""`,
		"|Str":       `""`,
		"Str|":       `""`,
		"Str||Undef": `""`,
		"Str|Strng":  `"Strng"`,
		"str":        `"str"`,
	} {
		_, err := FromName(bad)
		if err == nil {
			t.Errorf("FromName(%q) succeeded; want an error", bad)
			continue
		}
		if !strings.Contains(err.Error(), part) {
			t.Errorf("FromName(%q) error %q does not name %s", bad, err, part)
		}
	}
}

// TestTypeFromNameRejectsBool: the lattice says Boolean, so Bool is an
// unknown name.
func TestTypeFromNameRejectsBool(t *testing.T) {
	if _, err := FromName("Bool"); err == nil {
		t.Errorf("FromName(%q) succeeded; Bool is the retired spelling", "Bool")
	}
}

// TestFileHandleExcludesStrAndObject: RFC 0001 -- a string naming a handle
// dies under strict, and a blessed IO::File is an Object, so neither is a
// FileHandle; nor is a FileHandle an IO.
func TestFileHandleExcludesStrAndObject(t *testing.T) {
	for _, tc := range []struct{ child, parent Type }{
		{Str, FileHandle}, {Object, FileHandle}, {FileHandle, IO},
	} {
		if IsSubtype(tc.child, tc.parent) {
			t.Errorf("IsSubtype(%v, %v) = true", tc.child, tc.parent)
		}
	}
}
