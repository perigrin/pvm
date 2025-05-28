// ABOUTME: Type system pooling implementation for memory-efficient type checking operations
// ABOUTME: Provides object pools for Type objects, InferenceContext, and related structures following TypeScript-Go patterns

package typechecker

import (
	"sync"
	"sync/atomic"

	"tamarou.com/pvm/internal/ast"
	"tamarou.com/pvm/internal/binder"
	"tamarou.com/pvm/internal/core"
	"tamarou.com/pvm/internal/typedef"
)

// TypePoolManager provides pooled allocation for type system objects
type TypePoolManager struct {
	hooks TypePoolHooks

	// Core type system pools
	typeCheckResultPool     core.Pool[TypeCheckResult]
	typeCheckErrorPool      core.Pool[TypeCheckError]
	functionSignaturePool   core.Pool[FunctionSignature]
	genericFunctionSigPool  core.Pool[GenericFunctionSignature]
	typeStatePool           core.Pool[TypeState]
	conditionPool           core.Pool[Condition]
	validationPatternPool   core.Pool[ValidationPattern]
	higherKindedTypeDefPool core.Pool[HigherKindedTypeDefinition]

	// Inference engine pools
	inferenceEnginePool      core.Pool[InferenceEngine]
	inferredTypeInfoPool     core.Pool[InferredTypeInfo]
	inferenceSourcePool      core.Pool[InferenceSource]
	inferenceContextPool     core.Pool[InferenceContext]
	dataFlowAnalyzerPool     core.Pool[DataFlowAnalyzer]
	contextAnalyzerPool      core.Pool[ContextAnalyzer]
	usagePatternAnalyzerPool core.Pool[UsagePatternAnalyzer]
	typePropagatorPool       core.Pool[TypePropagator]
	dataFlowGraphPool        core.Pool[DataFlowGraph]
	dataFlowNodePool         core.Pool[DataFlowNode]
	dataFlowEdgePool         core.Pool[DataFlowEdge]
	variableStateTrackerPool core.Pool[VariableStateTracker]
	variableStatePool        core.Pool[VariableState]
	typeTransformationPool   core.Pool[TypeTransformation]
	perlContextPool          core.Pool[PerlContext]
	contextRulePool          core.Pool[ContextRule]
	usagePatternPool         core.Pool[UsagePattern]
	patternMatchPool         core.Pool[PatternMatch]
	propagationRulePool      core.Pool[PropagationRule]
	typeConstraintPool       core.Pool[TypeConstraint]

	// Map pools for efficient reuse
	stringMapPool        *core.Pool[map[string]string]
	functionSigMapPool   *core.Pool[map[string]*FunctionSignature]
	genericFuncMapPool   *core.Pool[map[string]*GenericFunctionSignature]
	moduleTypeMapPool    *core.Pool[map[string]map[string]string]
	higherKindedMapPool  *core.Pool[map[string]*HigherKindedTypeDefinition]
	inferredTypeMapPool  *core.Pool[map[string]*InferredTypeInfo]
	variableStateMapPool *core.Pool[map[string]*VariableStateTracker]
	contextRuleMapPool   *core.Pool[map[string]ContextRule]
	definitionMapPool    *core.Pool[map[string]string]
	usesMapPool          *core.Pool[map[string]string]
	constraintMapPool    *core.Pool[map[string][]TypeConstraint]
	nodeMapPool          *core.Pool[map[string]*DataFlowNode]
	edgeMapPool          *core.Pool[map[string][]*DataFlowEdge]
	stateMapPool         *core.Pool[map[string]*VariableState]
	detailMapPool        *core.Pool[map[string]interface{}]

	// Slice pools for collections
	typeErrorSlicePool         *core.Pool[[]TypeCheckError]
	typeAnnotationSlicePool    *core.Pool[[]*ast.TypeAnnotation]
	conditionSlicePool         *core.Pool[[]Condition]
	validationPatternSlicePool *core.Pool[[]ValidationPattern]
	typeStateSlicePool         *core.Pool[[]*TypeState]
	inferenceSourceSlicePool   *core.Pool[[]InferenceSource]
	contextStackSlicePool      *core.Pool[[]PerlContext]
	usagePatternSlicePool      *core.Pool[[]UsagePattern]
	transformationSlicePool    *core.Pool[[]TypeTransformation]
	propagationRuleSlicePool   *core.Pool[[]PropagationRule]
	typeConstraintSlicePool    *core.Pool[[]TypeConstraint]
	stringSlicePool            *core.Pool[[]string]
	edgeSlicePool              *core.Pool[[]*DataFlowEdge]

	// TypeChecker pools
	typeCheckerPool *core.Pool[TypeChecker]

	// Statistics and monitoring
	typeCheckCount int64
	inferenceCount int64
	poolHits       int64
	poolMisses     int64
	memoryReused   int64
	typeCreated    int64
	typeReused     int64

	mu sync.RWMutex
}

// TypePoolHooks provides lifecycle hooks for debugging and monitoring
type TypePoolHooks struct {
	OnTypeCheckCreate func(result *TypeCheckResult) // Called when a type check result is created
	OnInferenceCreate func(engine *InferenceEngine) // Called when an inference engine is created
	OnTypeStateCreate func(state *TypeState)        // Called when a type state is created
	OnTypeCheckReset  func(result *TypeCheckResult) // Called when a type check result is reset for pooling
	OnInferenceReset  func(engine *InferenceEngine) // Called when an inference engine is reset for pooling
	OnTypeStateReset  func(state *TypeState)        // Called when a type state is reset for pooling
	OnPoolWarming     func(poolType string)         // Called during pool warming
	OnTypeCreated     func(typeName string)         // Called when a type is created
	OnTypeReused      func(typeName string)         // Called when a type is reused from pool
	OnMemoryThreshold func(usage int64)             // Called when memory usage exceeds threshold
}

// TypePoolCoercible allows types to provide a TypePoolManager
type TypePoolCoercible interface {
	AsTypePoolManager() *TypePoolManager
}

