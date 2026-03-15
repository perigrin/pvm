# PVM Pure-Go Rewrite Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Rewrite PVM on a clean `pure-go` branch, copying forward proven core code and rebuilding PSC with gotreesitter (pure Go, no CGO).

**Architecture:** Copy-forward rewrite. Start with an empty Go module. Copy packages from the existing `pu` branch one tier at a time, verifying tests at each step. Build new parser (gotreesitter wrapper) and PSC (parse, LSP, analyze) from scratch. The existing codebase lives on `pu` and is the source of truth for copy-forward operations.

**Tech Stack:** Go 1.24+, gotreesitter (pure-Go tree-sitter), cobra (CLI), charmbracelet (UI), viper (config)

**Source branch:** `pu` (the current default branch, source for all copy-forward)
**Target branch:** `pure-go` (orphan branch, the fresh start)

---

## Chunk 1: Bootstrap and Foundation Packages

### Task 1: Create orphan branch and initialize Go module

**Files:**
- Create: `go.mod`
- Create: `go.sum`
- Create: `Makefile`
- Create: `CLAUDE.md`
- Create: `.gitignore`
- Create: `.pre-commit-config.yaml`

- [ ] **Step 1: Create the orphan branch**

```bash
cd /home/perigrin/dev/pvm
git checkout --orphan pure-go
git rm -rf .
git clean -fd
```

- [ ] **Step 2: Initialize the Go module**

```bash
go mod init tamarou.com/pvm
```

Then edit `go.mod` to set Go version:

```go
module tamarou.com/pvm

go 1.24.3
```

- [ ] **Step 3: Copy and adapt .gitignore from pu**

```bash
git show pu:.gitignore > .gitignore
```

Remove any tree-sitter-typed-perl entries. Keep the rest.

- [ ] **Step 4: Copy .pre-commit-config.yaml from pu**

```bash
git show pu:.pre-commit-config.yaml > .pre-commit-config.yaml
```

- [ ] **Step 5: Create minimal Makefile**

Create a `Makefile` that starts simple and will grow as we add packages:

```makefile
.PHONY: all build test clean

all: build

build:
	go build ./...

test:
	go test ./... -count=1

clean:
	go clean ./...
```

- [ ] **Step 6: Create CLAUDE.md**

Copy from pu and strip all tree-sitter, typed-perl, and type-system references:

```bash
git show pu:CLAUDE.md > CLAUDE.md
```

Edit to remove:
- Tree-sitter build commands and sections
- Type annotation references
- Typed-perl grammar references
- PSC type-checking references
- References to internal/typechecker, internal/binder, internal/compiler, internal/ast
- CGO dependency management section
- Tree-sitter integration principle section

Keep:
- Build/test commands (simplify to `make`, `make test`)
- Code style guidelines
- Test data format preference
- Repository configuration protection
- Pre-commit hook compliance
- PVM project patterns (adapt)
- Git workflow standards

- [ ] **Step 7: Verify clean state**

```bash
go build ./...
```

Expected: nothing to build yet, no errors.

- [ ] **Step 8: Commit**

```bash
git add .gitignore .pre-commit-config.yaml go.mod Makefile CLAUDE.md
git commit -m "feat: bootstrap pure-go branch with Go module and build infrastructure"
```

---

### Task 2: Copy zero-dependency foundation packages

These packages have no internal dependencies. Copy them directly from `pu`.

**Packages:**
- `internal/version/` -- semver parsing, version checking
- `internal/memory/` -- object pooling, lazy loading
- `internal/log/` -- logging framework
- `internal/xdg/` -- XDG directory standard
- `internal/platform/` -- platform detection

- [ ] **Step 1: Copy packages from pu branch**

For each package, use git to extract files:

```bash
mkdir -p internal/version internal/memory internal/log internal/xdg internal/platform

git show pu:internal/version/ | # use git checkout approach below
```

More reliable approach -- checkout files from pu:

```bash
git checkout pu -- internal/version/
git checkout pu -- internal/memory/
git checkout pu -- internal/log/
git checkout pu -- internal/xdg/
git checkout pu -- internal/platform/
```

- [ ] **Step 2: Run go mod tidy**

```bash
go mod tidy
```

- [ ] **Step 3: Verify compilation**

```bash
go build ./internal/version/... ./internal/memory/... ./internal/log/... ./internal/xdg/... ./internal/platform/...
```

Expected: builds cleanly.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/version/... ./internal/memory/... ./internal/log/... ./internal/xdg/... ./internal/platform/... -count=1
```

Expected: all tests pass. If any fail due to missing dependencies on other internal packages, identify what's missing and either copy the dependency or stub it.

- [ ] **Step 5: Commit**

```bash
git add internal/version/ internal/memory/ internal/log/ internal/xdg/ internal/platform/ go.mod go.sum
git commit -m "feat: copy foundation packages (version, memory, log, xdg, platform)"
```

---

### Task 3: Copy errors package

The errors package imports `internal/log` (already copied) and `internal/ast` for a Position type. We need to resolve the ast dependency.

**Files:**
- Copy: `internal/errors/` from pu
- Modify: `internal/errors/type_parse_error.go` -- remove or adapt ast import

- [ ] **Step 1: Copy errors package**

```bash
git checkout pu -- internal/errors/
```

- [ ] **Step 2: Check for ast dependency**

```bash
grep -r "internal/ast" internal/errors/
```

If `type_parse_error.go` imports `internal/ast` just for `Position`, define a local Position type in the errors package instead:

```go
// Position in source code, used for error reporting.
type Position struct {
    Line   int
    Column int
    File   string
}
```

Remove the `internal/ast` import.

- [ ] **Step 3: Remove any PSC-specific error types if present**

Check for `psc_generated.go` -- if it references type-checking error codes, keep the file but remove type-checking-specific entries.

- [ ] **Step 4: Run go mod tidy and verify**

```bash
go mod tidy
go build ./internal/errors/...
go test ./internal/errors/... -count=1
```

Expected: builds and tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/errors/ go.mod go.sum
git commit -m "feat: copy errors package with ast dependency resolved"
```

