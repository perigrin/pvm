// ABOUTME: The token cache (spec §6.3.2): a document's tokens and checkpoints, updated in place after an edit.
// ABOUTME: Re-lexes from the checkpoint before the edit until the stream resyncs with the old one, then shifts the old tail.

package lexer

import (
	"slices"
	"sort"
)

// TokenCache is one document's token stream and the checkpoints its lex
// passed. After every Edit it equals a fresh Checkpoints of the edited bytes'
// tokens, and each checkpoint resumes to that stream's suffix.
type TokenCache struct {
	Tokens      []Token
	Checkpoints []Checkpoint // sorted by Byte, the first at byte 0
	stride      int
}

// NewTokenCache lexes src from start, as the Tokenize* call start belongs to
// does.
func NewTokenCache(src []byte, start State) *TokenCache {
	return newTokenCache(src, start, checkpointStride)
}

func newTokenCache(src []byte, start State, stride int) *TokenCache {
	toks, cps := checkpoints(src, start, stride)
	return &TokenCache{Tokens: toks, Checkpoints: cps, stride: stride}
}

// Edit updates the cache for src, the old source with its bytes
// [start, oldEnd) replaced by src[start:newEnd], and returns how many tokens
// it re-lexed.
//
// It re-lexes from the last checkpoint before start until the new stream
// resyncs with the old one at an old checkpoint past the edit: the token
// before it equal in Kind, Flags and shifted span, and the lexer State equal
// to the checkpoint's. From there forward lexing is the old stream shifted,
// so the old tail is reused. A stream that never resyncs is lexed to EOF,
// once.
func (c *TokenCache) Edit(src []byte, start, oldEnd, newEnd int) int {
	delta := newEnd - oldEnd
	// Strictly before start: the tokens before a checkpoint depend on no byte
	// past it except, perhaps, the one a scanner peeks at to end its token.
	ci := max(sort.Search(len(c.Checkpoints), func(i int) bool { return c.Checkpoints[i].Byte >= start })-1, 0)
	from := c.Checkpoints[ci]
	firstOld := sort.Search(len(c.Tokens), func(i int) bool { return c.Tokens[i].Start >= from.Byte })

	l := from.State.Clone().l
	l.src, l.pos = src, from.Byte
	var fresh []Checkpoint
	last := from.Byte
	syncTok, syncCp := len(c.Tokens), len(c.Checkpoints)
	for k := ci + 1; l.pos < len(l.src); {
		l.step(scanOne)
		for k < len(c.Checkpoints) && c.Checkpoints[k].Byte+delta < l.pos {
			k++
		}
		if k < len(c.Checkpoints) && c.Checkpoints[k].Byte+delta == l.pos {
			if i, ok := c.resyncsAt(&l, k, oldEnd, delta); ok {
				syncTok, syncCp = i, k
				break
			}
		}
		if l.pos-last >= c.stride && l.atCheckpoint() {
			fresh = append(fresh, Checkpoint{Byte: l.pos, State: State{l}.Clone()})
			last = l.pos
		}
	}

	toks := slices.Concat(c.Tokens[:firstOld], l.toks, c.Tokens[syncTok:])
	for i := firstOld + len(l.toks); i < len(toks); i++ {
		toks[i].Start += delta
		toks[i].End += delta
	}
	cps := slices.Concat(c.Checkpoints[:ci+1], fresh, c.Checkpoints[syncCp:])
	for i := ci + 1 + len(fresh); i < len(cps); i++ {
		cps[i].Byte += delta
	}
	c.Tokens, c.Checkpoints = toks, cps
	return len(l.toks)
}

// resyncsAt reports whether the lexer l, just past the token it last emitted,
// has resynced with the old stream at old checkpoint k, and if so the index of
// the first old token after it. The old token ending at the checkpoint must
// lie wholly after the edit, so the bytes a resume may look back over are
// unchanged; Flags count because a `}` that closed a block and one that
// closed a hash leave different states, and State counts because equal
// tokens can be lexed in states that diverge later (spec §6.3).
func (c *TokenCache) resyncsAt(l *lexer, k, oldEnd, delta int) (int, bool) {
	cp := c.Checkpoints[k]
	i := sort.Search(len(c.Tokens), func(i int) bool { return c.Tokens[i].End >= cp.Byte })
	if i == len(c.Tokens) || c.Tokens[i].Start < oldEnd {
		return 0, false
	}
	old := c.Tokens[i]
	old.Start += delta
	old.End += delta
	return i + 1, old == l.toks[len(l.toks)-1] && State{*l}.Equal(cp.State)
}
