// ABOUTME: Aligns each corpus file's -DT token stream to internal/lexer's tokens and writes one fixture per file (spec §12).
// ABOUTME: Offline only, run by tdt-align.sh: go run ./internal/chart/offline/tdt/align <capture-dir> <t1-dir> <out-dir>.

// The capture directory holds one tdt-capture.sh stream per corpus file at
// its key: the repository path for in-tree files, T1/<name>.t for T1's. Each
// fixture is written to <out-dir>/<key>.tdt (TDT_FIXTURES), in the form
// offline/README.md states, and a summary goes to stdout: per file, the
// toke.c runs no pvm token matched, which is where module tokens leaking past
// the capture's bracketing show.
package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"tamarou.com/pvm/internal/lexer"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintln(os.Stderr, "usage: align <capture-dir> <t1-dir> <out-dir>")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2], os.Args[3]); err != nil {
		fmt.Fprintln(os.Stderr, "align:", err)
		os.Exit(2)
	}
}

func run(capDir, t1Dir, outDir string) error {
	begin := time.Now()
	for _, d := range []string{"internal", "T1"} {
		if err := os.RemoveAll(filepath.Join(outDir, d)); err != nil {
			return err
		}
	}
	var files, pvmToks, aligned, runs int
	err := filepath.WalkDir(capDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		key, _ := filepath.Rel(capDir, path)
		key = filepath.ToSlash(key)
		srcPath := key
		if name, ok := strings.CutPrefix(key, "T1/"); ok {
			srcPath = filepath.Join(t1Dir, name)
		}
		src, err := os.ReadFile(srcPath)
		if err != nil {
			return err
		}
		stream, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		toks, marker := parseStream(string(stream))
		var b strings.Builder
		fmt.Fprintf(&b, "# tdt: %s\n", key)
		if marker != "" {
			// A refused, timed-out or truncated capture is not aligned (perigrin: "Timeout, mark refused").
			fmt.Fprintf(&b, "# capture: %s\n", marker)
		} else {
			b.WriteString("# capture: whole\n")
			lines, st := align(src, toks)
			for _, l := range lines {
				b.WriteString(l + "\n")
			}
			pvmToks += st.pvm
			aligned += st.aligned
			runs += len(st.runs)
			for _, r := range st.runs {
				fmt.Printf("%s: unmatched toke.c run: %s\n", key, r)
			}
		}
		files++
		out := filepath.Join(outDir, key+".tdt")
		if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
			return err
		}
		return os.WriteFile(out, []byte(b.String()), 0o644)
	})
	if err != nil {
		return err
	}
	fmt.Printf("aligned %d files: %d of %d pvm tokens aligned, %d unmatched toke.c runs, %.1f s\n",
		files, aligned, pvmToks, runs, time.Since(begin).Seconds())
	return nil
}

// tok is one -DT token: its perly.y type and the value tdt-capture.sh printed.
type tok struct{ typ, val string }

// parseStream reads a tdt-capture.sh stream; marker is the text of a
// trailing `# refused`, `# timeout` or `# truncated` line, "" for a whole one.
func parseStream(s string) (toks []tok, marker string) {
	for _, l := range strings.Split(strings.TrimSuffix(s, "\n"), "\n") {
		if l == "" {
			continue
		}
		if m, ok := strings.CutPrefix(l, "# "); ok {
			marker = m
			continue
		}
		typ, val, _ := strings.Cut(l, "\t")
		toks = append(toks, tok{typ, val})
	}
	return toks, marker
}

// fold is a toke.c token folded onto a pvm token by a fitting family; an
// empty typ marks a token paired through a family with nothing folded.
type fold struct{ fam, typ string }

// unit is what one pvm token is matched against: the toke.c token carrying
// its text, the tokens folded onto it, and the pvm token it matches.
type unit struct {
	head  tok
	folds []fold
	match func(p pvmTok) bool
}

type pvmTok struct {
	start, end int
	kind       lexer.Kind
	text       string
}