---

### Task 4: Copy project and config packages

Config depends on: errors, project, xdg, version. Project depends on: xdg. Copy project first, then config.

**Packages:**
- `internal/project/` -- project configuration
- `internal/config/` -- XDG paths, pvm.toml, defaults

- [ ] **Step 1: Copy project package**

```bash
git checkout pu -- internal/project/
```

- [ ] **Step 2: Verify project compiles**

```bash
go mod tidy
go build ./internal/project/...
```

- [ ] **Step 3: Copy config package**

```bash
git checkout pu -- internal/config/
```

- [ ] **Step 4: Check config for type-system references**

```bash
grep -r "typechecker\|type_definition\|psc\." internal/config/
```

If the config has `[psc]` section handling with type-specific fields, simplify it but keep the `[psc]` section structure for future use.

- [ ] **Step 5: Run go mod tidy and verify**

```bash
go mod tidy
go build ./internal/config/... ./internal/project/...
go test ./internal/config/... ./internal/project/... -count=1
```

Expected: builds and tests pass.

- [ ] **Step 6: Commit**

```bash
git add internal/project/ internal/config/ go.mod go.sum
git commit -m "feat: copy project and config packages"
```

---

### Task 5: Copy remaining utility packages

These are small supporting packages used throughout the codebase.

**Packages:**
- `internal/download/` -- binary downloading
- `internal/archive/` -- archive extraction
- `internal/backup/` -- backup management
- `internal/build/` -- build system utilities
- `internal/cpan/` -- CPAN integration
- `internal/modules/` -- module information
- `internal/current/` -- current version tracking
- `internal/shell/` -- shell integration
- `internal/fortune/` -- fortune messages
- `internal/diskspace/` -- disk space utilities
- `internal/data/` -- data utilities
- `internal/testing/` -- test utilities

- [ ] **Step 1: Copy all utility packages**

```bash
for pkg in download archive backup build cpan modules current shell fortune diskspace data testing; do
    git checkout pu -- internal/$pkg/
done
```

- [ ] **Step 2: Run go mod tidy**

```bash
go mod tidy
```

- [ ] **Step 3: Verify compilation**

```bash
go build ./internal/...
```

Fix any import issues. These packages may have cross-dependencies that need other packages already copied. If a package depends on something not yet copied, check if it's a clean package we can copy now or if it needs to wait.

- [ ] **Step 4: Run all tests so far**

```bash
go test ./internal/... -count=1
```

Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/download/ internal/archive/ internal/backup/ internal/build/ internal/cpan/ internal/modules/ internal/current/ internal/shell/ internal/fortune/ internal/diskspace/ internal/data/ internal/testing/ go.mod go.sum
git commit -m "feat: copy utility packages (download, archive, cpan, shell, etc.)"
```

---

## Chunk 2: Perl Toolchain and UI

### Task 6: Copy UI framework

The UI framework (fang-based) is used by CLI and many commands.

**Packages:**
- `internal/ui/` -- terminal UI components (if exists separately)
- `internal/cli/` -- CLI framework with ui/, docs/, progress/ subpackages

- [ ] **Step 1: Copy CLI package with subpackages**

```bash
git checkout pu -- internal/cli/
```

- [ ] **Step 2: Check for typed-perl references**

```bash
grep -r "typechecker\|binder\|compiler\|typed.perl\|type_check\|psc\." internal/cli/
```

Remove any references found. The CLI package should be clean based on research, but verify.

- [ ] **Step 3: Verify compilation**

```bash
go mod tidy
go build ./internal/cli/...
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/cli/... -count=1
```

Expected: all tests pass.

- [ ] **Step 5: Commit**

```bash
git add internal/cli/ go.mod go.sum
git commit -m "feat: copy CLI framework with UI components"
```

---

### Task 7: Copy updater and Perl toolchain

**Packages:**
- `internal/updater/` -- self-update system
- `internal/perl/` -- Perl build, detection, patchperl, shims

- [ ] **Step 1: Copy updater**

```bash
git checkout pu -- internal/updater/
```

- [ ] **Step 2: Verify updater compiles and tests pass**

```bash
go mod tidy
go build ./internal/updater/...
go test ./internal/updater/... -count=1
```

- [ ] **Step 3: Copy perl package**

```bash
git checkout pu -- internal/perl/
```

- [ ] **Step 4: Verify perl compiles and tests pass**

```bash
go mod tidy
go build ./internal/perl/...
go test ./internal/perl/... -count=1
```

The perl package has many files (58). If tests fail, categorize failures: missing dependency vs actual bug. Copy missing dependencies if they're clean packages.

- [ ] **Step 5: Commit**

```bash
git add internal/updater/ internal/perl/ go.mod go.sum
git commit -m "feat: copy updater and perl toolchain packages"
```

---

### Task 8: Copy tool management and templates

**Packages:**
- `internal/tool/` -- tool management (with install/ and shim/ subpackages)
- `internal/templates/` -- embedded templates
- `internal/compat/` -- compatibility wrappers (plenv, perlbrew, cpanm, carton)

- [ ] **Step 1: Copy packages**

```bash
git checkout pu -- internal/tool/
git checkout pu -- internal/templates/
git checkout pu -- internal/compat/
```

- [ ] **Step 2: Verify compilation and tests**

```bash
go mod tidy
go build ./internal/tool/... ./internal/templates/... ./internal/compat/...
go test ./internal/tool/... ./internal/templates/... ./internal/compat/... -count=1
```

- [ ] **Step 3: Commit**

```bash
git add internal/tool/ internal/templates/ internal/compat/ go.mod go.sum
git commit -m "feat: copy tool management, templates, and compatibility wrappers"
```

---

## Chunk 3: PVM Core Commands

### Task 9: Copy PVM command package

**Packages:**
- `internal/pvm/` -- PVM commands (install, use, global, local, versions, etc.)
- `internal/integration/` -- workspace integration

- [ ] **Step 1: Copy packages**

```bash
git checkout pu -- internal/pvm/
git checkout pu -- internal/integration/
```

- [ ] **Step 2: Strip PSC and typed-perl references**

```bash
grep -rn "psc\.\|typechecker\|binder\|compiler\|typed.perl\|type_check" internal/pvm/
```

For each reference found:
- If it's an import, remove the import line
- If it's a command registration (e.g., `psc.NewCommand`), remove that line
- If it's a feature flag or config option about types, remove it
- If removal breaks compilation, comment out the broken code with `// TODO: re-add when PSC is ready`

