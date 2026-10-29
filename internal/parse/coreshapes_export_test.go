// ABOUTME: CoreShapes exports CORE.tpm's builtin shape table, read here from outside the package.
// ABOUTME: The caller gets a copy: mutating it changes neither a later call nor what the parser reads.

package parse_test

import (
	"testing"

	"tamarou.com/pvm/internal/parse"
)

func TestCoreShapesReadsBuiltin(t *testing.T) {
	shape, ok := parse.CoreShapes()["time"]
	if !ok || shape != parse.ShapeNiladic {
		t.Fatalf("time: got %v, %v; want niladic, true", shape, ok)
	}
}

// TestCoreShapesIsACopy: `time + 1` reads as `time() + 1` because time is
// niladic. Were the caller's map the parser's own, making time a list
// operator in it would make the parser read `time(+1)`.
func TestCoreShapesIsACopy(t *testing.T) {
	src := []byte(`time + 1;`)
	before := shape(parse.Parse(src))

	shapes := parse.CoreShapes()
	shapes["time"] = parse.ShapeList
	delete(shapes, "print")

	again := parse.CoreShapes()
	if again["time"] != parse.ShapeNiladic {
		t.Errorf("second call: time is %v; want niladic", again["time"])
	}
	if _, ok := again["print"]; !ok {
		t.Errorf("second call: print is missing")
	}
	if after := shape(parse.Parse(src)); after != before {
		t.Errorf("parser's table changed: %q parses as %s, was %s", src, after, before)
	}
}

func TestCoreShapesUnknownName(t *testing.T) {
	if shape, ok := parse.CoreShapes()["no_such_builtin"]; ok {
		t.Errorf("no_such_builtin: got %v; want no entry", shape)
	}
}
