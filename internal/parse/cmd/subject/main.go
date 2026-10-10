// ABOUTME: The hand-written parser as a parseoracle subject: reads one file, prints SubjectFacts as JSON.
// ABOUTME: A subprocess rather than a Go call, because the harness measures implementations that are not Go.

// Usage: go run ./internal/parse/cmd/subject FILE
//
// Prints one JSON object on stdout, the SubjectFacts the harness decodes.
// Diagnostics go to stderr, which the harness quotes when a subject fails:
// "exit status 2" says nothing, "cannot read op/sub.t" says which directory
// it ran from.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s FILE\n", os.Args[0])
		os.Exit(2)
	}
	path := os.Args[1]

	// Resolved the way perl's own test suite resolves, since that is what the
	// oracle runs the subject against: from t/, where `require './test.pl'`
	// names t/test.pl, `./t/test.pl` names the same file from the tree root,
	// and test.pl's set_up_inc('../lib') puts lib/ in @INC. Without these the
	// subject would measure a parse that resolves nothing, which is not the
	// parser the M1 ratchets gate.
	root, err := parse.ParseFileFrom(path, ".", "..", filepath.Join("..", "lib"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read %s: %v\n", path, err)
		os.Exit(1)
	}
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read %s: %v\n", path, err)
		os.Exit(1)
	}

	// OK is whether this parser believes the file is valid Perl, and the
	// honest answer is yes: it declines constructs rather than rejecting
	// files. An Unknown says "I could not read this", which is a statement
	// about the parser and not about the source -- the corpus is perl's own
	// test suite and parses under perl. Reporting !OK would move every file
	// with one Unknown into no-answer and hide the parser's real coverage
	// behind a claim it has no evidence for.
	facts := parseoracle.SubjectFacts{
		OK: true,

		// Never nil, and that is the contract: an empty map is an answer
		// ("I looked and found none"), nil marshals to JSON null, and null
		// decodes as absent -- which scores as "did not look".
		Prototypes: prototypes(root),

		CallSites: perlNumbered(parse.Sites(root, src), src, path),
	}

	out, err := json.Marshal(facts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "encoding facts for %s: %v\n", path, err)
		os.Exit(1)
	}
	if _, err := os.Stdout.Write(append(out, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "writing facts for %s: %v\n", path, err)
		os.Exit(1)
	}
}

// perlNumbered renumbers sites to the lines perl records, which a `#line`
// directive moves away from the physical ones, and drops the sites a
// directive moved into another file: the oracle reports only this file's.
func perlNumbered(sites []parseoracle.SubjectCallSite, src []byte, path string) []parseoracle.SubjectCallSite {
	line, own := parse.PerlLines(src, path)
	kept := []parseoracle.SubjectCallSite{}
	for _, site := range sites {
		end := max(site.EndLine, site.Line)
		if !own[site.Line] || !own[end] {
			continue
		}
		site.Line, site.EndLine = line[site.Line], line[end]
		kept = append(kept, site)
	}
	return kept
}

// prototypes is every prototype the tree carries: the subs this file
// declares and the ones its imports bring in, each read from a declaration
// the parse actually saw. A name whose prototype the parse could not see --
// an XS sub, whose body is in a .so -- is left out rather than guessed.
//
// Keyed as perl names them, `name` in main and `Pkg::name` elsewhere, and
// spelled as prototype() returns them, without the parens.
func prototypes(root *parse.Node) map[string]string {
	protos := map[string]string{}
	for name, imp := range parse.Imports(root) {
		if imp.PrototypeKnown && imp.Prototype != "" {
			protos[name] = strings.TrimSuffix(strings.TrimPrefix(imp.Prototype, "("), ")")
		}
	}
	return protos
}
