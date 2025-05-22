package tree_sitter_typed_perl_test

import (
	"testing"

	tree_sitter "github.com/smacker/go-tree-sitter"
)

func TestCanLoadGrammar(t *testing.T) {
	language := tree_sitter.NewLanguage(tree_sitter_perl.Language())
	if language == nil {
		t.Errorf("Error loading Perl grammar")
	}
}
