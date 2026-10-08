// ABOUTME: Derives a .pmt declaration's typed signature from its prototype, and its prototype from its types.
// ABOUTME: RFC 0001's table, "A typed signature and a prototype say the same thing", read both ways.
package parse

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"tamarou.com/pvm/internal/types"
)

// typesFromPrototype reads a prototype, without its parentheses, into the
// coarse typed signature RFC 0001's table gives each character:
//
//	$ Scalar $      @ List @      % List %      + Array|Hash|Scalar $
//	& Code &        * Glob *      _ Scalar $ = $_
//	\$ Scalar \$x   \@ Array \@a   \% Hash \%h   \& Code \&c
//
// and `;` makes each parameter after it optional. A parameter derived
// from a prototype has no name, since a prototype names none.
//
// derived is false for `\*` and `\[...]`. A glob slot has no
// spelling (RFC 0001, "The scalar container"), and one of several
// containers, `\[%@]`, is the prototype a multi's candidates derive, by
// 01a10dc1-3499-7e2b-9e6f-d935084437fa ("Derive one prototype from several
// multi candidates").
func typesFromPrototype(proto string) (sig types.Signature, derived bool, err error) {
	// The type each backslashed character's parameter takes: the one
	// container the caller writes.
	aliasTypes := map[byte]types.Type{'$': types.Scalar, '@': types.Array, '%': types.Hash, '&': types.Code}
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
			if i+1 < len(proto) && strings.IndexByte("$@%&", proto[i+1]) >= 0 {
				i++
				p = types.Param{Sigil: proto[i], Type: aliasTypes[proto[i]], Alias: true}
				break
			}
			if i+1 < len(proto) && strings.IndexByte("*[", proto[i+1]) >= 0 {
				return types.Signature{}, false, nil
			}
			return types.Signature{}, false, fmt.Errorf("prototype (%s) has %q, which is not a prototype character", proto, proto[i:min(i+2, len(proto))])
		default:
			return types.Signature{}, false, fmt.Errorf("prototype (%s) has %q, which is not a prototype character", proto, proto[i:i+1])
		}
		p.Required = !p.Slurpy() && c != '_' && !optional
		sig.Params = append(sig.Params, p)
	}
	return sig, true, nil
}

// prototypeFromTypes gives the prototype, without its parentheses, that a
// typed signature's parameters state, the table read the other way. Types
// are finer than prototypes, so `Str $x` and `Scalar $x` both give `$`;
// a `$` parameter typed exactly `Array|Hash|Scalar`, or that and Void,
// which is List, gives `+`, and one defaulting to `$_` gives `_`. A wider or
// other type stays `$`: measured on 5.42, `f(@a)` passes an ARRAY reference
// under `+` but the count under `$`. Void does not change the character:
// measured on 5.42, `f(())` under `+` passes one undef, the empty list in
// scalar context, and whether the slot may be left empty is its `;`.
// A trailing `;` is not a type, so `$;` and `;` come back as `$` and nothing;
// a declaration that writes them keeps its declared prototype. A `;` goes before the first optional
// parameter other than `_`, which is optional of itself.
func prototypeFromTypes(sig types.Signature) string {
	var b strings.Builder
	semicolon := false
	for _, p := range sig.Params {
		slurpy := p.Slurpy()
		topic := p.Sigil == '$' && p.Default == "$_"
		if !p.Required && !slurpy && !topic && !semicolon {
			b.WriteByte(';')
			semicolon = true
		}
		switch {
		case p.Alias:
			b.WriteByte('\\')
			b.WriteByte(p.Sigil)
		case p.Sigil == '$' && p.Type|types.Void == types.List:
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
// prototype the table does not derive is not compared (see
// typesFromPrototype), except a `\[...]`, which a multi's candidates
// derive: it is compared as written, since perl keeps it as written
// (measured on 5.42, `sub f (\[@%])` reports `\[@%]`), less a trailing
// `;`, which no type states.
func agreement(declared, fromTypes string) error {
	proto := strings.TrimSuffix(strings.TrimPrefix(declared, "("), ")")
	canon := strings.TrimRight(proto, ";")
	if !strings.Contains(proto, `\[`) {
		sig, derived, err := typesFromPrototype(proto)
		if err != nil || !derived {
			return err
		}
		canon = prototypeFromTypes(sig)
	}
	if canon != fromTypes {
		return fmt.Errorf(":prototype%s disagrees with its types, which give (%s)", declared, fromTypes)
	}
	return nil
}

// prototypeFromCandidates gives the one prototype, without its
// parentheses, that a multi's candidates state together. Candidates whose
// prototypes differ only in which container one aliased parameter takes,
// `\%` in one and `\@` in another, derive RFC 0001's `\[$@%]` row with
// those containers in declaration order: `each`'s `(Hash \%h)` and
// `(Array \@a)` derive `\[%@]`. Candidates that differ in more than one
// position, in arity, or in a position that is not aliased derive none,
// and nor does a candidate with an invocant colon; err says why, for a
// declaration that states a prototype they must give.
func prototypeFromCandidates(sigs []types.Signature) (proto string, err error) {
	// Candidates group by their parameters, as the prototype sees them:
	// `:context` and the return type never split a group, so `keys`'
	// four candidates are two groups and `localtime`'s two are one.
	var protos []string
	for _, s := range sigs {
		if s.Invocant != nil {
			return "", errors.New("an invocant colon derives no prototype")
		}
		if p := prototypeFromTypes(s); !slices.Contains(protos, p) {
			protos = append(protos, p)
		}
	}
	// A `\` is always followed by its sigil, so one differing byte after
	// a shared `\` is one aliased position whose container differs.
	inexpressible := fmt.Errorf("its candidates give (%s), which no one prototype states", strings.Join(protos, "), ("))
	at := -1
	for _, p := range protos[1:] {
		if len(p) != len(protos[0]) {
			return "", inexpressible
		}
	}
	for i := range len(protos[0]) {
		for _, p := range protos[1:] {
			if p[i] != protos[0][i] && at != i {
				if at >= 0 || i == 0 || protos[0][i-1] != '\\' {
					return "", inexpressible
				}
				at = i
			}
		}
	}
	if at < 0 {
		return protos[0], nil
	}
	var union strings.Builder
	for _, p := range protos {
		union.WriteByte(p[at])
	}
	return protos[0][:at-1] + `\[` + union.String() + "]" + protos[0][at+1:], nil
}
