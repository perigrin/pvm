---
category: typed-perl
subcategory: smartmatch
tags:
    - mixed-types
    - union-types
    - numeric-string
    - boolean-conversion
    - edge-cases
type_check: true
should_error: true
expected_error: "parse error"
---

# Mixed Type Union Pattern Matching

Tests complex scenarios with mixed string/numeric/boolean union types to validate that PSC's
type system handles controversial smartmatch behaviors predictably when types are constrained.

```perl
# Classic numeric vs string problem - PSC should make this predictable
type MixedId = 42 | '42' | 'none';
my MixedId $id1 = 42;      # Numeric
my MixedId $id2 = '42';    # String
my MixedId $id3 = 'none';  # String literal

given ($id1) {
    when (42)     { say 'numeric forty-two' }
    when ('42')   { say 'string forty-two' }  # Should this match numeric 42?
    when ('none') { say 'none value' }
}

given ($id2) {
    when (42)     { say 'numeric forty-two' }  # Should this match string '42'?
    when ('42')   { say 'string forty-two' }
    when ('none') { say 'none value' }
}

# Mixed smartmatch testing
my $num_vs_str = $id1 ~~ '42';     # Numeric 42 vs string '42' - what should happen?
my $str_vs_num = $id2 ~~ 42;       # String '42' vs numeric 42 - consistent with above?
my $cross_check = $id1 ~~ $id2;    # Mixed-type variable comparison

# Falsy value unions - truth/falsy behavior
type FalsyValue = 'none' | 0 | false | undef;
my FalsyValue $falsy1 = 0;       # Numeric zero
my FalsyValue $falsy2 = false;   # Boolean false
my FalsyValue $falsy3 = 'none';  # String literal

# Test falsy comparisons with type constraints
my $zero_match = $falsy1 ~~ 0;            # Direct match
my $false_match = $falsy2 ~~ false;       # Direct match
my $zero_false = $falsy1 ~~ false;        # 0 vs false - should this match?
my $false_zero = $falsy2 ~~ 0;            # false vs 0 - consistent with above?

# Array membership with mixed types
my $falsy_array = $falsy1 ~~ [0, false, undef];  # Mixed falsy array
my $truthy_check = $falsy3 ~~ [0, false];        # 'none' vs [0, false] - should not match

# Complex nested scenarios
type ComplexValue = 'success' | 200 | true | 1 | '1';
my ComplexValue $val1 = 1;        # Numeric one
my ComplexValue $val2 = '1';      # String one
my ComplexValue $val3 = true;     # Boolean true
my ComplexValue $val4 = 200;      # Success code

# Truth value comparisons - PSC should clarify these
my $one_true = $val1 ~~ true;     # 1 vs true - truthy but different types
my $str_one_true = $val2 ~~ true; # '1' vs true - string vs boolean
my $one_vs_str_one = $val1 ~~ '1'; # 1 vs '1' - numeric vs string

given ($val1) {
    when (1)       { say 'numeric one' }
    when ('1')     { say 'string one' }    # Should numeric 1 match this?
    when (true)    { say 'boolean true' }  # Should numeric 1 match this?
    when (200)     { say 'success code' }
    when ('success') { say 'success string' }
}

# Array references and complex structures
type DataValue = 'empty' | [] | {} | 0;
my DataValue $data1 = [];      # Empty array ref
my DataValue $data2 = {};      # Empty hash ref
my DataValue $data3 = 0;       # Numeric zero
my DataValue $data4 = 'empty'; # String literal

# Reference smartmatch behavior with type constraints
my $empty_array = $data1 ~~ [];           # Array ref vs array ref
my $empty_hash = $data2 ~~ {};            # Hash ref vs hash ref
my $ref_vs_string = $data1 ~~ 'empty';    # Array ref vs string - should not match
my $zero_vs_empty = $data3 ~~ [];         # Numeric 0 vs array ref - should not match

# Regex patterns with mixed types
type Pattern = 42 | '42' | 'forty-two' | qr/\d+/;
my Pattern $pat1 = 42;          # Numeric
my Pattern $pat2 = '42';        # String
my Pattern $pat3 = 'forty-two'; # String literal
my Pattern $pat4 = qr/\d+/;     # Regex object

# Pattern matching with mixed types
my $num_regex = $pat1 ~~ qr/\d+/;     # Does numeric 42 match digit regex?
my $str_regex = $pat2 ~~ qr/\d+/;     # Does string '42' match digit regex?
my $word_regex = $pat3 ~~ qr/\d+/;    # Does 'forty-two' match digit regex?

# Reverse pattern matching
my $regex_num = qr/\d+/ ~~ $pat1;     # Does regex match numeric? (controversial)
my $regex_str = qr/\d+/ ~~ $pat2;     # Does regex match string?

# Context-sensitive behavior
type ContextValue = 'file.txt' | 42 | '0644';
my ContextValue $ctx = 'file.txt';

# File operations context
my $filename_match = $ctx ~~ qr/\.txt$/;  # String context - should work
my $permission_match = $ctx ~~ qr/\d+/;   # Numeric context - should not match filename

# Multi-dimensional complexity
type NestedValue = 'null' | 0 | [] | [0] | ['null'];
my NestedValue $nested1 = [];        # Empty array
my NestedValue $nested2 = [0];       # Array with zero
my NestedValue $nested3 = ['null'];  # Array with string
my NestedValue $nested4 = 0;         # Just zero
my NestedValue $nested5 = 'null';    # Just string

# Deep structure matching
my $empty_vs_zero_array = $nested1 ~~ [0];       # [] vs [0]
my $zero_array_vs_zero = $nested2 ~~ 0;          # [0] vs 0 - should not match
my $string_array_vs_string = $nested3 ~~ 'null'; # ['null'] vs 'null' - should not match
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path:
  Source length: 4512 characters
  Type Annotations:
    TypeAnnotation: MixedId = 42 | '42' | 'none' at 1:1
    VarAnnotation: $id1 :: MixedId at 2:1
    VarAnnotation: $id2 :: MixedId at 3:1
    VarAnnotation: $id3 :: MixedId at 4:1
    TypeAnnotation: FalsyValue = 'none' | 0 | false | undef at 20:1
    VarAnnotation: $falsy1 :: FalsyValue at 21:1
    VarAnnotation: $falsy2 :: FalsyValue at 22:1
    VarAnnotation: $falsy3 :: FalsyValue at 23:1
    TypeAnnotation: ComplexValue = 'success' | 200 | true | 1 | '1' at 35:1
    VarAnnotation: $val1 :: ComplexValue at 36:1
    VarAnnotation: $val2 :: ComplexValue at 37:1
    VarAnnotation: $val3 :: ComplexValue at 38:1
    VarAnnotation: $val4 :: ComplexValue at 39:1
    TypeAnnotation: DataValue = 'empty' | [] | {} | 0 at 55:1
    VarAnnotation: $data1 :: DataValue at 56:1
    VarAnnotation: $data2 :: DataValue at 57:1
    VarAnnotation: $data3 :: DataValue at 58:1
    VarAnnotation: $data4 :: DataValue at 59:1
    TypeAnnotation: Pattern = 42 | '42' | 'forty-two' | qr/\d+/ at 67:1
    VarAnnotation: $pat1 :: Pattern at 68:1
    VarAnnotation: $pat2 :: Pattern at 69:1
    VarAnnotation: $pat3 :: Pattern at 70:1
    VarAnnotation: $pat4 :: Pattern at 71:1
    TypeAnnotation: ContextValue = 'file.txt' | 42 | '0644' at 81:1
    VarAnnotation: $ctx :: ContextValue at 82:1
    TypeAnnotation: NestedValue = 'null' | 0 | [] | [0] | ['null'] at 88:1
    VarAnnotation: $nested1 :: NestedValue at 89:1
    VarAnnotation: $nested2 :: NestedValue at 90:1
    VarAnnotation: $nested3 :: NestedValue at 91:1
    VarAnnotation: $nested4 :: NestedValue at 92:1
    VarAnnotation: $nested5 :: NestedValue at 93:1
  Root: source_file
  Tree Structure:
  source_file
    type_declaration
      type_name(MixedId)
      union_type
        literal(42)
        literal('42')
        literal('none')
    var_decl
      type_expression(MixedId)
      scalar($id1)
    var_decl
      type_expression(MixedId)
      scalar($id2)
    var_decl
      type_expression(MixedId)
      scalar($id3)
    given_statement
      condition(scalar($id1))
      given_block
        when_clause
          condition(literal(42))
          block
        when_clause
          condition(literal('42'))
          block
        when_clause
          condition(literal('none'))
          block
    given_statement
      condition(scalar($id2))
      given_block
        when_clause
          condition(literal(42))
          block
        when_clause
          condition(literal('42'))
          block
        when_clause
          condition(literal('none'))
          block
    var_decl
      scalar($num_vs_str)
      smartmatch_expression
        scalar($id1)
        literal('42')
    var_decl
      scalar($str_vs_num)
      smartmatch_expression
        scalar($id2)
        literal(42)
    var_decl
      scalar($cross_check)
      smartmatch_expression
        scalar($id1)
        scalar($id2)
    type_declaration
      type_name(FalsyValue)
      union_type
        literal('none')
        literal(0)
        literal(false)
        literal(undef)
    var_decl
      type_expression(FalsyValue)
      scalar($falsy1)
    var_decl
      type_expression(FalsyValue)
      scalar($falsy2)
    var_decl
      type_expression(FalsyValue)
      scalar($falsy3)
    var_decl
      scalar($zero_match)
      smartmatch_expression
        scalar($falsy1)
        literal(0)
    var_decl
      scalar($false_match)
      smartmatch_expression
        scalar($falsy2)
        literal(false)
    var_decl
      scalar($zero_false)
      smartmatch_expression
        scalar($falsy1)
        literal(false)
    var_decl
      scalar($false_zero)
      smartmatch_expression
        scalar($falsy2)
        literal(0)
    var_decl
      scalar($falsy_array)
      smartmatch_expression
        scalar($falsy1)
        array_constructor
          literal(0)
          literal(false)
          literal(undef)
    var_decl
      scalar($truthy_check)
      smartmatch_expression
        scalar($falsy3)
        array_constructor
          literal(0)
          literal(false)
    type_declaration
      type_name(ComplexValue)
      union_type
        literal('success')
        literal(200)
        literal(true)
        literal(1)
        literal('1')
    var_decl
      type_expression(ComplexValue)
      scalar($val1)
    var_decl
      type_expression(ComplexValue)
      scalar($val2)
    var_decl
      type_expression(ComplexValue)
      scalar($val3)
    var_decl
      type_expression(ComplexValue)
      scalar($val4)
    var_decl
      scalar($one_true)
      smartmatch_expression
        scalar($val1)
        literal(true)
    var_decl
      scalar($str_one_true)
      smartmatch_expression
        scalar($val2)
        literal(true)
    var_decl
      scalar($one_vs_str_one)
      smartmatch_expression
        scalar($val1)
        literal('1')
    given_statement
      condition(scalar($val1))
      given_block
        when_clause
          condition(literal(1))
          block
        when_clause
          condition(literal('1'))
          block
        when_clause
          condition(literal(true))
          block
        when_clause
          condition(literal(200))
          block
        when_clause
          condition(literal('success'))
          block
    type_declaration
      type_name(DataValue)
      union_type
        literal('empty')
        array_constructor
        hash_constructor
        literal(0)
    var_decl
      type_expression(DataValue)
      scalar($data1)
    var_decl
      type_expression(DataValue)
      scalar($data2)
    var_decl
      type_expression(DataValue)
      scalar($data3)
    var_decl
      type_expression(DataValue)
      scalar($data4)
    var_decl
      scalar($empty_array)
      smartmatch_expression
        scalar($data1)
        array_constructor
    var_decl
      scalar($empty_hash)
      smartmatch_expression
        scalar($data2)
        hash_constructor
    var_decl
      scalar($ref_vs_string)
      smartmatch_expression
        scalar($data1)
        literal('empty')
    var_decl
      scalar($zero_vs_empty)
      smartmatch_expression
        scalar($data3)
        array_constructor
    type_declaration
      type_name(Pattern)
      union_type
        literal(42)
        literal('42')
        literal('forty-two')
        regex_pattern(qr/\d+/)
    var_decl
      type_expression(Pattern)
      scalar($pat1)
    var_decl
      type_expression(Pattern)
      scalar($pat2)
    var_decl
      type_expression(Pattern)
      scalar($pat3)
    var_decl
      type_expression(Pattern)
      scalar($pat4)
    var_decl
      scalar($num_regex)
      smartmatch_expression
        scalar($pat1)
        regex_pattern(qr/\d+/)
    var_decl
      scalar($str_regex)
      smartmatch_expression
        scalar($pat2)
        regex_pattern(qr/\d+/)
    var_decl
      scalar($word_regex)
      smartmatch_expression
        scalar($pat3)
        regex_pattern(qr/\d+/)
    var_decl
      scalar($regex_num)
      smartmatch_expression
        regex_pattern(qr/\d+/)
        scalar($pat1)
    var_decl
      scalar($regex_str)
      smartmatch_expression
        regex_pattern(qr/\d+/)
        scalar($pat2)
    type_declaration
      type_name(ContextValue)
      union_type
        literal('file.txt')
        literal(42)
        literal('0644')
    var_decl
      type_expression(ContextValue)
      scalar($ctx)
    var_decl
      scalar($filename_match)
      smartmatch_expression
        scalar($ctx)
        regex_pattern(qr/\.txt$/)
    var_decl
      scalar($permission_match)
      smartmatch_expression
        scalar($ctx)
        regex_pattern(qr/\d+/)
    type_declaration
      type_name(NestedValue)
      union_type
        literal('null')
        literal(0)
        array_constructor
        array_constructor
          literal(0)
        array_constructor
          literal('null')
    var_decl
      type_expression(NestedValue)
      scalar($nested1)
    var_decl
      type_expression(NestedValue)
      scalar($nested2)
    var_decl
      type_expression(NestedValue)
      scalar($nested3)
    var_decl
      type_expression(NestedValue)
      scalar($nested4)
    var_decl
      type_expression(NestedValue)
      scalar($nested5)
    var_decl
      scalar($empty_vs_zero_array)
      smartmatch_expression
        scalar($nested1)
        array_constructor
          literal(0)
    var_decl
      scalar($zero_array_vs_zero)
      smartmatch_expression
        scalar($nested2)
        literal(0)
    var_decl
      scalar($string_array_vs_string)
      smartmatch_expression
        scalar($nested3)
        literal('null')
}
```

