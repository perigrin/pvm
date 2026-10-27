// ABOUTME: Exposes package internals to this package's external tests, and only to them.
// ABOUTME: Compiled only under go test.
package parse

import (
	"fmt"
	"strings"

	"tamarou.com/pvm/internal/types"
)

// CoreTable is coreTable, for TestCoreDeclarationsArePerls (coretable_test.go): its test asks
// perl through conformance, which imports this package.
var CoreTable = coreTable

// DerivedPrototypes reads a CORE.tpm as coreProtos does, for
// TestCoreDerivedPrototypesArePerls (coretypes_test.go), and gives each
// builtin the prototype its line gives, without parentheses: the one its
// types derive, or as written where the line is untyped or writes a
// trailing `;` its types cannot (RFC 0001, "A typed signature and a
// prototype say the same thing"). A derived `@` stays `@`, which perl
// reports for a builtin that parses as one with no prototype does.
func DerivedPrototypes(src []byte) (map[string]string, error) {
	protos, sigs, err := coreProtos(src)
	if err != nil {
		return nil, err
	}
	for name, s := range sigs {
		// A multi's candidates derive one prototype by
		// 01a10dc1-3499-7e2b-9e6f-d935084437fa; an invocant derives none,
		// and so do `:unary` and `:listop`.
		if _, ok := protos[name]; !ok && len(s) == 1 && s[0].Invocant == nil && !s[0].Unary && !s[0].ListOp {
			protos[name] = prototypeFromTypes(s[0])
		}
	}
	return protos, nil
}

// CoreSignatures is coreSignatures, for TestCoreTypesMatchMeasuredSignatures
// (coretypes_test.go).
var CoreSignatures = coreSignatures

// CoreSignaturesOf reads a CORE.tpm as coreProtos does and gives its typed
// signatures by builtin, for the batch checks in coretypes_test.go, which
// run over copies of the file as well as the file itself.
func CoreSignaturesOf(src []byte) (map[string][]types.Signature, error) {
	_, sigs, err := coreProtos(src)
	return sigs, err
}

// FaithfulTokens is faithfulTokens, for TestFaithfulTokens and the
// closed-set tests (fidelity_test.go).
var FaithfulTokens = faithfulTokens

// CanonTokens is canonTokens, for TestCanonMarksParenNodePairs
// (fidelity_test.go).
var CanonTokens = canonTokens

// DroppedKinds is droppedKinds, for TestFaithfulForgivenessIsClosed
// (fidelity_test.go).
var DroppedKinds = droppedKinds

// CoreShapes is coreShapes, for the faithful queue's word categories
// (faithful_queue_test.go).
var CoreShapes = coreShapes

// FaithfulStop is where Faithful stops on n: the source and canon token
// indexes of the first difference it could not forgive, for the faithful
// queue (faithful_queue_test.go). differs is false on a faithful tree.
//
// faithfulTokens names only the source index of a token difference, so the
// canon index is counted from what it forgave: an extra-nesting paren
// advances the source alone, a last semicolon the canon alone, and every
// other step both.
func FaithfulStop(n *Node, src []byte) (w, g int, differs bool) {
	ok, forgiven, why := faithfulTokens(significantTokens(src), canonTokens(n, src))
	if ok {
		return 0, 0, false
	}
	if _, err := fmt.Sscanf(why, "token %d:", &w); err == nil {
		g = w
		for _, f := range forgiven {
			switch f.Rule {
			case ruleExtraNesting:
				g--
			case ruleLastSemicolon:
				g++
			}
		}
		return w, g, true
	}
	if i := strings.LastIndex(why, "(matched "); i >= 0 {
		if _, err := fmt.Sscanf(why[i:], "(matched %d and %d)", &w, &g); err == nil {
			return w, g, true
		}
	}
	panic("FaithfulStop: unrecognised verdict " + why)
}

// Triage is triage's strategy name, for TestTriageMatchesDefinedInvalidation
// (incremental_fuzz_test.go), whose §7.7.3 table holds strategy names.
func Triage(old *Node, src []byte, e Edit) string { return triage(old, src, e).String() }

// WidenAnchor is the span of the node widen anchors e at, for the same test.
func WidenAnchor(old *Node, src []byte, e Edit) (start, end int) {
	_, n := widen(old, src, e)
	return n.Start, n.End
}
