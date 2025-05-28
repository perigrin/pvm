# Type Annotation Detection Improvements - Implementation Plan

## Project Context

This plan focuses on improving type annotation detection accuracy and expanding test coverage for the PVM parser. The goal is to create a robust test suite that covers both untyped Perl (as the baseline) and typed-Perl extensions, with incremental parser improvements driven by comprehensive test coverage.

## Target Architecture

**Goal**: Robust, accurate parsing of both standard Perl and typed-Perl with comprehensive test coverage
**Strategy**: Test-driven development with incremental complexity increases
**Target Outcome**: Parser capable of handling complex type expressions with strong regression protection

### Key Areas of Improvement
1. **Test Infrastructure**: Comprehensive acceptance testing framework
2. **Untyped Perl Coverage**: Complete standard Perl language support
3. **Type Annotation Coverage**: Full typed-Perl extension support
4. **Parser Enhancements**: Improved accuracy and error handling
5. **Integration Testing**: End-to-end validation and regression protection

---

## Step-by-Step Implementation Plan

### Step 1: Parser Testing Infrastructure ✅ COMPLETED

```
You are implementing comprehensive parser testing infrastructure for PVM.

CONTEXT: PVM has a working parser with tree-sitter integration, but needs robust test coverage and accuracy measurement tools for upcoming type annotation improvements.

TASK: Create foundational testing infrastructure for parser accuracy and regression testing.

REQUIREMENTS:
1. Create `internal/parser/testdata/` directory structure for test fixtures
2. Implement parser test framework with accuracy measurement
3. Add test fixture management for both input files and expected outputs
4. Create baseline measurement tools for parser accuracy
5. Add test data validation and comparison utilities
6. Implement test categorization (untyped-perl, typed-perl, error-cases)
7. Add performance benchmarking for parser operations

TECHNICAL DETAILS:
- Use table-driven tests for systematic coverage
- Support both positive (successful parse) and negative (error) test cases
- Include position accuracy testing for AST nodes
- Add test fixture generation utilities
- Implement AST comparison utilities for expected vs actual
- Support test data in multiple formats (individual files, test suites)

DELIVERABLES:
- internal/parser/testdata/ directory with structure
- internal/parser/test_framework.go with testing utilities
- internal/parser/accuracy_test.go with measurement tools
- Baseline test infrastructure ready for language feature testing

SUCCESS CRITERIA:
- Test framework can load fixtures and measure parser accuracy
- AST comparison utilities work correctly
- Baseline measurements establish current parser capabilities
- Infrastructure ready for incremental test addition

Focus on creating robust foundations that will support 20+ upcoming test implementation steps.
```

### Step 2: Core Variable Declarations (Untyped Perl) ✅ COMPLETED

```
You are continuing the parser testing implementation from Step 1.

CONTEXT: Parser testing infrastructure is complete. Now implement comprehensive test coverage for core Perl variable declarations without type annotations.

TASK: Create exhaustive test coverage for scalar, array, and hash variable declarations in standard Perl.

REQUIREMENTS:
1. Add test fixtures for scalar variable declarations ($var, $Package::var)
2. Add test fixtures for array variable declarations (@array, @Package::array)
3. Add test fixtures for hash variable declarations (%hash, %Package::hash)
4. Test variable scoping keywords (my, our, state, local)
5. Test variable assignments with different value types
6. Test package-qualified variable names
7. Test variable declarations in different contexts (statement, expression)

TEST EXAMPLES TO COVER:
```perl
# Scalar variables
my $simple;
my $assigned = 42;
our $package_var;
state $persistent = "initial";
local $localized;
$Package::qualified = "value";

# Array variables
my @simple_array;
my @assigned = (1, 2, 3);
our @package_array;
@Package::qualified = qw(a b c);

# Hash variables
my %simple_hash;
my %assigned = (key => 'value');
our %package_hash;
%Package::qualified = (a => 1, b => 2);
```

TECHNICAL DETAILS:
- Each test should verify AST structure matches expected output
- Test both successful parsing and position accuracy
- Include edge cases like empty declarations
- Test variable names with underscores, numbers, package qualifiers
- Verify scoping keyword detection in AST

DELIVERABLES:
- testdata/untyped-perl/variables/ directory with comprehensive fixtures
- Parser tests validating variable declaration parsing
- Expected AST outputs for all variable declaration patterns
- Documentation of variable parsing coverage

SUCCESS CRITERIA:
- All standard Perl variable declaration patterns parse correctly
- AST contains accurate variable information (name, scope, package)
- Position information is precise for all variable elements
- Test coverage provides baseline for type annotation additions

Focus on comprehensive coverage of standard Perl variable patterns before adding type extensions.
```

### Step 3: Basic Expressions and Operators (Untyped Perl) ✅ COMPLETED

```
You are continuing the parser testing implementation from Step 2.

CONTEXT: Variable declarations test coverage is complete. Now implement comprehensive test coverage for basic expressions and operators in standard Perl.

TASK: Create exhaustive test coverage for arithmetic, string, logical, and comparison operators.

REQUIREMENTS:
1. Add test fixtures for arithmetic operations (+, -, *, /, %, **)
2. Add test fixtures for string operations (., x, eq, ne, lt, gt, le, ge)
3. Add test fixtures for logical operations (&&, ||, !, and, or, not)
4. Add test fixtures for comparison operations (==, !=, <, >, <=, >=, <=>)
5. Add test fixtures for assignment operations (=, +=, -=, .=, etc.)
6. Test operator precedence and associativity
7. Test parenthesized expressions and complex nesting

TEST EXAMPLES TO COVER:
```perl
# Arithmetic expressions
$result = $a + $b * $c;
$power = $base ** $exponent;
$remainder = $dividend % $divisor;

# String operations
$combined = $first . $second;
$repeated = $string x $count;
$comparison = $left eq $right;

# Logical operations
$and_result = $a && $b;
$or_result = $x || $y;
$not_result = !$condition;

# Complex expressions
$complex = ($a + $b) * ($c - $d) / ($e || 1);
$chained = $x <=> $y;

# Assignment operators
$total += $increment;
$message .= $suffix;
```

TECHNICAL DETAILS:
- Test operator precedence matches Perl specification
- Verify AST structure represents operator hierarchy correctly
- Include test cases for each operator with different operand types
- Test edge cases like division by zero handling
- Verify position information for operators and operands

DELIVERABLES:
- testdata/untyped-perl/expressions/ directory with operator test fixtures
- Comprehensive operator parsing tests with expected AST outputs
- Precedence and associativity validation tests
- Documentation of expression parsing coverage

SUCCESS CRITERIA:
- All Perl operators parse with correct precedence and associativity
- AST accurately represents expression structure and operator types
- Complex nested expressions parse correctly
- Position information is accurate for all expression elements

Focus on complete operator coverage to establish expression parsing foundation.
```

### Step 4: Control Flow Structures (Untyped Perl) ✅ COMPLETED

```
You are continuing the parser testing implementation from Step 3.

CONTEXT: Expression and operator test coverage is complete. Now implement comprehensive test coverage for control flow structures in standard Perl.

TASK: Create exhaustive test coverage for conditional statements, loops, and control flow keywords.

REQUIREMENTS:
1. Add test fixtures for if/elsif/else conditional statements
2. Add test fixtures for unless conditionals
3. Add test fixtures for while and until loops
4. Add test fixtures for for and foreach loops
5. Add test fixtures for loop control (next, last, redo)
6. Add test fixtures for given/when statements (if supported)
7. Test nested control structures and complex conditions

TEST EXAMPLES TO COVER:
```perl
# Conditional statements
if ($condition) {
    do_something();
} elsif ($other) {
    do_other();
} else {
    do_default();
}

unless ($negative_condition) {
    execute();
}

# Loop structures
while ($condition) {
    process();
}

for my $i (0..$max) {
    handle($i);
}

foreach my $item (@list) {
    process($item);
    next if skip_condition($item);
    last if stop_condition($item);
}

# Complex nesting
for my $outer (@outer_list) {
    foreach my $inner (@{$outer->{items}}) {
        if ($inner->{valid}) {
            process($inner);
        }
    }
}
```

TECHNICAL DETAILS:
- Test control flow structure parsing and AST representation
- Verify block structure and statement grouping
- Test loop variable scoping and iteration patterns
- Include edge cases like empty blocks and single statements
- Verify condition expression parsing within control structures

DELIVERABLES:
- testdata/untyped-perl/control-flow/ directory with comprehensive fixtures
- Control flow parsing tests with expected AST structures
- Loop and conditional statement validation tests
- Documentation of control flow parsing coverage

SUCCESS CRITERIA:
- All Perl control flow structures parse correctly with proper AST representation
- Block boundaries and statement grouping are accurate
- Loop variables and conditions are properly captured
- Nested structures maintain correct hierarchy in AST

Focus on complete control flow coverage to support complex program structure parsing.
```

