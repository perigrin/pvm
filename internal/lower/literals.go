// ABOUTME: Lowering of literals and names (§4.14.2): NumLit, StrLit, QwList and Bareword from their CST terms.
// ABOUTME: Every fact is read off the term's Text -- the delimiter, the quote operator, the digits -- never re-parsed.

package lower

import (
	"regexp"
	"strings"

	"tamarou.com/pvm/internal/parse"
)

// literal lowers a childless term that is a number, a string or a qw list,
// and declines anything else.
func (l *lowerer) literal(n *parse.Node) Node {
	t := n.Text
	if len(n.Children) != 0 || t == "" {
		return nil
	}
	switch c := t[0]; {
	case c >= '0' && c <= '9':
		return &NumLit{Text: t, IsInt: isInt(t)}
	case c == '\'':
		return &StrLit{Value: unescape(t[1:len(t)-1], '\'', '\'')}
	case c == '"':
		return &StrLit{Value: t[1 : len(t)-1], Interpolated: true}
	case c == '`':
		return &StrLit{Value: t[1 : len(t)-1], Interpolated: true, Backtick: true}
	}
	m := quoteOp.FindStringSubmatch(t)
	if m == nil {
		return nil
	}
	open := m[2][0]
	body := t[len(m[0]) : len(t)-1]
	switch m[1] {
	case "q":
		return &StrLit{Value: unescape(body, open, closer(open))}
	case "qq":
		return &StrLit{Value: body, Interpolated: true}
	case "qx":
		// qx'' is the one spelling that does not interpolate. perl:
		// qx'echo $HOME' hands the shell $HOME, not perl's.
		return &StrLit{Value: body, Interpolated: open != '\'', Backtick: true}
	}
	// ponytail: qw's \\ and \<delim> escapes are not resolved; resolve
	// them with unescape if a word ever needs a literal delimiter.
	return &QwList{Words: strings.Fields(body)}
}

// quoteOp matches the head of a q, qq, qw or qx term up to and including
// its opening delimiter.
var quoteOp = regexp.MustCompile(`^(q[qwx]?)\s*([^\w\s])`)

// isInt reports whether a number's spelling is an integer: hex and binary
// always are, and a decimal is unless it has a point or an exponent.
func isInt(t string) bool {
	if len(t) > 1 && t[0] == '0' && strings.ContainsRune("xXbB", rune(t[1])) {
		return true
	}
	return !strings.ContainsAny(t, ".eE")
}

// closer is the delimiter that closes one opened with open: its pair for a
// bracket, itself otherwise.
func closer(open byte) byte {
	if i := strings.IndexByte("([{<", open); i >= 0 {
		return ")]}>"[i]
	}
	return open
}

// unescape resolves the two escapes a non-interpolating string has: a
// backslash before a backslash or before either delimiter stands for that
// byte, and every other backslash is itself. perl: q(a\)b) is a)b and
// q(a\\b) is a\b.
func unescape(body string, open, close byte) string {
	var b strings.Builder
	for i := 0; i < len(body); i++ {
		if body[i] == '\\' && i+1 < len(body) {
			if c := body[i+1]; c == '\\' || c == open || c == close {
				i++
			}
		}
		b.WriteByte(body[i])
	}
	return b.String()
}
