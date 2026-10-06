// ABOUTME: Exposes package internals to this package's external tests, and only to them.
// ABOUTME: Compiled only under go test.
package parse

// CoreTable is coreTable, for TestCoreDeclarationsArePerls (coretable_test.go): its test asks
// perl through conformance, which imports this package.
var CoreTable = coreTable

// DerivedPrototypes reads a CORE.pmt as coreProtos does, for
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
		// 01a10dc1-3499-7e2b-9e6f-d935084437fa; an invocant derives none.
		if _, ok := protos[name]; !ok && len(s) == 1 && s[0].Invocant == nil {
			protos[name] = prototypeFromTypes(s[0])
		}
	}
	return protos, nil
}

// CoreSignatures is coreSignatures, for TestCoreTypesMatchMeasuredSignatures
// (coretypes_test.go).
var CoreSignatures = coreSignatures
