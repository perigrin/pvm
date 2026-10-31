// ABOUTME: Writes the generator's tables as deterministic text files, and finds drift against committed ones.
// ABOUTME: BuildTables takes the generator's own input (input.go) and builds the LR(0) automaton from it.

package grammar

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// TableFile is one generated file.
type TableFile struct {
	Name string
	Data []byte
}

// Files renders the tables, one file per section, in a fixed order. Every
// line is space-separated, with no trailing space; internal/chart/README.md
// describes each file.
func (t *Tables) Files() []TableFile {
	var sym, rules, items, states, trans, expected, leo bytes.Buffer
	for i, s := range t.Symbols {
		kind := "nonterminal"
		switch {
		case s.Terminal:
			kind = "terminal"
		case s.Nullable:
			kind = "nullable"
		}
		fmt.Fprintf(&sym, "%d %s %s\n", i, kind, s.Name)
	}
	for i := range t.Rules {
		fmt.Fprintf(&rules, "%d %s\n", i, t.ruleText(i))
	}
	for i, it := range t.Items {
		fmt.Fprintf(&items, "%d %d %d\n", i, it.Rule, it.Dot)
	}
	fmt.Fprintf(&states, "start %d\n", t.Start)
	for i, s := range t.States {
		kind := "kernel"
		if s.Predicted {
			kind = "predicted"
		}
		fmt.Fprintf(&states, "%d %s %d items%s reduce%s\n", i, kind, s.Epsilon, ints(s.Items), ints(s.Reduces))
		fmt.Fprintf(&trans, "%d", i)
		for _, tr := range s.Transitions {
			fmt.Fprintf(&trans, " %d:%d", tr.Symbol, tr.Target)
		}
		trans.WriteByte('\n')
		fmt.Fprintf(&expected, "%d%s\n", i, ints(s.Expected))
		for _, it := range s.Leo {
			fmt.Fprintf(&leo, "%d %d\n", i, it)
		}
	}
	return []TableFile{
		{"symbols.txt", sym.Bytes()},
		{"rules.txt", rules.Bytes()},
		{"items.txt", items.Bytes()},
		{"states.txt", states.Bytes()},
		{"transitions.txt", trans.Bytes()},
		{"expected.txt", expected.Bytes()},
		{"leo.txt", leo.Bytes()},
	}
}

// ints writes each number after a space.
func ints(ns []int) string {
	var b strings.Builder
	for _, n := range ns {
		fmt.Fprintf(&b, " %d", n)
	}
	return b.String()
}

func (t *Tables) ruleText(r int) string {
	rule := t.Rules[r]
	parts := []string{t.Symbols[rule.LHS].Name, "::="}
	for _, s := range rule.RHS {
		parts = append(parts, t.Symbols[s].Name)
	}
	return strings.Join(parts, " ")
}

// itemText writes core item id as its rule with a dot: "A ::= B . C".
func (t *Tables) itemText(id int) string {
	it := t.Items[id]
	rule := t.Rules[it.Rule]
	parts := []string{t.Symbols[rule.LHS].Name, "::="}
	for i, s := range rule.RHS {
		if i == it.Dot {
			parts = append(parts, ".")
		}
		parts = append(parts, t.Symbols[s].Name)
	}
	if it.Dot == len(rule.RHS) {
		parts = append(parts, ".")
	}
	return strings.Join(parts, " ")
}

// Dump is every file, each under a "== name ==" header.
func (t *Tables) Dump() string {
	var b strings.Builder
	for _, f := range t.Files() {
		fmt.Fprintf(&b, "== %s ==\n%s", f.Name, f.Data)
	}
	return b.String()
}

// BuildTables builds the tables of g, from its first G1 rule's LHS. The
// terminals are the left-hand sides of g.G0.
func BuildTables(g *Grammar) (*Tables, error) {
	if g == nil || len(g.G1) == 0 {
		return nil, fmt.Errorf("lr0: the grammar has no G1 rules, so no start rule")
	}
	var terminals []string
	for _, r := range g.G0 {
		terminals = append(terminals, r.LHS)
	}
	return BuildLR0(g.G1[0].LHS, g.G1, terminals)
}

// tablePaths is the regular .txt files in dir: the files the generator owns.
// Anything else there, a directory named x.txt included, is foreign.
func tablePaths(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, e := range entries {
		if e.Type().IsRegular() && filepath.Ext(e.Name()) == ".txt" {
			paths = append(paths, filepath.Join(dir, e.Name()))
		}
	}
	return paths, nil
}

// WriteTables writes files into dir, creating it. The generator owns every
// regular .txt file in dir, as TablesDrift reads them: it replaces those it
// writes and removes the rest, so a table an earlier run wrote and this one
// does not cannot linger. Anything else there is foreign and untouched.
func WriteTables(dir string, files []TableFile) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	keep := map[string]bool{}
	for _, f := range files {
		keep[f.Name] = true
		if err := os.WriteFile(filepath.Join(dir, f.Name), f.Data, 0o644); err != nil {
			return err
		}
	}
	old, err := tablePaths(dir)
	if err != nil {
		return err
	}
	for _, p := range old {
		if !keep[filepath.Base(p)] {
			if err := os.Remove(p); err != nil {
				return err
			}
		}
	}
	return nil
}

// TablesDrift names, in name order, each file that differs between the
// committed and the generated directory, or is in only one of them.
func TablesDrift(committed, generated string) ([]string, error) {
	read := func(dir string) (map[string][]byte, error) {
		paths, err := tablePaths(dir)
		if err != nil {
			return nil, err
		}
		m := map[string][]byte{}
		for _, p := range paths {
			if m[filepath.Base(p)], err = os.ReadFile(p); err != nil {
				return nil, err
			}
		}
		return m, nil
	}
	a, err := read(committed)
	if err != nil {
		return nil, err
	}
	b, err := read(generated)
	if err != nil {
		return nil, err
	}
	var drift []string
	for name, data := range a {
		gen, ok := b[name]
		switch {
		case !ok:
			drift = append(drift, name+" is committed but not generated")
		case !bytes.Equal(data, gen):
			drift = append(drift, name+" differs from the generator's")
		}
	}
	for name := range b {
		if _, ok := a[name]; !ok {
			drift = append(drift, name+" is generated but not committed")
		}
	}
	slices.Sort(drift)
	return drift, nil
}
