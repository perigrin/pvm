// ABOUTME: The generator's refusals (spec §5.2, H5): an inconsistent grammar gets an error and no tables.
// ABOUTME: Citations, class literals, classless terminals, unreachable or unproductive nonterminals, unused precedence tokens.

package grammar

import (
	_ "embed"
	"errors"
	"fmt"
	"slices"
	"strings"

	"tamarou.com/pvm/internal/lexer"
)

//go:embed testdata/perly/alternatives.txt
var alternativesText string

// classExceptions are the terminals that need no class: the GRAM* start
// selectors are never lexed, PLUGEXPR and PLUGSTMT are keyword plugins, and the
// PLUGIN_* terminals are waypoint 2's. The exceptions are exact, not prefixes.
var classExceptions = []string{"GRAMPROG", "GRAMEXPR", "GRAMBLOCK", "GRAMBARESTMT", "GRAMFULLSTMT", "GRAMSTMTSEQ", "GRAMSUBSIGNATURE", "PLUGEXPR", "PLUGSTMT"}

func needsClass(terminal string) bool {
	return !strings.HasPrefix(terminal, "PLUGIN_") && !slices.Contains(classExceptions, terminal)
}

// Committed is the grammar perly.go holds, its terminals the symbols no rule
// defines, and the reference list its citations are held to.
func Committed() (*Grammar, []Entry, error) {
	rules, err := ParseRules(perlyRules)
	if err != nil {
		return nil, nil, err
	}
	ref, err := ParseEntries(alternativesText, false)
	if err != nil {
		return nil, nil, err
	}
	g := &Grammar{G1: rules, Classes: perlyClasses}
	defined := map[string]bool{}
	for _, r := range rules {
		defined[r.LHS] = true
	}
	for _, r := range rules {
		for _, s := range r.RHS {
			if !defined[s.Text] {
				g.G0 = append(g.G0, Rule{LHS: s.Text})
			}
		}
	}
	return g, ref, nil
}

// Generate builds g's tables and refuses an inconsistent grammar: it returns
// the error and no tables. A grammar with no rules generates none and no error.
func Generate(g *Grammar, ref []Entry) (*Tables, error) {
	if len(g.G1) == 0 {
		return nil, nil
	}
	t, err := BuildTables(g)
	if err != nil {
		return nil, err
	}
	if err := Refuse(g, t, ref); err != nil {
		return nil, err
	}
	return t, nil
}

// GenerateTables writes g's tables into dir. A refused grammar writes nothing.
func GenerateTables(dir string, g *Grammar, ref []Entry) error {
	t, err := Generate(g, ref)
	if err != nil || t == nil {
		return err
	}
	return WriteTables(dir, t.Files())
}

// Refuse reports every inconsistency of g, built to t, against the reference list ref.
func Refuse(g *Grammar, t *Tables, ref []Entry) error {
	var errs []error
	errs = append(errs, refuseCitations(g, ref)...)
	classes, err := g.ResolveClasses(t)
	if err != nil {
		errs = append(errs, err)
	}
	errs = append(errs, refuseLiterals(classes)...)
	errs = append(errs, refuseClassless(classes, t)...)
	errs = append(errs, refuseReach(t)...)
	errs = append(errs, refusePrecedence(g)...)
	return errors.Join(errs...)
}

func refuseCitations(g *Grammar, ref []Entry) (errs []error) {
	for _, r := range g.G1 {
		line, _, err := parseCite(r.Cite)
		switch {
		case err != nil:
			errs = append(errs, fmt.Errorf("%s has no citation of the form perly.y:<line>: %w", ruleText(r, "::="), err))
		case !slices.ContainsFunc(ref, func(e Entry) bool { return e.Line == line }):
			errs = append(errs, fmt.Errorf("%s cites %s%d, which is not in the reference list", ruleText(r, "::="), citePrefix, line))
		}
	}
	return errs
}

