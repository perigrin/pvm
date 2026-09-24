// ABOUTME: One reader for the whole corpus, so sixteen globs stop each knowing where cases live.
// ABOUTME: A case's key is `<topic>.md/<case title>` -- what a reader needs to find it now that a file holds several.
package conformance

import (
	"fmt"
	"os"
	"path/filepath"
)

// CorpusCase is one case with enough context to report on it.
type CorpusCase struct {
	// Key is `<topic>.md/<case title>`, which is how a failure names a
	// case a reader then has to open. It replaces the file name that
	// used to serve as both identity and address.
	Key string

	// Topic is the file it lives in, Tier the directory that file
	// declares.
	Topic string
	Tier  string

	*File
}

// AllCases reads every case in the corpus.
//
// SIXTEEN GLOBS BECAME ONE READER. Each of those globs spelled
// `filepath.Join(corpusDir, tier, "*.t")` for itself, so each knew that
// a case was a file and that a tier was a directory of them. Both facts
// changed with the topic format, and a corpus-wide check should not have
// to know either.
//
// Fails rather than returning empty. A check that runs over no cases
// passes vacuously, which this package has been bitten by before.
func AllCases(corpus string) ([]CorpusCase, error) {
	paths, err := filepath.Glob(filepath.Join(corpus, "mdtest", "*.md"))
	if err != nil {
		return nil, fmt.Errorf("globbing topics: %w", err)
	}

	var out []CorpusCase
	for _, path := range paths {
		base := filepath.Base(path)
		if base == "FORMAT.md" {
			continue
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", base, err)
		}
		tier, err := topicTier(string(raw))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", base, err)
		}
		cases, err := ParseTopic(string(raw))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", base, err)
		}
		for _, c := range cases {
			out = append(out, CorpusCase{
				Key:   base + "/" + c.Title,
				Topic: base,
				Tier:  tier,
				File:  c.asFile(),
			})
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("the corpus holds no cases")
	}
	return out, nil
}