Key file: `internal/pvm/command.go` likely registers PSC. Remove that registration since we'll re-add it when PSC is rebuilt.

- [ ] **Step 3: Handle MCP references**

`internal/pvm/mcp.go` may reference the MCP server. If MCP depends on typed-perl infrastructure, remove the mcp command registration. If it's independent, keep it.

```bash
grep -r "internal/mcp\|internal/parser\|internal/ast" internal/pvm/mcp.go
```

If it has parser/ast dependencies, either:
- Remove mcp.go entirely (can add back later)
- Stub it to return "not yet available"

- [ ] **Step 4: Copy MCP if clean, skip if not**

```bash
grep -r "parser\|typechecker\|binder\|compiler\|ast" internal/mcp/
```

If MCP depends on parser infrastructure, skip it for now. If it's clean, copy it.

- [ ] **Step 5: Verify compilation**

```bash
go mod tidy
go build ./internal/pvm/... ./internal/integration/...
```

Fix compilation errors iteratively. The goal is to get pvm commands building without any parser/type-system dependencies.

- [ ] **Step 6: Run tests**

```bash
go test ./internal/pvm/... ./internal/integration/... -count=1
```

- [ ] **Step 7: Commit**

```bash
git add internal/pvm/ internal/integration/ go.mod go.sum
git commit -m "feat: copy PVM core commands (stripped of type-system references)"
```

---

### Task 10: Create PVM entry point

**Files:**
- Create: `cmd/pvm/main.go`

- [ ] **Step 1: Create cmd/pvm directory and main.go**

```bash
mkdir -p cmd/pvm
```

Write `cmd/pvm/main.go` based on the existing one but without PSC:

```go
// ABOUTME: Main entry point for PVM (Perl Version Manager)
// ABOUTME: Handles Perl installation, version switching, and management

package main

import (
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/compat"
	"tamarou.com/pvm/internal/pvm"
	"tamarou.com/pvm/internal/templates"
)

func init() {
	pvm.GlobalTemplatesFS = templates.FS

	cli.GlobalRegistry.Register(cli.ComponentPVM, pvm.NewCommand)

	// Compatibility commands
	cli.GlobalRegistry.Register(cli.ComponentCpanm, compat.NewCpanmCommand)
	cli.GlobalRegistry.Register(cli.ComponentCarton, compat.NewCartonCommand)
	cli.GlobalRegistry.Register(cli.ComponentPerlbrew, compat.NewPerlbrewCommand)
	cli.GlobalRegistry.Register(cli.ComponentPlenv, compat.NewPlenvCommand)
}

func main() {
	component := cli.DetectComponent()
	rootCmd := cli.CreateRootCommand(component)
	cli.Execute(rootCmd)
}
```

Note: PM, PVX, and PSC registrations will be added in later tasks.

- [ ] **Step 2: Verify it builds**

```bash
go build ./cmd/pvm/
```

Expected: produces a `pvm` binary.

- [ ] **Step 3: Smoke test**

```bash
./pvm --help
./pvm versions
```

Expected: help output shows PVM commands. `versions` lists installed Perls.

- [ ] **Step 4: Commit**

```bash
git add cmd/pvm/
git commit -m "feat: add PVM entry point"
```

---

## Chunk 4: PVX and PM

### Task 11: Copy and adapt PVX

**Packages:**
- `internal/pvx/` -- Perl executor

- [ ] **Step 1: Copy PVX**

```bash
git checkout pu -- internal/pvx/
```

- [ ] **Step 2: Strip parser/compiler/ast references**

```bash
grep -rn "internal/parser\|internal/compiler\|internal/ast\|type.check\|TypeCheck" internal/pvx/
```

Remove:
- Import lines for parser, compiler, ast
- Any `--type-check` flag registration
- Any code paths that call parser.NewParser(), compiler.NewCleanPerlCompilerUnified(), or type-checking functions
- Replace removed functionality with direct execution (pvx should just run Perl, no type checking)

- [ ] **Step 3: Verify compilation**

```bash
go mod tidy
go build ./internal/pvx/...
```

- [ ] **Step 4: Run tests**

```bash
go test ./internal/pvx/... -count=1
```

Remove or adapt tests that test type-checking integration.

- [ ] **Step 5: Create PVX entry point**

```bash
mkdir -p cmd/pvx
```

Copy from pu and verify:

```bash
git checkout pu -- cmd/pvx/
```

- [ ] **Step 6: Register PVX in PVM entry point**

Edit `cmd/pvm/main.go` to add:

```go
import "tamarou.com/pvm/internal/pvx"
```

And in init():

```go
cli.GlobalRegistry.Register(cli.ComponentPVX, pvx.NewCommand)
```

- [ ] **Step 7: Verify both entry points build**

```bash
go build ./cmd/pvm/ ./cmd/pvx/
```

- [ ] **Step 8: Commit**

```bash
git add internal/pvx/ cmd/pvx/ cmd/pvm/main.go go.mod go.sum
git commit -m "feat: copy PVX executor (stripped of type-check integration)"
```

---

### Task 12: Copy and adapt PM

**Packages:**
- `internal/pm/` -- module management
- `internal/dependencies/` -- dependency utilities
- `internal/typedef/` -- type definitions (check if needed)

- [ ] **Step 1: Copy PM**

