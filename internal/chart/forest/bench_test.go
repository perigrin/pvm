// ABOUTME: BenchmarkForest over std5k.pl and the committed grammar's tables: the forest built after recognition.
// ABOUTME: Its body is in bench.go, so the ratchet runs the same code; make test never runs it.

package forest

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkForest(b *testing.B) {
	src, err := os.ReadFile(filepath.Join("..", "testdata", "std5k.pl"))
	if err != nil {
		b.Fatal(err)
	}
	Bench(b, committedTables(b), src)
}
