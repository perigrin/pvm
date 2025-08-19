// ABOUTME: Migration layer providing backward compatibility during tree-sitter transition
// ABOUTME: Enables gradual migration from traditional AST to tree-sitter shim architecture

package migration

import (
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/parser"
)

// MigrationMode controls how the migration layer behaves
type MigrationMode int

const (
	// ModePreferTraditional uses traditional AST when possible, tree-sitter as fallback
	ModePreferTraditional MigrationMode = iota

	// ModePreferTreeSitter uses tree-sitter when possible, traditional AST as fallback
	ModePreferTreeSitter

	// ModeTreeSitterOnly uses only tree-sitter, errors if not available
	ModeTreeSitterOnly

	// ModeTraditionalOnly uses only traditional AST, errors if not available
	ModeTraditionalOnly
)

// MigrationConfig controls the migration layer behavior
type MigrationConfig struct {
	// Mode determines the parsing strategy
	Mode MigrationMode

	// AllowFallback enables fallback to alternative parser when primary fails
	AllowFallback bool

	// PreferShimForTypes lists node types that should prefer tree-sitter shim
	PreferShimForTypes []string

	// LogMigrationChoices enables logging of parser selection decisions
	LogMigrationChoices bool
}

// DefaultMigrationConfig returns a sensible default configuration
func DefaultMigrationConfig() *MigrationConfig {
	return &MigrationConfig{
		Mode:          ModePreferTreeSitter,
		AllowFallback: true,
		PreferShimForTypes: []string{
			"variable_declaration",
			"type_annotation",
			"parameterized_type",
			"union_type",
			"intersection_type",
		},
		LogMigrationChoices: false,
	}
}

// MigrationParser wraps both traditional and tree-sitter parsers for compatibility
type MigrationParser struct {
	traditionalParser parser.Parser
	shimParser        parser.ShimParser
	config            *MigrationConfig
}

// NewMigrationParser creates a new migration parser with both backends
func NewMigrationParser(config *MigrationConfig) (*MigrationParser, error) {
	if config == nil {
		config = DefaultMigrationConfig()
	}

	var traditionalParser parser.Parser
	var shimParser parser.ShimParser
	var err error

	// Initialize traditional parser if needed
	if config.Mode == ModePreferTraditional || config.Mode == ModeTraditionalOnly || config.AllowFallback {
		traditionalParser, err = parser.NewParser()
		if err != nil && config.Mode == ModeTraditionalOnly {
			return nil, fmt.Errorf("failed to create traditional parser: %w", err)
		}
	}

	// Initialize tree-sitter shim parser if needed
	if config.Mode == ModePreferTreeSitter || config.Mode == ModeTreeSitterOnly || config.AllowFallback {
		shimParser, err = parser.NewShimParser()
		if err != nil && config.Mode == ModeTreeSitterOnly {
			return nil, fmt.Errorf("failed to create tree-sitter shim parser: %w", err)
		}
	}

	return &MigrationParser{
		traditionalParser: traditionalParser,
		shimParser:        shimParser,
		config:            config,
	}, nil
}

// ParseFile parses a file using the configured migration strategy
func (mp *MigrationParser) ParseFile(path string) (interface{}, error) {
	strategy := mp.selectParsingStrategy(path, "")

	switch strategy {
	case "tree-sitter":
		if mp.shimParser == nil {
			return nil, fmt.Errorf("tree-sitter parser not available")
		}
		return mp.shimParser.ParseFileShim(path)

	case "traditional":
		if mp.traditionalParser == nil {
			return nil, fmt.Errorf("traditional parser not available")
		}
		return mp.traditionalParser.ParseFile(path)

	default:
		return nil, fmt.Errorf("no suitable parsing strategy available")
	}
}