```bash
git checkout pu -- internal/pm/
git checkout pu -- internal/dependencies/
```

- [ ] **Step 2: Strip parser/ast/type references**

```bash
grep -rn "internal/parser\|internal/ast\|internal/typedef\|type_command\|type.definition" internal/pm/
```

Remove:
- `type_command.go` entirely (or gut it)
- Import lines for parser, ast, typedef
- Any code in analyzer.go that uses parser for type extraction
- `pvx_integration.go` references to parser if any

- [ ] **Step 3: Check if typedef package is needed**

```bash
grep -r "internal/typedef" internal/pm/
```

If PM references typedef, and typedef is purely about type definitions, skip copying typedef. Remove the references from PM.

- [ ] **Step 4: Verify compilation**

```bash
go mod tidy
go build ./internal/pm/...
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/pm/... -count=1
```

Remove tests that test type definition integration.

- [ ] **Step 6: Create PM entry point and register**

```bash
git checkout pu -- cmd/pm/
```

Edit `cmd/pvm/main.go` to add PM registration:

```go
import "tamarou.com/pvm/internal/pm"
```

And in init():

```go
cli.GlobalRegistry.Register(cli.ComponentPM, pm.NewCommand)
```

- [ ] **Step 7: Verify all entry points build**

```bash
go build ./cmd/pvm/ ./cmd/pvx/ ./cmd/pm/
```

- [ ] **Step 8: Commit**

```bash
git add internal/pm/ internal/dependencies/ cmd/pm/ cmd/pvm/main.go go.mod go.sum
git commit -m "feat: copy PM module manager (stripped of type definition integration)"
```

---

## Chunk 5: New Parser with gotreesitter

### Task 13: Add gotreesitter dependency and build parser package

**Files:**
- Create: `internal/parser/parser.go` -- gotreesitter wrapper
- Create: `internal/parser/parser_test.go` -- parser tests

- [ ] **Step 1: Add gotreesitter dependency**

```bash
go get github.com/odvcencio/gotreesitter@latest
```

- [ ] **Step 2: Write the failing test for basic parsing**

Create `internal/parser/parser_test.go`:

```go
// ABOUTME: Tests for the gotreesitter-based Perl parser wrapper.
// ABOUTME: Verifies parsing of standard Perl constructs.

package parser_test

import (
	"testing"

	"tamarou.com/pvm/internal/parser"
)

func TestParseSimpleVariable(t *testing.T) {
	p := parser.New()
	tree, err := p.Parse([]byte("my $x = 42;\n"))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	root := tree.RootNode()
	if root.Kind() != "source_file" {
		t.Errorf("expected root kind source_file, got %s", root.Kind())
	}
	if root.HasError() {
		t.Error("parse tree has errors")
	}
}

func TestParseString(t *testing.T) {
	p := parser.New()
	tree, err := p.Parse([]byte(`my $x = "hello world";` + "\n"))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("parse tree has errors for string literal")
	}
}

func TestParseSubroutine(t *testing.T) {
	p := parser.New()
	src := []byte("sub greet {\n    my ($name) = @_;\n    say \"Hello, $name\";\n}\n")
	tree, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("parse tree has errors for subroutine")
	}
}

func TestParseClass(t *testing.T) {
	p := parser.New()
	src := []byte("use v5.40;\nclass Point {\n    field $x;\n    field $y;\n}\n")
	tree, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("parse tree has errors for class")
	}
}

func TestParseHeredoc(t *testing.T) {
	p := parser.New()
	src := []byte("my $text = <<END;\nHello world\nEND\n")
	tree, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("parse tree has errors for heredoc")
	}
}

func TestParseRegex(t *testing.T) {
	p := parser.New()
	src := []byte("if ($x =~ /hello/) {\n    say \"matched\";\n}\n")
	tree, err := p.Parse(src)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if tree.RootNode().HasError() {
		t.Error("parse tree has errors for regex")
	}
}

func TestNodeNavigation(t *testing.T) {
	p := parser.New()
	tree, err := p.Parse([]byte("my $x = 42;\n"))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	root := tree.RootNode()
	if root.ChildCount() == 0 {
		t.Fatal("root has no children")
	}
	child := root.Child(0)
	if child == nil {
		t.Fatal("first child is nil")
	}
	// Should be an expression_statement or similar
	kind := child.Kind()
	if kind == "" {
		t.Error("child kind is empty")
	}
	// Verify text extraction
	text := child.Text([]byte("my $x = 42;\n"))
	if text == "" {
		t.Error("child text is empty")
	}
}
```

- [ ] **Step 3: Run test to verify it fails**

```bash
go test ./internal/parser/... -count=1 -v
```

Expected: FAIL -- package parser does not exist yet.

- [ ] **Step 4: Write the parser implementation**

Create `internal/parser/parser.go`:

