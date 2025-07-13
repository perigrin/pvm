// ABOUTME: Pipeline-based Perl compiler that uses the transformation pipeline architecture
// ABOUTME: Provides composable, flexible compilation using the new pipeline system

package compiler

import (
	"fmt"

	sitter "github.com/tree-sitter/go-tree-sitter"
	"tamarou.com/pvm/internal/compiler/pipeline"
)

// PipelineCompiler is a compiler that uses the transformation pipeline architecture
type PipelineCompiler struct {
	target   Target
	pipeline pipeline.TransformationPipeline
	options  CompilerOptions
}

// NewPipelineCompiler creates a new pipeline-based compiler
func NewPipelineCompiler(target Target, transformationPipeline pipeline.TransformationPipeline) *PipelineCompiler {
	return &PipelineCompiler{
		target:   target,
		pipeline: transformationPipeline,
		options: CompilerOptions{
			PreserveComments:   true,
			PreserveFormatting: true,
			StrictMode:         false,
			CustomPatterns:     nil,
		},
	}
}

// NewCleanPerlPipelineCompiler creates a pipeline compiler for clean Perl output
func NewCleanPerlPipelineCompiler() *PipelineCompiler {
	cleanPipeline := pipeline.CreateCleanPerlPipeline()
	return NewPipelineCompiler(TargetCleanPerl, cleanPipeline)
}

// NewTypedPerlPipelineCompiler creates a pipeline compiler for typed Perl output
func NewTypedPerlPipelineCompiler() *PipelineCompiler {
	typedPipeline := pipeline.CreateTypedFormatterPipeline()
	return NewPipelineCompiler(TargetTypedPerl, typedPipeline)
}

// NewFormatterPipelineCompiler creates a pipeline compiler for code formatting
func NewFormatterPipelineCompiler() *PipelineCompiler {
	formatterPipeline := pipeline.CreateFormatterPipeline()
	return NewPipelineCompiler("formatter", formatterPipeline)
}

// Target returns the compilation target
func (pc *PipelineCompiler) Target() Target {
	return pc.target
}

// Validate checks if the AST is suitable for compilation
func (pc *PipelineCompiler) Validate(ast AST) error {
	if ast == nil {
		return NewCompilerError(ErrInvalidAST, "AST cannot be nil")
	}

	if !ast.IsValid() {
		return NewCompilerError(ErrInvalidAST, "AST is not valid")
	}

	// Check if we can get content
	_, err := ast.GetContent()
	if err != nil {
		return NewCompilerError(ErrInvalidAST, "AST must have accessible source content").WithCause(err)
	}

	return nil
}

// Compile converts an AST to source code using the transformation pipeline
func (pc *PipelineCompiler) Compile(ast AST) (string, error) {
	// Validate the AST first
	if err := pc.Validate(ast); err != nil {
		return "", err
	}

	// Get the source content
	content, err := ast.GetContent()
	if err != nil {
		return "", NewCompilerError(ErrInvalidAST, "failed to get AST content").WithCause(err)
	}

	// Check if this is already a CST-based AST
	//nolint:go-critic
	if cstAST, ok := ast.(*CSTBasedAST); ok {
		return pc.compileFromCST(cstAST.GetCSTRoot(), []byte(content))
	}

	// For other AST types, we need to re-parse the content to get CST
	// This is for backward compatibility with existing AST types
	cstAST, err := NewCSTBasedAST(ast.GetPath(), content)
	if err != nil {
		return "", NewCompilerError(ErrInvalidAST, "failed to create CST from content").WithCause(err)
	}

	return pc.compileFromCST(cstAST.GetCSTRoot(), []byte(content))
}

// compileFromCST performs the actual compilation using the transformation pipeline
func (pc *PipelineCompiler) compileFromCST(root *sitter.Node, content []byte) (string, error) {
	if root == nil {
		return "", NewCompilerError(ErrInvalidCST, "CST root node is nil")
	}

	// Execute the transformation pipeline
	result, err := pc.pipeline.Execute(root, content)
	if err != nil {
		return "", NewCompilerError(ErrCompilationFailed, "pipeline execution failed").WithCause(err)
	}

	return result.Content, nil
}

