// ABOUTME: Core compiler interfaces and types for converting AST to various target formats
// ABOUTME: Provides extensible architecture for multiple compilation targets

//go:generate moq -out compiler_mock.go . Compiler

package compiler

// Target represents a compilation target type
type Target string

const (
	// TargetCleanPerl produces standard Perl code without type annotations
	TargetCleanPerl Target = "clean_perl"

	// TargetTypedPerl produces Perl code with type annotations preserved
	TargetTypedPerl Target = "typed_perl"

	// TargetInferredTypeAnnotations produces Perl code with inferred type annotations
	TargetInferredTypeAnnotations Target = "inferred_typed_perl"
)

// Compiler interface defines the contract for AST-to-code compilation
type Compiler interface {
	// Compile converts an AST to source code for the target format
	Compile(ast AST) (string, error)

	// Target returns the compilation target this compiler supports
	Target() Target

	// Validate checks if the AST is suitable for compilation with this compiler
	Validate(ast AST) error
}

// OptionsCompiler is a compiler that supports configuration options
type OptionsCompiler interface {
	Compiler
	SetOptions(options CompilerOptions)
}

// CompilerOptions holds configuration options for compilation
type CompilerOptions struct {
	// PreserveComments controls whether comments are preserved in output
	PreserveComments bool

	// PreserveFormatting controls whether original formatting is preserved
	PreserveFormatting bool

	// StrictMode enables stricter compilation checks
	StrictMode bool

	// PerlVersion specifies the Perl version to use for pragmas (e.g., "5.42.0")
	// If empty, the compiler will auto-detect from current environment
	PerlVersion string

	// CustomPatterns allows custom transformation patterns
	CustomPatterns map[string]string
}

// CompilerRegistry manages available compilers for different targets
type CompilerRegistry struct {
	compilers map[Target]Compiler
}

// NewCompilerRegistry creates a new compiler registry with default compilers
func NewCompilerRegistry() *CompilerRegistry {
	registry := &CompilerRegistry{
		compilers: make(map[Target]Compiler),
	}

	// Register unified compilers (CST-based, replacing legacy AST-based compilers)
	registry.Register(NewCleanPerlCompilerUnified())
	registry.Register(NewTypedPerlCompilerUnified())
	registry.Register(NewInferredTypedPerlCompiler()) // Keep this one as-is for now

	return registry
}

// Register adds a compiler to the registry
func (r *CompilerRegistry) Register(compiler Compiler) {
	r.compilers[compiler.Target()] = compiler
}

// GetCompiler returns the compiler for the specified target
func (r *CompilerRegistry) GetCompiler(target Target) (Compiler, bool) {
	compiler, exists := r.compilers[target]
	return compiler, exists
}

// AvailableTargets returns a list of all available compilation targets
func (r *CompilerRegistry) AvailableTargets() []Target {
	targets := make([]Target, 0, len(r.compilers))
	for target := range r.compilers {
		targets = append(targets, target)
	}
	return targets
}

// CompileWithOptions compiles an AST using the specified target and options
func (r *CompilerRegistry) CompileWithOptions(ast AST, target Target, options *CompilerOptions) (string, error) {
	compiler, exists := r.GetCompiler(target)
	if !exists {
		return "", &CompilerError{
			Code:    "UNKNOWN_TARGET",
			Message: "unknown compilation target: " + string(target),
		}
	}

	// Set options if the compiler supports them
	if optCompiler, ok := compiler.(OptionsCompiler); ok && options != nil {
		optCompiler.SetOptions(*options)
	}

	if err := compiler.Validate(ast); err != nil {
		return "", err
	}

	return compiler.Compile(ast)
}

// Compile compiles an AST using the specified target with default options
func (r *CompilerRegistry) Compile(ast AST, target Target) (string, error) {
	return r.CompileWithOptions(ast, target, &CompilerOptions{
		PreserveComments:   true,
		PreserveFormatting: true,
		StrictMode:         false,
		PerlVersion:        "", // Auto-detect from environment
		CustomPatterns:     nil,
	})
}