```go
// ABOUTME: Pure-Go Perl parser using gotreesitter's bundled Perl grammar.
// ABOUTME: Provides a clean wrapper hiding gotreesitter implementation details.

package parser

import (
	gotreesitter "github.com/odvcencio/gotreesitter"
	"github.com/odvcencio/gotreesitter/grammars"
)

// Parser wraps gotreesitter for parsing Perl source code.
type Parser struct {
	parser *gotreesitter.Parser
	lang   *gotreesitter.Language
}

// New creates a Parser using gotreesitter's bundled Perl grammar.
func New() *Parser {
	lang := grammars.PerlLanguage()
	return &Parser{
		parser: gotreesitter.NewParser(lang),
		lang:   lang,
	}
}

// Parse parses source code and returns a syntax tree.
func (p *Parser) Parse(source []byte) (*Tree, error) {
	tree, err := p.parser.Parse(source)
	if err != nil {
		return nil, err
	}
	return &Tree{tree: tree, lang: p.lang, source: source}, nil
}

// Tree represents a parsed syntax tree.
type Tree struct {
	tree   *gotreesitter.Tree
	lang   *gotreesitter.Language
	source []byte
}

// RootNode returns the root node of the parse tree.
func (t *Tree) RootNode() *Node {
	return &Node{node: t.tree.RootNode(), lang: t.lang}
}

// Node represents a node in the syntax tree.
type Node struct {
	node *gotreesitter.Node
	lang *gotreesitter.Language
}

// Kind returns the grammar rule name for this node (e.g. "source_file",
// "variable_declaration").
func (n *Node) Kind() string {
	return n.node.Type(n.lang)
}

// IsNamed returns true if this is a named node (not anonymous punctuation).
func (n *Node) IsNamed() bool {
	return n.node.IsNamed()
}

// HasError returns true if this node or any descendant is an error node.
func (n *Node) HasError() bool {
	return n.node.HasError()
}

// IsError returns true if this node itself is an error node.
func (n *Node) IsError() bool {
	return n.node.IsError()
}

// StartByte returns the byte offset where this node starts.
func (n *Node) StartByte() uint32 {
	return n.node.StartByte()
}

// EndByte returns the byte offset where this node ends.
func (n *Node) EndByte() uint32 {
	return n.node.EndByte()
}

// Text returns the source text covered by this node.
func (n *Node) Text(source []byte) string {
	return string(source[n.StartByte():n.EndByte()])
}

// ChildCount returns the number of children.
func (n *Node) ChildCount() int {
	return n.node.ChildCount()
}

// Child returns the child at the given index.
func (n *Node) Child(i int) *Node {
	child := n.node.Child(i)
	if child == nil {
		return nil
	}
	return &Node{node: child, lang: n.lang}
}

// NamedChildCount returns the number of named children.
func (n *Node) NamedChildCount() int {
	return n.node.NamedChildCount()
}

// NamedChild returns the named child at the given index.
func (n *Node) NamedChild(i int) *Node {
	child := n.node.NamedChild(i)
	if child == nil {
		return nil
	}
	return &Node{node: child, lang: n.lang}
}

// ChildByFieldName returns the child with the given field name.
func (n *Node) ChildByFieldName(name string) *Node {
	child := n.node.ChildByFieldName(name, n.lang)
	if child == nil {
		return nil
	}
	return &Node{node: child, lang: n.lang}
}

// Parent returns the parent node, or nil for the root.
func (n *Node) Parent() *Node {
	parent := n.node.Parent()
	if parent == nil {
		return nil
	}
	return &Node{node: parent, lang: n.lang}
}

// SExpr returns the S-expression representation of this node.
func (n *Node) SExpr() string {
	return n.node.SExpr(n.lang)
}
```

- [ ] **Step 5: Run tests to verify they pass**

```bash
go mod tidy
go test ./internal/parser/... -count=1 -v
```

Expected: all tests PASS. If heredoc or regex tests fail, it means the bundled Perl scanner has issues with those constructs. Document any failures as known limitations and mark tests as skipped with explanation.

- [ ] **Step 6: Commit**

```bash
git add internal/parser/ go.mod go.sum
git commit -m "feat: add gotreesitter-based Perl parser"
```

---

### Task 14: Add incremental parsing support

**Files:**
- Modify: `internal/parser/parser.go` -- add ParseIncremental, Edit type
- Create: `internal/parser/incremental_test.go` -- incremental parsing tests

- [ ] **Step 1: Write failing test for incremental parsing**

Create `internal/parser/incremental_test.go`:

```go
// ABOUTME: Tests for incremental parsing support.
// ABOUTME: Verifies that edits can be applied and trees re-parsed efficiently.

package parser_test

import (
	"testing"

	"tamarou.com/pvm/internal/parser"
)

func TestIncrementalParse(t *testing.T) {
	p := parser.New()

	// Initial parse
	src1 := []byte("my $x = 42;\n")
	tree1, err := p.Parse(src1)
	if err != nil {
		t.Fatalf("initial parse failed: %v", err)
	}
	if tree1.RootNode().HasError() {
		t.Fatal("initial parse has errors")
	}

	// Edit: change 42 to 99
	src2 := []byte("my $x = 99;\n")
	edit := parser.Edit{
		StartByte:  8,
		OldEndByte: 10,
		NewEndByte: 10,
	}

	tree2, err := p.ParseIncremental(src2, tree1, edit)
	if err != nil {
		t.Fatalf("incremental parse failed: %v", err)
	}
	if tree2.RootNode().HasError() {
		t.Error("incremental parse has errors")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/parser/... -run TestIncremental -count=1 -v
```

Expected: FAIL -- Edit type and ParseIncremental don't exist yet.

- [ ] **Step 3: Add Edit type and ParseIncremental to parser.go**

Add to `internal/parser/parser.go`:

```go
// Edit describes a change to source code for incremental re-parsing.
type Edit struct {
	StartByte  uint32
	OldEndByte uint32
	NewEndByte uint32
}

// ParseIncremental re-parses source after an edit, reusing unchanged subtrees
// from the previous parse for efficiency.
func (p *Parser) ParseIncremental(source []byte, oldTree *Tree, edit Edit) (*Tree, error) {
	oldTree.tree.Edit(gotreesitter.InputEdit{
		StartByte:  edit.StartByte,
		OldEndByte: edit.OldEndByte,
		NewEndByte: edit.NewEndByte,
	})
	tree, err := p.parser.ParseIncremental(source, oldTree.tree)
	if err != nil {
		return nil, err
	}
	return &Tree{tree: tree, lang: p.lang, source: source}, nil
}
```

Note: check the exact gotreesitter API for `InputEdit` and `ParseIncremental`. The field names and method signature may differ. Consult:

```bash
grep -r "func.*ParseIncremental\|type InputEdit\|func.*Edit(" /home/perigrin/dev/gotreesitter/*.go
```

Adapt the code to match the actual API.

- [ ] **Step 4: Run tests**

```bash
go test ./internal/parser/... -count=1 -v
```

