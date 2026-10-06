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
// container (`\@`, `\[$@%]`): aliased parameters are derived by
// 01a0fd9b-4850-704d-8c14-c15add6194f2 ("Parse aliased parameters"),
// which parses the backslash, and until then such a declaration states no
// typed signature.
func typesFromPrototype(proto string) (sig types.Signature, derived bool, err error) {
	optional, list := false, -1
	for i := 0; i < len(proto); i++ {
		c := proto[i]
		if c == ';' {
			optional = true
			continue
		}
		// The `@` or `%` takes every argument, so nothing after it can be
		// filled: the final-List rule, as the typed reader enforces it.
		if list >= 0 {
			container, kind := "Array", "array"
			if proto[list] == '%' {
				container, kind = "Hash", "hash"
			}
			return types.Signature{}, false, fmt.Errorf(`prototype (%s): List parameter %c is not last; a single %s followed by more parameters is %s \%c, as in (%s\%s)`,
				proto, proto[list], kind, container, proto[list], proto[:list], proto[list:])
		}
		var p types.Param
		switch c {
		case '$':
			p = types.Param{Sigil: '$', Type: types.Scalar}
		case '_':
			p = types.Param{Sigil: '$', Type: types.Scalar, Default: "$_"}
		case '+':
			p = types.Param{Sigil: '$', Type: types.Array | types.Hash | types.Scalar}
		case '@', '%':
			p = types.Param{Sigil: c, Type: types.List}
			list = i
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
// a `$` parameter typed exactly List (`Array|Hash|Scalar`) gives `+`, and one
// defaulting to `$_` gives `_`. A wider or other type stays `$`: measured on
// 5.42, `f(@a)` passes an ARRAY reference under `+` but the count under `$`.
// A trailing `;` is not a type, so `$;` and `;` come back as `$` and nothing;
// a declaration that writes them keeps its declared prototype. A `;` goes before the first optional
// parameter other than `_`, which is optional of itself.
func prototypeFromTypes(sig types.Signature) string {
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
		case p.Sigil == '$' && p.Type == types.List:
			b.WriteByte('+')
		case topic:
			b.WriteByte('_')
		default:
			b.WriteByte(p.Sigil)
		}
	}
	return b.String()
}

// agreement reports a declared prototype, in its parentheses, that
// disagrees with the one its types give. The declared one is read through
// the table first, so what the table does not tell apart agrees. A
// backslashed prototype is not compared: aliased parameters are not
// derived yet (see typesFromPrototype).
func agreement(declared, fromTypes string) error {
	sig, derived, err := typesFromPrototype(strings.TrimSuffix(strings.TrimPrefix(declared, "("), ")"))
	if err != nil || !derived {
		return err
	}
	if canon := prototypeFromTypes(sig); canon != fromTypes {
		return fmt.Errorf(":prototype%s disagrees with its types, which give (%s)", declared, fromTypes)
	}
	return nil
}