// ResolveClasses is g's classes as CompileTokens takes them, each admitting the id
// that its declared terminal has in t. It refuses a name no rule uses and a nonterminal.
func (g *Grammar) ResolveClasses(t *Tables) ([]TokenClass, error) {
	id := map[string]int{}
	for i, s := range t.Symbols {
		id[s.Name] = i
	}
	var classes []TokenClass
	var errs []error
	for _, d := range g.Classes {
		i, ok := id[d.Terminal]
		switch {
		case !ok:
			errs = append(errs, fmt.Errorf("a token class names %s, which no rule uses", d.Terminal))
		case !t.Symbols[i].Terminal:
			errs = append(errs, fmt.Errorf("a token class names %s, which is a nonterminal", d.Terminal))
		default:
			c := d.Key
			c.Terminals = []int{i}
			classes = append(classes, c)
		}
	}
	return classes, errors.Join(errs...)
}

// refuseLiterals holds each class that names a kind and a text to the lexer.
func refuseLiterals(classes []TokenClass) (errs []error) {
	for _, c := range classes {
		if c.Kind != nil && c.Text != nil && !lexesAs(*c.Kind, *c.Text) {
			errs = append(errs, fmt.Errorf("the literal %q is not lexed to one %v token with that text", *c.Text, *c.Kind))
		}
	}
	return errs
}

// lexesAs reports whether the lexer lexes text to one token of kind and the same text, in a
// short context where a class stands: between two terms, as an operator does, or from a
// statement's start, reached through the lexer's checkpoint and resume.
func lexesAs(kind lexer.Kind, text string) bool {
	one := func(src string, toks []lexer.Token, n, at int) bool {
		var sig []lexer.Token
		for _, t := range toks {
			if t.Kind != lexer.Whitespace {
				sig = append(sig, t)
			}
		}
		return len(sig) == n && sig[at].Kind == kind && src[sig[at].Start:sig[at].End] == text
	}
	infix := "$a " + text + " $b"
	_, cps := lexer.Checkpoints(nil, lexer.StartState())
	return one(infix, lexer.Tokenize([]byte(infix)), 3, 1) || one(text, lexer.Resume([]byte(text), cps[0]), 1, 0)
}

// refuseClassless refuses a terminal no class admits, but the excepted ones.
func refuseClassless(classes []TokenClass, t *Tables) (errs []error) {
	has := map[int]bool{}
	for _, c := range classes {
		for _, id := range c.Terminals {
			has[id] = true
		}
	}
	for id, s := range t.Symbols {
		if s.Terminal && !has[id] && needsClass(s.Name) {
			errs = append(errs, fmt.Errorf("the terminal %s has no token class", s.Name))
		}
	}
	return errs
}

// refuseReach refuses a nonterminal the start symbol does not reach, and one that derives no string of terminals.
func refuseReach(t *Tables) (errs []error) {
	reached, productive := make([]bool, len(t.Symbols)), make([]bool, len(t.Symbols))
	reached[t.Start] = true
	for i, s := range t.Symbols {
		productive[i] = s.Terminal
	}
	for changed := true; changed; {
		changed = false
		for _, r := range t.Rules {
			if reached[r.LHS] {
				for _, s := range r.RHS {
					changed, reached[s] = changed || !reached[s], true
				}
			}
			if !productive[r.LHS] && !slices.ContainsFunc(r.RHS, func(s int) bool { return !productive[s] }) {
				changed, productive[r.LHS] = true, true
			}
		}
	}
	for i, s := range t.Symbols {
		if !s.Terminal && !reached[i] {
			errs = append(errs, fmt.Errorf("the nonterminal %s is unreachable from the start symbol", s.Name))
		}
		if !s.Terminal && !productive[i] {
			errs = append(errs, fmt.Errorf("the nonterminal %s is unproductive: it derives no string of terminals", s.Name))
		}
	}
	return errs
}

// refusePrecedence refuses a precedence token no right-hand side and no %prec uses.
func refusePrecedence(g *Grammar) (errs []error) {
	for _, p := range g.Prec {
		if !slices.ContainsFunc(g.G1, func(r Rule) bool {
			return r.Prec == p || slices.ContainsFunc(r.RHS, func(s Symbol) bool { return s.Text == p })
		}) {
			errs = append(errs, fmt.Errorf("the precedence token %s is used by no rule and no %%prec", p))
		}
	}
	return errs
}