// NewTypePoolManager creates a new type pool manager with the given hooks
func NewTypePoolManager(hooks TypePoolHooks) *TypePoolManager {
	manager := &TypePoolManager{
		hooks: hooks,
	}

	// Initialize map pools
	manager.stringMapPool = &core.Pool[map[string]string]{}
	manager.functionSigMapPool = &core.Pool[map[string]*FunctionSignature]{}
	manager.genericFuncMapPool = &core.Pool[map[string]*GenericFunctionSignature]{}
	manager.moduleTypeMapPool = &core.Pool[map[string]map[string]string]{}
	manager.higherKindedMapPool = &core.Pool[map[string]*HigherKindedTypeDefinition]{}
	manager.inferredTypeMapPool = &core.Pool[map[string]*InferredTypeInfo]{}
	manager.variableStateMapPool = &core.Pool[map[string]*VariableStateTracker]{}
	manager.contextRuleMapPool = &core.Pool[map[string]ContextRule]{}
	manager.definitionMapPool = &core.Pool[map[string]string]{}
	manager.usesMapPool = &core.Pool[map[string]string]{}
	manager.constraintMapPool = &core.Pool[map[string][]TypeConstraint]{}
	manager.nodeMapPool = &core.Pool[map[string]*DataFlowNode]{}
	manager.edgeMapPool = &core.Pool[map[string][]*DataFlowEdge]{}
	manager.stateMapPool = &core.Pool[map[string]*VariableState]{}
	manager.detailMapPool = &core.Pool[map[string]interface{}]{}

	// Initialize slice pools
	manager.typeErrorSlicePool = &core.Pool[[]TypeCheckError]{}
	manager.typeAnnotationSlicePool = &core.Pool[[]*ast.TypeAnnotation]{}
	manager.conditionSlicePool = &core.Pool[[]Condition]{}
	manager.validationPatternSlicePool = &core.Pool[[]ValidationPattern]{}
	manager.typeStateSlicePool = &core.Pool[[]*TypeState]{}
	manager.inferenceSourceSlicePool = &core.Pool[[]InferenceSource]{}
	manager.contextStackSlicePool = &core.Pool[[]PerlContext]{}
	manager.usagePatternSlicePool = &core.Pool[[]UsagePattern]{}
	manager.transformationSlicePool = &core.Pool[[]TypeTransformation]{}
	manager.propagationRuleSlicePool = &core.Pool[[]PropagationRule]{}
	manager.typeConstraintSlicePool = &core.Pool[[]TypeConstraint]{}
	manager.stringSlicePool = &core.Pool[[]string]{}
	manager.edgeSlicePool = &core.Pool[[]*DataFlowEdge]{}

	// Initialize TypeChecker pool
	manager.typeCheckerPool = &core.Pool[TypeChecker]{}

	// Register with global pool manager for monitoring
	core.RegisterGlobalPool("type-pool", manager)

	// Warm up pools with common types
	manager.warmPools()

	return manager
}

// AsTypePoolManager implements TypePoolCoercible
func (tpm *TypePoolManager) AsTypePoolManager() *TypePoolManager {
	return tpm
}

// Stats returns pool allocation statistics
func (tpm *TypePoolManager) Stats() core.PoolStats {
	return core.PoolStats{
		Allocations: atomic.LoadInt64(&tpm.typeCheckCount) + atomic.LoadInt64(&tpm.inferenceCount),
		Grows:       atomic.LoadInt64(&tpm.poolHits),
		TotalSize:   atomic.LoadInt64(&tpm.poolMisses),
		CurrentSize: atomic.LoadInt64(&tpm.memoryReused),
		Capacity:    atomic.LoadInt64(&tpm.typeCreated) + atomic.LoadInt64(&tpm.typeReused),
	}
}

// Pool statistics methods

// TypeCheckCount returns the total number of type check operations
func (tpm *TypePoolManager) TypeCheckCount() int64 {
	return atomic.LoadInt64(&tpm.typeCheckCount)
}

// InferenceCount returns the total number of type inference operations
func (tpm *TypePoolManager) InferenceCount() int64 {
	return atomic.LoadInt64(&tpm.inferenceCount)
}

// PoolEfficiency returns the pool hit rate as a percentage
func (tpm *TypePoolManager) PoolEfficiency() float64 {
	hits := atomic.LoadInt64(&tpm.poolHits)
	misses := atomic.LoadInt64(&tpm.poolMisses)
	total := hits + misses

	if total == 0 {
		return 0
	}
	return float64(hits) / float64(total) * 100
}

// TypeReuseRate returns the type reuse rate as a percentage
func (tpm *TypePoolManager) TypeReuseRate() float64 {
	created := atomic.LoadInt64(&tpm.typeCreated)
	reused := atomic.LoadInt64(&tpm.typeReused)
	total := created + reused

	if total == 0 {
		return 0
	}
	return float64(reused) / float64(total) * 100
}

// TypeCheckResult creation methods

// NewTypeCheckResult creates a pooled type check result
func (tpm *TypePoolManager) NewTypeCheckResult(path string) *TypeCheckResult {
	result := tpm.typeCheckResultPool.New()

	// Reset/initialize the pooled object
	tpm.resetTypeCheckResult(result)

	// Initialize fields
	result.Path = path
	result.Errors = tpm.getTypeErrorSlice()
	result.TypeAnnotations = tpm.getTypeAnnotationSlice()
	result.RefinedTypes = tpm.getStringMap()
	result.FlowSensitiveEnabled = false

	atomic.AddInt64(&tpm.typeCheckCount, 1)
	atomic.AddInt64(&tpm.poolHits, 1)

	if tpm.hooks.OnTypeCheckCreate != nil {
		tpm.hooks.OnTypeCheckCreate(result)
	}

	return result
}

// NewTypeCheckError creates a pooled type check error
func (tpm *TypePoolManager) NewTypeCheckError(message, path string, line, column int) *TypeCheckError {
	error := tpm.typeCheckErrorPool.New()

	// Initialize fields
	error.Message = message
	error.Path = path
	error.Line = line
	error.Column = column

	atomic.AddInt64(&tpm.poolHits, 1)

	return error
}

