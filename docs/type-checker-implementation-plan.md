# Type Checker Implementation Plan

## Overview

This document outlines the incremental implementation plan for the Typed Perl type checker, based on the comprehensive specification in `typed-perl-spec.md`. The plan follows Test-Driven Development (TDD) principles and builds complexity incrementally.

## Core Design Principles (from Spec)

1. **Gradual** - Type annotations are optional and can be added incrementally
2. **Bidirectional** - Type information flows both from declarations to expressions and from usage to variables
3. **Flow-sensitive** - Type refinements based on runtime checks are tracked through control flow
4. **Zero runtime overhead** - Types are stripped after verification
5. **Separate from runtime** - Type checking happens in a separate process from Perl execution

## Implementation Strategy

- **TDD First**: Write failing tests before implementation
- **Incremental Complexity**: Each phase builds on the previous one
- **Working Increments**: Every phase leaves the system in a working state
- **Specification Alignment**: Each phase maps to specific spec requirements

## Implementation Phases

### Phase 1: Pure Type Inference ✓ Spec Sections 3.1, 3.2
*Focus: Basic type synthesis from literals and expressions*

**Specification Requirements:**
- Type Synthesis (↑) - Generate types from expressions upward (Section 3.1.2)
- Synthesis Priority - Generate most specific type possible (Section 3.1.3)
- Basic literal type inference for Int, Float, Str, Bool, Undef (Section 2.1)

**TDD Implementation:**
```perl
# Test cases to drive implementation:
42        → Int
3.14      → Float
"hello"   → Str
'world'   → Str
1         → Bool (in boolean context)
0         → Bool (in boolean context)
undef     → Undef
```

**Success Criteria:**
- All basic literals correctly infer their most specific type
- Type hierarchy respects Int ⊆ Num ⊆ Str ⊆ Scalar ⊆ Any
- No variable dependencies yet

### Phase 2: Basic Type Validation ✓ Spec Section 2.1, 6.2
*Focus: Validating primitive types exist and are well-formed*

**Specification Requirements:**
- Core Type Hierarchy validation (Section 2.1)
- Type name validation and error reporting (Section 6.2.4)
- Built-in type recognition

**TDD Implementation:**
```perl
# Valid types that must pass validation:
Int, Str, Bool, Num, Float, Undef, Scalar, Any
ArrayRef, HashRef, CodeRef, ScalarRef, Ref
List, Array, Hash, Code, Glob

# Invalid types that must be rejected:
invalidType (lowercase), "", custom types without definition
```

**Success Criteria:**
- All built-in types validate correctly
- Invalid types rejected with clear error messages
- Type names follow specification conventions

### Phase 3: Variable Type Annotations ✓ Spec Section 4.1, 7.1
*Focus: Simple variable type checking*

**Specification Requirements:**
- Variable Type Annotations syntax (Section 4.1)
- Variables and Assignment rules (Section 7.1)
- Assignment compatibility checking (Section 2.2.2)

**TDD Implementation:**
```perl
# Test cases for variable annotations:
my Int $x = 42;        # Valid
my Str $name = "Bob";  # Valid
my Int $bad = "text";  # Type error
my $inferred = 42;     # Infers Int type
```

**Success Criteria:**
- Variables can be annotated with types
- Assignment type mismatches detected
- Type inference for unallocated variables works
- Subtype assignment compatibility (Int → Num → Str)

### Phase 4: Operator Type Checking ✓ Spec Section 3.2, 7.2, Appendix B
*Focus: Type checking for basic operators*

**Specification Requirements:**
- Operator-Based Type Inference (Section 3.2)
- Operator type rules (Section 7.2)
- Specific operator type mappings (Appendix B)