### Step 5: Subroutines and Calls (Untyped Perl) ✅ COMPLETED

```
You are continuing the parser testing implementation from Step 4.

CONTEXT: Control flow test coverage is complete. Now implement comprehensive test coverage for subroutine definitions, calls, and related constructs in standard Perl.

TASK: Create exhaustive test coverage for subroutine definition and invocation patterns.

REQUIREMENTS:
1. Add test fixtures for basic subroutine definitions (sub name)
2. Add test fixtures for anonymous subroutines and references
3. Add test fixtures for subroutine calls with various argument patterns
4. Add test fixtures for method calls and arrow notation
5. Add test fixtures for subroutine prototypes and attributes
6. Test package-qualified subroutine names
7. Test subroutine references and code references

TEST EXAMPLES TO COVER:
```perl
# Basic subroutine definitions
sub simple_sub {
    return "result";
}

sub with_params {
    my ($first, $second) = @_;
    return $first + $second;
}

# Anonymous subroutines
my $code_ref = sub {
    my $param = shift;
    return $param * 2;
};

# Subroutine calls
my $result = simple_sub();
my $sum = with_params(10, 20);
my $doubled = $code_ref->(5);

# Method calls
my $object = Package->new();
$object->method($arg1, $arg2);
Package::function($args);

# Prototypes and attributes
sub prototype_sub ($$) {
    my ($a, $b) = @_;
    return $a + $b;
}

sub attributed_sub : lvalue {
    return $global_var;
}
```

TECHNICAL DETAILS:
- Test subroutine definition parsing with parameter handling
- Verify call site argument parsing and expression handling
- Test method call syntax and arrow operator parsing
- Include edge cases like calls without parentheses
- Test subroutine reference creation and dereferencing

DELIVERABLES:
- testdata/untyped-perl/subroutines/ directory with comprehensive fixtures
- Subroutine definition and call parsing tests
- Method invocation and reference handling tests
- Documentation of subroutine parsing coverage

SUCCESS CRITERIA:
- All subroutine definition patterns parse correctly
- Call site argument lists and expressions are properly captured
- Method calls and package qualification work correctly
- AST accurately represents subroutine structure and calls

Focus on complete subroutine coverage to support complex program organization.
```

### Step 6: Package and Module Constructs (Untyped Perl) ✅ COMPLETED

```
You are continuing the parser testing implementation from Step 5.

CONTEXT: Subroutine test coverage is complete. Now implement comprehensive test coverage for package declarations, module imports, and namespace constructs in standard Perl.

TASK: Create exhaustive test coverage for package organization and module system features.

REQUIREMENTS:
1. Add test fixtures for package declarations and namespace changes
2. Add test fixtures for use and require statements
3. Add test fixtures for import and export functionality
4. Add test fixtures for package variables and qualification
5. Test version specifications in use statements
6. Test pragma usage (strict, warnings, etc.)
7. Test complex module loading patterns

TEST EXAMPLES TO COVER:
```perl
# Package declarations
package MyPackage;
package MyPackage::Subspace;
package MyPackage 1.23;

# Module imports
use strict;
use warnings;
use Data::Dumper;
use MyModule qw(function1 function2);
use AnotherModule 1.5 qw(:all);
require DynamicModule;

# Package qualification
$MyPackage::variable = "value";
MyPackage::function();
my $ref = \&MyPackage::function;

# Complex patterns
{
    package LocalPackage;
    use parent 'BaseClass';

    sub new {
        my $class = shift;
        return bless {}, $class;
    }
}

# Version specifications
use 5.010;
use MyModule v1.2.3;
```

TECHNICAL DETAILS:
- Test package declaration parsing and namespace tracking
- Verify use/require statement parsing with import lists
- Test version specification parsing in various formats
- Include edge cases like bareword imports and version ranges
- Test package qualification resolution in different contexts

DELIVERABLES:
- testdata/untyped-perl/packages/ directory with comprehensive fixtures
- Package and module parsing tests with expected AST structures
- Import/export and version specification tests
- Documentation of package system parsing coverage

SUCCESS CRITERIA:
- All package declaration patterns parse correctly
- Module import statements are properly parsed with import lists
- Version specifications are accurately captured
- Package qualification works correctly in all contexts

This completes the untyped Perl baseline coverage. Next steps will add type annotations.
```

### Step 7: Simple Type Annotations ✅ COMPLETED

```
You are continuing the parser testing implementation from Step 6.

CONTEXT: Complete untyped Perl baseline coverage is established. Now begin implementing typed-Perl extensions starting with simple type annotations.

TASK: Create comprehensive test coverage for basic type annotations on variable declarations.

REQUIREMENTS:
1. Add test fixtures for typed scalar declarations (my Int $var)
2. Add test fixtures for typed array declarations (my ArrayRef[Int] @array)
3. Add test fixtures for typed hash declarations (my HashRef[Str] %hash)
4. Test built-in type names (Int, Str, Bool, Num, ArrayRef, HashRef)
5. Test custom type names and package-qualified types
6. Test type annotations with scoping keywords (my, our, state)
7. Verify type information is captured in AST alongside variable information

TEST EXAMPLES TO COVER:
```perl
# Basic typed variables
my Int $count = 42;
my Str $name = "example";
my Bool $flag = 1;
my Num $pi = 3.14159;

# Typed arrays and hashes
my ArrayRef[Int] @numbers = (1, 2, 3);
my HashRef[Str] %config = (key => 'value');

# Custom types
my MyType $custom;
my Package::CustomType $qualified;

# Different scoping
our Int $global_counter;
state Str $persistent_cache;

# Complex assignments
my Int $calculated = $base + $increment;
my Str $interpolated = "Value: $count";
```

TECHNICAL DETAILS:
- Extend existing variable declaration parsing to include type information
- Ensure type annotations are optional and don't break untyped code
- Test that type information is properly stored in AST nodes
- Verify position information for both type and variable elements
- Ensure backward compatibility with all untyped variable patterns

DELIVERABLES:
- testdata/typed-perl/simple-annotations/ directory with type annotation fixtures
- Parser tests validating type annotation parsing and AST integration
- Type information validation in AST structures
- Documentation of type annotation parsing coverage

SUCCESS CRITERIA:
- Simple type annotations parse correctly and are captured in AST
- Type information doesn't interfere with untyped variable parsing
- All built-in type names are recognized correctly
- Custom and package-qualified types work properly

This establishes the foundation for more complex type expressions in subsequent steps.
```

### Step 8: Method and Field Annotations ✅ COMPLETED

```
You are continuing the parser testing implementation from Step 7.

CONTEXT: Simple type annotations are working. Now implement comprehensive test coverage for method and field type annotations.

TASK: Create exhaustive test coverage for typed method definitions and field declarations.

REQUIREMENTS:
1. Add test fixtures for typed method definitions with parameter types
2. Add test fixtures for method return type annotations
3. Add test fixtures for field declarations with types
4. Test method signature parsing with multiple typed parameters
5. Test optional parameter and return type annotations
6. Test method type annotations in class contexts
7. Verify method and field type information in AST

TEST EXAMPLES TO COVER:
```perl
# Typed method definitions
method calculate(Int $a, Int $b) -> Int {
    return $a + $b;
}

method process(Str $input, Bool $validate = 1) -> ArrayRef[Str] {
    my @result = split /,/, $input;
    return \@result;
}

# Field declarations
field Int $count = 0;
field Str $name;
field ArrayRef[MyType] $items;

# Class context
class Calculator {
    field Num $precision = 0.001;

    method add(Num $a, Num $b) -> Num {
        return $a + $b;
    }

    method get_precision() -> Num {
        return $precision;
    }
}

# Complex parameter types
method complex_method(
    HashRef[ArrayRef[Int]] $data,
    CodeRef $callback,
    Optional[Str] $name
) -> Bool {
    return 1;
}
```

TECHNICAL DETAILS:
- Extend method parsing to capture parameter and return type information
- Test field declaration parsing with type annotations
- Verify method signature parsing handles complex type expressions
- Include edge cases like methods without types mixed with typed methods
- Test that method and field types are properly integrated into AST

DELIVERABLES:
- testdata/typed-perl/methods-fields/ directory with comprehensive fixtures
- Method and field type annotation parsing tests
- Method signature and return type validation tests
- Documentation of method/field type parsing coverage

SUCCESS CRITERIA:
- Method parameter and return types are correctly parsed and stored
- Field type annotations are properly captured in AST
- Complex method signatures with multiple typed parameters work
- Type information for methods and fields integrates cleanly with AST

This builds on simple type annotations to support object-oriented typed programming.
```

### Step 9: Type Assertions and Constraints ✅ COMPLETED

