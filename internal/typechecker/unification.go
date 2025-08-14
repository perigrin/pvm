// ABOUTME: Implements type unification algorithm for generic type inference
// ABOUTME: Based on Hindley-Milner type system with Perl-specific extensions

package typechecker

import (
	"fmt"

	"tamarou.com/pvm/internal/typedef"
)

// UnificationContext tracks the state of type unification
type UnificationContext struct {
	// Substitutions maps type variables to their unified types
	Substitutions map[string]typedef.Type
	// Constraints that must be satisfied
	Constraints []UnificationConstraint
	// Track occurs check to prevent infinite types
	occursCheck map[string]bool
}

// UnificationConstraint represents a constraint between two types
type UnificationConstraint struct {
	Left  typedef.Type
	Right typedef.Type
	Kind  ConstraintKind
}

// ConstraintKind represents the type of constraint
type ConstraintKind int

const (
	// Equality constraint - types must be equal
	EqualityConstraint ConstraintKind = iota
	// Subtype constraint - left must be subtype of right
	SubtypeConstraint
	// Protocol constraint - left must implement protocol
	ProtocolConstraint
)

// TypeUnifier implements the unification algorithm
type TypeUnifier struct {
	hierarchy            *typedef.TypeHierarchy
	constraintPropagator *ConstraintPropagator
}

// NewTypeUnifier creates a new type unifier
func NewTypeUnifier(hierarchy *typedef.TypeHierarchy) *TypeUnifier {
	return &TypeUnifier{
		hierarchy:            hierarchy,
		constraintPropagator: NewConstraintPropagator(hierarchy),
	}
}

// Unify attempts to unify two types, returning substitutions if successful
func (u *TypeUnifier) Unify(t1, t2 typedef.Type) (map[string]typedef.Type, error) {
	ctx := &UnificationContext{
		Substitutions: make(map[string]typedef.Type),
		Constraints:   []UnificationConstraint{},
		occursCheck:   make(map[string]bool),
	}

	// Add initial constraint
	ctx.Constraints = append(ctx.Constraints, UnificationConstraint{
		Left:  t1,
		Right: t2,
		Kind:  EqualityConstraint,
	})

	// Solve constraints
	if err := u.solveConstraints(ctx); err != nil {
		return nil, err
	}

	return ctx.Substitutions, nil
}