**TDD Implementation:**
```perl
# String operators require Str operands, return appropriate types:
$a . $b     # Str, Str → Str
$a eq $b    # Str, Str → Bool
$a lt $b    # Str, Str → Bool

# Numeric operators require Num operands:
$a + $b     # Num, Num → Num
$a == $b    # Num, Num → Bool
$a < $b     # Num, Num → Bool

# Boolean operators:
$a && $b    # Any, Any → Bool
$a || $b    # Any, Any → Bool
!$a         # Any → Bool
```

**Success Criteria:**
- Operators enforce correct operand types
- Result types properly inferred per specification
- Type coercion follows Perl semantics (Int → Num → Str)

### Phase 5: Function Signatures ✓ Spec Section 4.2, 7.4
*Focus: Parameter and return type checking*

**Specification Requirements:**
- Function Signatures and Return Types syntax (Section 4.2)
- Function Calls validation (Section 7.4)
- Method vs subroutine distinction

**TDD Implementation:**
```perl
# Function signature validation:
sub add(Int $a, Int $b) -> Int { return $a + $b; }
method getName() -> Str { return $self->{name}; }

# Call-site type checking:
my $result = add(1, 2);      # Valid - Int, Int
my $bad = add("a", "b");     # Error - Str, Str not compatible
```

**Success Criteria:**
- Function parameters type-checked at call sites
- Return types match function signatures
- Method vs subroutine distinction works correctly

### Phase 6: Container Types ✓ Spec Section 2.4, 4.3
*Focus: ArrayRef, HashRef, and parameterized types*

**Specification Requirements:**
- Parameterized Types (Section 2.4.1)
- Container Types syntax (Section 4.3)
- Generic constraints and variance (Section 2.4.2, 2.4.4)

**TDD Implementation:**
```perl
# Basic parameterized types:
my ArrayRef[Int] $numbers = [1, 2, 3];
my HashRef[Str, Int] %grades = {alice => 95, bob => 87};

# Nested containers:
my ArrayRef[ArrayRef[Int]] $matrix = [[1, 2], [3, 4]];

# Covariance: ArrayRef[Int] ⊆ ArrayRef[Num]
```

**Success Criteria:**
- Container types work with proper element types
- Nested containers supported
- Covariant subtyping for containers
- Type parameters validated correctly

### Phase 7: Union and Intersection Types ✓ Spec Section 2.3, 4.4
*Focus: Complex type combinations*

**Specification Requirements:**
- Type Operations (Section 2.3)
- Type Combinations syntax (Section 4.4)
- Union, intersection, and negation types

**TDD Implementation:**
```perl
# Union types - value can be either type:
my Int|Str $flexible = 42;      # Valid
$flexible = "hello";            # Also valid

# Intersection types - value must satisfy both:
my Serializable & Printable $obj;

# Type negation - value cannot be this type:
my !Ref $primitive = 42;        # Valid
$primitive = [];                # Error - is a Ref
```

**Success Criteria:**
- Union types accept either alternative
- Intersection types require both constraints
- Type negation properly excludes types
- Complex type expressions parse and validate

### Phase 8: Flow-Sensitive Analysis ✓ Spec Section 3.4
*Focus: Type refinement based on conditions*

**Specification Requirements:**
- Flow-Sensitive Type Refinement (Section 3.4)
- Pattern Recognition for type narrowing
- Key validation patterns (defined, ref checks, regex)

**TDD Implementation:**
```perl
# Type refinement in conditionals:
my Maybe[Str] $data = get_input();
if (defined($data)) {
    # $data is now Str, not Maybe[Str]
    my $length = length($data);  # Valid
}

# Reference type refinement:
my Ref $thing = get_ref();
if (ref($thing) eq 'ARRAY') {
    # $thing is now ArrayRef
    my $first = $thing->[0];     # Valid
}

# Regex-based type refinement:
my Str $input = get_string();
if ($input =~ /^\d+$/) {
    # $input refined to Int
    my $doubled = $input * 2;    # Valid
}
```