```
You are continuing the parser testing implementation from Step 8.

CONTEXT: Method and field type annotations are working. Now implement comprehensive test coverage for type assertions and constraint expressions.

TASK: Create exhaustive test coverage for type assertion syntax and runtime type checking expressions.

REQUIREMENTS:
1. Add test fixtures for type assertion expressions ($value as Type)
2. Add test fixtures for type constraint expressions (where Type)
3. Add test fixtures for conditional type assertions
4. Test type assertions in various expression contexts
5. Test nested type assertions and complex constraints
6. Test error handling for invalid type assertions
7. Verify type assertion information is captured in AST

TEST EXAMPLES TO COVER:
```perl
# Basic type assertions
my $number = $input as Int;
my $text = $data as Str;
my $ref = $object as MyClass;

# Type assertions in expressions
my $result = ($calculation + $offset) as Num;
my $item = $array->[$index] as ItemType;

# Conditional type assertions
my $value = $maybe_number as Int // 0;
my $obj = $input as MyClass or die "Wrong type";

# Complex constraints
my $validated = $input as (Int where { $_ > 0 });
my $range = $number as (Num where { $_ >= 0 && $_ <= 100 });

# Method context
method process($input) {
    my $typed = $input as ArrayRef[Str];
    return $typed->map(sub { uc($_) });
}

# Assignment with assertion
$self->{count} = $new_value as Int;
```

TECHNICAL DETAILS:
- Extend expression parsing to handle 'as' type assertion operator
- Test type assertion precedence with other operators
- Verify constraint expressions ('where' clauses) parse correctly
- Include edge cases like chained assertions and complex expressions
- Test that type assertion information is preserved in AST

DELIVERABLES:
- testdata/typed-perl/assertions/ directory with type assertion fixtures
- Type assertion and constraint parsing tests
- Expression context and precedence validation tests
- Documentation of type assertion parsing coverage

SUCCESS CRITERIA:
- Type assertion expressions parse correctly with proper precedence
- Constraint expressions and 'where' clauses are captured in AST
- Type assertions work correctly in all expression contexts
- Complex nested assertions and constraints parse properly

This enables runtime type checking and validation in typed-Perl code.
```

### Step 10: Simple Union Types ✅ COMPLETED

```
You are continuing the parser testing implementation from Step 9.

CONTEXT: Type assertions are working. Now implement comprehensive test coverage for union type expressions (Type1|Type2).

TASK: Create exhaustive test coverage for union type syntax in variable declarations and method signatures.

REQUIREMENTS:
1. Add test fixtures for simple union types (Int|Str)
2. Add test fixtures for multi-way unions (Int|Str|Bool)
3. Add test fixtures for union types in method parameters and returns
4. Add test fixtures for union types with custom types
5. Test union type parsing with whitespace variations
6. Test union types in complex type expressions
7. Verify union type information is properly structured in AST

TEST EXAMPLES TO COVER:
```perl
# Simple union types
my Int|Str $flexible = 42;
my Bool|Undef $maybe_flag;

# Multi-way unions
my Int|Str|Bool $multi = "text";
my Num|ArrayRef|HashRef $complex;

# Method signatures with unions
method process(Int|Str $input) -> Bool|Str {
    if (ref $input) {
        return "Invalid";
    }
    return 1;
}

# Union types with custom types
my MyClass|OtherClass $object;
my Package::Type1|Package::Type2 $qualified;

# Whitespace variations
my Int | Str $spaced;
my Int|Str|Bool $compact;

# Complex expressions
my ArrayRef[Int|Str] @mixed_array;
field HashRef[Int|Bool] $mixed_hash;

# Nested contexts
method complex(
    ArrayRef[Int|Str] $data,
    CodeRef|Undef $callback
) -> HashRef[Bool|Str] {
    return {};
}
```

TECHNICAL DETAILS:
- Extend type expression parsing to handle pipe (|) union operator
- Test union type precedence and associativity
- Verify union types work correctly in all type annotation contexts
- Include edge cases like single-element unions and nested unions
- Test that union type structure is properly represented in AST

DELIVERABLES:
- testdata/typed-perl/union-types/ directory with union type fixtures
- Union type parsing tests with AST structure validation
- Precedence and whitespace handling tests
- Documentation of union type parsing coverage

SUCCESS CRITERIA:
- Union type expressions parse correctly with proper operator handling
- Multi-way unions are properly structured in AST
- Union types work in all contexts (variables, methods, fields)
- Whitespace and formatting variations are handled correctly

This enables flexible type specifications that can accept multiple types.
```

### Step 11: Basic Parameterized Types ✅ COMPLETED

```
You are continuing the parser testing implementation from Step 10.

CONTEXT: Union types are working. Now implement comprehensive test coverage for parameterized type expressions like ArrayRef[Type] and HashRef[Type].

TASK: Create exhaustive test coverage for parameterized type syntax and nested type parameters.

REQUIREMENTS:
1. Add test fixtures for basic parameterized types (ArrayRef[Int], HashRef[Str])
2. Add test fixtures for multiple type parameters (Map[Str, Int])
3. Add test fixtures for nested parameterized types (ArrayRef[ArrayRef[Int]])
4. Add test fixtures for parameterized types with union parameters
5. Test custom parameterized types (MyContainer[Type])
6. Test whitespace handling in parameterized type expressions
7. Verify parameterized type structure is correctly represented in AST

TEST EXAMPLES TO COVER:
```perl
# Basic parameterized types
my ArrayRef[Int] @numbers;
my HashRef[Str] %strings;
my CodeRef[Int, Str] $function;

# Multiple parameters
my Map[Str, Int] %mapping;
my Tuple[Int, Str, Bool] $triple;

# Nested parameterization
my ArrayRef[ArrayRef[Int]] @matrix;
my HashRef[ArrayRef[Str]] %grouped_strings;
my ArrayRef[HashRef[Int]] @array_of_hashes;

# Parameterized with unions
my ArrayRef[Int|Str] @mixed;
my HashRef[Bool|Undef] %flags;

# Custom parameterized types
my Container[MyType] $custom_container;
my Package::Generic[Int] $qualified;

# Method signatures
method process(ArrayRef[Str] $input) -> HashRef[Int] {
    my %result;
    return \%result;
}

# Field declarations
field ArrayRef[MyClass] $objects;
field HashRef[ArrayRef[Str]] $nested_data;

# Complex nesting
my Map[Str, ArrayRef[HashRef[Int|Bool]]] $complex;
```

TECHNICAL DETAILS:
- Extend type expression parsing to handle bracket notation for parameters
- Test parameter list parsing with comma separation
- Verify nested parameterization parses correctly with proper precedence
- Include edge cases like empty parameter lists and single parameters
- Test that parameter type information is preserved in AST hierarchy

DELIVERABLES:
- testdata/typed-perl/parameterized-types/ directory with comprehensive fixtures
- Parameterized type parsing tests with nested structure validation
- Parameter list and bracket notation tests
- Documentation of parameterized type parsing coverage

SUCCESS CRITERIA:
- Basic parameterized types parse correctly with parameters captured
- Nested parameterization works with proper AST structure
- Multiple parameters are correctly parsed and ordered
- Complex combinations with unions and custom types work properly

This enables generic programming with type-safe containers and collections.
```

### Step 12: Enhanced Type Keyword Recognition ✅ COMPLETED

```
You are continuing the parser implementation from Step 11.

CONTEXT: Comprehensive test coverage for basic type features is complete. Now begin parser improvements to handle all the type expressions we've tested.

TASK: Enhance the scanner and parser to properly recognize and tokenize all type-related keywords and operators.

REQUIREMENTS:
1. Update scanner to recognize all built-in type keywords (Int, Str, Bool, Num, etc.)
2. Add proper tokenization for type operators (|, as, where)
3. Enhance bracket and parenthesis handling for parameterized types
4. Improve whitespace handling in type expressions
5. Add better error recovery for malformed type expressions
6. Ensure type keywords don't conflict with variable names
7. Update token type definitions and string representations

TECHNICAL DETAILS:
- Review existing TokenType enum and add missing type-related tokens
- Update scanner.go to properly tokenize type keywords and operators
- Ensure type keyword recognition is context-aware (not in strings, etc.)
- Test that existing non-type code still tokenizes correctly
- Add position tracking for all type-related tokens

TOKEN UPDATES NEEDED:
```go
TokenTypeKeyword    // for 'type' keyword
TokenFieldKeyword   // for 'field' keyword
TokenMethodKeyword  // for 'method' keyword
TokenAsKeyword      // for 'as' type assertion
TokenWhereKeyword   // for 'where' constraints
TokenPipe           // for '|' union operator
TokenArrow          // for '->' return type annotation
// Built-in type names as identifiers or keywords
```

DELIVERABLES:
- Updated scanner token definitions with all type-related tokens
- Enhanced tokenization logic for type keywords and operators
- Improved bracket and operator precedence handling
- Scanner tests validating correct tokenization of type expressions
- Documentation of token enhancements

SUCCESS CRITERIA:
- All type keywords are properly recognized and tokenized
- Type operators have correct precedence and associativity
- Existing non-type code tokenization is unaffected
- Error recovery for malformed type expressions works correctly

This provides the tokenization foundation for improved type expression parsing.
```

