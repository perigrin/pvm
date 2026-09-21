// ABOUTME: Checks a declared lexical fact against the token stream.
// ABOUTME: Facts are written in the glossary's vocabulary, not our enum's.
package conformance

import (
	"fmt"
	"regexp"
	"strconv"
)

// Two sentence forms, which is all the corpus asserts so far. Both name a
// GLOSSARY category rather than one of our token kinds, so the same file
// means the same thing to a lexer whose kinds differ from ours.
//
//	one numeric literal whose text is ".5"
//	no operator whose text is "."
//
// A third form will be needed the first time a file counts something other
// than one; `countWords` is where that goes.
var (
	reCount = regexp.MustCompile(`^(one|no) ([a-z ]+?) whose text is (".*")$`)
)

// checkTokenFact asserts one declared fact about the token stream.
func checkTokenFact(t reporter, fact string, src []byte) {
	m := reCount.FindStringSubmatch(fact)
	if m == nil {
		t.Errorf("unparsable token fact %q\n\tsee conformance/GLOSSARY.md for the forms", fact)
		return
	}
	want, category, quoted := m[1], m[2], m[3]

	text, err := strconv.Unquote(quoted)
	if err != nil {
		t.Errorf("token fact %q: bad quoted text %s: %v", fact, quoted, err)
		return
	}

	match, ok := categories[category]
	if !ok {
		t.Errorf("token fact %q: unknown category %q\n\tadd it to conformance/GLOSSARY.md first", fact, category)
		return
	}

	got := 0
	for _, tk := range significant(src) {
		if match(tk.Kind) && string(src[tk.Start:tk.End]) == text {
			got++
		}
	}

	wantN := map[string]int{"one": 1, "no": 0}[want]
	if got != wantN {
		t.Errorf("%s\n\tgot %d, want %d\n\tactual tokens: %s",
			fact, got, wantN, describe(src))
	}
}

// describe renders the token stream for a failure message, so a refusal
// says what the lexer actually produced rather than only that it was wrong.
func describe(src []byte) string {
	out := ""
	for _, tk := range significant(src) {
		out += fmt.Sprintf("%s(%q) ", tk.Kind, string(src[tk.Start:tk.End]))
	}
	return out
}