type stats struct {
	pvm, aligned int
	runs         []string
}

// align pairs a whole -DT stream with src's pvm tokens and returns the
// fixture lines, in pvm order.
func align(src []byte, toks []tok) ([]string, stats) {
	var ps []pvmTok
	for _, t := range lexer.Tokenize(src) {
		switch t.Kind {
		case lexer.Whitespace, lexer.Comment, lexer.Pod, lexer.DataSection:
			continue // toke.c delivers no token for these
		}
		ps = append(ps, pvmTok{t.Start, t.End, t.Kind, string(src[t.Start:t.End])})
	}
	us := units(toks)
	pair, unmatched := matchInOrder(us, ps)

	// Families that explain pvm tokens with no toke.c token of their own.
	fam := make([]string, len(ps))
	var openers []int
	owner := -1 // the last pair whose head owns the unaligned tokens after it
	for j, p := range ps {
		if pair[j] >= 0 {
			owner = -1
			switch us[pair[j]].head.typ {
			case "LABEL", "ATTRLIST":
				owner = j
			}
			if p.kind == lexer.HeredocOpen {
				openers = append(openers, j)
			}
			continue
		}
		switch {
		case p.kind == lexer.HeredocOpen:
			openers = append(openers, j)
		case p.kind == lexer.HeredocBody && len(openers) > 0:
			// toke.c delivers a heredoc's tokens at its marker; the body is the marker's.
			o := openers[0]
			openers = openers[1:]
			if pair[o] >= 0 {
				pair[j] = pair[o]
				fam[j] = "heredoc"
			}
		case p.kind == lexer.Word && (p.text == "return" || p.text == "CORE::return"):
			fam[j] = "unreported-return" // OLDLOP returns LSTOP without -DT's report
		case owner >= 0 && us[pair[owner]].head.typ == "LABEL":
			fam[j] = "label"
		case owner >= 0:
			fam[j] = "attribute"
		}
	}

	var lines []string
	var st stats
	st.pvm = len(ps)
	for j, p := range ps {
		for _, r := range unmatched[j] {
			lines = append(lines, fmt.Sprintf("unmatched %d %s", p.start, r))
			st.runs = append(st.runs, fmt.Sprintf("at byte %d: %s", p.start, r))
		}
		if pair[j] < 0 {
			lines = append(lines, strings.TrimSpace(fmt.Sprintf("%d %d - %s", p.start, p.end, fam[j])))
			continue
		}
		st.aligned++
		u := us[pair[j]]
		l := fmt.Sprintf("%d %d %s", p.start, p.end, u.head.typ)
		if fam[j] != "" {
			l += " " + fam[j]
		} else {
			last := ""
			for _, f := range u.folds {
				if f.fam != last {
					l += " " + f.fam
					last = f.fam
				}
				if f.typ != "" {
					l += " " + f.typ
				}
			}
		}
		lines = append(lines, l)
	}
	for _, r := range unmatched[len(ps)] {
		lines = append(lines, fmt.Sprintf("unmatched %d %s", len(src), r))
		st.runs = append(st.runs, fmt.Sprintf("at the end: %s", r))
	}
	return lines, st
}

// window bounds how far matchInOrder looks ahead for a match.
const window = 2000