// ParseString parses source code using the configured migration strategy
func (mp *MigrationParser) ParseString(content string) (interface{}, error) {
	strategy := mp.selectParsingStrategy("", content)

	switch strategy {
	case "tree-sitter":
		if mp.shimParser == nil {
			return nil, fmt.Errorf("tree-sitter parser not available")
		}
		return mp.shimParser.ParseStringShim(content)

	case "traditional":
		if mp.traditionalParser == nil {
			return nil, fmt.Errorf("traditional parser not available")
		}
		return mp.traditionalParser.ParseString(content)

	default:
		return nil, fmt.Errorf("no suitable parsing strategy available")
	}
}

// selectParsingStrategy determines which parser to use based on configuration and content
func (mp *MigrationParser) selectParsingStrategy(path string, content string) string {
	// Check for type-specific preferences
	if mp.hasPreferredTypes(content) {
		if mp.shimParser != nil {
			mp.logChoice("tree-sitter", "content contains preferred types")
			return "tree-sitter"
		}
	}

	// Apply mode-based selection
	switch mp.config.Mode {
	case ModeTreeSitterOnly:
		mp.logChoice("tree-sitter", "tree-sitter only mode")
		return "tree-sitter"

	case ModeTraditionalOnly:
		mp.logChoice("traditional", "traditional only mode")
		return "traditional"

	case ModePreferTreeSitter:
		if mp.shimParser != nil {
			mp.logChoice("tree-sitter", "prefer tree-sitter mode")
			return "tree-sitter"
		}
		if mp.config.AllowFallback && mp.traditionalParser != nil {
			mp.logChoice("traditional", "tree-sitter fallback to traditional")
			return "traditional"
		}

	case ModePreferTraditional:
		if mp.traditionalParser != nil {
			mp.logChoice("traditional", "prefer traditional mode")
			return "traditional"
		}
		if mp.config.AllowFallback && mp.shimParser != nil {
			mp.logChoice("tree-sitter", "traditional fallback to tree-sitter")
			return "tree-sitter"
		}
	}

	mp.logChoice("none", "no suitable parser available")
	return "none"
}

// hasPreferredTypes checks if content contains types that prefer tree-sitter
func (mp *MigrationParser) hasPreferredTypes(content string) bool {
	for _, nodeType := range mp.config.PreferShimForTypes {
		// Simple heuristic: check for type annotation patterns
		switch nodeType {
		case "variable_declaration":
			if containsTypedVariables(content) {
				return true
			}
		case "type_annotation", "parameterized_type":
			if containsTypeAnnotations(content) {
				return true
			}
		case "union_type":
			if containsUnionTypes(content) {
				return true
			}
		}
	}
	return false
}

// logChoice logs parser selection decisions if enabled
func (mp *MigrationParser) logChoice(choice, reason string) {
	if mp.config.LogMigrationChoices {
		fmt.Printf("[Migration] Selected %s parser: %s\n", choice, reason)
	}
}

// Utility functions for content analysis

// containsTypedVariables checks for typed variable declarations
func containsTypedVariables(content string) bool {
	// Look for patterns like "my Int $var" or "my ArrayRef[Str] $items"
	return containsPattern(content, []string{
		"my Int ",
		"my Str ",
		"my ArrayRef",
		"my HashRef",
		"my CodeRef",
		"my Object ",
	})
}

// containsTypeAnnotations checks for explicit type annotations
func containsTypeAnnotations(content string) bool {
	// Look for parameterized types and complex type expressions
	return containsPattern(content, []string{
		"[Int]",
		"[Str]",
		"[Any]",
		"ArrayRef[",
		"HashRef[",
		"Maybe[",
	})
}

// containsUnionTypes checks for union type syntax
func containsUnionTypes(content string) bool {
	// Look for union type patterns
	return containsPattern(content, []string{
		"Int|Str",
		"Str|Int",
		"|Undef",
		"Maybe[",
	})
}

// containsPattern checks if content contains any of the given patterns
func containsPattern(content string, patterns []string) bool {
	for _, pattern := range patterns {
		if len(content) >= len(pattern) {
			for i := 0; i <= len(content)-len(pattern); i++ {
				if content[i:i+len(pattern)] == pattern {
					return true
				}
			}
		}
	}
	return false
}

// Conversion utilities for migrating between AST types

