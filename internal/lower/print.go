// ABOUTME: Prints a lowered tree as an indented outline or an s-expression: each node's kind, set fields and children.
// ABOUTME: Read by reflection over node.go's structs, so a kind added there prints with no edit here.

package lower

import (
	"fmt"
	"reflect"
	"strings"
)

// Tree renders n as an outline, one node per line: its kind, its set scalar
// fields, and each child indented under it, named by the field it sits in.
// A Hole prints as Hole, an Unknown's one trace in the lowered tree.
func Tree(n Node) string {
	var b strings.Builder
	var walk func(prefix string, n Node, depth int)
	walk = func(prefix string, n Node, depth int) {
		b.WriteString(strings.Repeat("  ", depth) + prefix + label(n) + "\n")
		for _, c := range roles(n) {
			walk(c.Name+": ", c.Node, depth+1)
		}
	}
	walk("", n, 0)
	return b.String()
}

// Sexpr renders n as Tree does, on one line: `(Key Arrow Base:(Var ...) ...)`.
func Sexpr(n Node) string {
	var b strings.Builder
	var walk func(n Node)
	walk = func(n Node) {
		b.WriteString("(" + label(n))
		for _, c := range roles(n) {
			b.WriteString(" " + c.Name + ":")
			walk(c.Node)
		}
		b.WriteString(")")
	}
	walk(n)
	return b.String()
}

// role is one child and the field it sits in.
type role struct {
	Name string
	Node Node
}

var (
	nodeType = reflect.TypeOf((*Node)(nil)).Elem()
	byteType = reflect.TypeOf(byte(0))
)

// roles is n's non-nil children in field order, each named by its field:
// Children's walk, with the names it does not keep.
func roles(n Node) []role {
	var out []role
	v := reflect.ValueOf(n).Elem()
	for i := 0; i < v.NumField(); i++ {
		f, fv := v.Type().Field(i), v.Field(i)
		switch {
		case f.Type == nodeType:
			if !fv.IsNil() {
				out = append(out, role{f.Name, fv.Interface().(Node)})
			}
		case f.Type.Kind() == reflect.Slice && f.Type.Elem() == nodeType:
			for j := 0; j < fv.Len(); j++ {
				if e := fv.Index(j); !e.IsNil() {
					out = append(out, role{f.Name, e.Interface().(Node)})
				}
			}
		}
	}
	return out
}

// label is n's kind and its set scalar fields: a true bool by its name,
// anything else as Name=value. Zero values are left out.
func label(n Node) string {
	v := reflect.ValueOf(n).Elem()
	parts := []string{v.Type().Name()}
	if n.Meta().Paren {
		parts = append(parts, "Paren")
	}
	for i := 0; i < v.NumField(); i++ {
		f, fv := v.Type().Field(i), v.Field(i)
		if f.Anonymous || f.Type == nodeType || f.Type.Kind() == reflect.Slice && f.Type.Elem() == nodeType || fv.IsZero() {
			continue
		}
		switch {
		case f.Type.Kind() == reflect.Bool:
			parts = append(parts, f.Name)
		case f.Type.Kind() == reflect.String:
			parts = append(parts, fmt.Sprintf("%s=%q", f.Name, fv.String()))
		case f.Type == byteType:
			parts = append(parts, fmt.Sprintf("%s=%q", f.Name, rune(fv.Uint())))
		default:
			parts = append(parts, fmt.Sprintf("%s=%v", f.Name, fv.Interface()))
		}
	}
	return strings.Join(parts, " ")
}