Expected: all tests PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/parser/
git commit -m "feat: add incremental parsing support"
```

---

## Chunk 6: New PSC Commands

### Task 15: Build psc parse command

**Files:**
- Create: `internal/psc/command.go` -- PSC root command
- Create: `internal/psc/parse_command.go` -- parse subcommand
- Create: `internal/psc/parse_command_test.go` -- tests

- [ ] **Step 1: Write failing test**

Create `internal/psc/parse_command_test.go`:

```go
// ABOUTME: Tests for the psc parse command.
// ABOUTME: Verifies AST output in various formats.

package psc_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/psc"
)

func TestParseCommand_TreeOutput(t *testing.T) {
	// Create a temp Perl file
	dir := t.TempDir()
	file := filepath.Join(dir, "test.pl")
	err := os.WriteFile(file, []byte("my $x = 42;\n"), 0644)
	if err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	cmd := psc.NewCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"parse", file})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected non-empty output")
	}
}

func TestParseCommand_SExprOutput(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.pl")
	err := os.WriteFile(file, []byte("my $x = 42;\n"), 0644)
	if err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	cmd := psc.NewCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"parse", "--format", "sexpr", file})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected non-empty S-expression output")
	}
}

func TestParseCommand_NoFile(t *testing.T) {
	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"parse"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no file specified")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/psc/... -count=1 -v
```

Expected: FAIL -- package psc does not exist.

- [ ] **Step 3: Create PSC root command**

Create `internal/psc/command.go`:

```go
// ABOUTME: PSC root command and subcommand registration.
// ABOUTME: PSC provides Perl structural analysis, parsing, and LSP.

package psc

import (
	"github.com/spf13/cobra"
)

// NewCommand creates the PSC root command with all subcommands.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "psc",
		Short: "Perl Structural Checker - Parse, analyze, and serve Perl code",
		Long: `PSC provides structural analysis tools for Perl code.

Commands:
  psc parse <file>          Parse a file and display its AST
  psc analyze <file|dir>    Analyze project structure and dependencies
  psc lsp                   Start the Language Server Protocol server`,
	}

	cmd.AddCommand(
		newParseCommand(),
	)

	return cmd
}
```

- [ ] **Step 4: Create parse command**

Create `internal/psc/parse_command.go`:

```go
// ABOUTME: Implements the 'psc parse' command for AST inspection.
// ABOUTME: Parses Perl files and outputs syntax trees in various formats.

package psc

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/parser"
)

func newParseCommand() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "parse <file>",
		Short: "Parse a Perl file and display its syntax tree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("read file: %w", err)
			}

			p := parser.New()
			tree, err := p.Parse(source)
			if err != nil {
				return fmt.Errorf("parse: %w", err)
			}

			root := tree.RootNode()
			if root.HasError() {
				fmt.Fprintln(cmd.ErrOrStderr(), "warning: parse tree contains errors")
			}

			switch format {
			case "sexpr":
				fmt.Fprintln(cmd.OutOrStdout(), root.SExpr())
			case "tree":
				printTree(cmd, root, source, 0)
			default:
				return fmt.Errorf("unknown format: %s (use 'tree' or 'sexpr')", format)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "format", "f", "tree", "Output format: tree, sexpr")

	return cmd
}

func printTree(cmd *cobra.Command, node *parser.Node, source []byte, depth int) {
	if !node.IsNamed() {
		return
	}
	prefix := strings.Repeat("  ", depth)
	text := node.Text(source)
	if len(text) > 60 {
		text = text[:57] + "..."
	}
	text = strings.ReplaceAll(text, "\n", "\\n")
	fmt.Fprintf(cmd.OutOrStdout(), "%s(%s %q)\n", prefix, node.Kind(), text)
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child != nil {
			printTree(cmd, child, source, depth+1)
		}
	}
}
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/psc/... -count=1 -v
```

Expected: all tests PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/psc/
git commit -m "feat: add psc parse command with tree and sexpr output"
```

---

### Task 16: Build psc analyze command

**Files:**
- Create: `internal/psc/analyze_command.go` -- analyze subcommand
- Create: `internal/psc/analyze_command_test.go` -- tests

- [ ] **Step 1: Write failing test**

Create `internal/psc/analyze_command_test.go`:

```go
// ABOUTME: Tests for the psc analyze command.
// ABOUTME: Verifies dependency extraction and module analysis.

package psc_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"tamarou.com/pvm/internal/psc"
)

func TestAnalyzeCommand_SingleFile(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.pl")
	src := []byte("use strict;\nuse warnings;\nuse File::Basename;\nrequire Carp;\n")
	err := os.WriteFile(file, src, 0644)
	if err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	cmd := psc.NewCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"analyze", file})
	err = cmd.Execute()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Error("expected non-empty analysis output")
	}
}

func TestAnalyzeCommand_NoArgs(t *testing.T) {
	cmd := psc.NewCommand()
	cmd.SetArgs([]string{"analyze"})
	err := cmd.Execute()
	if err == nil {
		t.Error("expected error when no file/dir specified")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/psc/... -run TestAnalyze -count=1 -v
```

Expected: FAIL -- analyze command not registered.

- [ ] **Step 3: Create analyze command**

Create `internal/psc/analyze_command.go`:

