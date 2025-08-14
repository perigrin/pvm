---
category: error-cases
subcategory: smartmatch
tags:
    - type-safety
    - violations
    - error-prevention
    - unsafe-patterns
    - warning-cases
type_check: true
should_error: true
---

# Type Safety Violations in Smartmatch Operations

Tests cases where PSC should prevent unsafe smartmatch operations or provide warnings
to ensure type safety and prevent runtime errors that could occur with untyped values.

```perl
# SHOULD ERROR: Type-constrained vs untyped variable
type Status = 'active' | 'inactive' | 'pending';
my Status $status = 'active';
my $user_input;  # Untyped - could be anything

# This should be an error or strong warning
my $unsafe_match = $status ~~ $user_input;  # Type-constrained vs untyped

# SHOULD ERROR: Impossible type combinations
type StrictInt = 1 | 2 | 3;
my StrictInt $number = 2;

given ($number) {
    when ('string') { say 'impossible' }  # Int can never match string literal
    when (4) { say 'also impossible' }    # Value not in union type
    when ([]) { say 'reference impossible' }  # Array ref can never match Int
}

# SHOULD ERROR: Incompatible union type assignments
type ColorEnum = 'red' | 'green' | 'blue';
type SizeEnum = 'small' | 'medium' | 'large';
my ColorEnum $color = 'red';
my SizeEnum $size = 'small';

# This should error - incompatible union types
my $incompatible = $color ~~ $size;  # red/green/blue vs small/medium/large

# SHOULD WARN: Potentially dangerous reference comparisons
my @array1 = (1, 2, 3);
my @array2 = (1, 2, 3);

# Reference identity vs content - should warn about unexpected behavior
my $ref_danger = \@array1 ~~ \@array2;  # Same content, different identity

# SHOULD ERROR: Uninitialized variable in pattern matching
type InitCheck = 'initialized' | 'ready';
my InitCheck $init;  # Uninitialized typed variable

# Should error - using uninitialized variable in pattern match
given ($init) {
    when ('initialized') { say 'ready' }
    when ('ready') { say 'go' }
}

# SHOULD ERROR: Invalid pattern syntax
type ValidEnum = 'one' | 'two' | 'three';
my ValidEnum $valid = 'one';

given ($valid) {
    when ('one' | 'two') { say 'invalid syntax' }  # Union in when clause - error
    when (qr/on|tw/) { say 'regex too broad' }     # Could match outside union
}

# SHOULD ERROR: Type constraint violation in array membership
type SmallNum = 1 | 2 | 3;
my SmallNum $small = 2;

# Array contains values outside the type constraint
my $unsafe_array = $small ~~ [1, 2, 3, 4, 5];  # Should warn: array contains 4,5

# SHOULD ERROR: Function return type mismatch in pattern
sub get_unknown() { return int(rand(100)); }  # Returns random int

type KnownRange = 10 | 20 | 30;
my KnownRange $known = 10;

# Function could return any int, not just the union values
my $function_unsafe = $known ~~ get_unknown();  # Should error

# SHOULD ERROR: Regex pattern too broad for type constraint
type FileExt = '.txt' | '.md' | '.pl';
my FileExt $ext = '.txt';

# Regex could match more than the union allows
my $broad_regex = $ext ~~ qr/\.\w+/;  # Too broad - matches any extension

# SHOULD ERROR: Nested type constraint violations
type Inner = 'a' | 'b';
type Outer = Inner | 'c';
my Outer $outer = 'a';

# Should error - trying to match against type not in union
my $nested_violation = $outer ~~ 'd';  # 'd' not in Inner or 'c'

# SHOULD ERROR: Cross-type smartmatch without coercion rules
type NumericId = 1 | 2 | 3;
type StringId = 'one' | 'two' | 'three';
my NumericId $num_id = 1;
my StringId $str_id = 'one';

# Should error - no defined coercion between numeric and string IDs
my $cross_type = $num_id ~~ $str_id;  # 1 vs 'one' - incompatible

# SHOULD ERROR: Object method dispatch without proper method
class NoSmartmatch {
    method new() { bless {}, __CLASS__ }
    # No smartmatch overload or special methods
}

type ObjNoMatch = NoSmartmatch | 'string';
my ObjNoMatch $obj_no = NoSmartmatch->new();

# Should error - object has no smartmatch capability
my $obj_unsafe = $obj_no ~~ 'string';  # Unpredictable object behavior

# SHOULD ERROR: Circular dependency in type checking
type CircularA = CircularB | 'value_a';
type CircularB = CircularA | 'value_b';  # Circular type definition
my CircularA $circular = 'value_a';

# Should error during type checking - circular dependency
my $circular_match = $circular ~~ 'value_b';

# SHOULD ERROR: Unicode normalization issues
type UnicodeStrict = 'café' | 'naïve';  # Specific Unicode forms
my UnicodeStrict $unicode = 'café';

# Different Unicode normalization forms - could be unsafe
my $norm_unsafe = $unicode ~~ "cafe\x{301}";  # Combining form vs precomposed

# SHOULD ERROR: Floating point precision issues in types
type PreciseFloat = 3.14159265359 | 2.71828182846;
my PreciseFloat $pi = 3.14159265359;

# Should warn - floating point comparison issues
my $float_unsafe = $pi ~~ 3.14159265358;  # Very close but not exact

# SHOULD ERROR: Type coercion without explicit rules
type MixedCoercion = 42 | '42' | 0 | '0' | 1 | '1';
my MixedCoercion $mixed = 42;

# Should require explicit coercion rules or error
my $coercion_unsafe = $mixed ~~ '42';  # Numeric vs string without rules

# SHOULD ERROR: Overloaded operator conflicts
class ConflictedOverload {
    method new() { bless {}, __CLASS__ }

    # Multiple conflicting overloads
    use overload
        '~~' => sub { return 'smartmatch' },
        'eq' => sub { return $_[1] eq 'conflict' },
        '""' => sub { return 'stringified' };
}

type ConflictType = ConflictedOverload | 'string';
my ConflictType $conflict = ConflictedOverload->new();

# Should error - unclear which overload takes precedence
my $overload_unsafe = $conflict ~~ 'conflict';

# SHOULD ERROR: Thread safety violations (if threading enabled)
# use threads;
# use threads::shared;

# my shared %shared_data : shared;
# type ThreadUnsafe = %shared_data | {};
# my ThreadUnsafe $thread_var = %shared_data;

# # Should error - shared data access without proper locking
# my $thread_unsafe = $thread_var ~~ {};

# SHOULD ERROR: Tainted data in pattern matching (if taint mode)
# use tainting;

# type TaintSafe = 'safe_value' | 'another_safe';
# my TaintSafe $safe = 'safe_value';

# # Simulated tainted input
# my $tainted_input = $ENV{USER};  # Tainted in taint mode

# # Should error - tainted data in pattern match with typed variable
# my $taint_unsafe = $safe ~~ $tainted_input;

# SHOULD ERROR: Readonly variable modifications
use Readonly;
Readonly my $readonly_value => 'constant';

type ReadonlyTest = 'constant' | 'variable';
my ReadonlyTest $readonly_typed = 'constant';

# Should error - attempting to use readonly in mutable context
given ($readonly_typed) {
    when ('constant') {
        $readonly_value = 'changed';  # Should error - readonly modification
    }
}

# SHOULD ERROR: Memory leak potential with circular references
type CircularLeak = {} | [];
my CircularLeak $leak_test = {};

# Create circular reference
$leak_test->{self} = $leak_test;

# Should warn - potential memory leak without weak references
my $leak_match = $leak_test ~~ {};  # Circular reference in comparison

# SHOULD ERROR: Stack overflow potential with deep recursion
type DeepNested = [DeepNested] | 'base';

sub create_deep($depth) {
    return 'base' if $depth <= 0;
    return [create_deep($depth - 1)];
}

my DeepNested $deep = create_deep(1000);  # Very deep structure

# Should error - potential stack overflow in deep comparison
my $deep_unsafe = $deep ~~ create_deep(1000);

# SHOULD ERROR: Resource exhaustion with large data structures
type LargeData = [Int] | 'small';
my LargeData $large = [1..1000000];  # Very large array

# Should warn - performance and memory concerns
my $size_unsafe = $large ~~ [1..1000000];  # Massive comparison

# SHOULD ERROR: Platform-specific behavior dependencies
type PlatformSpecific = \*STDIN | \*STDOUT | 42;
my PlatformSpecific $platform = \*STDIN;

# Should warn - platform-dependent behavior
my $platform_unsafe = $platform ~~ 0;  # File descriptor comparison

# SHOULD ERROR: Version compatibility issues
use version;

type VersionStrict = version->parse('v1.0.0');
my VersionStrict $ver = version->parse('v1.0.0');

# Should error - version comparison without proper version object
my $version_unsafe = $ver ~~ '1.0.0';  # String vs version object

# SHOULD ERROR: Locale-dependent string comparisons
use locale;

type LocaleString = 'Straße' | 'Muller';  # German strings
my LocaleString $locale_str = 'Straße';

# Should warn - locale-dependent comparison
my $locale_unsafe = $locale_str ~~ 'STRASSE';  # Case/locale sensitivity
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path:
  Source length: 9247 characters
  Type Annotations:
    TypeAnnotation: Status = 'active' | 'inactive' | 'pending' at 1:1
    VarAnnotation: $status :: Status at 2:1
    TypeAnnotation: StrictInt = 1 | 2 | 3 at 8:1
    VarAnnotation: $number :: StrictInt at 9:1
    TypeAnnotation: ColorEnum = 'red' | 'green' | 'blue' at 17:1
    TypeAnnotation: SizeEnum = 'small' | 'medium' | 'large' at 18:1
    VarAnnotation: $color :: ColorEnum at 19:1
    VarAnnotation: $size :: SizeEnum at 20:1
    TypeAnnotation: InitCheck = 'initialized' | 'ready' at 29:1
    VarAnnotation: $init :: InitCheck at 30:1
    TypeAnnotation: ValidEnum = 'one' | 'two' | 'three' at 38:1
    VarAnnotation: $valid :: ValidEnum at 39:1
    TypeAnnotation: SmallNum = 1 | 2 | 3 at 46:1
    VarAnnotation: $small :: SmallNum at 47:1
    TypeAnnotation: KnownRange = 10 | 20 | 30 at 53:1
    VarAnnotation: $known :: KnownRange at 54:1
  Root: source_file
  Tree Structure:
  source_file
    comment("# SHOULD ERROR: Type-constrained vs untyped variable")
    type_declaration
      type_name(Status)
      union_type
        literal('active')
        literal('inactive')
        literal('pending')
    var_decl
      type_expression(Status)
      scalar($status)
    var_decl
      scalar($user_input)
    comment("# This should be an error or strong warning")
    var_decl
      scalar($unsafe_match)
      smartmatch_expression
        scalar($status)
        scalar($user_input)
    comment("# SHOULD ERROR: Impossible type combinations")
    type_declaration
      type_name(StrictInt)
      union_type
        literal(1)
        literal(2)
        literal(3)
    var_decl
      type_expression(StrictInt)
      scalar($number)
    given_statement
      condition(scalar($number))
      given_block
        when_clause
          condition(literal('string'))
          block
        when_clause
          condition(literal(4))
          block
        when_clause
          condition(array_constructor)
          block
    comment("# SHOULD ERROR: Incompatible union type assignments")
    type_declaration
      type_name(ColorEnum)
      union_type
        literal('red')
        literal('green')
        literal('blue')
    type_declaration
      type_name(SizeEnum)
      union_type
        literal('small')
        literal('medium')
        literal('large')
    var_decl
      type_expression(ColorEnum)
      scalar($color)
    var_decl
      type_expression(SizeEnum)
      scalar($size)
    comment("# This should error - incompatible union types")
    var_decl
      scalar($incompatible)
      smartmatch_expression
        scalar($color)
        scalar($size)
    comment("# SHOULD WARN: Potentially dangerous reference comparisons")
    var_decl
      array(@array1)
      array_constructor(literal(1), literal(2), literal(3))
    var_decl
      array(@array2)
      array_constructor(literal(1), literal(2), literal(3))
    comment("# Reference identity vs content - should warn about unexpected behavior")
    var_decl
      scalar($ref_danger)
      smartmatch_expression
        reference_expression(\@array1)
        reference_expression(\@array2)
    comment("# SHOULD ERROR: Uninitialized variable in pattern matching")
    type_declaration
      type_name(InitCheck)
      union_type
        literal('initialized')
        literal('ready')
    var_decl
      type_expression(InitCheck)
      scalar($init)
    comment("# Should error - using uninitialized variable in pattern match")
    given_statement
      condition(scalar($init))
      given_block
        when_clause
          condition(literal('initialized'))
          block
        when_clause
          condition(literal('ready'))
          block
}
```