// matchInOrder pairs units with pvm tokens in order. At a mismatch it skips
// whichever side reaches a match sooner; when both reach one at the same
// distance the order is ambiguous and both are skipped, so no pair is made
// by reordering. pair[j] is the unit paired with ps[j], or -1. A skipped
// unit an invented-token family explains folds onto the next pair; the rest
// are reported as runs before the pvm token they precede (unmatched[j];
// unmatched[len(ps)] after the last).
func matchInOrder(us []unit, ps []pvmTok) (pair []int, unmatched [][]string) {
	pair = make([]int, len(ps))
	for j := range pair {
		pair[j] = -1
	}
	unmatched = make([][]string, len(ps)+1)
	var pending []int
	flush := func(j int) { // settle the skipped units before ps[j]
		var run []string
		for _, k := range pending {
			if f := invented(us, k); f != "" && j < len(ps) && pair[j] >= 0 {
				u := &us[pair[j]]
				u.folds = append(u.folds, fold{f, us[k].head.typ})
				continue
			}
			run = append(run, us[k].head.typ)
			for _, f := range us[k].folds {
				if f.typ != "" {
					run = append(run, f.typ)
				}
			}
		}
		if len(run) > 0 {
			unmatched[j] = append(unmatched[j], strings.Join(run, " "))
		}
		pending = pending[:0]
	}
	i, j := 0, 0
	for i < len(us) && j < len(ps) {
		if us[i].match(ps[j]) {
			pair[j] = i
			flush(j)
			i, j = i+1, j+1
			continue
		}
		skipU, skipP := 1, 1
		for d := 1; d <= window; d++ {
			p := j+d < len(ps) && us[i].match(ps[j+d])
			u := i+d < len(us) && us[i+d].match(ps[j])
			if p && !u {
				skipU = 0
				break
			}
			if u && !p {
				skipP = 0
				break
			}
			if p && u {
				break
			}
		}
		for ; skipU > 0; skipU-- {
			pending = append(pending, i)
			i++
		}
		if skipP > 0 {
			j++
		}
	}
	for ; i < len(us); i++ {
		pending = append(pending, i)
	}
	flush(len(ps))
	return pair, unmatched
}

// invented names the family of a token toke.c invents with no pvm
// counterpart, folded onto the pvm token after it; "" for any other.
func invented(us []unit, k int) string {
	var next tok
	if k+1 < len(us) {
		next = us[k+1].head
	}
	switch us[k].head.typ {
	case "NOAMP":
		return "noamp"
	case "PERLY_AMPERSAND":
		if next.typ == "BAREWORD" || next.typ == "PRIVATEREF" {
			return "noamp"
		}
	case "KW_SUB_named", "KW_SUB_named_sig":
		if next.typ == "BAREWORD" && phasers[pvText(next.val)] {
			return "phaser"
		}
	}
	return ""
}

var phasers = map[string]bool{"BEGIN": true, "END": true, "INIT": true, "CHECK": true, "UNITCHECK": true, "AUTOLOAD": true, "DESTROY": true}