### Step 13: Improved Type Expression Parsing ✅ COMPLETED

```
You are continuing the parser implementation from Step 12.

CONTEXT: Enhanced token recognition for type expressions is complete. Now improve the parser to correctly build AST nodes for all type expression patterns.

TASK: Enhance the parser to correctly parse and build AST structures for union types, parameterized types, and type assertions.

REQUIREMENTS:
1. Implement proper parsing for union type expressions (Type1|Type2)
2. Implement parameterized type parsing with bracket notation
3. Enhance type assertion parsing with 'as' keyword
4. Add type constraint parsing with 'where' clauses
5. Implement proper precedence for type operators
6. Ensure type expressions integrate correctly with variable declarations
7. Update AST node structures to properly represent type information

TECHNICAL DETAILS:
- Update parseTypeExpression() to handle union and parameterized types
- Implement parseTypeParameterList() for parameterized type arguments
- Add parseTypeAssertion() for runtime type checking expressions
- Ensure type expression parsing integrates with existing variable parsing
- Update AST node types to include comprehensive type information

AST NODE UPDATES:
```go
type TypeExpression struct {
    Kind       TypeExpressionKind  // Simple, Union, Parameterized, etc.
    Name       string              // Base type name
    Parameters []TypeExpression    // For parameterized types
    UnionTypes []TypeExpression    // For union types
    Constraint Expression          // For where clauses
    Position   Position            // Source position
}

type TypeAssertion struct {
    Expression Expression         // The expression being asserted
    Type       TypeExpression     // The target type
    Position   Position          // Source position
}
```

DELIVERABLES:
- Enhanced type expression parsing in parser.go
- Updated AST node definitions for comprehensive type representation
- Type expression integration with variable and method parsing
- Parser tests validating correct AST structure for all type patterns
- Documentation of type expression parsing improvements

SUCCESS CRITERIA:
- All tested type expression patterns produce correct AST structures
- Type information is properly integrated with variable and method nodes
- Parser handles complex nested type expressions correctly
- Existing untyped code parsing is unaffected

This enables the parser to correctly handle all the type expressions we've tested.
```

### Step 14: Better Error Recovery and Position Tracking ✅ COMPLETED