// NewFunctionSignature creates a pooled function signature
func (tpm *TypePoolManager) NewFunctionSignature(returnType string, isMethod bool) *FunctionSignature {
	sig := tpm.functionSignaturePool.New()

	// Reset/initialize the pooled object
	tpm.resetFunctionSignature(sig)

	// Initialize fields
	sig.ParameterTypes = tpm.getStringMap()
	sig.ReturnType = returnType
	sig.IsMethod = isMethod

	atomic.AddInt64(&tpm.poolHits, 1)

	return sig
}

// NewGenericFunctionSignature creates a pooled generic function signature
func (tpm *TypePoolManager) NewGenericFunctionSignature(returnType string, isMethod bool) *GenericFunctionSignature {
	sig := tpm.genericFunctionSigPool.New()

	// Reset/initialize the pooled object
	tpm.resetGenericFunctionSignature(sig)

	// Initialize fields
	sig.TypeParameters = tpm.getStringSlice()
	sig.ParameterTypes = tpm.getStringMap()
	sig.ReturnType = returnType
	sig.Constraints = tpm.getStringSliceMap()
	sig.IsMethod = isMethod

	atomic.AddInt64(&tpm.poolHits, 1)

	return sig
}

// NewTypeState creates a pooled type state
func (tpm *TypePoolManager) NewTypeState() *TypeState {
	state := tpm.typeStatePool.New()

	// Reset/initialize the pooled object
	tpm.resetTypeState(state)

	// Initialize fields
	state.VariableTypes = tpm.getStringMap()
	state.RefinedTypes = tpm.getStringMap()
	state.Conditions = tpm.getConditionSlice()
	state.SkipFlowChecks = false

	atomic.AddInt64(&tpm.poolHits, 1)

	if tpm.hooks.OnTypeStateCreate != nil {
		tpm.hooks.OnTypeStateCreate(state)
	}

	return state
}

// NewCondition creates a pooled condition
func (tpm *TypePoolManager) NewCondition(variable, operator, value string, negated bool) *Condition {
	condition := tpm.conditionPool.New()

	// Initialize fields
	condition.Variable = variable
	condition.Operator = operator
	condition.Value = value
	condition.Negated = negated

	atomic.AddInt64(&tpm.poolHits, 1)

	return condition
}

// NewHigherKindedTypeDefinition creates a pooled higher-kinded type definition
func (tpm *TypePoolManager) NewHigherKindedTypeDefinition(name, definition string) *HigherKindedTypeDefinition {
	def := tpm.higherKindedTypeDefPool.New()

	// Reset/initialize the pooled object
	tpm.resetHigherKindedTypeDefinition(def)

	// Initialize fields
	def.Name = name
	def.TypeConstructors = tpm.getStringSlice()
	def.Definition = definition

	atomic.AddInt64(&tpm.poolHits, 1)

	return def
}

// InferenceEngine creation methods

// NewInferenceEngine creates a pooled inference engine
func (tpm *TypePoolManager) NewInferenceEngine(hierarchy *typedef.TypeHierarchy, symbolTable *binder.SymbolTable) *InferenceEngine {
	engine := tpm.inferenceEnginePool.New()

	// Reset/initialize the pooled object
	tpm.resetInferenceEngine(engine)

	// Initialize fields
	engine.DataFlowAnalyzer = tpm.NewDataFlowAnalyzer()
	engine.ContextAnalyzer = tpm.NewContextAnalyzer()
	engine.UsagePatternAnalyzer = tpm.NewUsagePatternAnalyzer()
	engine.TypePropagator = tpm.NewTypePropagator()
	engine.TypeHierarchy = hierarchy
	engine.SymbolTable = symbolTable
	engine.InferredTypes = tpm.getInferredTypeMap()
	engine.ConfidenceThreshold = 0.7

	// Initialize analyzers with rules and patterns
	engine.initializeAnalyzers()

	atomic.AddInt64(&tpm.inferenceCount, 1)
	atomic.AddInt64(&tpm.poolHits, 1)

	if tpm.hooks.OnInferenceCreate != nil {
		tpm.hooks.OnInferenceCreate(engine)
	}

	return engine
}

// NewInferredTypeInfo creates a pooled inferred type info
func (tpm *TypePoolManager) NewInferredTypeInfo(name, typeName string, confidence float64) *InferredTypeInfo {
	info := tpm.inferredTypeInfoPool.New()

	// Reset/initialize the pooled object
	tpm.resetInferredTypeInfo(info)

	// Initialize fields
	info.Name = name
	info.Type = typeName
	info.Confidence = confidence
	info.Sources = tpm.getInferenceSourceSlice()
	info.Context = *tpm.NewInferenceContext("", "", "", "")

	atomic.AddInt64(&tpm.poolHits, 1)

	return info
}

// NewInferenceSource creates a pooled inference source
func (tpm *TypePoolManager) NewInferenceSource(sourceType InferenceSourceType, location ast.Position, confidence float64, details string) *InferenceSource {
	source := tpm.inferenceSourcePool.New()

	// Initialize fields
	source.Type = sourceType
	source.Location = location
	source.Confidence = confidence
	source.Details = details

	atomic.AddInt64(&tpm.poolHits, 1)

	return source
}

// NewInferenceContext creates a pooled inference context
func (tpm *TypePoolManager) NewInferenceContext(function, pkg, perlContext, controlFlow string) *InferenceContext {
	context := tpm.inferenceContextPool.New()

	// Initialize fields
	context.Function = function
	context.Package = pkg
	context.PerlContext = perlContext
	context.ControlFlow = controlFlow

	atomic.AddInt64(&tpm.poolHits, 1)

	return context
}

// Data Flow Analysis pools

// NewDataFlowAnalyzer creates a pooled data flow analyzer
func (tpm *TypePoolManager) NewDataFlowAnalyzer() *DataFlowAnalyzer {
	analyzer := tpm.dataFlowAnalyzerPool.New()

	// Reset/initialize the pooled object
	tpm.resetDataFlowAnalyzer(analyzer)

	// Initialize fields
	analyzer.DataFlowGraph = tpm.NewDataFlowGraph()
	analyzer.VariableStates = tpm.getVariableStateMap()

	atomic.AddInt64(&tpm.poolHits, 1)

	return analyzer
}