// units groups the stream into what pvm tokens are matched against, folding
// the split, invented and sublexed tokens the fitting families name.
func units(toks []tok) []unit {
	var us []unit
	var eof bool
	if n := len(toks); n > 0 && toks[n-1].typ == "PERLY_SEMICOLON" {
		toks, eof = toks[:n-1], true // toke.c's ; at end of file
	}
	typ := func(i int) string {
		if i < len(toks) {
			return toks[i].typ
		}
		return ""
	}
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		switch {
		case typ(i+1) == "SUBLEXSTART":
			// A quote-like: the head, then its sublexed interior through the matching SUBLEXEND.
			u := unit{head: t, match: kinds(lexer.Quote, lexer.HeredocOpen, lexer.Readline)}
			depth := 0
			for i++; i < len(toks); i++ {
				u.folds = append(u.folds, fold{"quote", toks[i].typ})
				switch toks[i].typ {
				case "SUBLEXSTART":
					depth++
				case "SUBLEXEND":
					depth--
				}
				if depth == 0 {
					break
				}
			}
			us = append(us, u)
		case t.typ == "PERLY_EQUAL_SIGN" && typ(i+1) == "PERLY_SEMICOLON" && typ(i+2) == "FORMRBRACK":
			// A format's picture lines and their arguments, through the PERLY_DOT toke.c makes of the closing line.
			us = append(us, unit{head: t, match: texts("=")})
			u := unit{head: toks[i+2], folds: []fold{{"format", "PERLY_SEMICOLON"}}, match: kinds(lexer.FormatBody)}
			for i += 3; i < len(toks); i++ {
				u.folds = append(u.folds, fold{"format", toks[i].typ})
				if toks[i].typ == "PERLY_DOT" {
					break
				}
			}
			us = append(us, u)
		case (t.typ == "KW_USE_or_NO" || t.typ == "KW_PACKAGE" || t.typ == "KW_CLASS") && typ(i+1) == "BAREWORD":
			// S_force_version's version slot, forced before the module or package name.
			us = append(us, unit{head: t, match: single(t)})
			v := toks[i+1]
			if v.val == "" {
				us[len(us)-1].folds = []fold{{"forced-version", "BAREWORD"}}
				i++
				if m := toks[i+1:]; len(m) > 0 && m[0].typ == "BAREWORD" && m[0].val != "" && !strings.HasPrefix(m[0].val, "PV(") {
					// `use v5.36`: the version in the module slot
					us = append(us, unit{head: m[0], folds: []fold{{"forced-version", ""}}, match: single(m[0])})
					i++
				}
				continue
			}
			if typ(i+2) == "BAREWORD" && strings.HasPrefix(toks[i+2].val, "PV(") {
				us = append(us, unit{head: toks[i+2], match: single(toks[i+2])},
					unit{head: v, folds: []fold{{"forced-version", ""}}, match: single(v)})
				i += 2
			}
		case t.typ == "PERLY_SEMICOLON" && typ(i+1) == "PERLY_BRACE_CLOSE":
			us = append(us, unit{head: toks[i+1], folds: []fold{{"brace-semicolon", "PERLY_SEMICOLON"}}, match: texts("}")})
			i++
		case t.typ == "DOLSHARP" && (typ(i+1) == "PERLY_DOLLAR" || typ(i+1) == "PERLY_BRACE_OPEN"):
			// pvm lexes `$#$x` and `$#{EXPR}` as one Variable; toke.c delivers $# and the expression.
			u := unit{head: t, match: func(p pvmTok) bool { return p.kind == lexer.Variable && strings.HasPrefix(p.text, "$#") }}
			depth := 0
			for i++; i < len(toks); i++ {
				u.folds = append(u.folds, fold{"variable", toks[i].typ})
				switch toks[i].typ {
				case "PERLY_BRACE_OPEN", "HASHBRACK":
					depth++
				case "PERLY_BRACE_CLOSE":
					depth--
				}
				if depth == 0 && toks[i].typ != "PERLY_DOLLAR" {
					break
				}
			}
			us = append(us, u)
		case sigils[t.typ] != "" && i > 0 && toks[i-1].typ == "ARROW" && typ(i+1) == "PERLY_STAR":
			us = append(us, unit{head: t, folds: []fold{{"postfix-deref", "PERLY_STAR"}}, match: texts(sigils[t.typ] + "*")})
			i++
		case sigils[t.typ] != "" && t.typ != "PERLY_STAR" && t.typ != "PERLY_AMPERSAND" &&
			(typ(i+1) == "PRIVATEREF" || typ(i+1) == "BAREWORD" && strings.HasPrefix(toks[i+1].val, "PV(")):
			// toke.c delivers a variable's sigil and its name as two tokens; pvm as one Variable.
			name := varName(toks[i+1])
			us = append(us, unit{head: t, folds: []fold{{"variable", toks[i+1].typ}}, match: func(p pvmTok) bool {
				return p.kind == lexer.Variable && sameName(p.text, name)
			}})
			i++
		default:
			us = append(us, unit{head: t, match: single(t)})
			switch {
			case scopeDecided(t):
				us[len(us)-1].folds = []fold{{"scope-sub", ""}}
			case t.typ == "BAREWORD" && t.val != "" && !strings.HasPrefix(t.val, "PV("):
				us[len(us)-1].folds = []fold{{"constant-sub", ""}}
			}
		}
	}
	if eof && len(us) > 0 {
		u := &us[len(us)-1]
		u.folds = append(u.folds, fold{"eof-semicolon", "PERLY_SEMICOLON"})
	}
	return us
}

