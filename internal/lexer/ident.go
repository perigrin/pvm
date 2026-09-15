// ABOUTME: The identifier character class, and the `use utf8` pragma that widens it.
// ABOUTME: §0.13 rank 1 — 34 corpus files, every one of them a lexer failure rather than a grammar one.

package lexer

import (
	"unicode/utf8"

	"unicode"
)

// identStart reports whether r may begin an identifier.
//
// perldata.pod:179-180 gives perl's rule as an INTERSECTION:
//
//	(\p{Word} & \p{XID_Start}) + [_]
//
// A restriction of XID, not an extension of it. This matters because an
// earlier draft of the issue said "XID with perl-specific additions, and the
// additions are where the corpus files live" -- which would have sent the
// implementer hunting for extensions that do not exist. All 34 corpus files
// use plain XID letters.
//
// Without the pragma the class is ASCII only. See the measurement in the
// body: perl rejects every high byte, Latin-1 letters included.
func identStart(r rune, utf8Pragma bool) bool {
	if r == '_' {
		return true
	}
	if r < utf8.RuneSelf {
		return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
	}
	// Without the pragma NO high byte may appear in an identifier. Measured
	// on perl 5.42 against raw bytes in a file, which is the only way to ask
	// the question honestly:
	//
	//	my $F\xC3\xB8  ->  Unrecognized character \xC3
	//	my $F\xE9      ->  Unrecognized character \xE9
	//
	// Both rejected, and \xE9 is Latin-1 `é`. An earlier draft of this file
	// allowed Latin-1 letters through on the strength of `perl -e 'my $á'`
	// failing in the COMPILER rather than the lexer -- but that error came
	// after the lexer had already refused the byte, so it proved the
	// opposite of what it appeared to.
	if !utf8Pragma {
		return false
	}
	return unicode.IsLetter(r) && isXIDStart(r)
}

// identContinue reports whether r may appear after the first character.
//
//	(\p{Word} & \p{XID_Continue})*
func identContinue(r rune, utf8Pragma bool) bool {
	if r == '_' {
		return true
	}
	if r < utf8.RuneSelf {
		return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9'
	}
	if !utf8Pragma {
		return false
	}
	return (unicode.IsLetter(r) || unicode.IsDigit(r)) && isXIDContinue(r)
}

// isXIDStart and isXIDContinue approximate the Unicode XID properties with
// the tables the standard library ships.
//
// Go has no unicode.XID_Start table. The difference from L+Nl+Other_ID_Start
// is a handful of characters that are excluded for normalisation stability,
// none of which appear in the corpus; treating the L categories as XID is
// therefore right for every file this milestone measures and wrong only in
// cases nothing tests. Recorded rather than hidden: if a corpus file ever
// turns up needing the exact table, this is the function to fix.
func isXIDStart(r rune) bool {
	return unicode.IsOneOf([]*unicode.RangeTable{
		unicode.L, unicode.Nl, unicode.Other_ID_Start,
	}, r) && !unicode.Is(unicode.Pattern_Syntax, r)
}

func isXIDContinue(r rune) bool {
	if isXIDStart(r) {
		return true
	}
	return unicode.IsOneOf([]*unicode.RangeTable{
		unicode.Mn, unicode.Mc, unicode.Nd, unicode.Pc, unicode.Other_ID_Continue,
	}, r)
}

// scanIdentRunes consumes an identifier at the cursor, including both package
// separators, and reports whether it consumed anything.
//
// BOTH separators. §2.6.3 and §2.15 item 4; §0.13 rank 10. `$main'a` is
// `$main::a`, measured:
//
//	$ perl -e '$main::a = 5; print $main'"'"'a'
//	5
//
// The apostrophe separates only BETWEEN identifier characters, which is what
// keeps `$'` -- the postmatch variable -- from being read as the start of a
// package name. Both appear on one corpus line, comp/package.t:17:
//
//	$main'a = $'b;
func (l *lexer) scanIdentRunes() bool {
	started := l.pos
	first := true
	for l.pos < len(l.src) {
		// A separator joins two name parts. `::` always; `'` only when a name
		// character follows, or `$'` would start a package name and swallow
		// the rest of the line.
		if !first {
			if l.src[l.pos] == ':' && l.pos+1 < len(l.src) && l.src[l.pos+1] == ':' {
				l.pos += 2
				continue
			}
			if l.src[l.pos] == '\'' && l.pos+1 < len(l.src) {
				r, _ := utf8.DecodeRune(l.src[l.pos+1:])
				if identStart(r, l.utf8Pragma) {
					l.pos++
					continue
				}
			}
		}

		r, size := utf8.DecodeRune(l.src[l.pos:])
		if r == utf8.RuneError && size <= 1 {
			break
		}
		ok := identContinue(r, l.utf8Pragma)
		if first {
			ok = identStart(r, l.utf8Pragma)
		}
		if !ok {
			break
		}
		l.pos += size
		first = false
	}
	return l.pos > started
}

// notePragma watches for `use utf8` and `no utf8` as they are lexed.
//
// The pragma is file-level here, deliberately. It is lexically scoped in
// perl, but spec §2.3.1 says "a file-level flag is a reasonable v1 with a
// recorded limitation", and block scoping needs a brace stack plus the
// ability to tell a real `{` from one inside a string -- neither of which
// exists yet. An earlier draft of this issue demanded block scoping and would
// have blocked on two later issues.
//
// LIMITATION, recorded: a `no utf8` inside a block turns the pragma off for
// the rest of the FILE rather than the rest of the block. No corpus file does
// that; when one does, this is where scoping goes.
func (l *lexer) notePragma(start int) {
	word := string(l.src[start:l.pos])
	switch word {
	case "use":
		l.pendingPragma = 1
	case "no":
		l.pendingPragma = 2
	case "utf8":
		switch l.pendingPragma {
		case 1:
			l.utf8Pragma = true
		case 2:
			l.utf8Pragma = false
		}
		l.pendingPragma = 0
	default:
		l.pendingPragma = 0
	}
}