// solveConstraints solves all constraints in the context
func (u *TypeUnifier) solveConstraints(ctx *UnificationContext) error {
	// Process constraints until fixed point
	for len(ctx.Constraints) > 0 {
		// Pop constraint
		constraint := ctx.Constraints[0]
		ctx.Constraints = ctx.Constraints[1:]

		// Apply current substitutions
		left := u.applySubstitutions(constraint.Left, ctx.Substitutions)
		right := u.applySubstitutions(constraint.Right, ctx.Substitutions)

		// Solve based on constraint kind
		switch constraint.Kind {
		case EqualityConstraint:
			if err := u.unifyTypes(left, right, ctx); err != nil {
				return err
			}
		case SubtypeConstraint:
			if err := u.checkSubtype(left, right, ctx); err != nil {
				return err
			}
		case ProtocolConstraint:
			if err := u.checkProtocol(left, right, ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

// unifyTypes unifies two types for equality
func (u *TypeUnifier) unifyTypes(t1, t2 typedef.Type, ctx *UnificationContext) error {
	// Handle type variables
	if typeVar, ok := t1.(*TypeVariable); ok {
		return u.unifyVariable(typeVar.Name, t2, ctx)
	}
	if typeVar, ok := t2.(*TypeVariable); ok {
		return u.unifyVariable(typeVar.Name, t1, ctx)
	}

	// Handle simple types
	if simple1, ok := t1.(*typedef.SimpleType); ok {
		if simple2, ok := t2.(*typedef.SimpleType); ok {
			if simple1.Name == simple2.Name {
				return nil // Already unified
			}
			// Check if types are compatible through hierarchy
			if u.hierarchy.IsSubtypeOf(simple1.Name, simple2.Name) ||
				u.hierarchy.IsSubtypeOf(simple2.Name, simple1.Name) {
				return nil
			}
			return fmt.Errorf("cannot unify %s with %s", simple1.Name, simple2.Name)
		}
	}

	// Handle parameterized types
	if param1, ok := t1.(*typedef.ParameterizedType); ok {
		if param2, ok := t2.(*typedef.ParameterizedType); ok {
			// Unify base types
			ctx.Constraints = append(ctx.Constraints, UnificationConstraint{
				Left:  param1.BaseType,
				Right: param2.BaseType,
				Kind:  EqualityConstraint,
			})

			// Unify arguments
			if len(param1.Arguments) != len(param2.Arguments) {
				return fmt.Errorf("parameter count mismatch: %d vs %d",
					len(param1.Arguments), len(param2.Arguments))
			}

			for i := range param1.Arguments {
				ctx.Constraints = append(ctx.Constraints, UnificationConstraint{
					Left:  param1.Arguments[i],
					Right: param2.Arguments[i],
					Kind:  EqualityConstraint,
				})
			}

			return nil
		}
	}

	// Handle generic types
	if gen1, ok := t1.(*typedef.GenericType); ok {
		if gen2, ok := t2.(*typedef.GenericType); ok {
			return u.unifyGenerics(gen1, gen2, ctx)
		}
		// Try to instantiate generic with concrete type
		return u.instantiateGeneric(gen1, t2, ctx)
	}
	if gen2, ok := t2.(*typedef.GenericType); ok {
		return u.instantiateGeneric(gen2, t1, ctx)
	}

	return fmt.Errorf("cannot unify %T with %T", t1, t2)
}

// unifyVariable unifies a type variable with a type
func (u *TypeUnifier) unifyVariable(varName string, typ typedef.Type, ctx *UnificationContext) error {
	// Check if already bound
	if existing, ok := ctx.Substitutions[varName]; ok {
		// Unify with existing binding
		ctx.Constraints = append(ctx.Constraints, UnificationConstraint{
			Left:  existing,
			Right: typ,
			Kind:  EqualityConstraint,
		})
		return nil
	}

	// Occurs check - prevent infinite types like T = List[T]
	if u.occursIn(varName, typ, ctx) {
		return fmt.Errorf("occurs check failed: %s occurs in %v", varName, typ)
	}

	// Add substitution
	ctx.Substitutions[varName] = typ
	return nil
}

// occursIn checks if a type variable occurs in a type
func (u *TypeUnifier) occursIn(varName string, typ typedef.Type, ctx *UnificationContext) bool {
	// Apply substitutions first
	typ = u.applySubstitutions(typ, ctx.Substitutions)

	switch t := typ.(type) {
	case *TypeVariable:
		return t.Name == varName
	case *typedef.ParameterizedType:
		for _, arg := range t.Arguments {
			if u.occursIn(varName, arg, ctx) {
				return true
			}
		}
	case *typedef.GenericType:
		// Check in constraints
		for _, constraint := range t.Constraints {
			if u.occursInConstraint(varName, constraint, ctx) {
				return true
			}
		}
	}

	return false
}

// occursInConstraint checks if a variable occurs in a constraint
func (u *TypeUnifier) occursInConstraint(varName string, constraint typedef.TypeConstraint, ctx *UnificationContext) bool {
	// Check if the constraint expression contains the type variable
	if constraint.Expression != nil {
		return u.occursIn(varName, constraint.Expression, ctx)
	}
	return false
}

// applySubstitutions applies all current substitutions to a type
func (u *TypeUnifier) applySubstitutions(typ typedef.Type, subs map[string]typedef.Type) typedef.Type {
	switch t := typ.(type) {
	case *TypeVariable:
		if sub, ok := subs[t.Name]; ok {
			// Recursively apply substitutions
			return u.applySubstitutions(sub, subs)
		}
		return t
	case *typedef.ParameterizedType:
		// Apply to arguments
		newArgs := make([]typedef.Type, len(t.Arguments))
		for i, arg := range t.Arguments {
			newArgs[i] = u.applySubstitutions(arg, subs)
		}
		return &typedef.ParameterizedType{
			BaseType:  u.applySubstitutions(t.BaseType, subs),
			Arguments: newArgs,
		}
	default:
		return typ
	}
}

// unifyGenerics unifies two generic types
func (u *TypeUnifier) unifyGenerics(g1, g2 *typedef.GenericType, ctx *UnificationContext) error {
	if g1.Name != g2.Name {
		return fmt.Errorf("cannot unify generics %s and %s", g1.Name, g2.Name)
	}

	if len(g1.TypeParameters) != len(g2.TypeParameters) {
		return fmt.Errorf("type parameter count mismatch")
	}

	// Create mapping between type parameters
	paramMap := make(map[string]string)
	for i := range g1.TypeParameters {
		paramMap[g1.TypeParameters[i].Name] = g2.TypeParameters[i].Name
	}

	// Unify constraints
	if err := u.unifyConstraints(g1.Constraints, g2.Constraints, paramMap, ctx); err != nil {
		return fmt.Errorf("constraint unification failed: %w", err)
	}

	return nil
}

// instantiateGeneric attempts to instantiate a generic type with a concrete type
func (u *TypeUnifier) instantiateGeneric(generic *typedef.GenericType, concrete typedef.Type, ctx *UnificationContext) error {
	// If the concrete type is a parameterized type, try to match it against the generic pattern
	if paramType, ok := concrete.(*typedef.ParameterizedType); ok {
		return u.matchGenericPattern(generic, paramType, ctx)
	}

	// If the concrete type is a simple type, check if it matches the generic base type
	if simpleType, ok := concrete.(*typedef.SimpleType); ok {
		return u.matchGenericWithSimple(generic, simpleType, ctx)
	}

	return fmt.Errorf("cannot instantiate generic %s with %T", generic.Name, concrete)
}

// checkSubtype checks if left is a subtype of right
func (u *TypeUnifier) checkSubtype(left, right typedef.Type, ctx *UnificationContext) error {
	// Use type hierarchy to check subtyping
	if simple1, ok := left.(*typedef.SimpleType); ok {
		if simple2, ok := right.(*typedef.SimpleType); ok {
			if u.hierarchy.IsSubtypeOf(simple1.Name, simple2.Name) {
				return nil
			}
			return fmt.Errorf("%s is not a subtype of %s", simple1.Name, simple2.Name)
		}
	}

	// For other types, fall back to equality for now
	ctx.Constraints = append(ctx.Constraints, UnificationConstraint{
		Left:  left,
		Right: right,
		Kind:  EqualityConstraint,
	})

	return nil
}

// checkProtocol checks if a type implements a protocol/role
func (u *TypeUnifier) checkProtocol(typ, protocol typedef.Type, ctx *UnificationContext) error {
	// TODO: Implement protocol checking
	return fmt.Errorf("protocol checking not yet implemented")
}

// TypeVariable represents a type variable like T, U, etc.
type TypeVariable struct {
	Name        string
	Constraints []typedef.TypeConstraint
}

// GetName implements the Type interface
func (tv *TypeVariable) GetName() string {
	return tv.Name
}

// IsCompatible implements the Type interface
func (tv *TypeVariable) IsCompatible(other typedef.Type) bool {
	// Type variables are compatible with anything during unification
	return true
}

// String returns string representation
func (tv *TypeVariable) String() string {
	if len(tv.Constraints) > 0 {
		return fmt.Sprintf("%s where %v", tv.Name, tv.Constraints)
	}
	return tv.Name
}

// applyTypeParameterSubstitutions applies type parameter substitutions to a type
func (u *TypeUnifier) applyTypeParameterSubstitutions(typ typedef.Type, substitutions map[string]typedef.Type) typedef.Type {
	if typ == nil {
		return nil
	}

	// If this is a simple type that matches a type parameter, substitute it
	if simple, ok := typ.(*typedef.SimpleType); ok {
		if replacement, exists := substitutions[simple.Name]; exists {
			return replacement
		}
		return typ
	}

	// For parameterized types, recursively apply substitutions
	if param, ok := typ.(*typedef.ParameterizedType); ok {
		newArgs := make([]typedef.Type, len(param.Arguments))
		for i, arg := range param.Arguments {
			newArgs[i] = u.applyTypeParameterSubstitutions(arg, substitutions)
		}
		return &typedef.ParameterizedType{
			BaseType:  u.applyTypeParameterSubstitutions(param.BaseType, substitutions),
			Arguments: newArgs,
		}
	}

	return typ
}

// convertConstraintKind converts typedef.ConstraintKind to UnificationConstraintKind
func convertConstraintKind(kind typedef.ConstraintKind) ConstraintKind {
	switch kind {
	case typedef.TraitConstraint:
		return ProtocolConstraint // Map trait constraints to protocol constraints
	case typedef.ProtocolConstraint:
		return ProtocolConstraint
	case typedef.ValueConstraint:
		return SubtypeConstraint // Map value constraints to subtype constraints
	default:
		return EqualityConstraint // Default to equality
	}
}

// matchGenericPattern attempts to match a generic type pattern against a parameterized type
func (u *TypeUnifier) matchGenericPattern(generic *typedef.GenericType, paramType *typedef.ParameterizedType, ctx *UnificationContext) error {
	// Check if the base types are compatible
	if paramType.BaseType != nil && generic.Name == paramType.BaseType.GetName() {
		// The base types match, so we can try to unify type parameters with arguments
		if len(generic.TypeParameters) != len(paramType.Arguments) {
			return fmt.Errorf("type parameter count mismatch: expected %d, got %d",
				len(generic.TypeParameters), len(paramType.Arguments))
		}

		// Create unification constraints for each type parameter
		for i, param := range generic.TypeParameters {
			typeVar := &TypeVariable{Name: param.Name}
			ctx.Constraints = append(ctx.Constraints, UnificationConstraint{
				Left:  typeVar,
				Right: paramType.Arguments[i],
				Kind:  EqualityConstraint,
			})
		}

		return nil
	}

	return fmt.Errorf("generic pattern %s does not match %s", generic.Name, paramType.GetName())
}

// matchGenericWithSimple attempts to match a generic type with a simple concrete type
func (u *TypeUnifier) matchGenericWithSimple(generic *typedef.GenericType, simpleType *typedef.SimpleType, ctx *UnificationContext) error {
	// If the generic has a base type that's compatible with the simple type
	if generic.BaseType != nil && generic.BaseType.IsCompatible(simpleType) {
		// For a simple match, we can't infer specific type parameters
		// This would need more sophisticated pattern matching
		// For now, just succeed if the types are compatible
		return nil
	}

	// Check if the simple type name matches the generic name directly
	if generic.Name == simpleType.Name {
		return nil
	}

	return fmt.Errorf("simple type %s does not match generic %s", simpleType.Name, generic.Name)
}

// UnifyWithConstraintPropagation unifies types with full constraint propagation
func (u *TypeUnifier) UnifyWithConstraintPropagation(generic *typedef.GenericType, concrete typedef.Type) (*PropagationResult, error) {
	if generic == nil {
		return nil, fmt.Errorf("generic type cannot be nil")
	}
	if concrete == nil {
		return nil, fmt.Errorf("concrete type cannot be nil")
	}

	// First, perform basic unification to get initial type assignments
	substitutions, err := u.Unify(generic, concrete)
	if err != nil {
		return nil, fmt.Errorf("initial unification failed: %w", err)
	}

	// Convert substitutions to type slice for constraint propagation
	inferredTypes := make([]typedef.Type, len(generic.TypeParameters))
	for i, param := range generic.TypeParameters {
		if sub, exists := substitutions[param.Name]; exists {
			inferredTypes[i] = sub
		} else {
			// If no substitution found, use a type variable
			inferredTypes[i] = &TypeVariable{Name: param.Name}
		}
	}

	// Propagate constraints
	result, err := u.constraintPropagator.PropagateConstraints(generic, inferredTypes)
	if err != nil {
		return nil, fmt.Errorf("constraint propagation failed: %w", err)
	}

	return result, nil
}

// unifyConstraints unifies constraints between two generic types.
// This is the core constraint unification algorithm that ensures constraints
// are compatible when unifying Container<T: Serializable> with Container<U: Serializable>.
//
// The algorithm works by:
// 1. Mapping constraints from the second generic to the first's parameter namespace
// 2. For each type parameter, checking that constraints from both generics are compatible
// 3. Requiring bidirectional compatibility (each constraint must have a match)
//
// Examples:
//
//	Container<T: Serializable> ∪ Container<U: Serializable> → success (same constraint)
//	Container<T: Serializable> ∪ Container<U: Display>      → failure (Serializable ≠ Display)
//	Container<T: Serializable> ∪ Container<U>               → success (no constraints on U)
func (u *TypeUnifier) unifyConstraints(
	constraints1, constraints2 []typedef.TypeConstraint,
	paramMap map[string]string,
	ctx *UnificationContext,
) error {
	// Create a map of constraints by parameter for easy lookup
	c1Map := make(map[string][]typedef.TypeConstraint)
	for _, c := range constraints1 {
		c1Map[c.Parameter] = append(c1Map[c.Parameter], c)
	}

	c2Map := make(map[string][]typedef.TypeConstraint)
	for _, c := range constraints2 {
		// Map parameter names from g2 to g1's namespace
		mappedParam := c.Parameter
		// Reverse lookup: find g1's parameter that maps to g2's parameter
		for k, v := range paramMap {
			if v == c.Parameter {
				mappedParam = k
				break
			}
		}
		c2Map[mappedParam] = append(c2Map[mappedParam], c)
	}

	// Check constraint compatibility for each parameter
	allParams := make(map[string]bool)
	for param := range c1Map {
		allParams[param] = true
	}
	for param := range c2Map {
		allParams[param] = true
	}

	for param := range allParams {
		c1List := c1Map[param]
		c2List := c2Map[param]

		if err := u.unifyParameterConstraints(param, c1List, c2List, ctx); err != nil {
			return fmt.Errorf("failed to unify constraints for parameter %s: %w", param, err)
		}
	}

	return nil
}

// unifyParameterConstraints unifies constraints for a single type parameter.
// Checks bidirectional compatibility: every constraint in the first set must have
// a compatible constraint in the second set, and vice versa.
func (u *TypeUnifier) unifyParameterConstraints(
	param string,
	constraints1, constraints2 []typedef.TypeConstraint,
	ctx *UnificationContext,
) error {
	// If one side has no constraints, unification succeeds
	if len(constraints1) == 0 || len(constraints2) == 0 {
		return nil
	}

	// For each constraint in the first set, find a compatible constraint in the second set
	for _, c1 := range constraints1 {
		compatible := false
		for _, c2 := range constraints2 {
			if isConstraintCompatible(c1, c2) {
				compatible = true
				break
			}
		}
		if !compatible {
			return fmt.Errorf("constraint %s:%s from first generic has no compatible constraint in second generic",
				c1.Parameter, c1.Expression.GetName())
		}
	}

	// Check the reverse direction
	for _, c2 := range constraints2 {
		compatible := false
		for _, c1 := range constraints1 {
			if isConstraintCompatible(c1, c2) {
				compatible = true
				break
			}
		}
		if !compatible {
			return fmt.Errorf("constraint %s:%s from second generic has no compatible constraint in first generic",
				c2.Parameter, c2.Expression.GetName())
		}
	}

	return nil
}

// isConstraintCompatible checks if two constraints are compatible for unification.
// Two constraints are compatible if:
// 1. They have the same constraint kind (trait, protocol, etc.)
// 2. For trait/protocol constraints, they reference the same trait/protocol name
// 3. Value constraints are currently considered always compatible
//
// Future enhancements could add trait inheritance checking where Serializable
// might be compatible with a base trait like Object.
func isConstraintCompatible(c1, c2 typedef.TypeConstraint) bool {
	// Same constraint kind is required
	if c1.Kind != c2.Kind {
		return false
	}

	// For trait and protocol constraints, check if the types are the same
	// In a more sophisticated implementation, this would check for trait inheritance
	if c1.Kind == typedef.TraitConstraint || c1.Kind == typedef.ProtocolConstraint {
		return c1.Expression.GetName() == c2.Expression.GetName()
	}

	// Value constraints are always considered compatible for now
	if c1.Kind == typedef.ValueConstraint {
		return true
	}

	return false
}