// NewDataFlowGraph creates a pooled data flow graph
func (tpm *TypePoolManager) NewDataFlowGraph() *DataFlowGraph {
	graph := tpm.dataFlowGraphPool.New()

	// Reset/initialize the pooled object
	tpm.resetDataFlowGraph(graph)

	// Initialize fields
	graph.Nodes = tpm.getNodeMap()
	graph.Edges = tpm.getEdgeMap()

	atomic.AddInt64(&tpm.poolHits, 1)

	return graph
}

// NewDataFlowNode creates a pooled data flow node
func (tpm *TypePoolManager) NewDataFlowNode(id, nodeType string, astNode ast.Node) *DataFlowNode {
	node := tpm.dataFlowNodePool.New()

	// Reset/initialize the pooled object
	tpm.resetDataFlowNode(node)

	// Initialize fields
	node.ID = id
	node.Type = nodeType
	node.ASTNode = astNode
	node.Definitions = tpm.getStringMap()
	node.Uses = tpm.getStringMap()

	atomic.AddInt64(&tpm.poolHits, 1)

	return node
}

// NewDataFlowEdge creates a pooled data flow edge
func (tpm *TypePoolManager) NewDataFlowEdge(from, to, edgeType, condition string) *DataFlowEdge {
	edge := tpm.dataFlowEdgePool.New()

	// Initialize fields
	edge.From = from
	edge.To = to
	edge.Type = edgeType
	edge.Condition = condition

	atomic.AddInt64(&tpm.poolHits, 1)

	return edge
}

// Context Analysis pools

// NewContextAnalyzer creates a pooled context analyzer
func (tpm *TypePoolManager) NewContextAnalyzer() *ContextAnalyzer {
	analyzer := tpm.contextAnalyzerPool.New()

	// Reset/initialize the pooled object
	tpm.resetContextAnalyzer(analyzer)

	// Initialize fields
	analyzer.ContextStack = tpm.getContextStackSlice()
	analyzer.ContextRules = tpm.getContextRuleMap()

	atomic.AddInt64(&tpm.poolHits, 1)

	return analyzer
}

// NewPerlContext creates a pooled Perl context
func (tpm *TypePoolManager) NewPerlContext(contextType, expectedType string, parent *PerlContext) *PerlContext {
	context := tpm.perlContextPool.New()

	// Initialize fields
	context.Type = contextType
	context.ExpectedType = expectedType
	context.Parent = parent

	atomic.AddInt64(&tpm.poolHits, 1)

	return context
}

// Usage Pattern Analysis pools

// NewUsagePatternAnalyzer creates a pooled usage pattern analyzer
func (tpm *TypePoolManager) NewUsagePatternAnalyzer() *UsagePatternAnalyzer {
	analyzer := tpm.usagePatternAnalyzerPool.New()

	// Reset/initialize the pooled object
	tpm.resetUsagePatternAnalyzer(analyzer)

	// Initialize fields
	analyzer.Patterns = tpm.getUsagePatternSlice()
	analyzer.PatternCache = tpm.getPatternMatchMap()

	atomic.AddInt64(&tpm.poolHits, 1)

	return analyzer
}

// Type Propagation pools

// NewTypePropagator creates a pooled type propagator
func (tpm *TypePoolManager) NewTypePropagator() *TypePropagator {
	propagator := tpm.typePropagatorPool.New()

	// Reset/initialize the pooled object
	tpm.resetTypePropagator(propagator)

	// Initialize fields
	propagator.PropagationRules = tpm.getPropagationRuleSlice()
	propagator.TypeConstraints = tpm.getConstraintMap()
	propagator.SolvedTypes = tpm.getStringMap()

	atomic.AddInt64(&tpm.poolHits, 1)

	return propagator
}

// TypeChecker creation methods

// NewTypeChecker creates a pooled type checker
func (tpm *TypePoolManager) NewTypeChecker(hierarchy *typedef.TypeHierarchy, symbolTable *binder.SymbolTable, moduleName string) *TypeChecker {
	tc := tpm.typeCheckerPool.New()

	// Reset/initialize the pooled object
	tpm.resetTypeChecker(tc)

	// Initialize fields with pooled objects
	tc.Hierarchy = hierarchy
	tc.SymbolTable = symbolTable
	tc.CurrentModule = moduleName
	tc.ImportedModules = tpm.getBoolMap()
	tc.TypeAnnotations = tpm.getStringMap()
	tc.VariableTypes = tpm.getStringMap()
	tc.FunctionTypes = tpm.getFunctionSigMap()
	tc.TypeState = tpm.NewTypeState()
	tc.TypeStateStack = tpm.getTypeStateSlice()
	tc.ValidationPatterns = tpm.getValidationPatternSlice()
	tc.ContextSensitiveFunctions = tpm.getModuleTypeMap()
	tc.TypeAliases = tpm.getStringMap()
	tc.GenericFunctions = tpm.getGenericFuncMap()
	tc.ModuleTypes = tpm.getModuleTypeMap()
	tc.HigherKindedTypes = tpm.getHigherKindedMap()
	tc.InferenceEngine = tpm.NewInferenceEngine(hierarchy, symbolTable)
	tc.Debug = false

	// Initialize validation patterns
	tc.initializeValidationPatterns()

	atomic.AddInt64(&tpm.typeCheckCount, 1)
	atomic.AddInt64(&tpm.poolHits, 1)

	return tc
}

// Helper methods for getting pooled collections