## After Type Inference

### Text Format

```
# Same as before type inference - type inference not yet fully implemented
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $id1 = 42;
my $id2 = '42';
my $id3 = 'none';

given ($id1) {
    when (42)     { say 'numeric forty-two' }
    when ('42')   { say 'string forty-two' }
    when ('none') { say 'none value' }
}

given ($id2) {
    when (42)     { say 'numeric forty-two' }
    when ('42')   { say 'string forty-two' }
    when ('none') { say 'none value' }
}

my $num_vs_str = $id1 ~~ '42';
my $str_vs_num = $id2 ~~ 42;
my $cross_check = $id1 ~~ $id2;

my $falsy1 = 0;
my $falsy2 = false;
my $falsy3 = 'none';

my $zero_match = $falsy1 ~~ 0;
my $false_match = $falsy2 ~~ false;
my $zero_false = $falsy1 ~~ false;
my $false_zero = $falsy2 ~~ 0;

my $falsy_array = $falsy1 ~~ [0, false, undef];
my $truthy_check = $falsy3 ~~ [0, false];

my $val1 = 1;
my $val2 = '1';
my $val3 = true;
my $val4 = 200;

my $one_true = $val1 ~~ true;
my $str_one_true = $val2 ~~ true;
my $one_vs_str_one = $val1 ~~ '1';

given ($val1) {
    when (1)       { say 'numeric one' }
    when ('1')     { say 'string one' }
    when (true)    { say 'boolean true' }
    when (200)     { say 'success code' }
    when ('success') { say 'success string' }
}

my $data1 = [];
my $data2 = {};
my $data3 = 0;
my $data4 = 'empty';

my $empty_array = $data1 ~~ [];
my $empty_hash = $data2 ~~ {};
my $ref_vs_string = $data1 ~~ 'empty';
my $zero_vs_empty = $data3 ~~ [];

my $pat1 = 42;
my $pat2 = '42';
my $pat3 = 'forty-two';
my $pat4 = qr/\d+/;

my $num_regex = $pat1 ~~ qr/\d+/;
my $str_regex = $pat2 ~~ qr/\d+/;
my $word_regex = $pat3 ~~ qr/\d+/;

my $regex_num = qr/\d+/ ~~ $pat1;
my $regex_str = qr/\d+/ ~~ $pat2;

my $ctx = 'file.txt';

my $filename_match = $ctx ~~ qr/\.txt$/;
my $permission_match = $ctx ~~ qr/\d+/;

my $nested1 = [];
my $nested2 = [0];
my $nested3 = ['null'];
my $nested4 = 0;
my $nested5 = 'null';

my $empty_vs_zero_array = $nested1 ~~ [0];
my $zero_array_vs_zero = $nested2 ~~ 0;
my $string_array_vs_string = $nested3 ~~ 'null';
```