// ASTConverter provides utilities for converting between traditional and tree-sitter ASTs
type ASTConverter struct {
	shimParser        parser.ShimParser
	traditionalParser parser.Parser
}

// NewASTConverter creates a new AST converter
func NewASTConverter() (*ASTConverter, error) {
	shimParser, err := parser.NewShimParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create shim parser: %w", err)
	}

	traditionalParser, err := parser.NewParser()
	if err != nil {
		return nil, fmt.Errorf("failed to create traditional parser: %w", err)
	}

	return &ASTConverter{
		shimParser:        shimParser,
		traditionalParser: traditionalParser,
	}, nil
}

// ToTreeSitterAST converts a traditional AST to TreeSitterAST
func (c *ASTConverter) ToTreeSitterAST(traditional *ast.AST) (*ast.TreeSitterAST, error) {
	if traditional == nil {
		return nil, fmt.Errorf("traditional AST cannot be nil")
	}

	// Get the source content from traditional AST
	var content string
	if traditional.Source != "" {
		content = traditional.Source
	} else if traditional.Root != nil {
		content = traditional.Root.Text()
	} else {
		return nil, fmt.Errorf("cannot extract source content from traditional AST")
	}

	// Parse with tree-sitter shim
	return c.shimParser.ParseStringShim(content)
}

// ToTraditionalAST converts a TreeSitterAST to traditional AST
func (c *ASTConverter) ToTraditionalAST(treeSitter *ast.TreeSitterAST) (*ast.AST, error) {
	if treeSitter == nil {
		return nil, fmt.Errorf("tree-sitter AST cannot be nil")
	}

	// Parse with traditional parser
	return c.traditionalParser.ParseString(treeSitter.Source)
}

// ExtractTypeAnnotations extracts type annotations from either AST type
func (c *ASTConverter) ExtractTypeAnnotations(astInput interface{}) ([]*ast.TypeAnnotation, error) {
	switch a := astInput.(type) {
	case *ast.TreeSitterAST:
		return a.TypeAnnotations, nil

	case *ast.AST:
		return a.TypeAnnotations, nil

	default:
		return nil, fmt.Errorf("unsupported AST type: %T", astInput)
	}
}

// ConvertPosition converts between position types
func ConvertPosition(pos interface{}) ast.Position {
	switch p := pos.(type) {
	case ast.Position:
		return p

	case sitter.Point:
		return ast.ConvertTreeSitterPosition(p, 0)

	default:
		return ast.Position{}
	}
}

// Migration helpers for specific components

// MigrateVarDecl converts a traditional VarDecl to tree-sitter backed VarDeclNode
// This is a simplified implementation for demonstration purposes
func MigrateVarDecl(traditional *ast.VarDecl, treeSitterAST *ast.TreeSitterAST) (*ast.VarDeclNode, error) {
	if traditional == nil || treeSitterAST == nil {
		return nil, fmt.Errorf("both traditional VarDecl and TreeSitterAST are required")
	}

	// For now, just return the first VarDeclNode found in the tree-sitter AST
	// In a real implementation, this would match based on position and content
	var result *ast.VarDeclNode
	if treeSitterAST.Root != nil {
		treeSitterAST.Root.WalkNodes(func(node *ast.TreeSitterNode) bool {
			if node.Type() == "variable_declaration" {
				result = node.AsVarDecl()
				return false // Stop at first match for simplicity
			}
			return true
		})
	}

	if result == nil {
		return nil, fmt.Errorf("could not find variable declaration in tree-sitter AST")
	}

	return result, nil
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// MigrationError represents errors that occur during migration
type MigrationError struct {
	Operation string
	Cause     error
	Context   string
}

// Error implements the error interface
func (e *MigrationError) Error() string {
	return fmt.Sprintf("migration error in %s: %s (context: %s)", e.Operation, e.Cause, e.Context)
}

// NewMigrationError creates a new migration error
func NewMigrationError(operation string, cause error, context string) *MigrationError {
	return &MigrationError{
		Operation: operation,
		Cause:     cause,
		Context:   context,
	}
}