```go
// ABOUTME: Implements the 'psc analyze' command for project analysis.
// ABOUTME: Extracts dependencies, maps module structure, and reports project layout.

package psc

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/parser"
)

func newAnalyzeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze <file|dir>",
		Short: "Analyze Perl project structure and dependencies",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]
			info, err := os.Stat(target)
			if err != nil {
				return fmt.Errorf("stat %s: %w", target, err)
			}

			var files []string
			if info.IsDir() {
				err = filepath.Walk(target, func(path string, fi os.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if !fi.IsDir() && (strings.HasSuffix(path, ".pl") || strings.HasSuffix(path, ".pm") || strings.HasSuffix(path, ".t")) {
						files = append(files, path)
					}
					return nil
				})
				if err != nil {
					return fmt.Errorf("walk %s: %w", target, err)
				}
			} else {
				files = []string{target}
			}

			p := parser.New()
			for _, file := range files {
				source, err := os.ReadFile(file)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: skip %s: %v\n", file, err)
					continue
				}
				tree, err := p.Parse(source)
				if err != nil {
					fmt.Fprintf(cmd.ErrOrStderr(), "warning: parse %s: %v\n", file, err)
					continue
				}

				deps := extractDependencies(tree.RootNode(), source)
				if len(deps) > 0 {
					fmt.Fprintf(cmd.OutOrStdout(), "%s:\n", file)
					for _, dep := range deps {
						fmt.Fprintf(cmd.OutOrStdout(), "  %s %s\n", dep.Kind, dep.Module)
					}
				}
			}
			return nil
		},
	}

	return cmd
}

// Dependency represents a module dependency found in source code.
type Dependency struct {
	Kind   string // "use" or "require"
	Module string
}

// extractDependencies walks the AST looking for use/require statements.
func extractDependencies(node *parser.Node, source []byte) []Dependency {
	var deps []Dependency
	for i := 0; i < node.ChildCount(); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		kind := child.Kind()
		if kind == "use_no_statement" || kind == "use_version_statement" {
			text := child.Text(source)
			parts := strings.Fields(text)
			if len(parts) >= 2 && (parts[0] == "use" || parts[0] == "no") {
				module := strings.TrimSuffix(parts[1], ";")
				deps = append(deps, Dependency{Kind: parts[0], Module: module})
			}
		} else if kind == "require_expression" {
			text := child.Text(source)
			parts := strings.Fields(text)
			if len(parts) >= 2 {
				module := strings.TrimSuffix(parts[1], ";")
				deps = append(deps, Dependency{Kind: "require", Module: module})
			}
		}
		// Recurse into children for nested structures
		nested := extractDependencies(child, source)
		deps = append(deps, nested...)
	}
	return deps
}
```

- [ ] **Step 4: Register analyze command in command.go**

Edit `internal/psc/command.go`, add to the `cmd.AddCommand()` call:

```go
cmd.AddCommand(
    newParseCommand(),
    newAnalyzeCommand(),
)
```

- [ ] **Step 5: Run tests**

```bash
go test ./internal/psc/... -count=1 -v
```

Expected: all tests PASS. The dependency extraction uses text parsing of AST nodes which is straightforward. If the exact node kinds differ from what gotreesitter produces for `use` statements, adjust the `extractDependencies` function by inspecting actual parse output:

```bash
go run ./cmd/psc parse --format sexpr /path/to/test.pl
```

- [ ] **Step 6: Commit**

```bash
git add internal/psc/
git commit -m "feat: add psc analyze command for dependency extraction"
```

---

### Task 17: Build psc lsp command (skeleton)

The LSP is the most complex PSC feature. Start with a skeleton that handles initialization and document sync, then grow it.

**Files:**
- Create: `internal/psc/lsp_command.go` -- lsp subcommand
- Create: `internal/psc/lsp.go` -- LSP server implementation
- Create: `internal/psc/lsp_test.go` -- tests

- [ ] **Step 1: Write failing test**

Create `internal/psc/lsp_test.go`:

```go
// ABOUTME: Tests for the PSC LSP server.
// ABOUTME: Verifies initialization and basic document sync.

package psc_test

import (
	"testing"

	"tamarou.com/pvm/internal/psc"
)

func TestLSPServer_Creation(t *testing.T) {
	server := psc.NewLSPServer()
	if server == nil {
		t.Fatal("NewLSPServer returned nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
go test ./internal/psc/... -run TestLSP -count=1 -v
```

Expected: FAIL.

- [ ] **Step 3: Create LSP server skeleton**

Create `internal/psc/lsp.go`:

```go
// ABOUTME: LSP server implementation for Perl structural analysis.
// ABOUTME: Provides document sync, diagnostics, and symbol navigation.

package psc

import (
	"sync"

	"tamarou.com/pvm/internal/parser"
)

// LSPServer provides Language Server Protocol support for Perl.
type LSPServer struct {
	parser    *parser.Parser
	documents map[string]*document
	mu        sync.RWMutex
}

type document struct {
	uri    string
	source []byte
	tree   *parser.Tree
}

// NewLSPServer creates a new LSP server instance.
func NewLSPServer() *LSPServer {
	return &LSPServer{
		parser:    parser.New(),
		documents: make(map[string]*document),
	}
}

// OpenDocument parses and tracks a document.
func (s *LSPServer) OpenDocument(uri string, source []byte) error {
	tree, err := s.parser.Parse(source)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.documents[uri] = &document{uri: uri, source: source, tree: tree}
	return nil
}

// CloseDocument removes a document from tracking.
func (s *LSPServer) CloseDocument(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.documents, uri)
}
```

- [ ] **Step 4: Create LSP command**

Create `internal/psc/lsp_command.go`:

```go
// ABOUTME: Implements the 'psc lsp' command to start the LSP server.
// ABOUTME: Communicates over stdin/stdout using the LSP protocol.

package psc

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLSPCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lsp",
		Short: "Start the Language Server Protocol server",
		Long:  "Starts a Perl LSP server on stdin/stdout for editor integration.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// LSP protocol implementation will be added incrementally.
			// For now, create the server to verify the infrastructure works.
			_ = NewLSPServer()
			return fmt.Errorf("LSP server not yet implemented -- coming soon")
		},
	}
	return cmd
}
```

- [ ] **Step 5: Register LSP command**

Edit `internal/psc/command.go`:

```go
cmd.AddCommand(
    newParseCommand(),
    newAnalyzeCommand(),
    newLSPCommand(),
)
```

- [ ] **Step 6: Run tests**

```bash
go test ./internal/psc/... -count=1 -v
```

Expected: all tests PASS.

- [ ] **Step 7: Commit**

```bash
git add internal/psc/
git commit -m "feat: add psc lsp command skeleton with document tracking"
```

---