## Typed Perl Output

```perl
type MixedId = 42 | '42' | 'none';
my MixedId $id1 = 42;
my MixedId $id2 = '42';
my MixedId $id3 = 'none';

given ($id1) {
    when (42)     { say 'numeric forty-two' }
    when ('42')   { say 'string forty-two' }
    when ('none') { say 'none value' }
}

given ($id2) {
    when (42)     { say 'numeric forty-two' }
    when ('42')   { say 'string forty-two' }
    when ('none') { say 'none value' }
}

my $num_vs_str = $id1 ~~ '42';
my $str_vs_num = $id2 ~~ 42;
my $cross_check = $id1 ~~ $id2;

type FalsyValue = 'none' | 0 | false | undef;
my FalsyValue $falsy1 = 0;
my FalsyValue $falsy2 = false;
my FalsyValue $falsy3 = 'none';

my $zero_match = $falsy1 ~~ 0;
my $false_match = $falsy2 ~~ false;
my $zero_false = $falsy1 ~~ false;
my $false_zero = $falsy2 ~~ 0;

my $falsy_array = $falsy1 ~~ [0, false, undef];
my $truthy_check = $falsy3 ~~ [0, false];

type ComplexValue = 'success' | 200 | true | 1 | '1';
my ComplexValue $val1 = 1;
my ComplexValue $val2 = '1';
my ComplexValue $val3 = true;
my ComplexValue $val4 = 200;

my $one_true = $val1 ~~ true;
my $str_one_true = $val2 ~~ true;
my $one_vs_str_one = $val1 ~~ '1';

given ($val1) {
    when (1)       { say 'numeric one' }
    when ('1')     { say 'string one' }
    when (true)    { say 'boolean true' }
    when (200)     { say 'success code' }
    when ('success') { say 'success string' }
}

type DataValue = 'empty' | [] | {} | 0;
my DataValue $data1 = [];
my DataValue $data2 = {};
my DataValue $data3 = 0;
my DataValue $data4 = 'empty';

my $empty_array = $data1 ~~ [];
my $empty_hash = $data2 ~~ {};
my $ref_vs_string = $data1 ~~ 'empty';
my $zero_vs_empty = $data3 ~~ [];

type Pattern = 42 | '42' | 'forty-two' | qr/\d+/;
my Pattern $pat1 = 42;
my Pattern $pat2 = '42';
my Pattern $pat3 = 'forty-two';
my Pattern $pat4 = qr/\d+/;

my $num_regex = $pat1 ~~ qr/\d+/;
my $str_regex = $pat2 ~~ qr/\d+/;
my $word_regex = $pat3 ~~ qr/\d+/;

my $regex_num = qr/\d+/ ~~ $pat1;
my $regex_str = qr/\d+/ ~~ $pat2;

type ContextValue = 'file.txt' | 42 | '0644';
my ContextValue $ctx = 'file.txt';

my $filename_match = $ctx ~~ qr/\.txt$/;
my $permission_match = $ctx ~~ qr/\d+/;

type NestedValue = 'null' | 0 | [] | [0] | ['null'];
my NestedValue $nested1 = [];
my NestedValue $nested2 = [0];
my NestedValue $nested3 = ['null'];
my NestedValue $nested4 = 0;
my NestedValue $nested5 = 'null';

my $empty_vs_zero_array = $nested1 ~~ [0];
my $zero_array_vs_zero = $nested2 ~~ 0;
my $string_array_vs_string = $nested3 ~~ 'null';
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none - but PSC should document expected behavior for mixed-type comparisons)
```
