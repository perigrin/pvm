// ABOUTME: Reads a module's XS source for the subs it implements in C and their
// ABOUTME: prototypes, by xsubpp's rules -- what perl knows once the module loads.

package parse

import (
	"regexp"
	"strings"
)

var (
	xsModuleLine = regexp.MustCompile(`^MODULE\s*=\s*\S+\s+PACKAGE\s*=\s*(\S+)(?:\s+PREFIX\s*=\s*(\S+))?`)
	xsPrototypes = regexp.MustCompile(`^\s*PROTOTYPES:\s*(\w+)`)
	xsPrototype  = regexp.MustCompile(`^\s*PROTOTYPE:\s*(.*?)\s*$`)
	xsDecl       = regexp.MustCompile(`^([A-Za-z_]\w*)\s*\((.*)\)\s*$`)
)

// XSSubs maps each sub an XS file defines, by qualified name, to its
// prototype as this package spells one: "(...)", or "" for none.
//
// The rules are xsubpp's (ExtUtils::ParseXS). Nothing before the first
// MODULE line is XS. An XSUB is its name and parameters at column 0, on the
// line after its return type; PACKAGE names where it goes and PREFIX is
// stripped from its name. Prototypes are off until `PROTOTYPES: ENABLE`;
// with them on, each parameter is `$`, a `;` goes before the first one with
// a default and `...` adds `;@` (Node.pm's proto_string); a `PROTOTYPE:`
// line in the XSUB's body replaces that, and an empty one is `()`.
//
// ponytail: a typemap can give its type a prototype character other than
// `$`, and ALIAS names and BOOT-time newXS calls add subs; none is read.
// Read them if a module whose calls matter uses them.
func XSSubs(src []byte) map[string]string {
	subs := map[string]string{}
	var pkg, prefix, cur, prev string
	enabled, inXS, inPod := false, false, false
	for _, line := range strings.Split(string(src), "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case inPod:
			inPod = !strings.HasPrefix(line, "=cut")
		case strings.HasPrefix(line, "="):
			inPod = true
		case xsModuleLine.MatchString(line):
			m := xsModuleLine.FindStringSubmatch(line)
			pkg, prefix, cur, inXS = m[1], m[2], "", true
		case !inXS:
		case xsPrototypes.MatchString(line):
			enabled = strings.HasPrefix(xsPrototypes.FindStringSubmatch(line)[1], "ENABLE")
		case cur != "" && xsPrototype.MatchString(line):
			switch proto := xsPrototype.FindStringSubmatch(line)[1]; proto {
			case "DISABLE":
				subs[cur] = ""
			case "ENABLE":
			default:
				subs[cur] = "(" + strings.Join(strings.Fields(proto), "") + ")"
			}
		case xsDecl.MatchString(line) && isXSReturnType(prev):
			m := xsDecl.FindStringSubmatch(line)
			cur = pkg + "::" + strings.TrimPrefix(m[1], prefix)
			subs[cur] = ""
			if enabled {
				subs[cur] = xsDerivedPrototype(m[2])
			}
		}
		prev = line
	}
	return subs
}

// isXSReturnType reports whether line can be the return type an XSUB's name
// follows: at column 0, and C type words and stars only.
func isXSReturnType(line string) bool {
	if line == "" || line[0] == ' ' || line[0] == '\t' || line[0] == '#' {
		return false
	}
	return !strings.ContainsAny(line, "(){};:=")
}

// xsDerivedPrototype is xsubpp's prototype for a parameter list when
// prototypes are on.
func xsDerivedPrototype(params string) string {
	var b strings.Builder
	optional := false
	for _, p := range strings.Split(params, ",") {
		p = strings.TrimSpace(p)
		switch {
		case p == "":
		case p == "...":
			if !optional {
				b.WriteByte(';')
			}
			b.WriteByte('@')
			optional = true
		case strings.Contains(p, "="):
			if !optional {
				b.WriteByte(';')
				optional = true
			}
			b.WriteByte('$')
		default:
			b.WriteByte('$')
		}
	}
	return "(" + b.String() + ")"
}
