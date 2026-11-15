// ABOUTME: Resolve, after the forest: a scope-decided word keeps only the readings toke.c gives it under the subs in scope at its byte.
// ABOUTME: Declared subs and packages come from the file's own forest; imported subs from the caller's Scope, parse.Imports' file-wide answer.

package forest

import "tamarou.com/pvm/internal/parse"

// Scope is what resolve reads from outside the file: the subs its use
// statements import, as parse.Imports gives them over a tree from
// parse.ParseFileFrom or parse.ParseWithLoader. The caller decides how they
// were obtained; the forest never loads a module. A nil map imports nothing.
// shortcut: imports are file-wide, so a call before its use reads as
// imported (perigrin, 2026-11-10, "File-wide for now"): parse.Imports does
// not say which use brought a name in; upgrade when it carries each import's position.
type Scope struct {
	Imports map[string]parse.Import
}