```
You are continuing the parser implementation from Step 13.

CONTEXT: Type expression parsing is enhanced. Now improve error recovery and position tracking for better debugging and user experience.

TASK: Enhance parser error recovery for malformed type expressions and improve position tracking accuracy.

REQUIREMENTS:
1. Implement better error recovery for incomplete type expressions
2. Add specific error messages for common type syntax mistakes
3. Improve position tracking for all type-related AST nodes
4. Implement error recovery for malformed parameterized types
5. Add helpful error suggestions for type-related syntax errors
6. Ensure errors don't cause parser to fail catastrophically
7. Test error recovery with intentionally malformed input

ERROR RECOVERY SCENARIOS:
```perl
# Missing closing bracket
my ArrayRef[Int $var;

# Invalid union syntax
my Int||Str $bad_union;

# Malformed type assertion
my $val as ;

# Invalid parameterized type
my ArrayRef[ $incomplete;

# Missing type in annotation
my $var;

# Invalid where clause
my Int where $invalid;
```

TECHNICAL DETAILS:
- Implement synchronization points for error recovery
- Add specific error messages for type-related syntax errors
- Ensure position information is accurate for both valid and error cases
- Test that error recovery allows parsing to continue correctly
- Add error reporting utilities for type expression context

ERROR HANDLING IMPROVEMENTS:
```go
type TypeError struct {
    Message   string
    Position  Position
    Suggestion string    // Helpful suggestion for fix
    Context   string     // Type expression context
}

func (p *Parser) recoverTypeExpression() {
    // Skip to next statement or known synchronization point
    // Provide helpful error message about type syntax
}
```

DELIVERABLES:
- Enhanced error recovery logic for type expressions
- Improved position tracking for all type-related AST nodes
- Specific error messages and suggestions for type syntax errors
- Error recovery tests with malformed type expressions
- Documentation of error handling improvements

SUCCESS CRITERIA:
- Parser recovers gracefully from malformed type expressions
- Error messages are helpful and include position information
- Position tracking is accurate for both valid and error cases
- Error recovery allows parsing to continue successfully

This improves the development experience when working with type annotations.
```

### Step 15: Complex Type Expression Support ✅ COMPLETED

```
You are continuing the parser implementation from Step 14.

CONTEXT: Error recovery and position tracking are improved. Now add support for complex type expressions that combine multiple type features.

TASK: Enhance parser to handle complex combinations of type features in single expressions.

REQUIREMENTS:
1. Support nested union types within parameterized types
2. Support parameterized types within union expressions
3. Handle deep nesting of parameterized types
4. Support type assertions with complex type expressions
5. Handle method signatures with complex parameter and return types
6. Test performance with deeply nested type expressions
7. Ensure complex types integrate correctly with all language constructs

COMPLEX TYPE EXAMPLES:
```perl
# Nested unions in parameterized types
my ArrayRef[Int|Str|Bool] @complex_array;
my HashRef[ArrayRef[Int]|HashRef[Str]] %nested_complex;

# Parameterized unions
my (ArrayRef[Int]|HashRef[Str]) $param_union;

# Deep nesting
my ArrayRef[HashRef[ArrayRef[Int|Str]]] @deep_nested;
my Map[Str, ArrayRef[Tuple[Int, Bool|Str]]] %complex_map;

# Complex method signatures
method transform(
    ArrayRef[HashRef[Int|Str]] $input,
    CodeRef[Str, Bool] $validator
) -> HashRef[ArrayRef[Int]|Str] {
    return {};
}

# Complex type assertions
my $result = $data as ArrayRef[HashRef[Int|Bool]];
my $complex = ($input->process()) as Map[Str, ArrayRef[MyType]];

# Field with complex types
field Map[Str, ArrayRef[Int|Str]|HashRef[Bool]] $complex_field;
```

TECHNICAL DETAILS:
- Ensure recursive type expression parsing handles arbitrary nesting depth
- Test memory usage and performance with complex nested types
- Verify AST structure correctly represents deep nesting
- Include stress tests with very complex type expressions
- Ensure complex types work in all contexts (variables, methods, fields)

PERFORMANCE CONSIDERATIONS:
- Test parsing performance with deeply nested types
- Ensure memory usage remains reasonable
- Add safeguards against infinite recursion in type parsing
- Test with realistic complex type expressions from real code

DELIVERABLES:
- Enhanced parser support for arbitrary type expression complexity
- Performance tests for complex type expression parsing
- Stress tests with deeply nested and complex type combinations
- AST validation for complex type structure representation
- Documentation of complex type expression support

SUCCESS CRITERIA:
- Parser handles arbitrarily complex type expressions correctly
- Performance remains acceptable with realistic complex types
- AST structure accurately represents complex type hierarchies
- All language constructs work correctly with complex types

This enables sophisticated type expressions for advanced typed-Perl programming.
```

### Step 16: AST Enhancements for Type Information ✅ COMPLETED

```
You are continuing the parser implementation from Step 15.

CONTEXT: Complex type expression parsing is complete. Now enhance AST node structures to comprehensively represent type information for tool integration.

TASK: Update AST node definitions to include complete type information that supports advanced tooling and analysis.

REQUIREMENTS:
1. Update all AST nodes to include comprehensive type information
2. Add type information to variable declaration nodes
3. Enhance method and field nodes with complete type signatures
4. Add type assertion information to expression nodes
5. Implement type information visitor patterns for AST traversal
6. Add type information serialization for tool integration
7. Ensure type information is optional and doesn't break untyped code

AST ENHANCEMENTS NEEDED:
```go
// Enhanced variable declaration node
type VariableDeclaration struct {
    BaseNode
    Scope      string           // my, our, state, etc.
    Name       string           // variable name
    Type       *TypeExpression  // type annotation (optional)
    Value      Expression       // initial value (optional)
    Position   Position         // source position
}

// Enhanced method definition node
type MethodDefinition struct {
    BaseNode
    Name         string              // method name
    Parameters   []ParameterInfo     // typed parameters
    ReturnType   *TypeExpression     // return type (optional)
    Body         *BlockStatement     // method body
    Position     Position            // source position
}

type ParameterInfo struct {
    Name     string           // parameter name
    Type     *TypeExpression  // parameter type (optional)
    Default  Expression       // default value (optional)
    Position Position         // source position
}

// Enhanced expression node for type assertions
type TypeAssertionExpression struct {
    BaseNode
    Expression Expression      // expression being asserted
    TargetType TypeExpression  // target type
    Position   Position        // source position
}
```

TECHNICAL DETAILS:
- Update all relevant AST node types to include type information
- Ensure type information is optional to maintain backward compatibility
- Add visitor patterns for traversing type information in AST
- Test that type information serializes correctly for tool integration
- Verify that untyped code produces AST nodes without type information

VISITOR PATTERN SUPPORT:
```go
type TypeVisitor interface {
    VisitTypeExpression(node *TypeExpression) error
    VisitTypedVariable(node *VariableDeclaration) error
    VisitTypedMethod(node *MethodDefinition) error
    VisitTypeAssertion(node *TypeAssertionExpression) error
}

func (ast *AST) WalkTypes(visitor TypeVisitor) error {
    // Traverse AST and call visitor methods for nodes with type information
}
```

DELIVERABLES:
- Updated AST node definitions with comprehensive type information
- Type information visitor patterns for AST traversal
- Type information serialization and deserialization support
- Tests validating type information in AST nodes
- Documentation of AST type information enhancements

SUCCESS CRITERIA:
- All AST nodes include appropriate type information where relevant
- Type information is optional and doesn't affect untyped code
- Visitor patterns enable easy traversal of type information
- Type information serializes correctly for external tool integration

This provides the AST foundation for advanced type-aware tooling and analysis.
```

### Step 17: Complex Union and Intersection Types ✅ COMPLETED

```
You are continuing the parser implementation from Step 16.

CONTEXT: AST enhancements for type information are complete. Now add support for advanced type expressions including intersection types and complex union combinations.

TASK: Implement parsing support for intersection types (Type1&Type2) and advanced union type patterns.

REQUIREMENTS:
1. Add support for intersection type syntax (Type1&Type2)
2. Implement parsing for complex union and intersection combinations
3. Add support for negation types (!Type)
4. Handle precedence between union, intersection, and negation operators
5. Test complex combinations of all type operators
6. Ensure intersection types work in all type annotation contexts
7. Update AST nodes to represent intersection and negation type information

ADVANCED TYPE EXAMPLES:
```perl
# Intersection types
my Object&Serializable $serializable_object;
my Readable&Writable $file_handle;

# Complex combinations
my (Int|Str)&Defined $defined_value;
my ArrayRef[Object&Clonable] @clonable_objects;

# Negation types
my !Undef $definitely_defined;
my ArrayRef[!Str] @non_strings;

# Complex precedence
my Int|Str&Defined $complex;  # (Int|(Str&Defined))
my (Int|Str)&Defined $grouped;

# Method signatures with advanced types
method process(
    Object&Serializable $input,
    !Undef $required
) -> ArrayRef[Int|Str]&Defined {
    return [];
}

# Field declarations
field Readable&Writable $handle;
field ArrayRef[Object&Clonable] $objects;

# Type assertions
my $obj = $input as Object&Serializable;
my $list = $data as ArrayRef[!Undef];
```

TECHNICAL DETAILS:
- Add intersection (&) and negation (!) operators to type expression parsing
- Implement proper precedence: negation > intersection > union
- Update TypeExpression AST node to include intersection and negation info
- Test operator precedence and associativity extensively
- Ensure parenthesized grouping works correctly

AST UPDATES:
```go
type TypeExpression struct {
    Kind           TypeExpressionKind  // Simple, Union, Intersection, Negation, etc.
    Name           string              // Base type name
    Parameters     []TypeExpression    // For parameterized types
    UnionTypes     []TypeExpression    // For union types
    IntersectionTypes []TypeExpression // For intersection types
    NegatedType    *TypeExpression     // For negation types
    Constraint     Expression          // For where clauses
    Position       Position            // Source position
}
```

DELIVERABLES:
- Intersection and negation type parsing implementation
- Updated AST nodes for advanced type expressions
- Operator precedence and associativity tests
- Complex type combination parsing tests
- Documentation of advanced type expression support

SUCCESS CRITERIA:
- Intersection types parse correctly with proper precedence
- Negation types work in all type annotation contexts
- Complex combinations of type operators parse correctly
- AST accurately represents advanced type expression structure

This enables sophisticated type constraints and requirements for advanced typing.
```

### Step 18: Nested Parameterized Types ✅ TODO

```
You are continuing the parser implementation from Step 17.

CONTEXT: Advanced union and intersection types are working. Now enhance support for deeply nested parameterized types and complex generic expressions.

TASK: Improve parsing of nested parameterized types and complex generic type expressions.

REQUIREMENTS:
1. Enhance support for deeply nested parameterized types
2. Implement better bracket matching and error recovery for nested generics
3. Add support for multiple type parameter constraints
4. Handle complex generic method signatures
5. Test performance and memory usage with deep nesting
6. Support generic type aliases and custom parameterized types
7. Ensure nested types work correctly with union and intersection operators

COMPLEX NESTED TYPE EXAMPLES:
```perl
# Deep nesting
my ArrayRef[HashRef[ArrayRef[Int|Str]]] @deep_structure;
my Map[Str, Tuple[ArrayRef[Int], HashRef[Bool|Undef]]] %complex_mapping;

# Generic methods with constraints
method transform<T, U>(
    ArrayRef[T] $input,
    CodeRef[T, U] $transformer
) -> ArrayRef[U] where T: Serializable, U: Defined {
    return [];
}

# Custom parameterized types
type Container[T] = ArrayRef[T]|HashRef[T];
my Container[MyClass] $flexible_container;

# Nested with advanced operators
my ArrayRef[Object&Serializable] @serializable_list;
my HashRef[ArrayRef[Int|Str]&Defined] %defined_arrays;

# Complex type aliases
type EventHandler[T] = CodeRef[T, Bool|Str];
type DataStore[K, V] = HashRef[ArrayRef[Tuple[K, V]]];

my DataStore[Str, MyClass] %store;
my EventHandler[ClickEvent] $click_handler;

# Recursive type definitions
type Tree[T] = T|ArrayRef[Tree[T]];
my Tree[Int] $number_tree;
```

TECHNICAL DETAILS:
- Improve bracket matching and nesting depth tracking
- Add performance limits to prevent excessive nesting
- Test memory usage and parsing speed with complex nested types
- Ensure error recovery works correctly with malformed nested types
- Support generic type constraints and bounds

PERFORMANCE AND SAFETY:
- Add maximum nesting depth limits to prevent stack overflow
- Implement efficient bracket matching algorithms
- Test parsing performance with realistic complex types
- Add memory usage monitoring for complex type parsing
- Ensure error recovery doesn't cause infinite loops

DELIVERABLES:
- Enhanced nested parameterized type parsing support
- Performance and safety limits for complex type expressions
- Generic type constraint and bounds support
- Complex nested type parsing tests and benchmarks
- Documentation of nested type capabilities and limits

SUCCESS CRITERIA:
- Deep nested parameterized types parse correctly and efficiently
- Performance remains acceptable with realistic complex nesting
- Error recovery works correctly with malformed nested types
- Generic constraints and bounds are properly supported

This enables sophisticated generic programming patterns in typed-Perl.
```

### Step 19: Method Signature Parsing ✅ TODO

```
You are continuing the parser implementation from Step 18.

CONTEXT: Nested parameterized types are working well. Now enhance method signature parsing to handle complex parameter lists and return type specifications.

TASK: Implement comprehensive method signature parsing with advanced parameter handling and return type specifications.

REQUIREMENTS:
1. Enhance method parameter parsing to handle complex type expressions
2. Add support for optional parameters with default values
3. Implement named parameter syntax
4. Add support for variadic parameters and parameter unpacking
5. Enhance return type specification parsing
6. Support method type constraints and generic parameters
7. Test method signature parsing with all type expression combinations

ADVANCED METHOD SIGNATURE EXAMPLES:
```perl
# Complex parameter types
method process(
    ArrayRef[HashRef[Int|Str]] $data,
    Optional[CodeRef[Str, Bool]] $validator = undef,
    Slurpy[HashRef[Any]] %options
) -> Result[ArrayRef[ProcessedItem], ErrorCode] {
    return success([]);
}

# Named parameters
method configure(
    :$host as Str,
    :$port as Int = 8080,
    :$ssl as Bool = 0,
    :$timeout as Optional[Num]
) -> ConnectionConfig {
    return ConnectionConfig->new;
}

# Generic method signatures
method map<T, U>(
    ArrayRef[T] $input,
    CodeRef[T, U] $transformer
) -> ArrayRef[U] where T: Defined, U: Serializable {
    return [];
}

# Variadic parameters
method sum(Int *@numbers) -> Int {
    my $total = 0;
    $total += $_ for @numbers;
    return $total;
}

# Complex return types
method get_data() -> (
    ArrayRef[UserRecord]|ErrorResponse,
    Optional[MetaData]
) {
    return ([], undef);
}

# Method with all features
method complex_method<T>(
    Required[T] $input,
    Optional[CodeRef[T, Bool]] $validator = undef,
    :$timeout as Num = 30.0,
    Slurpy[Any] *@rest
) -> Result[T, ProcessingError] where T: Serializable {
    return success($input);
}
```

TECHNICAL DETAILS:
- Extend method parsing to handle complex parameter type expressions
- Implement optional parameter parsing with default value expressions
- Add named parameter syntax support (:$param)
- Support variadic parameters (*@args) and parameter unpacking
- Parse complex return type specifications including tuples

METHOD SIGNATURE AST UPDATES:
```go
type MethodSignature struct {
    Name           string               // method name
    TypeParameters []TypeParameter      // generic type parameters
    Parameters     []MethodParameter    // method parameters
    ReturnType     *TypeExpression      // return type specification
    Constraints    []TypeConstraint     // type constraints
    Position       Position             // source position
}

type MethodParameter struct {
    Name       string           // parameter name
    Type       *TypeExpression  // parameter type
    IsOptional bool            // optional parameter flag
    IsNamed    bool            // named parameter flag (:$param)
    IsVariadic bool            // variadic parameter flag (*@args)
    Default    Expression      // default value expression
    Position   Position        // source position
}
```

DELIVERABLES:
- Enhanced method signature parsing with all parameter features
- Support for generic method parameters and type constraints
- Complex return type specification parsing
- Method signature AST node enhancements
- Comprehensive method signature parsing tests

SUCCESS CRITERIA:
- All method parameter types parse correctly including complex expressions
- Optional, named, and variadic parameters work properly
- Generic method signatures with constraints parse correctly
- Complex return type specifications are properly captured

This enables sophisticated method definitions with advanced parameter handling.
```

### Step 20: Class and Role Declarations ✅ TODO

```
You are continuing the parser implementation from Step 19.

CONTEXT: Method signature parsing is comprehensive. Now implement parsing for class and role declarations with type information.

TASK: Add comprehensive parsing support for class and role declarations with typed fields, methods, and inheritance.

REQUIREMENTS:
1. Implement class declaration parsing with typed fields
2. Add role declaration parsing with type signatures
3. Support inheritance and role composition with type constraints
4. Parse class and role type parameters (generic classes)
5. Handle access modifiers and field visibility
6. Support constructor and destructor method parsing
7. Test class and role parsing with all previously implemented type features

CLASS AND ROLE EXAMPLES:
```perl
# Basic class with typed fields
class User {
    field Str $name;
    field Int $age;
    field Optional[Email] $email;

    method new(Str $name, Int $age) -> User {
        return bless {
            name => $name,
            age => $age
        }, __PACKAGE__;
    }

    method get_name() -> Str {
        return $name;
    }
}

# Generic class
class Container<T> where T: Serializable {
    field ArrayRef[T] $items = [];

    method add(T $item) -> Void {
        push @{$items}, $item;
    }

    method get_all() -> ArrayRef[T] {
        return $items;
    }
}

# Role with type signatures
role Serializable {
    method serialize() -> Str;
    method deserialize(Str $data) -> Self;
}

# Class with inheritance and roles
class Document : BaseDocument does Serializable, Cacheable {
    field Str $content;
    field DateTime $created;
    field Optional[UserRef] $author;

    method serialize() -> Str {
        return encode_json({
            content => $content,
            created => $created->iso8601,
            author => $author ? $author->id : undef
        });
    }
}

# Complex inheritance with type constraints
class ProcessingQueue<T> : BaseQueue<T>
    where T: Serializable&Processable {

    field CodeRef[T, ProcessResult] $processor;
    field ArrayRef[T] $pending = [];

    method process_all() -> ArrayRef[ProcessResult] {
        return [ map { $processor->($_) } @{$pending} ];
    }
}
```

TECHNICAL DETAILS:
- Implement class and role declaration parsing
- Support generic type parameters on classes and roles
- Parse inheritance relationships with type constraints
- Handle field declarations with access modifiers
- Support role composition and multiple inheritance

CLASS/ROLE AST NODES:
```go
type ClassDeclaration struct {
    BaseNode
    Name           string              // class name
    TypeParameters []TypeParameter     // generic parameters
    Superclass     *TypeExpression     // parent class
    Roles          []TypeExpression    // implemented roles
    Fields         []FieldDeclaration  // class fields
    Methods        []MethodDefinition  // class methods
    Constraints    []TypeConstraint    // type constraints
    Position       Position            // source position
}

type RoleDeclaration struct {
    BaseNode
    Name           string              // role name
    TypeParameters []TypeParameter     // generic parameters
    RequiredMethods []MethodSignature  // required method signatures
    ProvidedMethods []MethodDefinition // provided method implementations
    Fields         []FieldDeclaration  // role fields
    Position       Position            // source position
}
```

DELIVERABLES:
- Class and role declaration parsing implementation
- Support for generic classes and roles with type constraints
- Inheritance and role composition parsing
- Class and role AST node definitions
- Comprehensive class and role parsing tests

SUCCESS CRITERIA:
- Class declarations parse correctly with all type features
- Role declarations and composition work properly
- Generic classes and roles with constraints parse correctly
- Inheritance relationships are properly captured in AST

This enables full object-oriented programming with sophisticated typing in typed-Perl.
```

### Step 21: Advanced Type Constraints ✅ TODO

```
You are continuing the parser implementation from Step 20.

CONTEXT: Class and role declarations are working. Now implement advanced type constraint parsing for sophisticated type system features.

TASK: Add comprehensive support for type constraints, bounds, and advanced type system features.

REQUIREMENTS:
1. Implement where clause parsing for type constraints
2. Add support for multiple constraint types (subtype, protocol, value)
3. Parse constraint expressions with complex boolean logic
4. Support type bounds on generic parameters
5. Handle constraint inheritance and composition
6. Add constraint validation expressions
7. Test constraint parsing with all type expression combinations

ADVANCED TYPE CONSTRAINT EXAMPLES:
```perl
# Basic type constraints
method process<T>(ArrayRef[T] $data) -> ArrayRef[T]
    where T: Serializable {
    return $data;
}

# Multiple constraints
method transform<T, U>(T $input) -> U
    where T: Serializable&Defined,
          U: Deserializable&!Undef {
    return deserialize($input->serialize());
}

# Value constraints
method create_array<T>(Int $size) -> ArrayRef[T]
    where T: Any,
          $size > 0 && $size < 1000 {
    return [(undef) x $size];
}

# Protocol constraints
method handle<T>(T $object) -> ProcessResult
    where T does EventHandler,
          T can 'process',
          T->VERSION >= 1.5 {
    return $object->process();
}

# Complex constraint expressions
method complex<T, U>(T $input, CodeRef[T, U] $transform) -> U
    where T: (Serializable|Storable)&Defined,
          U: !Undef,
          $transform isa 'CodeRef',
          T->can('serialize') || T->can('store') {
    return $transform->($input);
}

# Class constraints
class Container<T> where T: Clonable&Serializable {
    field ArrayRef[T] $items;

    method add_cloned(T $item) -> Void
        where $item->can('clone') {
        push @{$items}, $item->clone();
    }
}

# Constraint inheritance
role Processable<T> where T: Defined {
    method process(T $input) -> ProcessResult;
}

class DataProcessor<T> does Processable<T>
    where T: Serializable&Defined {
    # Inherits T: Defined from role
    # Adds T: Serializable requirement
}
```

TECHNICAL DETAILS:
- Implement where clause parsing with constraint expressions
- Support multiple constraint types (type, protocol, value, capability)
- Parse complex boolean logic in constraint expressions
- Handle constraint inheritance from roles and parent classes
- Support runtime constraint validation expressions

CONSTRAINT AST NODES:
```go
type TypeConstraint struct {
    Parameter   string              // type parameter being constrained
    Kind        ConstraintKind      // type, protocol, value, capability
    Expression  Expression          // constraint expression
    Position    Position            // source position
}

type ConstraintKind int
const (
    TypeConstraint ConstraintKind = iota    // T: SomeType
    ProtocolConstraint                      // T does SomeRole
    CapabilityConstraint                    // T can 'method'
    ValueConstraint                         // $param > 0
    VersionConstraint                       // T->VERSION >= 1.0
)

type WhereClause struct {
    Constraints []TypeConstraint    // list of constraints
    Position    Position            // source position
}
```

DELIVERABLES:
- Advanced type constraint parsing implementation
- Support for multiple constraint types and complex expressions
- Constraint inheritance and composition handling
- Constraint AST node definitions and validation
- Comprehensive constraint parsing tests

SUCCESS CRITERIA:
- All constraint types parse correctly with proper AST representation
- Complex constraint expressions with boolean logic work properly
- Constraint inheritance from roles and classes works correctly
- Runtime constraint validation expressions are properly captured

This enables sophisticated type system features for advanced type safety and validation.
```

### Step 22: End-to-End Integration Testing ✅ TODO

```
You are completing the parser implementation improvements from Step 21.

CONTEXT: All individual type annotation features are implemented and tested. Now create comprehensive end-to-end integration tests that validate the complete type system works together.

TASK: Create comprehensive integration tests that validate complete typed-Perl programs parse correctly with all features working together.

REQUIREMENTS:
1. Create realistic typed-Perl program examples using all features
2. Test integration between type annotations, classes, roles, and methods
3. Validate complex programs with nested types and inheritance
4. Test mixed typed and untyped code in the same program
5. Create performance tests with large typed programs
6. Validate AST correctness for complete programs
7. Test tool integration (type checker, LSP, etc.) with enhanced parser

COMPREHENSIVE INTEGRATION EXAMPLES:
```perl
# Complete typed-Perl program
use v5.38;
use strict;
use warnings;

# Type definitions
type UserId = Int where { $_ > 0 };
type Email = Str where { $_ =~ /\@/ };
type Result<T, E> = Success<T> | Failure<E>;

# Role definitions
role Serializable {
    method serialize() -> Str;
    method deserialize(Str $data) -> Self;
}

role Cacheable<K> where K: Serializable {
    field Optional[DateTime] $cached_at;
    method cache_key() -> K;
    method is_stale() -> Bool;
}

# Class hierarchy
class User does Serializable, Cacheable<UserId> {
    field UserId $id;
    field Str $name;
    field Email $email;
    field ArrayRef[Role] $roles = [];

    method new(UserId $id, Str $name, Email $email) -> User {
        return bless {
            id => $id,
            name => $name,
            email => $email,
            roles => []
        }, __PACKAGE__;
    }

    method add_role(Role $role) -> Void where $role->is_valid() {
        push @{$roles}, $role;
    }

    method serialize() -> Str {
        return encode_json({
            id => $id,
            name => $name,
            email => $email,
            roles => [map { $_->serialize() } @{$roles}]
        });
    }

    method cache_key() -> UserId {
        return $id;
    }
}

# Generic service class
class UserService<T> where T: User&Cacheable<UserId> {
    field HashRef[UserId, T] $cache = {};
    field CodeRef[UserId, Optional[T]] $loader;

    method new(CodeRef[UserId, Optional[T]] $loader) -> UserService<T> {
        return bless { cache => {}, loader => $loader }, __PACKAGE__;
    }

    method get(UserId $id) -> Result<T, Str> {
        # Check cache first
        if (exists $cache->{$id} && !$cache->{$id}->is_stale()) {
            return Success->new($cache->{$id});
        }

        # Load from source
        my $user = $loader->($id);
        return Failure->new("User not found") unless defined $user;

        # Cache and return
        $cache->{$id} = $user;
        return Success->new($user);
    }

    method invalidate(UserId $id) -> Void {
        delete $cache->{$id};
    }
}

# Complex method with all features
method process_users<T>(
    ArrayRef[T] $users,
    CodeRef[T, Bool] $filter,
    Optional[CodeRef[T, T]] $transform = undef,
    :$batch_size as Int = 100,
    Slurpy[Any] *@options
) -> Result<ArrayRef[ProcessedUser], ProcessingError>
    where T: User&Serializable,
          $batch_size > 0 && $batch_size <= 1000 {

    my @results;
    my @batch;

    for my $user (@{$users}) {
        next unless $filter->($user);

        my $processed = $transform ? $transform->($user) : $user;
        push @batch, ProcessedUser->from_user($processed as T);

        if (@batch >= $batch_size) {
            push @results, @batch;
            @batch = ();
        }
    }

    push @results, @batch if @batch;
    return Success->new(\@results);
}
```

TECHNICAL DETAILS:
- Test complete programs that use all implemented type features
- Validate AST structure for complex programs
- Test parsing performance with large typed programs
- Ensure mixed typed/untyped code works correctly
- Test integration with type checker and other tools

INTEGRATION TEST CATEGORIES:
1. **Feature Integration**: All type features working together
2. **Performance Tests**: Large programs with complex types
3. **Mixed Code Tests**: Typed and untyped code in same program
4. **Tool Integration**: Parser output works with type checker, LSP
5. **Regression Tests**: Ensure improvements don't break existing functionality
6. **Error Recovery**: Complex programs with syntax errors

DELIVERABLES:
- Comprehensive end-to-end integration test suite
- Realistic typed-Perl program examples
- Performance benchmarks for complex program parsing
- Tool integration validation tests
- Integration test documentation and examples

SUCCESS CRITERIA:
- Complete typed-Perl programs parse correctly with all features
- AST structure accurately represents complex program structure
- Performance is acceptable for realistic program sizes
- Tool integration works correctly with enhanced parser output

This validates that all implemented improvements work together correctly in real-world scenarios.
```

### Step 23: Performance and Regression Testing ✅ TODO

```
You are completing the parser implementation improvements from Step 22.

CONTEXT: End-to-end integration testing is complete. Now implement comprehensive performance testing and regression detection to ensure the improvements maintain good performance characteristics.

TASK: Create performance benchmarking and regression testing infrastructure to validate parser improvements don't degrade performance.

REQUIREMENTS:
1. Implement performance benchmarks for parsing speed and memory usage
2. Create regression tests comparing old vs new parser performance
3. Add stress testing with large files and complex type expressions
4. Implement memory usage monitoring and leak detection
5. Create performance baselines for different code patterns
6. Add automated performance regression detection
7. Test parsing performance across different complexity levels

PERFORMANCE TESTING SCENARIOS:
```perl
# Benchmark test cases of increasing complexity

# 1. Simple untyped Perl (baseline)
my $simple = "hello";
my @array = (1, 2, 3);
sub simple_function { return 42; }

# 2. Basic type annotations
my Int $typed_var = 42;
my ArrayRef[Str] @typed_array = ("a", "b");
method typed_method(Int $param) -> Str { return "$param"; }

# 3. Complex type expressions
my ArrayRef[HashRef[Int|Str]] @complex;
method complex_sig(
    ArrayRef[Object&Serializable] $input,
    CodeRef[Int, Bool|Str] $processor
) -> HashRef[ArrayRef[Int]|ErrorCode] { return {}; }

# 4. Large program simulation (generated test)
# - 1000+ variable declarations with types
# - 100+ method definitions with complex signatures
# - 50+ class definitions with inheritance
# - Deep nesting and complex type expressions

# 5. Stress test patterns
# - Very deep type nesting (10+ levels)
# - Very long union types (20+ alternatives)
# - Large method signatures (50+ parameters)
# - Complex constraint expressions
```

PERFORMANCE TESTING INFRASTRUCTURE:
```go
type PerformanceTest struct {
    Name        string
    InputFile   string
    MaxDuration time.Duration
    MaxMemory   int64
    Iterations  int
}

type PerformanceResult struct {
    TestName       string
    ParseDuration  time.Duration
    MemoryUsage    int64
    AllocCount     int64
    Success        bool
    Error          error
}

func BenchmarkParserPerformance(tests []PerformanceTest) []PerformanceResult {
    // Run performance tests and collect metrics
}

func DetectPerformanceRegression(current, baseline []PerformanceResult) bool {
    // Compare current results against baseline
    // Return true if significant regression detected
}
```

TECHNICAL DETAILS:
- Use Go's built-in benchmarking and profiling tools
- Create test files of various sizes and complexity levels
- Measure parsing time, memory allocation, and garbage collection impact
- Compare performance before and after type annotation improvements
- Set performance regression thresholds and automated detection

PERFORMANCE METRICS TO TRACK:
1. **Parsing Speed**: Lines per second, tokens per second
2. **Memory Usage**: Peak memory, allocation count, GC pressure
3. **Complexity Scaling**: How performance scales with program complexity
4. **Feature Impact**: Performance cost of each type annotation feature
5. **Regression Detection**: Automated alerts for performance degradation

DELIVERABLES:
- Comprehensive performance testing infrastructure
- Performance benchmarks for all complexity levels
- Regression detection and alerting system
- Performance optimization recommendations
- Performance testing documentation and baselines

SUCCESS CRITERIA:
- Parser performance meets or exceeds baseline expectations
- Memory usage remains within acceptable bounds
- Performance scales reasonably with program complexity
- Regression detection catches performance degradation automatically
- Performance testing can be run automatically in CI/CD

This ensures the type annotation improvements maintain good performance characteristics.
```

### Step 24: Backward Compatibility Validation ✅ TODO

```
You are completing the parser implementation improvements from Step 23.

CONTEXT: Performance and regression testing infrastructure is complete. Now validate that all improvements maintain complete backward compatibility with existing untyped Perl code.

TASK: Create comprehensive backward compatibility tests to ensure enhanced parser doesn't break existing functionality.

REQUIREMENTS:
1. Test existing untyped Perl code parses identically to before improvements
2. Validate that type annotations are completely optional
3. Test mixed typed and untyped code in same files
4. Ensure existing tools and integrations continue to work
5. Test that AST structure for untyped code is unchanged
6. Validate error handling for untyped code remains consistent
7. Test compatibility with existing Perl syntax edge cases

BACKWARD COMPATIBILITY TEST SCENARIOS:
```perl
# 1. Existing untyped Perl patterns that must continue working

# Variable declarations without types
my $var = "value";
our @array = (1, 2, 3);
local %hash = (key => 'value');

# Subroutines without type annotations
sub existing_function {
    my ($param1, $param2) = @_;
    return $param1 + $param2;
}

# Complex expressions and operators
my $result = ($a + $b) * ($c || 1) / ($d && $e);
my $string = "Hello " . $name . "!";

# Control structures
if ($condition) {
    do_something();
} elsif ($other) {
    do_other();
}

for my $item (@list) {
    process($item);
}

# Package and module usage
package MyPackage;
use strict;
use warnings;
use Data::Dumper;

# 2. Edge cases that might be affected by type parsing

# Variables that look like types
my $Int = "not a type";
my $ArrayRef = [];

# Methods that might conflict with type keywords
sub type { return "method"; }
sub method { return "function"; }

# Complex expressions that might confuse type parser
my $complex = $obj->method() + $other->as() * $value;

# Heredocs and string literals with type-like content
my $code = <<'END';
my Int $typed_var = 42;
method foo() -> Str { return "test"; }
END

# Comments with type annotations
# This is a comment with Int and ArrayRef[Str]
my $var = 42; # Not typed despite comment

# 3. Error cases that should still produce same errors
# Syntax errors in untyped code should have same error messages
my $unclosed = "string
my @bad_array = (1, 2, 3
```

COMPATIBILITY TESTING INFRASTRUCTURE:
```go
type CompatibilityTest struct {
    Name              string
    InputCode         string
    ExpectedAST       *AST           // AST from original parser
    ExpectedErrors    []string       // Expected error messages
    ShouldParse       bool           // Whether code should parse successfully
}

func ValidateBackwardCompatibility(tests []CompatibilityTest) []CompatibilityResult {
    // Parse with enhanced parser and compare results
    // Ensure AST structure matches original for untyped code
    // Verify error messages are consistent
}

type CompatibilityResult struct {
    TestName        string
    ASTMatches      bool              // AST structure identical
    ErrorsMatch     bool              // Error messages consistent
    Compatible      bool              // Overall compatibility
    Differences     []string          // List of detected differences
}
```

TECHNICAL DETAILS:
- Compare AST output between original and enhanced parser for untyped code
- Ensure error messages and error recovery behavior is identical
- Test that existing tools (type checker, LSP) work with enhanced parser
- Verify that parser performance for untyped code is not degraded
- Test edge cases where type-like syntax appears in untyped contexts

COMPATIBILITY VALIDATION AREAS:
1. **AST Structure**: Untyped code produces identical AST nodes
2. **Error Handling**: Same error messages and recovery behavior
3. **Performance**: No performance regression for untyped code
4. **Tool Integration**: Existing tools work without modification
5. **Edge Cases**: Complex syntax patterns continue to work
6. **Mixed Code**: Typed and untyped code can coexist

DELIVERABLES:
- Comprehensive backward compatibility test suite
- AST comparison utilities for detecting changes
- Compatibility validation infrastructure
- Documentation of compatibility guarantees
- Migration guide for any breaking changes (should be none)

SUCCESS CRITERIA:
- All existing untyped Perl code parses identically to before
- Error messages and recovery behavior is unchanged for untyped code
- Existing tools and integrations work without modification
- Performance for untyped code is maintained or improved
- Mixed typed/untyped code works seamlessly

This ensures the enhanced parser maintains complete backward compatibility while adding new type annotation capabilities.
```

### Step 25: Documentation and Final Integration ✅ TODO

```
You are completing the parser implementation improvements from Step 24.

CONTEXT: All parser improvements and testing are complete. Now create comprehensive documentation and perform final integration of all enhancements.

TASK: Document all improvements, create usage guides, and perform final integration to complete the type annotation detection enhancement project.

REQUIREMENTS:
1. Create comprehensive documentation for all type annotation features
2. Write usage guides and examples for developers
3. Document parser enhancement architecture and design decisions
4. Create migration guide for projects adopting type annotations
5. Integrate all improvements with existing PVM components
6. Update tool integrations (LSP, type checker) to use enhanced parser
7. Create final validation that all components work together

DOCUMENTATION DELIVERABLES:

## 1. Type Annotation Reference Guide
```markdown
# PVM Type Annotation Reference

## Basic Type Annotations
- Simple types: `my Int $var`
- Arrays and hashes: `my ArrayRef[Str] @array`
- Custom types: `my MyClass $object`

## Advanced Type Expressions
- Union types: `Int|Str|Bool`
- Parameterized types: `ArrayRef[Int]`, `HashRef[Str]`
- Intersection types: `Object&Serializable`
- Negation types: `!Undef`

## Method Signatures
- Parameter types: `method foo(Int $a, Str $b)`
- Return types: `method bar() -> ArrayRef[Int]`
- Generic methods: `method process<T>(T $input) -> T`

## Classes and Roles
- Typed fields: `field Str $name`
- Generic classes: `class Container<T>`
- Role composition: `class User does Serializable`

## Type Constraints
- Where clauses: `where T: Serializable`
- Multiple constraints: `where T: Defined&Clonable`
- Value constraints: `where $count > 0`
```

## 2. Parser Enhancement Architecture Documentation
```markdown
# Parser Enhancement Architecture

## Overview
The enhanced parser maintains full backward compatibility while adding comprehensive type annotation support.

## Key Components
1. **Enhanced Scanner**: Recognizes type keywords and operators
2. **Type Expression Parser**: Handles complex type expressions
3. **AST Integration**: Type information in all relevant nodes
4. **Error Recovery**: Graceful handling of malformed types

## Design Principles
- Backward compatibility: All existing code continues to work
- Optional typing: Type annotations are completely optional
- Incremental adoption: Projects can adopt types gradually
- Tool integration: Enhanced AST supports advanced tooling
```

## 3. Usage Examples and Best Practices
```perl
# Progressive typing adoption example

# Phase 1: Add basic type annotations
my Int $count = 0;
my Str $name = "example";

# Phase 2: Add method type signatures
method calculate(Int $a, Int $b) -> Int {
    return $a + $b;
}

# Phase 3: Use advanced type features
method process<T>(ArrayRef[T] $input) -> ArrayRef[T]
    where T: Serializable {
    return $input->map(sub { $_->clone() });
}

# Phase 4: Full class-based architecture
class DataProcessor<T> does Cacheable
    where T: Serializable&Defined {

    field ArrayRef[T] $items = [];

    method add(T $item) -> Void {
        push @{$items}, $item;
    }
}
```

FINAL INTEGRATION TASKS:
1. **Update PVM Components**:
   - Integrate enhanced parser with PSC type checker
   - Update LSP to use new type information
   - Enhance error reporting with type context

2. **Tool Integration**:
   - Update editor integrations to support type annotations
   - Enhance completion and navigation features
   - Add type-aware refactoring capabilities

3. **Testing and Validation**:
   - Run complete test suite including all new tests
   - Validate tool integrations work correctly
   - Perform end-to-end testing with real projects

4. **Performance Optimization**:
   - Apply any identified performance optimizations
   - Ensure memory usage is optimized
   - Validate performance meets all benchmarks

DELIVERABLES:
- Complete type annotation reference documentation
- Parser enhancement architecture documentation
- Usage guides and best practices
- Migration guides for existing projects
- Updated tool integrations and component integration
- Final validation and testing results

SUCCESS CRITERIA:
- All documentation is comprehensive and accurate
- Tool integrations work seamlessly with enhanced parser
- Complete test suite passes including all new functionality
- Performance meets or exceeds baseline requirements
- Project is ready for production use with type annotations

This completes the comprehensive type annotation detection improvement project, providing a robust foundation for advanced typed-Perl development.
```

---

## Implementation Benefits

### Performance Targets
- **Parsing Accuracy**: 95%+ type annotation detection accuracy
- **Performance**: No degradation for untyped code, <10% overhead for typed code
- **Memory Usage**: Efficient AST representation without memory bloat
- **Error Recovery**: Graceful handling of malformed type expressions

### Technical Advantages
- **Comprehensive Coverage**: Both untyped and typed Perl fully supported
- **Incremental Adoption**: Projects can adopt types gradually
- **Tool Integration**: Enhanced AST enables advanced tooling
- **Future-Proof**: Architecture supports future type system extensions

## Success Criteria

The type annotation improvement project is successful when:

1. **Parsing Accuracy**: 95%+ accuracy for all type expression patterns
2. **Backward Compatibility**: 100% compatibility with existing untyped Perl code
3. **Performance**: Acceptable performance for both typed and untyped code
4. **Tool Integration**: Enhanced parser supports advanced LSP and type checking features
5. **Test Coverage**: Comprehensive test suite provides regression protection
6. **Documentation**: Complete documentation enables developer adoption

This implementation will provide a robust, high-performance parser capable of handling sophisticated type expressions while maintaining complete backward compatibility with existing Perl code.
