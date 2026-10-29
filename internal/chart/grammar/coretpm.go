// ABOUTME: The CORE.tpm cross-check: each builtin's shape from G1's per-keyword rules, against internal/parse's table.
// ABOUTME: Lists every builtin whose shapes differ, or that only one side has, for review in the pull request.

package grammar

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"tamarou.com/pvm/internal/parse"
)

// The shapes in order of width. G1's shape for a builtin is the widest any
// of its alternatives takes. niladic, unary and list are parse.Shape's
// names; assign is Guacamole's assign keyword operators (goto $a = $b is
// goto($a = $b)), which CORE.tpm has no shape for.
const (
	niladic = iota
	unary
	assign
	list
)

var shapeNames = []string{"niladic", "unary", "assign", "list"}

// noShape marks the side of a difference that lacks the builtin.
const noShape = "none"

// argShapes is the width each symbol of a keyword's argument gives it. A
// symbol not listed here is an error, so a new argument form is classified
// on purpose rather than by default.
var argShapes = map[string]int{
	"LParen": niladic, "RParen": niladic, "Label": niladic,
	"OpUnaryKeywordArg": unary, "OpUnaryKeywordArgNonBrace": unary, "Value": unary,
	"BlockNonEmpty": unary, "BuiltinFilehandle": unary, "SubDefinition": unary,
	"VarScalar": unary, "ExprAssignR": unary,
	"OpAssignKeywordArg": assign,
	"OpListKeywordArg":   list, "OpListKeywordArgNonBrace": list, "Expression": list, "OpComma": list,
}

var keywordExpr = regexp.MustCompile(`^(OpKeyword[A-Za-z0-9_]+)Expr$`)

// G1Shapes is each builtin's shape from G1's OpKeyword<X>Expr rules, by the
// text of the OpKeyword<X> terminal each rule starts with.
func G1Shapes(c *CutGrammar) (map[string]string, error) {
	width := map[string]int{}
	for _, r := range c.G1 {
		m := keywordExpr.FindStringSubmatch(r.LHS)
		if m == nil {
			continue
		}
		if len(r.RHS) == 0 || r.RHS[0].Text != m[1] {
			return nil, fmt.Errorf("%s (line %d) does not start with %s", ruleText(r, "::="), r.Line, m[1])
		}
		preds := c.Terminals[m[1]]
		name, ok := "", len(preds) == 1
		if ok {
			name, ok = literalText(preds[0].Def)
		}
		if !ok {
			return nil, fmt.Errorf("%s: keyword %s is not a single literal", ruleText(r, "::="), m[1])
		}
		w := niladic
		for _, s := range r.RHS[1:] {
			sw, known := argShapes[s.Text]
			if !known {
				return nil, fmt.Errorf("%s (line %d): no shape for argument symbol %s", ruleText(r, "::="), r.Line, s.Text)
			}
			w = max(w, sw)
		}
		width[name] = max(width[name], w) // niladic is 0, the missing value
	}
	shapes := make(map[string]string, len(width))
	for name, w := range width {
		shapes[name] = shapeNames[w]
	}
	return shapes, nil
}

// CoreTpmDifferences lists, in name order, "<builtin> <G1 shape> <CORE.tpm
// shape>" for every builtin whose two shapes differ, with none for the side
// that lacks it.
func CoreTpmDifferences(g1, core map[string]string) []string {
	names := slices.Collect(maps.Keys(g1))
	for n := range core {
		if _, ok := g1[n]; !ok {
			names = append(names, n)
		}
	}
	slices.Sort(names)
	var out []string
	for _, name := range names {
		a, b := orNone(g1, name), orNone(core, name)
		if a != b {
			out = append(out, strings.Join([]string{name, a, b}, " "))
		}
	}
	return out
}

func orNone(m map[string]string, k string) string {
	if v, ok := m[k]; ok {
		return v
	}
	return noShape
}

// GuacamoleCoreTpmDifferences cross-checks the grammar at slifPath, after
// GuacamoleCut, against internal/parse's CORE.tpm shape table.
func GuacamoleCoreTpmDifferences(slifPath string) ([]string, error) {
	c, err := readGuacamoleCut(slifPath)
	if err != nil {
		return nil, err
	}
	g1, err := G1Shapes(c)
	if err != nil {
		return nil, err
	}
	core := map[string]string{}
	for name, s := range parse.CoreShapes() {
		core[name] = s.String()
	}
	return CoreTpmDifferences(g1, core), nil
}