// sigils are the sigil tokens and their source text.
var sigils = map[string]string{
	"PERLY_DOLLAR": "$", "PERLY_SNAIL": "@", "PERLY_PERCENT_SIGN": "%",
	"DOLSHARP": "$#", "PERLY_AMPERSAND": "&", "PERLY_STAR": "*",
}

// scopeDecided reports a *SUB or name-less token whose sub toke.c resolved
// by scope: its value is op_rv2cv or op_padany, not the name, so it pairs
// with any word.
func scopeDecided(t tok) bool {
	switch t.typ {
	case "FUNC0SUB", "UNIOPSUB", "LSTOPSUB", "BAREWORD", "PRIVATEREF":
		return t.val == "op_rv2cv" || t.val == "op_padany"
	}
	return false
}

func kinds(ks ...lexer.Kind) func(pvmTok) bool {
	return func(p pvmTok) bool { return slices.Contains(ks, p.kind) }
}

// texts matches a pvm token by its text, a builtin's `CORE::` or
// `CORE::GLOBAL::` qualifier dropped.
func texts(ts ...string) func(pvmTok) bool {
	return func(p pvmTok) bool {
		w := strings.TrimPrefix(strings.TrimPrefix(p.text, "CORE::"), "GLOBAL::")
		return slices.Contains(ts, p.text) || slices.Contains(ts, w)
	}
}

func never(pvmTok) bool { return false }

// single is the match for a token standing alone: its source text where
// the type and value give it, else the pvm kinds it can only be.
func single(t tok) func(pvmTok) bool {
	if s, ok := punct[t.typ]; ok {
		return texts(strings.Split(s, " ")...)
	}
	if ops, ok := opTexts[t.typ]; ok {
		if s, ok := ops[strings.TrimPrefix(t.val, "op_")]; ok {
			return texts(strings.Split(s, " ")...)
		}
		return never
	}
	if w, ok := strings.CutPrefix(t.typ, "KW_"); ok {
		if s, ok := kwTexts[w]; ok {
			return texts(strings.Split(s, " ")...)
		}
		return texts(strings.ToLower(w))
	}
	switch t.typ {
	case "THING", "QWLIST":
		// shortcut: a literal pairs with any pvm literal rather than comparing its decoded value; compare values if a corpus file mispairs one.
		return kinds(lexer.Quote, lexer.Number, lexer.HeredocOpen, lexer.Readline)
	case "PMFUNC":
		return kinds(lexer.Quote)
	case "PROTOTYPE":
		return kinds(lexer.Prototype)
	case "FUNC0OP":
		if t.val == "op_runcv" {
			return texts("__SUB__")
		}
		return texts("__LINE__", "__FILE__", "__PACKAGE__", "__CLASS__")
	case "ATTRLIST":
		if strings.HasPrefix(t.val, "PV(") {
			name, _, _ := strings.Cut(pvText(t.val), "(")
			return texts(strings.Fields(name + " ")[0])
		}
		return kinds(lexer.Word)
	case "PRIVATEREF":
		if scopeDecided(t) {
			return kinds(lexer.Word)
		}
		name := varName(t)
		if strings.HasPrefix(t.val, "&") {
			return texts(name) // a lexical sub's name after sub
		}
		return func(p pvmTok) bool { return p.kind == lexer.Variable && sameName(p.text, name) }
	case "BAREWORD", "METHCALL0", "METHCALL", "LABEL":
		switch {
		case scopeDecided(t):
			return kinds(lexer.Word)
		case strings.HasPrefix(t.val, "PV("):
			pv := pvText(t.val)
			return texts(pv, strings.TrimPrefix(pv, "CORE::GLOBAL::"), pv+"::")
		case t.val != "":
			// a version (NV, a v-string's PVMG, a version object) or a constant sub's inlined value
			return kinds(lexer.Number, lexer.Word, lexer.Quote)
		}
		return never
	case "FUNC0SUB", "UNIOPSUB", "LSTOPSUB":
		return kinds(lexer.Word)
	case "LSTOP", "FUNC", "UNIOP", "FUNC1", "FUNC0", "BLKLSTOP", "LOOPEX":
		if t.val == "" {
			return kinds(lexer.Word)
		}
		w := strings.TrimPrefix(t.val, "op_")
		if a, ok := opWords[w]; ok {
			return texts(strings.Split(a, " ")...)
		}
		return texts(w)
	}
	return never
}

