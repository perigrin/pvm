// ABOUTME: Derives a .pmt declaration's typed signature from its prototype, and its prototype from its types.
// ABOUTME: RFC 0001's table, "A typed signature and a prototype say the same thing", read both ways.
package parse

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/types"
)

// typesFromPrototype reads a prototype, without its parentheses, into the
// coarse typed signature RFC 0001's table gives each character:
//
//	$ Scalar $      @ List @      % List %      + Array|Hash|Scalar $
//	& Code &        * Glob *      _ Scalar $ = $_
//
// and `;` makes each parameter after it optional. A parameter derived
// from a prototype has no name, since a prototype names none.
//
// derived is false for a prototype holding a backslash before a
// container (`\@`, `\[$@%]`): aliased parameters are derived by the
// aliased-parameter issue (01a0fd9b, "Parse aliased parameters"), which
// parses the backslash, and until then such a declaration states no
// typed signature.
func typesFromPrototype(proto string) (sig types.Signature, derived bool, err error) {
	optional := false
	for i := 0; i < len(proto); i++ {
		c := proto[i]
		var p types.Param
		switch c {
		case ';':
			optional = true
			continue
		case '$':
			p = types.Param{Sigil: '$', Type: types.Scalar}
		case '_':
			p = types.Param{Sigil: '$', Type: types.Scalar, Default: "$_"}
		case '+':
			p = types.Param{Sigil: '$', Type: types.Array | types.Hash | types.Scalar}
		case '@', '%':
			p = types.Param{Sigil: c, Type: types.List}
		case '&':
			p = types.Param{Sigil: '&', Type: types.Code}
		case '*':
			p = types.Param{Sigil: '*', Type: types.Glob}
		case '\\':
			if i+1 < len(proto) && strings.IndexByte("$@%&*[", proto[i+1]) >= 0 {
				return types.Signature{}, false, nil
			}
			return types.Signature{}, false, fmt.Errorf("prototype (%s) has %q, which is not a prototype character", proto, proto[i:min(i+2, len(proto))])
		default:
			return types.Signature{}, false, fmt.Errorf("prototype (%s) has %q, which is not a prototype character", proto, proto[i:i+1])
		}
		p.Required = c != '@' && c != '%' && c != '_' && !optional
		sig.Params = append(sig.Params, p)
	}
	return sig, true, nil
}

// prototypeFromTypes gives the prototype, without its parentheses, that a
// typed signature's parameters state, the table read the other way. Types
// are finer than prototypes, so `Str $x` and `Scalar $x` both give `$`;
// a `$` parameter whose type admits an array or a hash gives `+`, and one
// defaulting to `$_` gives `_`. A `;` goes before the first optional
// parameter other than `_`, which is optional of itself.
func prototypeFromTypes(sig types.Signature) (string, error) {
	var b strings.Builder
	semicolon := false
	for _, p := range sig.Params {
		slurpy := p.Sigil == '@' || p.Sigil == '%'
		topic := p.Sigil == '$' && p.Default == "$_"
		if !p.Required && !slurpy && !topic && !semicolon {
			b.WriteByte(';')
			semicolon = true
		}
		switch {
		case p.Sigil == '$' && p.Type&(types.Array|types.Hash) != 0:
			b.WriteByte('+')
		case topic:
			b.WriteByte('_')
		default:
			b.WriteByte(p.Sigil)
		}
	}
	return b.String(), nil
}