**Success Criteria:**
- Types refined in conditional branches
- Maybe[T] narrows to T after defined() check
- Ref types narrow after ref() validation
- Regex patterns enable Str → Int refinement

### Phase 9: Context-Sensitive Types ✓ Spec Section 3.3
*Focus: Scalar vs list context and type coercion*

**Specification Requirements:**
- Context Sensitivity (Section 3.3)
- List vs. Scalar Context handling
- Union type approach for context-dependent returns

**TDD Implementation:**
```perl
# Context-dependent return types:
sub get_data() -> List[Int]|Str { ... }

my @array_context = get_data();  # List[Int] applies
my $scalar_context = get_data(); # Str applies

# Built-in context handling:
my @keys = keys %hash;           # List[Str]
my $count = keys %hash;          # Int
```

**Success Criteria:**
- Functions return appropriate types based on context
- Union types resolve correctly in different contexts
- Built-in functions respect context sensitivity

### Phase 10: Advanced Features ✓ Spec Section 4.3, 4.5, 5
*Focus: Generic types, type aliases, and module imports*

**Specification Requirements:**
- Generic Type Syntax (Section 4.3)
- Type Aliases (Section 4.5)
- Type Definition Files (Section 5)
- Higher-Kinded Types (Section 2.5)

**TDD Implementation:**
```perl
# Generic functions:
sub identity<T>(T $value) -> T { return $value; }
my $str_result = identity("hello");  # T = Str
my $int_result = identity(42);       # T = Int

# Type aliases:
type UserID = Str;
type Point = {x: Int, y: Int};

# Module type imports:
use DBI;  # Automatically imports DBI.ptd type definitions
my DBI::db $dbh = DBI->connect(...);
```

**Success Criteria:**
- Generic functions work with type parameters
- Type aliases can be defined and used
- Module types imported from .ptd files
- Higher-kinded types supported (advanced)

## Implementation Notes

### Error Reporting
Each phase must include clear, actionable error messages with:
- Source location (file, line, column)
- Expected vs actual types
- Suggested fixes where appropriate
- Error codes for programmatic handling

### Performance Considerations
- Start with correctness over performance
- Profile and optimize in later phases
- Consider incremental compilation strategies
- Cache type information appropriately

### Integration Points
- Parser integration for type annotation syntax
- AST enhancement for type information
- Error reporting system integration
- Test suite integration with existing tests

## Testing Strategy

Each phase follows strict TDD:

1. **Red**: Write failing test that defines required behavior
2. **Green**: Implement minimum code to make test pass
3. **Refactor**: Clean up implementation while keeping tests green
4. **Integration**: Verify new feature works with all existing functionality

### Test Categories
- **Unit Tests**: Individual type checker components
- **Integration Tests**: Type checker + parser interaction
- **End-to-End Tests**: Complete files with type annotations
- **Error Tests**: Verify proper error reporting
- **Performance Tests**: Ensure acceptable speeds

## Specification Compliance

This plan maps directly to the typed-perl-spec.md requirements:

- ✅ **Section 1**: Core design principles embedded throughout
- ✅ **Section 2**: Type hierarchy implemented in Phases 1-7
- ✅ **Section 3**: Type inference system in Phases 1, 4, 8-9
- ✅ **Section 4**: Type syntax in Phases 3, 5-7, 10
- ✅ **Section 5**: Type definition files in Phase 10
- ✅ **Section 6**: Implementation requirements covered across all phases
- ✅ **Section 7**: Type checking rules distributed across relevant phases
- ✅ **Section 8**: Error handling integrated throughout
- ✅ **Section 9**: Future extensions noted for post-implementation

## Success Metrics

- All specification requirements implemented
- Comprehensive test coverage (>95%)
- Clear, helpful error messages
- Performance acceptable for typical Perl files
- Integration with existing PSC architecture
- Documentation and examples for users

This incremental approach ensures that each phase delivers working functionality while building toward the complete specification compliance.