// punct is the source text of each token that carries no op, space-separated
// where it has several.
var punct = map[string]string{
	"PERLY_SEMICOLON": ";", "PERLY_BRACE_OPEN": "{", "PERLY_BRACE_CLOSE": "}",
	"PERLY_BRACKET_OPEN": "[", "PERLY_BRACKET_CLOSE": "]", "PERLY_PAREN_OPEN": "(",
	"PERLY_PAREN_CLOSE": ")", "PERLY_COMMA": ", =>", "PERLY_DOT": ".", "PERLY_EQUAL_SIGN": "=",
	"PERLY_MINUS": "-", "PERLY_PLUS": "+", "PERLY_SLASH": "/", "PERLY_STAR": "*",
	"PERLY_AMPERSAND": "&", "PERLY_COLON": ":", "PERLY_QUESTION_MARK": "?",
	"PERLY_EXCLAMATION_MARK": "!", "PERLY_TILDE": "~", "PERLY_SNAIL": "@",
	"PERLY_PERCENT_SIGN": "%", "PERLY_DOLLAR": "$", "DOLSHARP": "$#",
	"HASHBRACK": "{", "ARROW": "->", "ANDAND": "&&", "OROR": "||", "DORDOR": "//",
	"OROP": "or xor", "ANDOP": "and", "NOTOP": "not", "DOTDOT": ".. ...", "REFGEN": "\\",
	"POSTINC": "++", "PREINC": "++", "POSTDEC": "--", "PREDEC": "--", "YADAYADA": "...",
	"COLONATTR": ":", "POWOP": "**", "POSTJOIN": "@*",
}

// opTexts is the source text of each operator token by its op.
var opTexts = map[string]map[string]string{
	"ASSIGNOP": {
		"null": "=", "sassign": "=", "add": "+=", "subtract": "-=", "multiply": "*=", "divide": "/=",
		"modulo": "%=", "pow": "**=", "concat": ".=", "repeat": "x=", "left_shift": "<<=",
		"right_shift": ">>=", "bit_and": "&=", "bit_or": "|=", "bit_xor": "^=", "nbit_and": "&=",
		"nbit_or": "|=", "nbit_xor": "^=", "sbit_and": "&.=", "sbit_or": "|.=", "sbit_xor": "^.=",
		"andassign": "&&=", "orassign": "||=", "dorassign": "//=", "xor": "^^=",
	},
	"ADDOP":    {"add": "+", "subtract": "-", "concat": "."},
	"MULOP":    {"multiply": "*", "divide": "/", "modulo": "%", "repeat": "x"},
	"CHEQOP":   {"eq": "==", "ne": "!=", "seq": "eq", "sne": "ne"},
	"NCEQOP":   {"ncmp": "<=>", "scmp": "cmp", "smartmatch": "~~"},
	"CHRELOP":  {"lt": "<", "gt": ">", "le": "<=", "ge": ">=", "slt": "lt", "sgt": "gt", "sle": "le", "sge": "ge"},
	"NCRELOP":  {"isa": "isa"},
	"MATCHOP":  {"match": "=~", "not": "!~"},
	"SHIFTOP":  {"left_shift": "<<", "right_shift": ">>"},
	"POWOP":    {"pow": "**"},
	"BITANDOP": {"bit_and": "&", "nbit_and": "&", "sbit_and": "&."},
	"BITOROP":  {"bit_or": "|", "bit_xor": "^", "nbit_or": "|", "nbit_xor": "^", "sbit_or": "|.", "sbit_xor": "^."},
}