## After Type Inference

### Text Format

```
# Type inference would fail due to multiple type errors
```

# Expected Compilation Outcomes

## Clean Perl Output

```
# Compilation should fail with type safety errors
```

## Typed Perl Output

```
# Compilation should fail with type safety errors
```

## Inferred Perl Output

```
# Compilation should fail with type safety errors
```

# Expected Type Errors

```
ERROR: Type safety violation at line 6: Cannot match type-constrained variable ($status :: Status) with untyped variable ($user_input)
ERROR: Impossible type combination at line 12: StrictInt cannot match string literal 'string'
ERROR: Impossible type combination at line 13: StrictInt cannot match numeric literal 4 (not in union)
ERROR: Impossible type combination at line 14: StrictInt cannot match array reference
ERROR: Incompatible union types at line 23: Cannot compare ColorEnum with SizeEnum (no common values)
WARNING: Potentially dangerous reference comparison at line 28: Array references with same content but different identity
ERROR: Uninitialized variable usage at line 33: Variable $init used in pattern match before initialization
ERROR: Invalid pattern syntax at line 42: Union operators not allowed in when clauses
WARNING: Regex pattern too broad at line 43: Pattern could match values outside ValidEnum
WARNING: Array contains values outside type constraint at line 50: Array [1,2,3,4,5] contains 4,5 not in SmallNum
ERROR: Function return type mismatch at line 57: get_unknown() return type incompatible with KnownRange
WARNING: Regex pattern too broad at line 62: Pattern /\.\w+/ could match extensions outside FileExt
ERROR: Value not in union type at line 69: 'd' is not a valid value for Outer type
ERROR: Incompatible types without coercion at line 76: No defined coercion between NumericId and StringId
ERROR: Object lacks smartmatch capability at line 85: NoSmartmatch class has no smartmatch overload
ERROR: Circular type dependency at line 89: CircularA and CircularB have circular reference
WARNING: Unicode normalization issue at line 96: Different Unicode forms may not match predictably
WARNING: Floating point precision issue at line 101: Very close floating point values may not match exactly
ERROR: Type coercion without explicit rules at line 106: Numeric vs string comparison requires explicit coercion rules
ERROR: Conflicting operator overloads at line 116: Multiple overloads create ambiguous behavior
ERROR: Readonly variable modification at line 140: Cannot modify readonly variable $readonly_value
WARNING: Potential memory leak at line 148: Circular reference without weak references
ERROR: Potential stack overflow at line 158: Deep recursive structure may cause stack overflow
WARNING: Performance concern at line 165: Very large data structure comparison
WARNING: Platform-dependent behavior at line 171: File descriptor comparison is platform-specific
ERROR: Version comparison type mismatch at line 178: String vs version object comparison without proper handling
WARNING: Locale-dependent comparison at line 185: String comparison may vary by locale settings
```