// WithOptions returns a new compiler with modified options
func (pc *PipelineCompiler) WithOptions(options CompilerOptions) *PipelineCompiler {
	// Convert compiler options to pipeline options
	pipelineOptions := pipeline.PipelineOptions{
		PreserveComments:    options.PreserveComments,
		PreserveWhitespace:  options.PreserveFormatting,
		EnableOptimizations: !options.StrictMode, // In strict mode, disable optimizations
		Debug:               false,
		MaxTransformers:     50,
	}

	newPipeline := pc.pipeline.WithOptions(pipelineOptions)

	return &PipelineCompiler{
		target:   pc.target,
		pipeline: newPipeline,
		options:  options,
	}
}

// GetPipeline returns the underlying transformation pipeline
func (pc *PipelineCompiler) GetPipeline() pipeline.TransformationPipeline {
	return pc.pipeline
}

// GetTransformationSteps executes the pipeline and returns detailed transformation steps
func (pc *PipelineCompiler) GetTransformationSteps(ast AST) (*pipeline.TransformationResult, error) {
	// Validate the AST first
	if err := pc.Validate(ast); err != nil {
		return nil, err
	}

	// Get the source content
	content, err := ast.GetContent()
	if err != nil {
		return nil, NewCompilerError(ErrInvalidAST, "failed to get AST content").WithCause(err)
	}

	// Get CST root
	var root *sitter.Node
	//nolint:go-critic
	if cstAST, ok := ast.(*CSTBasedAST); ok {
		root = cstAST.GetCSTRoot()
	} else {
		// Re-parse for backward compatibility
		cstAST, err := NewCSTBasedAST(ast.GetPath(), content)
		if err != nil {
			return nil, NewCompilerError(ErrInvalidAST, "failed to create CST from content").WithCause(err)
		}
		root = cstAST.GetCSTRoot()
	}

	// Execute the pipeline and return full result
	return pc.pipeline.Execute(root, []byte(content))
}

// PipelineCompilerFactory provides factory methods for creating pipeline compilers
type PipelineCompilerFactory struct{}

// NewPipelineCompilerFactory creates a new factory instance
func NewPipelineCompilerFactory() *PipelineCompilerFactory {
	return &PipelineCompilerFactory{}
}

// CreateForTarget creates a pipeline compiler for a specific target
func (f *PipelineCompilerFactory) CreateForTarget(target Target) (*PipelineCompiler, error) {
	switch target {
	case TargetCleanPerl:
		return NewCleanPerlPipelineCompiler(), nil
	case TargetTypedPerl:
		return NewTypedPerlPipelineCompiler(), nil
	default:
		return nil, fmt.Errorf("unsupported target: %s", target)
	}
}

// CreateWithPreset creates a pipeline compiler using a preset configuration
func (f *PipelineCompilerFactory) CreateWithPreset(presetName string) (*PipelineCompiler, error) {
	presets := pipeline.GetAllPresets()

	for _, preset := range presets {
		if preset.Name == presetName {
			return NewPipelineCompiler(Target(preset.Name), preset.Pipeline), nil
		}
	}

	return nil, fmt.Errorf("unknown preset: %s", presetName)
}

// CreateCustom creates a pipeline compiler with a custom pipeline
func (f *PipelineCompilerFactory) CreateCustom(target Target, customPipeline pipeline.TransformationPipeline) *PipelineCompiler {
	return NewPipelineCompiler(target, customPipeline)
}

// Error codes for pipeline compiler
const (
	ErrInvalidCST = "INVALID_CST"
)

// Integration with existing compiler registry

// RegisterPipelineCompilers registers pipeline-based compilers with the registry
func RegisterPipelineCompilers(registry *CompilerRegistry) {
	// Register pipeline compilers alongside existing ones
	registry.Register(NewCleanPerlPipelineCompiler())
	registry.Register(NewTypedPerlPipelineCompiler())
	registry.Register(NewFormatterPipelineCompiler())

	// Create pipeline variants of existing targets
	factory := NewPipelineCompilerFactory()

	if cleanCompiler, err := factory.CreateForTarget(TargetCleanPerl); err == nil {
		registry.Register(cleanCompiler)
	}

	if typedCompiler, err := factory.CreateForTarget(TargetTypedPerl); err == nil {
		registry.Register(typedCompiler)
	}
}