## Chunk 7: PSC Entry Point, E2E Tests, and Polish

### Task 18: Create PSC entry point and wire into PVM

**Files:**
- Create: `cmd/psc/main.go`
- Modify: `cmd/pvm/main.go` -- register PSC

- [ ] **Step 1: Create PSC entry point**

```bash
mkdir -p cmd/psc
```

Create `cmd/psc/main.go`:

```go
// ABOUTME: Standalone entry point for PSC (Perl Structural Checker).
// ABOUTME: Can be invoked directly or via pvm psc subcommand.

package main

import (
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/psc"
)

func init() {
	cli.GlobalRegistry.Register(cli.ComponentPSC, psc.NewCommand)
}

func main() {
	component := cli.DetectComponent()
	rootCmd := cli.CreateRootCommand(component)
	cli.Execute(rootCmd)
}
```

- [ ] **Step 2: Register PSC in PVM entry point**

Edit `cmd/pvm/main.go` to add:

```go
import "tamarou.com/pvm/internal/psc"
```

And in init():

```go
cli.GlobalRegistry.Register(cli.ComponentPSC, psc.NewCommand)
```

- [ ] **Step 3: Build all entry points**

```bash
go build ./cmd/pvm/ ./cmd/pvx/ ./cmd/pm/ ./cmd/psc/
```

Expected: all four binaries build.

- [ ] **Step 4: Smoke test PSC**

```bash
./psc parse --help
./psc analyze --help
./psc lsp --help
```

Also test via PVM:

```bash
./pvm psc parse --help
```

- [ ] **Step 5: Commit**

```bash
git add cmd/psc/ cmd/pvm/main.go
git commit -m "feat: add PSC entry point and register all four components"
```

---

### Task 19: Adapt Makefile for full build

**Files:**
- Modify: `Makefile`

- [ ] **Step 1: Update Makefile**

Replace the minimal Makefile with one that builds all components:

```makefile
.PHONY: all build test clean pvm pvx pm psc cross-compile

BINARIES = pvm pvx pm psc

all: build

build: $(BINARIES)

pvm:
	go build -o pvm ./cmd/pvm/

pvx:
	go build -o pvx ./cmd/pvx/

pm:
	go build -o pm ./cmd/pm/

psc:
	go build -o psc ./cmd/psc/

test:
	go test ./... -count=1

clean:
	rm -f $(BINARIES)
	go clean ./...

cross-compile:
	GOOS=linux GOARCH=amd64 go build -o pvm-linux-amd64 ./cmd/pvm/
	GOOS=linux GOARCH=arm64 go build -o pvm-linux-arm64 ./cmd/pvm/
	GOOS=darwin GOARCH=amd64 go build -o pvm-darwin-amd64 ./cmd/pvm/
	GOOS=darwin GOARCH=arm64 go build -o pvm-darwin-arm64 ./cmd/pvm/
```

- [ ] **Step 2: Verify**

```bash
make clean && make
make test
```

Expected: all binaries build, all tests pass.

- [ ] **Step 3: Commit**

```bash
git add Makefile
git commit -m "feat: update Makefile for full four-component build"
```

---

### Task 20: Copy and adapt E2E tests

**Files:**
- Copy: `test/e2e/` from pu (selectively)

- [ ] **Step 1: Copy e2e test directory**

```bash
git checkout pu -- test/e2e/
```

- [ ] **Step 2: Identify and remove type-system tests**

```bash
grep -rl "psc check\|psc strip\|psc infer\|psc compile\|psc def\|type.check\|type_annotation\|typed.perl" test/e2e/
```

Remove or gut any test files that exclusively test type-checking functionality.

- [ ] **Step 3: Adapt remaining tests**

For tests that reference PSC commands that still exist (parse, analyze), keep them. For tests that reference removed commands, delete them.

- [ ] **Step 4: Run e2e tests**

```bash
go test ./test/e2e/... -count=1 -v
```

Fix failures. Some may need PVM to be installed or configured. Mark environment-dependent tests with appropriate build tags or skip conditions.

- [ ] **Step 5: Commit**

```bash
git add test/e2e/ go.mod go.sum
git commit -m "feat: copy and adapt e2e tests for pure-go branch"
```

---

### Task 21: Final verification and cross-compilation check

- [ ] **Step 1: Run full test suite**

```bash
make test
```

Expected: 100% pass rate (or document any pre-existing flaky tests).

- [ ] **Step 2: Verify cross-compilation (no CGO)**

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ./cmd/pvm/
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ./cmd/pvm/
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ./cmd/pvm/
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ./cmd/pvm/
```

Expected: all cross-compilations succeed with `CGO_ENABLED=0`. This is the whole point -- no CGO.

- [ ] **Step 3: Check binary sizes**

```bash
ls -lh pvm pvx pm psc
```

Document sizes. Compare with current pu branch binaries if available.

- [ ] **Step 4: Verify no tree-sitter-typed-perl remnants**

```bash
grep -r "tree-sitter-typed-perl\|typed_perl\|TypedPerl" .
grep -r "CGO_\|cgo " Makefile go.mod
```

Expected: no matches.

- [ ] **Step 5: Commit any final fixes**

```bash
git add -A
git commit -m "feat: complete pure-go rewrite verification"
```

---

## Summary

| Chunk | Tasks | What it delivers |
|-------|-------|-----------------|
| 1: Bootstrap | 1-5 | Empty Go module with all foundation/utility packages building and tested |
| 2: Perl Toolchain | 6-8 | CLI framework, updater, Perl toolchain, tool management |
| 3: PVM Core | 9-10 | Working `pvm` binary with all version management commands |
| 4: PVX and PM | 11-12 | Working `pvx` and `pm` binaries |
| 5: Parser | 13-14 | gotreesitter-based Perl parser with incremental support |
| 6: PSC | 15-17 | Working `psc` binary with parse, analyze, lsp commands |
| 7: Polish | 18-21 | All four entry points wired, e2e tests, cross-compilation verified |