// kwTexts is the source text of the KW_ tokens not named by their lowercased suffix.
var kwTexts = map[string]string{
	"MY": "my our state", "USE_or_NO": "use no", "FOR": "for foreach",
	"SUB_named": "sub", "SUB_named_sig": "sub", "SUB_anon": "sub", "SUB_anon_sig": "sub",
	"METHOD_named": "method", "METHOD_anon": "method",
}

// opWords is the source word of each keyword op not named by the op itself.
var opWords = map[string]string{
	"grepstart": "grep", "mapstart": "map", "prtf": "printf", "entereval": "eval",
	"entertry": "eval", "open_dir": "opendir", "pipe_op": "pipe", "ghbyname": "gethostbyname",
	"gsockopt": "getsockopt", "ssockopt": "setsockopt", "sockpair": "socketpair",
	"enterwrite": "write", "ftis": "-e", "ftfile": "-f", "ftdir": "-d", "ftsize": "-s",
	"fteexec": "-x", "fttty": "-t", "ftzero": "-z", "ftsock": "-S", "ftrread": "-r",
	"ftrwrite": "-w", "ftrexec": "-x", "fteread": "-r", "ftewrite": "-w", "ftlink": "-l",
	"ftpipe": "-p", "ftblk": "-b", "ftchr": "-c", "ftmtime": "-M", "ftatime": "-A", "ftctime": "-C",
	"ftsuid": "-u", "ftsgid": "-g", "ftsvtx": "-k", "fttext": "-T", "ftbinary": "-B", "fteowned": "-O",
	"ftrowned": "-o", "lstat": "lstat", "readline": "readline", "backtick": "readpipe qx",
	"undef": "undef", "not": "not", "lc": "lc",
}

// varName is a variable token's name: a PRIVATEREF's without its sigil, a
// BAREWORD's PV.
func varName(t tok) string {
	if t.typ == "PRIVATEREF" {
		return strings.TrimLeft(t.val, "$@%&*")
	}
	return pvText(t.val)
}

// sameName reports whether pvm Variable text names name: its sigil and any
// braces dropped, `^X` as the control character toke.c names, and a name
// toke.c qualified by its package (`our`, `$::x`) compared by its last part.
func sameName(text, name string) bool {
	if strings.HasPrefix(text, "$#") && len(text) > 2 {
		text = text[2:]
	} else if text != "" {
		text = text[1:]
	}
	if strings.HasPrefix(text, "{") && strings.HasSuffix(text, "}") {
		text = strings.TrimSpace(text[1 : len(text)-1])
	}
	if len(text) >= 2 && text[0] == '^' {
		text = string(text[1]-'@') + text[2:]
	}
	text = strings.TrimPrefix(strings.TrimPrefix(text, "main"), "::")
	if text == name {
		return true
	}
	return strings.HasSuffix(name, "::"+text) || strings.TrimPrefix(name, "main::") == text
}

// pvText decodes the string inside SvPEEK's `PV("...")`: C escapes, and
// octal for the rest.
func pvText(v string) string {
	_, s, ok := strings.Cut(v, `("`)
	if !ok {
		return ""
	}
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '"' {
			break // the closing quote; a `[UTF8 "..."]` after it repeats the value
		}
		if s[i] != '\\' || i+1 == len(s) {
			b.WriteByte(s[i])
			continue
		}
		i++
		switch c := s[i]; {
		case c >= '0' && c <= '7':
			j := i
			for j < len(s) && j < i+3 && s[j] >= '0' && s[j] <= '7' {
				j++
			}
			n, _ := strconv.ParseUint(s[i:j], 8, 8)
			b.WriteByte(byte(n))
			i = j - 1
		case c == 'n':
			b.WriteByte('\n')
		case c == 't':
			b.WriteByte('\t')
		case c == 'r':
			b.WriteByte('\r')
		case c == 'f':
			b.WriteByte('\f')
		case c == 'e':
			b.WriteByte(27)
		case c == 'a':
			b.WriteByte(7)
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}
