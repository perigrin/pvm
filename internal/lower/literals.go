// ABOUTME: Lowering of literals and names (§4.14.2): NumLit, StrLit, QwList and Bareword from their CST terms.
// ABOUTME: Every fact is read off the term's Text -- the delimiter, the quote operator, the digits -- never re-parsed.

package lower

import (
	"regexp"
	"strconv"
	"strings"

	"tamarou.com/pvm/internal/parse"
)

// literal lowers a childless term that is a number, a string, a qw list or a
// bare word, and declines anything else.
func (l *lowerer) literal(n *parse.Node) Node {
	t := n.Text
	if len(n.Children) != 0 || t == "" {
		return nil
	}
	switch c := t[0]; {
	case vstring.MatchString(t):
		return &StrLit{Value: vstringValue(t)}
	case c >= '0' && c <= '9', c == '.' && len(t) > 1 && t[1] >= '0' && t[1] <= '9':
		// perl: .5 + 1 is 1.5.
		return &NumLit{Text: t, IsInt: isInt(t)}
	case c == '\'':
		return &StrLit{Value: unescape(t[1:len(t)-1], '\'', '\'')}
	case c == '"':
		return &StrLit{Value: t[1 : len(t)-1], Interpolated: true}
	case c == '`':
		return &StrLit{Value: t[1 : len(t)-1], Interpolated: true, Backtick: true}
	}
	if strings.HasPrefix(t, "<<") && t != "<<>>" {
		return heredoc(t)
	}
	if word.MatchString(t) {
		// A class name or the stat buffer `_`. A parent whose position
		// quotes the word says so: see hashKey and fatWord.
		return &Bareword{Name: t}
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

// heredoc lowers a heredoc's opener, `<<"EOT"`, to the string it stands
// for. The quote after `<<`, and after `~` and spaces, decides as a
// string's delimiter does: `'` and `\` do not interpolate, a backtick
// runs a command, and `"` or a bare terminator interpolates.
func heredoc(t string) Node {
	q := strings.TrimLeft(t[2:], "~ \t")
	s := &StrLit{Interpolated: true, Heredoc: true}
	if q != "" {
		switch q[0] {
		case '\'', '\\':
			s.Interpolated = false
		case '`':
			s.Backtick = true
		}
	}
	return s
}

// vstring matches a v-string: v5.36.0, v1.2 and v65, or a number with two
// or more points, 5.36.0. perl 5.42: `5.36.0 eq v5.36.0`, and
// `ref \v5.36.0` is VSTRING.
var vstring = regexp.MustCompile(`^(v\d[\d_]*(\.\d[\d_]*)*|\d[\d_]*(\.\d[\d_]*){2,})$`)

// vstringValue is the string a v-string spells: one character per part,
// numbered by it. perl: v65 is "A".
func vstringValue(t string) string {
	var b strings.Builder
	for _, part := range strings.Split(strings.TrimPrefix(t, "v"), ".") {
		n, _ := strconv.Atoi(strings.ReplaceAll(part, "_", ""))
		b.WriteRune(rune(n))
	}
	return b.String()
}

// vstringCall lowers a `call` that is a dotless v-string, `v65`: a word the
// parser could not resolve, standing alone. perl 5.42 reads `my $v = v65`
// as "A", and calls the sub when `sub v65` is declared before it.
func (l *lowerer) vstringCall(n *parse.Node) Node {
	if n.Resolved || n.Indirect || len(n.Children) != 0 || parenthesized(n) || !vstring.MatchString(n.Text) {
		return nil
	}
	return &StrLit{Value: vstringValue(n.Text)}
}

// quoteOp matches the head of a q, qq, qw or qx term up to and including
// its opening delimiter.
var quoteOp = regexp.MustCompile(`^(q[qwx]?)\s*([^\w\s])`)

// word matches a bare word, package-qualified or not: Foo, Foo::Bar, _.
var word = regexp.MustCompile(`^[A-Za-z_]\w*(::\w+)*(::)?$`)

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

// hashKey lowers the subscript of an `index "{"`, where a bare word is a
// quoted string: perl reads $h{k} as $h{"k"}.
func (l *lowerer) hashKey(n *parse.Node) Node {
	if n.Kind == parse.Term && len(n.Children) == 0 && word.MatchString(n.Text) {
		return l.stamp(n, &Bareword{Name: n.Text, Quoted: true})
	}
	return l.expr(n)
}

// fatWord lowers a `call` with Fat set and no arguments: the left of a fat
// comma, which perl reads as a string (`(a => 1)` is ("a", 1)). The Fat flag
// is on the word itself, so no position is needed.
func (l *lowerer) fatWord(n *parse.Node) Node {
	if !n.Fat || len(n.Children) != 0 {
		return nil
	}
	return &Bareword{Name: n.Text, Quoted: true}
}
