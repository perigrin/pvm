// ABOUTME: A bareword in a builtin's filehandle slot is the handle unless a sub of that name is declared.
// ABOUTME: `open(FH, "<", $f)` opens FH; `sub FH {1} open(FH, ...)` calls FH first.

package parse_test

import (
	"os/exec"
	"testing"

	"tamarou.com/pvm/internal/conformance"
	"tamarou.com/pvm/internal/parse"
)

// TestBarewordFilehandleSlotIsNotACall: perl reads a bareword in a builtin's
// filehandle slot as the handle, parenthesised or not, unless a sub of that
// name is declared at that point. Measured on 5.42.0 with -MO=Deparse:
//
//	open(FH, "<", $f); open FH, ...    open FH, '<', $f;
//	close(FH); binmode(FH); eof(FH)    close FH; binmode FH; eof FH;
//	seek(FH,0,0); read(FH,$b,1)        seek FH, 0, 0; read FH, $b, 1;
//	opendir(DH,"."); readdir(DH)       opendir DH, '.'; readdir DH;
//	select(FH); flock(FH,1); stat FH   select FH; flock FH, 1; stat FH;
//	-f FH; -s(FH); pipe(R,W)           -f FH; -s FH; pipe R, W;
//	getc FH; write(FH)                 getc FH; write FH;
//	sub FH {1} open(FH, "<", $f)       open FH(), '<', $f;
//	sub FH {1} close FH; -f FH         close FH(); -f FH();
//	open(FH(), "<", $f)                open FH(), '<', $f;
//	sub f {1} seek(FH, f(), 0)         seek FH, f(), 0;
//
// The tree said `open(FH() , ...)` -- call a sub FH and open whatever it
// returns -- which is a different program wherever no such sub exists.
func TestBarewordFilehandleSlotIsNotACall(t *testing.T) {
	perl, perlErr := conformance.PerlPath()
	deparse := func(src string) string {
		out, err := exec.Command(perl, "-MO=Deparse", "-e", src).CombinedOutput()
		if err != nil {
			t.Errorf("perl refuses %q: %v\n%s", src, err, out)
		}
		return string(out)
	}

	for _, tc := range []struct {
		src  string
		word string
		call bool
	}{
		// The handle: no sub of that name is declared.
		{`open(FH, "<", $f);`, "FH", false},
		{`open FH, "<", $f;`, "FH", false},
		{`open(FH);`, "FH", false},
		{`close(FH);`, "FH", false},
		{`close FH;`, "FH", false},
		{`close(FH) or die;`, "FH", false},
		{`close FH or die;`, "FH", false},
		{`binmode(FH);`, "FH", false},
		{`binmode FH, ":raw";`, "FH", false},
		{`eof(FH);`, "FH", false},
		{`eof FH;`, "FH", false},
		{`fileno(FH);`, "FH", false},
		{`fileno FH;`, "FH", false},
		{`seek(FH, 0, 0);`, "FH", false},
		{`seek FH, 0, 0;`, "FH", false},
		{`tell(FH);`, "FH", false},
		{`tell FH;`, "FH", false},
		{`read(FH, $b, 1);`, "FH", false},
		{`read FH, $b, 1;`, "FH", false},
		{`sysread(FH, $b, 1);`, "FH", false},
		{`sysread FH, $b, 1;`, "FH", false},
		{`opendir(DH, ".");`, "DH", false},
		{`opendir DH, ".";`, "DH", false},
		{`readdir(DH);`, "DH", false},
		{`readdir DH;`, "DH", false},
		{`closedir(DH);`, "DH", false},
		{`closedir DH;`, "DH", false},
		{`print FH 1;`, "FH", false},
		{`print(FH 1);`, "FH", false},
		{`printf FH "x";`, "FH", false},
		{`use feature "say"; say FH 1;`, "FH", false},
		{`select(FH);`, "FH", false},
		{`select FH;`, "FH", false},
		{`flock(FH, 1);`, "FH", false},
		{`flock FH, 1;`, "FH", false},
		{`truncate(FH, 0);`, "FH", false},
		{`truncate FH, 0;`, "FH", false},
		{`stat(FH);`, "FH", false},
		{`stat FH;`, "FH", false},
		{`lstat(FH);`, "FH", false},
		{`lstat FH;`, "FH", false},
		{`-f FH;`, "FH", false},
		{`-s(FH);`, "FH", false},
		{`getc FH;`, "FH", false},
		{`write(FH);`, "FH", false},
		{`pipe(R, W);`, "W", false},
		{`pipe(R, W);`, "R", false},
		{`open(FH, "<", $f); close FH; sub FH {1}`, "FH", false},

		// A call: a sub of that name is declared before the slot, or the
		// word is written with parentheses.
		{`sub FH {1} open(FH, "<", $f);`, "FH", true},
		{`sub FH {1} open FH, "<", $f;`, "FH", true},
		{`sub FH {1} close FH;`, "FH", true},
		{`sub FH {1} eof(FH);`, "FH", true},
		{`sub FH {1} print FH 1;`, "FH", true},
		{`sub FH {1} -f FH;`, "FH", true},
		{`open(FH(), "<", $f);`, "FH", true},

		// A call in a slot that takes no handle stays a call.
		{`sub f {1} seek(FH, f(), 0);`, "f", true},
		{`sub f {1} seek(FH, f(), 0);`, "FH", false},
		{`sub mode {"<"} open(FH, mode, $f);`, "mode", true},
		{`sub mode {"<"} open(FH, mode, $f);`, "FH", false},
	} {
		root := parse.Parse([]byte(tc.src))
		if n := countUnknown(root); n != 0 {
			t.Errorf("%s: Unknown=%d\n%s", tc.src, n, dumpTree(root, 2))
			continue
		}
		if got := callsWord(root, tc.word); got != tc.call {
			t.Errorf("%s: %s is a call = %v, perl reads call = %v\n%s",
				tc.src, tc.word, got, tc.call, dumpTree(root, 2))
		}
		if ok, why := canonMatches([]byte(tc.src)); !ok {
			t.Errorf("%s: canon is not a fixpoint: %s", tc.src, why)
		}
		if perlErr != nil {
			continue
		}
		canon := parse.Canon(root, []byte(tc.src))
		if a, b := deparse(tc.src), deparse(canon); a != b {
			t.Errorf("perl reads %q and its canon %q differently:\n%s\n%s", tc.src, canon, a, b)
		}
	}

	if perlErr != nil {
		t.Logf("no perl 5.42 to compare readings with: %v", perlErr)
	}
}
