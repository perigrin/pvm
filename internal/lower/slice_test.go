// ABOUTME: Tests for lowering slices to Slice{Kind}: the five kinds, told apart by the base's sigil and the bracket.
// ABOUTME: Each fragment's parse is measured, and each claim about what a slice gives is perl 5.42's.

package lower_test

import (
	"reflect"
	"testing"

	"tamarou.com/pvm/internal/lower"
)

// sliceKind returns n's Kind if n is a Slice. It reads the field through
// reflect, so the test asserts no lowered type.
func sliceKind(n lower.Node) (lower.SliceKind, bool) {
	v := reflect.ValueOf(n)
	if v.Type() != reflect.TypeOf(&lower.Slice{}) {
		return 0, false
	}
	return lower.SliceKind(v.Elem().FieldByName("Kind").Uint()), true
}

// The five kinds give four result types (§4.4.7). perl 5.42, with @a =
// (10,20,30) and %h = (x=>1, y=>2): @a[0,2] is 10,30 (elements), @h{qw(x
// y)} is 1,2 (values), %a[0,1] is 0,10,1,20 (index/element pairs), %h{"x"}
// is x,1 (key/value pairs), and (1,2)[0] is 1.
func TestLowerSliceKind(t *testing.T) {
	for _, c := range []struct {
		src  string
		want lower.SliceKind
	}{
		{"@a[0,1];", lower.SliceArray},
		{"@h{'a','b'};", lower.SliceHash},
		{"%a[0,1];", lower.SliceKVArray},
		{"%h{'a'};", lower.SliceKVHash},
		{"(1,2)[0];", lower.SliceList},
	} {
		got := lowerOne(t, c.src)
		if k, ok := sliceKind(got); !ok || k != c.want {
			t.Errorf("%q: lowered to %#v, want a Slice of kind %d", c.src, got, c.want)
		}
	}
}