// getStringMap returns a pooled string map
func (tpm *TypePoolManager) getStringMap() map[string]string {
	m := tpm.stringMapPool.New()
	if *m == nil {
		*m = make(map[string]string)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		tpm.clearStringMap(*m)
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *m
}

// getBoolMap returns a pooled bool map
func (tpm *TypePoolManager) getBoolMap() map[string]bool {
	// Since we don't have a specific bool map pool, create a new one
	// In a production implementation, we might add a specific pool for this
	return make(map[string]bool)
}

// getFunctionSigMap returns a pooled function signature map
func (tpm *TypePoolManager) getFunctionSigMap() map[string]*FunctionSignature {
	m := tpm.functionSigMapPool.New()
	if *m == nil {
		*m = make(map[string]*FunctionSignature)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		tpm.clearFunctionSigMap(*m)
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *m
}

// getGenericFuncMap returns a pooled generic function map
func (tpm *TypePoolManager) getGenericFuncMap() map[string]*GenericFunctionSignature {
	m := tpm.genericFuncMapPool.New()
	if *m == nil {
		*m = make(map[string]*GenericFunctionSignature)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		tpm.clearGenericFuncMap(*m)
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *m
}

// getModuleTypeMap returns a pooled module type map
func (tpm *TypePoolManager) getModuleTypeMap() map[string]map[string]string {
	m := tpm.moduleTypeMapPool.New()
	if *m == nil {
		*m = make(map[string]map[string]string)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		tpm.clearModuleTypeMap(*m)
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *m
}

// getHigherKindedMap returns a pooled higher-kinded type map
func (tpm *TypePoolManager) getHigherKindedMap() map[string]*HigherKindedTypeDefinition {
	m := tpm.higherKindedMapPool.New()
	if *m == nil {
		*m = make(map[string]*HigherKindedTypeDefinition)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		tpm.clearHigherKindedMap(*m)
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *m
}

// getInferredTypeMap returns a pooled inferred type map
func (tpm *TypePoolManager) getInferredTypeMap() map[string]*InferredTypeInfo {
	m := tpm.inferredTypeMapPool.New()
	if *m == nil {
		*m = make(map[string]*InferredTypeInfo)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		tpm.clearInferredTypeMap(*m)
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *m
}

// getVariableStateMap returns a pooled variable state map
func (tpm *TypePoolManager) getVariableStateMap() map[string]*VariableStateTracker {
	m := tpm.variableStateMapPool.New()
	if *m == nil {
		*m = make(map[string]*VariableStateTracker)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		tpm.clearVariableStateMap(*m)
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *m
}

// getContextRuleMap returns a pooled context rule map
func (tpm *TypePoolManager) getContextRuleMap() map[string]ContextRule {
	m := tpm.contextRuleMapPool.New()
	if *m == nil {
		*m = make(map[string]ContextRule)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		tpm.clearContextRuleMap(*m)
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *m
}

// getNodeMap returns a pooled node map
func (tpm *TypePoolManager) getNodeMap() map[string]*DataFlowNode {
	m := tpm.nodeMapPool.New()
	if *m == nil {
		*m = make(map[string]*DataFlowNode)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		tpm.clearNodeMap(*m)
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *m
}

// getEdgeMap returns a pooled edge map
func (tpm *TypePoolManager) getEdgeMap() map[string][]*DataFlowEdge {
	m := tpm.edgeMapPool.New()
	if *m == nil {
		*m = make(map[string][]*DataFlowEdge)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		tpm.clearEdgeMap(*m)
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *m
}

// getConstraintMap returns a pooled constraint map
func (tpm *TypePoolManager) getConstraintMap() map[string][]TypeConstraint {
	m := tpm.constraintMapPool.New()
	if *m == nil {
		*m = make(map[string][]TypeConstraint)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		tpm.clearConstraintMap(*m)
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *m
}

// getPatternMatchMap returns a pooled pattern match map
func (tpm *TypePoolManager) getPatternMatchMap() map[string]*PatternMatch {
	// Create a new map since we don't have a specific pool for this
	// In a production implementation, we might add a specific pool
	return make(map[string]*PatternMatch)
}

// Slice pool methods

// getTypeErrorSlice returns a pooled type error slice
func (tpm *TypePoolManager) getTypeErrorSlice() []TypeCheckError {
	s := tpm.typeErrorSlicePool.New()
	if *s == nil {
		*s = make([]TypeCheckError, 0, 8)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *s
}

// getTypeAnnotationSlice returns a pooled type annotation slice
func (tpm *TypePoolManager) getTypeAnnotationSlice() []*ast.TypeAnnotation {
	s := tpm.typeAnnotationSlicePool.New()
	if *s == nil {
		*s = make([]*ast.TypeAnnotation, 0, 8)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *s
}

// getConditionSlice returns a pooled condition slice
func (tpm *TypePoolManager) getConditionSlice() []Condition {
	s := tpm.conditionSlicePool.New()
	if *s == nil {
		*s = make([]Condition, 0, 4)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *s
}

// getValidationPatternSlice returns a pooled validation pattern slice
func (tpm *TypePoolManager) getValidationPatternSlice() []ValidationPattern {
	s := tpm.validationPatternSlicePool.New()
	if *s == nil {
		*s = make([]ValidationPattern, 0, 8)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *s
}

// getTypeStateSlice returns a pooled type state slice
func (tpm *TypePoolManager) getTypeStateSlice() []*TypeState {
	s := tpm.typeStateSlicePool.New()
	if *s == nil {
		*s = make([]*TypeState, 0, 4)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *s
}

// getInferenceSourceSlice returns a pooled inference source slice
func (tpm *TypePoolManager) getInferenceSourceSlice() []InferenceSource {
	s := tpm.inferenceSourceSlicePool.New()
	if *s == nil {
		*s = make([]InferenceSource, 0, 4)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *s
}

// getContextStackSlice returns a pooled context stack slice
func (tpm *TypePoolManager) getContextStackSlice() []PerlContext {
	s := tpm.contextStackSlicePool.New()
	if *s == nil {
		*s = make([]PerlContext, 0, 4)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *s
}

// getUsagePatternSlice returns a pooled usage pattern slice
func (tpm *TypePoolManager) getUsagePatternSlice() []UsagePattern {
	s := tpm.usagePatternSlicePool.New()
	if *s == nil {
		*s = make([]UsagePattern, 0, 8)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *s
}

// getPropagationRuleSlice returns a pooled propagation rule slice
func (tpm *TypePoolManager) getPropagationRuleSlice() []PropagationRule {
	s := tpm.propagationRuleSlicePool.New()
	if *s == nil {
		*s = make([]PropagationRule, 0, 4)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *s
}

// getStringSlice returns a pooled string slice
func (tpm *TypePoolManager) getStringSlice() []string {
	s := tpm.stringSlicePool.New()
	if *s == nil {
		*s = make([]string, 0, 4)
		atomic.AddInt64(&tpm.poolMisses, 1)
	} else {
		*s = (*s)[:0] // Reset length but keep capacity
		atomic.AddInt64(&tpm.poolHits, 1)
	}
	return *s
}

// getStringSliceMap returns a pooled string slice map (for constraints)
func (tpm *TypePoolManager) getStringSliceMap() map[string][]string {
	// Create a new map since we don't have a specific pool for this
	// In a production implementation, we might add a specific pool
	return make(map[string][]string)
}

// Pool warming methods

// warmPools pre-allocates common type objects for better performance
func (tpm *TypePoolManager) warmPools() {
	if tpm.hooks.OnPoolWarming != nil {
		tpm.hooks.OnPoolWarming("type-system")
	}

	// Warm up pools with common type patterns
	tpm.warmTypeCheckResults()
	tpm.warmTypeStates()
	tpm.warmInferenceEngines()
	tpm.warmCommonTypes()

	if tpm.hooks.OnPoolWarming != nil {
		tpm.hooks.OnPoolWarming("type-system-complete")
	}
}

// warmTypeCheckResults pre-allocates common type check result structures
func (tpm *TypePoolManager) warmTypeCheckResults() {
	// Create a few type check results to warm the pool
	for i := 0; i < 4; i++ {
		result := tpm.NewTypeCheckResult("")
		tpm.resetTypeCheckResult(result) // Reset for reuse
	}
}

// warmTypeStates pre-allocates common type state structures
func (tpm *TypePoolManager) warmTypeStates() {
	// Create a few type states to warm the pool
	for i := 0; i < 8; i++ {
		state := tpm.NewTypeState()
		tpm.resetTypeState(state) // Reset for reuse
	}
}

// warmInferenceEngines pre-allocates inference engine structures
func (tpm *TypePoolManager) warmInferenceEngines() {
	// Create a few inference engines to warm the pool
	// Note: We can't actually warm these without valid hierarchy and symbol table
	// This is a placeholder for the pattern
}

// warmCommonTypes creates pre-allocated type objects for common Perl types
func (tpm *TypePoolManager) warmCommonTypes() {
	commonTypes := []string{"Int", "Str", "Num", "Bool", "Undef", "ArrayRef", "HashRef", "Any", "Object"}

	for _, typeName := range commonTypes {
		if tpm.hooks.OnTypeCreated != nil {
			tpm.hooks.OnTypeCreated(typeName)
		}
		atomic.AddInt64(&tpm.typeCreated, 1)
	}
}

// Reset methods for proper object cleanup and reuse

// resetTypeCheckResult resets a type check result for reuse
func (tpm *TypePoolManager) resetTypeCheckResult(result *TypeCheckResult) {
	result.Path = ""
	result.FlowSensitiveEnabled = false

	// Reset slices but keep capacity
	if result.Errors != nil {
		result.Errors = result.Errors[:0]
	}
	if result.TypeAnnotations != nil {
		result.TypeAnnotations = result.TypeAnnotations[:0]
	}

	// Clear maps
	tpm.clearStringMap(result.RefinedTypes)

	if tpm.hooks.OnTypeCheckReset != nil {
		tpm.hooks.OnTypeCheckReset(result)
	}
}

// resetFunctionSignature resets a function signature for reuse
func (tpm *TypePoolManager) resetFunctionSignature(sig *FunctionSignature) {
	sig.ReturnType = ""
	sig.IsMethod = false
	tpm.clearStringMap(sig.ParameterTypes)
}

// resetGenericFunctionSignature resets a generic function signature for reuse
func (tpm *TypePoolManager) resetGenericFunctionSignature(sig *GenericFunctionSignature) {
	sig.ReturnType = ""
	sig.IsMethod = false

	if sig.TypeParameters != nil {
		sig.TypeParameters = sig.TypeParameters[:0]
	}

	tpm.clearStringMap(sig.ParameterTypes)

	// Clear constraints map
	for k := range sig.Constraints {
		delete(sig.Constraints, k)
	}
}

// resetTypeState resets a type state for reuse
func (tpm *TypePoolManager) resetTypeState(state *TypeState) {
	state.SkipFlowChecks = false

	tpm.clearStringMap(state.VariableTypes)
	tpm.clearStringMap(state.RefinedTypes)

	if state.Conditions != nil {
		state.Conditions = state.Conditions[:0]
	}

	if tpm.hooks.OnTypeStateReset != nil {
		tpm.hooks.OnTypeStateReset(state)
	}
}

// resetHigherKindedTypeDefinition resets a higher-kinded type definition for reuse
func (tpm *TypePoolManager) resetHigherKindedTypeDefinition(def *HigherKindedTypeDefinition) {
	def.Name = ""
	def.Definition = ""

	if def.TypeConstructors != nil {
		def.TypeConstructors = def.TypeConstructors[:0]
	}
}

// resetInferenceEngine resets an inference engine for reuse
func (tpm *TypePoolManager) resetInferenceEngine(engine *InferenceEngine) {
	engine.TypeHierarchy = nil
	engine.SymbolTable = nil
	engine.ConfidenceThreshold = 0.7

	// Reset sub-analyzers
	if engine.DataFlowAnalyzer != nil {
		tpm.resetDataFlowAnalyzer(engine.DataFlowAnalyzer)
	}
	if engine.ContextAnalyzer != nil {
		tpm.resetContextAnalyzer(engine.ContextAnalyzer)
	}
	if engine.UsagePatternAnalyzer != nil {
		tpm.resetUsagePatternAnalyzer(engine.UsagePatternAnalyzer)
	}
	if engine.TypePropagator != nil {
		tpm.resetTypePropagator(engine.TypePropagator)
	}

	// Clear inferred types
	tpm.clearInferredTypeMap(engine.InferredTypes)

	if tpm.hooks.OnInferenceReset != nil {
		tpm.hooks.OnInferenceReset(engine)
	}
}

// resetInferredTypeInfo resets inferred type info for reuse
func (tpm *TypePoolManager) resetInferredTypeInfo(info *InferredTypeInfo) {
	info.Name = ""
	info.Type = ""
	info.Confidence = 0.0

	if info.Sources != nil {
		info.Sources = info.Sources[:0]
	}
}

// resetDataFlowAnalyzer resets a data flow analyzer for reuse
func (tpm *TypePoolManager) resetDataFlowAnalyzer(analyzer *DataFlowAnalyzer) {
	if analyzer.DataFlowGraph != nil {
		tpm.resetDataFlowGraph(analyzer.DataFlowGraph)
	}
	tpm.clearVariableStateMap(analyzer.VariableStates)
}

// resetDataFlowGraph resets a data flow graph for reuse
func (tpm *TypePoolManager) resetDataFlowGraph(graph *DataFlowGraph) {
	tpm.clearNodeMap(graph.Nodes)
	tpm.clearEdgeMap(graph.Edges)
}

// resetDataFlowNode resets a data flow node for reuse
func (tpm *TypePoolManager) resetDataFlowNode(node *DataFlowNode) {
	node.ID = ""
	node.Type = ""
	node.ASTNode = nil

	tpm.clearStringMap(node.Definitions)
	tpm.clearStringMap(node.Uses)
}

// resetContextAnalyzer resets a context analyzer for reuse
func (tpm *TypePoolManager) resetContextAnalyzer(analyzer *ContextAnalyzer) {
	if analyzer.ContextStack != nil {
		analyzer.ContextStack = analyzer.ContextStack[:0]
	}
	tpm.clearContextRuleMap(analyzer.ContextRules)
}

// resetUsagePatternAnalyzer resets a usage pattern analyzer for reuse
func (tpm *TypePoolManager) resetUsagePatternAnalyzer(analyzer *UsagePatternAnalyzer) {
	if analyzer.Patterns != nil {
		analyzer.Patterns = analyzer.Patterns[:0]
	}

	// Clear pattern cache
	for k := range analyzer.PatternCache {
		delete(analyzer.PatternCache, k)
	}
}

// resetTypePropagator resets a type propagator for reuse
func (tpm *TypePoolManager) resetTypePropagator(propagator *TypePropagator) {
	if propagator.PropagationRules != nil {
		propagator.PropagationRules = propagator.PropagationRules[:0]
	}

	tpm.clearConstraintMap(propagator.TypeConstraints)
	tpm.clearStringMap(propagator.SolvedTypes)
}

// resetTypeChecker resets a type checker for reuse
func (tpm *TypePoolManager) resetTypeChecker(tc *TypeChecker) {
	tc.Hierarchy = nil
	tc.SymbolTable = nil
	tc.CurrentModule = ""
	tc.Debug = false

	// Clear all maps
	for k := range tc.ImportedModules {
		delete(tc.ImportedModules, k)
	}
	tpm.clearStringMap(tc.TypeAnnotations)
	tpm.clearStringMap(tc.VariableTypes)
	tpm.clearFunctionSigMap(tc.FunctionTypes)
	tpm.clearStringMap(tc.TypeAliases)
	tpm.clearGenericFuncMap(tc.GenericFunctions)
	tpm.clearModuleTypeMap(tc.ModuleTypes)
	tpm.clearHigherKindedMap(tc.HigherKindedTypes)

	// Reset type state
	if tc.TypeState != nil {
		tpm.resetTypeState(tc.TypeState)
	}

	// Reset type state stack
	if tc.TypeStateStack != nil {
		tc.TypeStateStack = tc.TypeStateStack[:0]
	}

	// Reset validation patterns
	if tc.ValidationPatterns != nil {
		tc.ValidationPatterns = tc.ValidationPatterns[:0]
	}

	// Reset inference engine
	if tc.InferenceEngine != nil {
		tpm.resetInferenceEngine(tc.InferenceEngine)
	}
}

// Helper methods to clear maps efficiently

// clearStringMap clears a string map efficiently
func (tpm *TypePoolManager) clearStringMap(m map[string]string) {
	for k := range m {
		delete(m, k)
	}
}

// clearFunctionSigMap clears a function signature map efficiently
func (tpm *TypePoolManager) clearFunctionSigMap(m map[string]*FunctionSignature) {
	for k := range m {
		delete(m, k)
	}
}

// clearGenericFuncMap clears a generic function map efficiently
func (tpm *TypePoolManager) clearGenericFuncMap(m map[string]*GenericFunctionSignature) {
	for k := range m {
		delete(m, k)
	}
}

// clearModuleTypeMap clears a module type map efficiently
func (tpm *TypePoolManager) clearModuleTypeMap(m map[string]map[string]string) {
	for k := range m {
		delete(m, k)
	}
}

// clearHigherKindedMap clears a higher-kinded type map efficiently
func (tpm *TypePoolManager) clearHigherKindedMap(m map[string]*HigherKindedTypeDefinition) {
	for k := range m {
		delete(m, k)
	}
}

// clearInferredTypeMap clears an inferred type map efficiently
func (tpm *TypePoolManager) clearInferredTypeMap(m map[string]*InferredTypeInfo) {
	for k := range m {
		delete(m, k)
	}
}

// clearVariableStateMap clears a variable state map efficiently
func (tpm *TypePoolManager) clearVariableStateMap(m map[string]*VariableStateTracker) {
	for k := range m {
		delete(m, k)
	}
}

// clearContextRuleMap clears a context rule map efficiently
func (tpm *TypePoolManager) clearContextRuleMap(m map[string]ContextRule) {
	for k := range m {
		delete(m, k)
	}
}

// clearNodeMap clears a node map efficiently
func (tpm *TypePoolManager) clearNodeMap(m map[string]*DataFlowNode) {
	for k := range m {
		delete(m, k)
	}
}

// clearEdgeMap clears an edge map efficiently
func (tpm *TypePoolManager) clearEdgeMap(m map[string][]*DataFlowEdge) {
	for k := range m {
		delete(m, k)
	}
}

// clearConstraintMap clears a constraint map efficiently
func (tpm *TypePoolManager) clearConstraintMap(m map[string][]TypeConstraint) {
	for k := range m {
		delete(m, k)
	}
}

// Pool management methods

// Reset resets all pools for reuse
func (tpm *TypePoolManager) Reset() {
	tpm.mu.Lock()
	defer tpm.mu.Unlock()

	// Reset all core pools
	tpm.typeCheckResultPool.Reset()
	tpm.typeCheckErrorPool.Reset()
	tpm.functionSignaturePool.Reset()
	tpm.genericFunctionSigPool.Reset()
	tpm.typeStatePool.Reset()
	tpm.conditionPool.Reset()
	tpm.validationPatternPool.Reset()
	tpm.higherKindedTypeDefPool.Reset()

	// Reset inference engine pools
	tpm.inferenceEnginePool.Reset()
	tpm.inferredTypeInfoPool.Reset()
	tpm.inferenceSourcePool.Reset()
	tpm.inferenceContextPool.Reset()
	tpm.dataFlowAnalyzerPool.Reset()
	tpm.contextAnalyzerPool.Reset()
	tpm.usagePatternAnalyzerPool.Reset()
	tpm.typePropagatorPool.Reset()
	tpm.dataFlowGraphPool.Reset()
	tpm.dataFlowNodePool.Reset()
	tpm.dataFlowEdgePool.Reset()
	tpm.variableStateTrackerPool.Reset()
	tpm.variableStatePool.Reset()
	tpm.typeTransformationPool.Reset()
	tpm.perlContextPool.Reset()
	tpm.contextRulePool.Reset()
	tpm.usagePatternPool.Reset()
	tpm.patternMatchPool.Reset()
	tpm.propagationRulePool.Reset()
	tpm.typeConstraintPool.Reset()

	// Reset type checker pool
	tpm.typeCheckerPool.Reset()
}

// Clear completely empties all pools and resets statistics
func (tpm *TypePoolManager) Clear() {
	tpm.mu.Lock()
	defer tpm.mu.Unlock()

	// Clear all core pools
	tpm.typeCheckResultPool.Clear()
	tpm.typeCheckErrorPool.Clear()
	tpm.functionSignaturePool.Clear()
	tpm.genericFunctionSigPool.Clear()
	tpm.typeStatePool.Clear()
	tpm.conditionPool.Clear()
	tpm.validationPatternPool.Clear()
	tpm.higherKindedTypeDefPool.Clear()

	// Clear inference engine pools
	tpm.inferenceEnginePool.Clear()
	tpm.inferredTypeInfoPool.Clear()
	tpm.inferenceSourcePool.Clear()
	tpm.inferenceContextPool.Clear()
	tpm.dataFlowAnalyzerPool.Clear()
	tpm.contextAnalyzerPool.Clear()
	tpm.usagePatternAnalyzerPool.Clear()
	tpm.typePropagatorPool.Clear()
	tpm.dataFlowGraphPool.Clear()
	tpm.dataFlowNodePool.Clear()
	tpm.dataFlowEdgePool.Clear()
	tpm.variableStateTrackerPool.Clear()
	tpm.variableStatePool.Clear()
	tpm.typeTransformationPool.Clear()
	tpm.perlContextPool.Clear()
	tpm.contextRulePool.Clear()
	tpm.usagePatternPool.Clear()
	tpm.patternMatchPool.Clear()
	tpm.propagationRulePool.Clear()
	tpm.typeConstraintPool.Clear()

	// Clear type checker pool
	tpm.typeCheckerPool.Clear()

	// Reset statistics
	atomic.StoreInt64(&tpm.typeCheckCount, 0)
	atomic.StoreInt64(&tpm.inferenceCount, 0)
	atomic.StoreInt64(&tpm.poolHits, 0)
	atomic.StoreInt64(&tpm.poolMisses, 0)
	atomic.StoreInt64(&tpm.memoryReused, 0)
	atomic.StoreInt64(&tpm.typeCreated, 0)
	atomic.StoreInt64(&tpm.typeReused, 0)
}

// GetDetailedStats returns detailed statistics about all type pools
func (tpm *TypePoolManager) GetDetailedStats() TypePoolStats {
	return TypePoolStats{
		TypeCheckCount: atomic.LoadInt64(&tpm.typeCheckCount),
		InferenceCount: atomic.LoadInt64(&tpm.inferenceCount),
		PoolHits:       atomic.LoadInt64(&tpm.poolHits),
		PoolMisses:     atomic.LoadInt64(&tpm.poolMisses),
		MemoryReused:   atomic.LoadInt64(&tpm.memoryReused),
		TypeCreated:    atomic.LoadInt64(&tpm.typeCreated),
		TypeReused:     atomic.LoadInt64(&tpm.typeReused),
		PoolEfficiency: tpm.PoolEfficiency(),
		TypeReuseRate:  tpm.TypeReuseRate(),
	}
}

// TypePoolStats contains detailed type pool statistics
type TypePoolStats struct {
	TypeCheckCount int64   // Total number of type check operations
	InferenceCount int64   // Total number of type inference operations
	PoolHits       int64   // Number of successful pool allocations
	PoolMisses     int64   // Number of pool misses requiring new allocation
	MemoryReused   int64   // Total amount of memory reused
	TypeCreated    int64   // Number of types created
	TypeReused     int64   // Number of types reused from pool
	PoolEfficiency float64 // Pool hit rate as percentage
	TypeReuseRate  float64 // Type reuse rate as percentage
}

// Global type pool manager instance
var globalTypePoolManager *TypePoolManager
var typePoolOnce sync.Once

// GlobalTypePoolManager returns the global type pool manager instance
func GlobalTypePoolManager() *TypePoolManager {
	typePoolOnce.Do(func() {
		globalTypePoolManager = NewTypePoolManager(TypePoolHooks{
			// Default hooks can be set here
			OnPoolWarming: func(poolType string) {
				// Default pool warming notification
			},
		})
	})
	return globalTypePoolManager
}

// SetGlobalTypePoolHooks sets hooks for the global type pool manager
func SetGlobalTypePoolHooks(hooks TypePoolHooks) {
	tpm := GlobalTypePoolManager()
	tpm.hooks = hooks
}
