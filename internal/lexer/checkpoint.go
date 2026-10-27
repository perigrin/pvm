// ABOUTME: The lexer's resumable State, the safe checkpoints a full lex records, and Resume (spec §6.3.1).
// ABOUTME: Lexing from a checkpoint to EOF reproduces the full Tokenize* stream's suffix from that byte.

package lexer

import (
	"bytes"
	"maps"
	"reflect"
	"slices"
)

// State is everything the lexer carries from one token to the next: the
// expect state, the bracket stack, the heredoc queue, the pragmas, the
// declared names and every one-token carry. If two states are Equal, lexing
// forward from the same byte of the same source produces identical tokens.
//
// It is the lexer itself with its cursor, source and output cleared, so a
// field added to lexer is part of the state without anyone remembering to add
// it. TestLexStateClonesEveryReference holds Clone to the same rule.
//
// Quotes, regexes, POD, heredoc bodies and format bodies are each one token,
// scanned within a single step, so the lexer has no sublex or delimiter stack
// for a state to carry: a step boundary is never inside one.
type State struct{ l lexer }

// Checkpoint is a byte offset and the state the lexer had there.
type Checkpoint struct {
	Byte  int
	State State
}

// checkpointStride is the minimum spacing between recorded checkpoints, so a
// 5,000-line file holds hundreds rather than thousands (spec §6.3.1).
const checkpointStride = 512

// StartState is Tokenize's state at byte 0.
func StartState() State { return StartStateBarewords(BarewordOperators) }

// StartStateBarewords is TokenizeBarewords's state at byte 0.
func StartStateBarewords(barewordOps []string) State {
	return State{lexer{expect: XState, barewordOps: barewordOps}}
}

// StartStateTyped is TokenizeTyped's state at byte 0.
func StartStateTyped() State {
	return State{lexer{expect: XState, signatures: true, typed: true, barewordOps: BarewordOperators}}
}

// Checkpoints lexes src from start, as the Tokenize* call start belongs to
// does, and returns its tokens with the safe checkpoints it passed: byte 0,
// then at most one per checkpointStride bytes.
func Checkpoints(src []byte, start State) ([]Token, []Checkpoint) {
	return checkpoints(src, start, checkpointStride)
}

func checkpoints(src []byte, start State, stride int) ([]Token, []Checkpoint) {
	l := start.Clone().l
	l.src = src
	cps := []Checkpoint{{Byte: 0, State: start.Clone()}}
	for l.pos < len(l.src) {
		l.step(scanOne)
		if l.pos-cps[len(cps)-1].Byte >= stride && l.atCheckpoint() {
			cps = append(cps, Checkpoint{Byte: l.pos, State: State{l}.Clone()})
		}
	}
	return l.toks, cps
}

// Resume lexes src from cp to the end of input. src must hold the same bytes
// before cp.Byte as the source cp was recorded on: a scanner reads a few
// bytes and tokens behind the cursor, and a checkpoint only follows a `;` or
// a block's `}`, which none of those reads looks past.
func Resume(src []byte, cp Checkpoint) []Token {
	l := cp.State.Clone().l
	l.src, l.pos = src, cp.Byte
	return tokenize(&l)
}

// atCheckpoint reports whether the cursor sits just after a `;` or a block's
// `}` at bracket depth 0, in a safe state.
func (l *lexer) atCheckpoint() bool {
	t := l.toks[len(l.toks)-1]
	return t.End == l.pos && (t.Kind == Semicolon || t.Kind == CloseBracket && l.src[t.Start] == '}') &&
		l.safe()
}

// safe is spec §6.3.1's safe-checkpoint condition in this lexer's terms: a
// statement is starting, no bracket is open, no heredoc is queued and no
// format body is about to begin.
func (l *lexer) safe() bool {
	return l.expect == XState && len(l.brackets) == 0 && len(l.pending) == 0 &&
		!l.pendingFormat && !l.inFormat
}

// Equal reports whether s and o lex forward identically. An empty collection
// equals a missing one.
func (s State) Equal(o State) bool {
	return reflect.DeepEqual(s.Clone().l, o.Clone().l)
}

// Clone is a deep copy, so the lexer going on from a checkpoint cannot write
// into it. Empty collections become nil, which is what lets Equal compare
// with reflect.DeepEqual.
func (s State) Clone() State {
	c := s.l
	c.src, c.pos, c.toks = nil, 0, nil
	c.barewordOps = cloneOrNil(c.barewordOps)
	c.brackets = cloneOrNil(c.brackets)
	c.lexSubs = cloneOrNil(c.lexSubs)
	c.pending = cloneOrNil(c.pending)
	for i := range c.pending {
		c.pending[i].term = bytes.Clone(c.pending[i].term)
	}
	if len(c.constSubs) == 0 {
		c.constSubs = nil
	} else {
		c.constSubs = maps.Clone(c.constSubs)
	}
	return State{c}
}

func cloneOrNil[T any](s []T) []T {
	if len(s) == 0 {
		return nil
	}
	return slices.Clone(s)
}
