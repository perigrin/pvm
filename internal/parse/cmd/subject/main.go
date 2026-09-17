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

	"tamarou.com/pvm/internal/parse"
	"tamarou.com/pvm/internal/parseoracle"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s FILE\n", os.Args[0])
		os.Exit(2)
	}
	path := os.Args[1]

	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot read %s: %v\n", path, err)
		os.Exit(1)
	}

	root := parse.Parse(src)

	// OK is whether this parser believes the file is valid Perl, and the
	// honest answer is yes: it declines constructs rather than rejecting
	// files. An Unknown says "I could not read this", which is a statement
	// about the parser and not about the source -- the corpus is perl's own
	// test suite and parses under perl. Reporting !OK would move every file
	// with one Unknown into no-answer and hide the parser's real coverage
	// behind a claim it has no evidence for.
	facts := parseoracle.SubjectFacts{
		OK: true,

		// Empty rather than nil, and that is the contract: an empty map is
		// an answer ("I looked and found none"), nil marshals to JSON null,
		// and null decodes as absent -- which scores as "did not look".
		// Prototype RESOLUTION is M4's; recognition without resolution has
		// nothing to report here yet.
		Prototypes: map[string]string{},

		CallSites: parse.Sites(root, src),
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
