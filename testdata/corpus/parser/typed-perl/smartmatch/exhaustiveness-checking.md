---
category: typed-perl
subcategory: smartmatch
tags:
    - exhaustiveness
    - pattern-coverage
    - completeness
    - missing-cases
    - unreachable-patterns
type_check: true
should_error: true
expected_error: "parse error"
---

# Exhaustiveness Checking for Pattern Matching

Tests pattern matching completeness validation to ensure PSC can detect missing cases,
unreachable patterns, and provide helpful warnings for incomplete coverage of union types.

```perl
# Complete exhaustiveness - should not need default clause
type Color = 'red' | 'green' | 'blue';
my Color $color = 'red';

given ($color) {
    when ('red')   { say 'warm color' }
    when ('green') { say 'cool color' }
    when ('blue')  { say 'cool color' }
    # Complete coverage - no default needed
}

# Incomplete coverage - missing 'blue' case (should warn)
given ($color) {
    when ('red')   { say 'warm' }
    when ('green') { say 'cool' }
    # Missing 'blue' - PSC should detect this
    default { say 'unknown color' }  # Default required due to incomplete coverage
}

# Over-complete coverage - unreachable default (should warn)
given ($color) {
    when ('red')   { say 'red' }
    when ('green') { say 'green' }
    when ('blue')  { say 'blue' }
    default { say 'impossible' }  # Unreachable - should warn
}

# Duplicate patterns - unreachable case (should warn)
given ($color) {
    when ('red')   { say 'first red' }
    when ('green') { say 'green' }
    when ('red')   { say 'second red' }  # Unreachable - should warn
    when ('blue')  { say 'blue' }
}

# HTTP status code exhaustiveness
type HttpStatus = '200' | '201' | '404' | '500';
my HttpStatus $status = '200';

# Complete success/error coverage
given ($status) {
    when ('200') { say 'OK' }
    when ('201') { say 'Created' }
    when ('404') { say 'Not Found' }
    when ('500') { say 'Server Error' }
    # Complete - no default needed
}

# Partial coverage with logical default
given ($status) {
    when ('200') { say 'Success' }
    when ('201') { say 'Success' }
    default { say 'Error' }  # Covers '404', '500' - logical grouping
}

# Numeric priority exhaustiveness
type Priority = 1 | 2 | 3 | 4 | 5;
my Priority $pri = 3;

# Range-based complete coverage
given ($pri) {
    when ([1, 2])    { say 'high priority' }     # Array pattern covers multiple
    when (3)         { say 'medium priority' }
    when ([4, 5])    { say 'low priority' }      # Array pattern covers remaining
    # Complete via array patterns
}

# Individual complete coverage
given ($pri) {
    when (1) { say 'critical' }
    when (2) { say 'high' }
    when (3) { say 'medium' }
    when (4) { say 'low' }
    when (5) { say 'minimal' }
    # Individually complete
}

# Partial coverage with gap
given ($pri) {
    when (1) { say 'critical' }
    when (2) { say 'high' }
    when (4) { say 'low' }
    when (5) { say 'minimal' }
    # Missing case 3 - should warn
    default { say 'unknown priority' }
}

# Boolean exhaustiveness
type BoolChoice = true | false;
my BoolChoice $choice = true;

# Simple boolean coverage
given ($choice) {
    when (true)  { say 'yes' }
    when (false) { say 'no' }
    # Complete boolean coverage
}

# Boolean with redundant patterns
given ($choice) {
    when (true)  { say 'affirmative' }
    when (1)     { say 'numeric true' }   # Should warn: true already covered
    when (false) { say 'negative' }
}

# Complex nested union exhaustiveness
type Result[T] = Success[T] | Error[Str];
type IntResult = Result[Int];
my IntResult $result = Success[42];

# Generic exhaustiveness (if PSC supports this syntax)
given ($result) {
    when (Success) { say 'got success' }    # Pattern matches Success[T] for any T
    when (Error)   { say 'got error' }      # Pattern matches Error[Str]
    # Complete for parameterized types
}

# Mixed union type exhaustiveness
type MixedResult = 'success' | 42 | true | [];
my MixedResult $mixed = 'success';

# Type-based exhaustiveness
given ($mixed) {
    when ('success') { say 'string success' }
    when (42)        { say 'numeric success' }
    when (true)      { say 'boolean success' }
    when ([])        { say 'array success' }
    # Complete mixed-type coverage
}

# Partial mixed type with type grouping
given ($mixed) {
    when (Str)    { say 'string type' }     # Should match 'success'
    when (Int)    { say 'integer type' }    # Should match 42
    when (Bool)   { say 'boolean type' }    # Should match true
    when (Array)  { say 'array type' }     # Should match []
    # Type-based complete coverage
}

# Regex pattern exhaustiveness
type FileType = '.txt' | '.md' | '.pl' | '.pm';
my FileType $ext = '.txt';

# Pattern-based coverage
given ($ext) {
    when (qr/\.txt$/) { say 'text file' }
    when (qr/\.md$/)  { say 'markdown file' }
    when (qr/\.p[lm]$/) { say 'perl file' }  # Covers both .pl and .pm
    # Complete via regex patterns
}

# Overlapping pattern coverage (should warn about precedence)
given ($ext) {
    when (qr/\.p/) { say 'perl-related' }   # Matches .pl and .pm
    when ('.pl')   { say 'perl script' }    # Unreachable due to regex above
    when ('.pm')   { say 'perl module' }    # Unreachable due to regex above
    when ('.txt')  { say 'text file' }
    when ('.md')   { say 'markdown file' }
}

# Smartmatch exhaustiveness with arrays
type SmallInt = 1 | 2 | 3;
my SmallInt $num = 2;

# Array membership exhaustiveness
my $in_range = $num ~~ [1, 2, 3];  # Complete coverage in array
my $partial = $num ~~ [1, 2];      # Partial coverage - could miss 3

# Complex condition exhaustiveness
given ($num) {
    when ($_ <= 2) { say 'small' }    # Covers 1, 2
    when ($_ == 3) { say 'medium' }   # Covers 3
    # Complete via conditional logic
}

# Range-based exhaustiveness
given ($num) {
    when (1..2) { say 'low range' }   # Range covers 1, 2
    when (3)    { say 'high value' }  # Individual covers 3
    # Complete via range and individual
}

# Multi-dimensional exhaustiveness
type Coordinate = [0,0] | [0,1] | [1,0] | [1,1];
my Coordinate $coord = [0,0];

# Structure-based exhaustiveness
given ($coord) {
    when ([0,0]) { say 'origin' }
    when ([0,1]) { say 'north' }
    when ([1,0]) { say 'east' }
    when ([1,1]) { say 'northeast' }
    # Complete coordinate coverage
}
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path:
  Source length: 5847 characters
  Type Annotations:
    TypeAnnotation: Color = 'red' | 'green' | 'blue' at 1:1
    VarAnnotation: $color :: Color at 2:1
    TypeAnnotation: HttpStatus = '200' | '201' | '404' | '500' at 25:1
    VarAnnotation: $status :: HttpStatus at 26:1
    TypeAnnotation: Priority = 1 | 2 | 3 | 4 | 5 at 40:1
    VarAnnotation: $pri :: Priority at 41:1
    TypeAnnotation: BoolChoice = true | false at 66:1
    VarAnnotation: $choice :: BoolChoice at 67:1
    TypeAnnotation: Result[T] = Success[T] | Error[Str] at 83:1
    TypeAnnotation: IntResult = Result[Int] at 84:1
    VarAnnotation: $result :: IntResult at 85:1
    TypeAnnotation: MixedResult = 'success' | 42 | true | [] at 94:1
    VarAnnotation: $mixed :: MixedResult at 95:1
    TypeAnnotation: FileType = '.txt' | '.md' | '.pl' | '.pm' at 113:1
    VarAnnotation: $ext :: FileType at 114:1
    TypeAnnotation: SmallInt = 1 | 2 | 3 at 134:1
    VarAnnotation: $num :: SmallInt at 135:1
    TypeAnnotation: Coordinate = [0,0] | [0,1] | [1,0] | [1,1] at 153:1
    VarAnnotation: $coord :: Coordinate at 154:1
  Root: source_file
  Tree Structure:
  source_file
    type_declaration
      type_name(Color)
      union_type
        literal('red')
        literal('green')
        literal('blue')
    var_decl
      type_expression(Color)
      scalar($color)
    given_statement
      condition(scalar($color))
      given_block
        when_clause
          condition(literal('red'))
          block
        when_clause
          condition(literal('green'))
          block
        when_clause
          condition(literal('blue'))
          block
    given_statement
      condition(scalar($color))
      given_block
        when_clause
          condition(literal('red'))
          block
        when_clause
          condition(literal('green'))
          block
        default_clause
          block
    given_statement
      condition(scalar($color))
      given_block
        when_clause
          condition(literal('red'))
          block
        when_clause
          condition(literal('green'))
          block
        when_clause
          condition(literal('blue'))
          block
        default_clause
          block
    given_statement
      condition(scalar($color))
      given_block
        when_clause
          condition(literal('red'))
          block
        when_clause
          condition(literal('green'))
          block
        when_clause
          condition(literal('red'))
          block
        when_clause
          condition(literal('blue'))
          block
    type_declaration
      type_name(HttpStatus)
      union_type
        literal('200')
        literal('201')
        literal('404')
        literal('500')
    var_decl
      type_expression(HttpStatus)
      scalar($status)
    given_statement
      condition(scalar($status))
      given_block
        when_clause
          condition(literal('200'))
          block
        when_clause
          condition(literal('201'))
          block
        when_clause
          condition(literal('404'))
          block
        when_clause
          condition(literal('500'))
          block
    given_statement
      condition(scalar($status))
      given_block
        when_clause
          condition(literal('200'))
          block
        when_clause
          condition(literal('201'))
          block
        default_clause
          block
    type_declaration
      type_name(Priority)
      union_type
        literal(1)
        literal(2)
        literal(3)
        literal(4)
        literal(5)
    var_decl
      type_expression(Priority)
      scalar($pri)
    given_statement
      condition(scalar($pri))
      given_block
        when_clause
          condition(array_constructor(literal(1), literal(2)))
          block
        when_clause
          condition(literal(3))
          block
        when_clause
          condition(array_constructor(literal(4), literal(5)))
          block
    given_statement
      condition(scalar($pri))
      given_block
        when_clause
          condition(literal(1))
          block
        when_clause
          condition(literal(2))
          block
        when_clause
          condition(literal(3))
          block
        when_clause
          condition(literal(4))
          block
        when_clause
          condition(literal(5))
          block
    given_statement
      condition(scalar($pri))
      given_block
        when_clause
          condition(literal(1))
          block
        when_clause
          condition(literal(2))
          block
        when_clause
          condition(literal(4))
          block
        when_clause
          condition(literal(5))
          block
        default_clause
          block
    type_declaration
      type_name(BoolChoice)
      union_type
        literal(true)
        literal(false)
    var_decl
      type_expression(BoolChoice)
      scalar($choice)
    given_statement
      condition(scalar($choice))
      given_block
        when_clause
          condition(literal(true))
          block
        when_clause
          condition(literal(false))
          block
    given_statement
      condition(scalar($choice))
      given_block
        when_clause
          condition(literal(true))
          block
        when_clause
          condition(literal(1))
          block
        when_clause
          condition(literal(false))
          block
    type_declaration
      type_name(Result[T])
      parameterized_union_type
        Success[T]
        Error[Str]
    type_alias
      type_name(IntResult)
      type_expression(Result[Int])
    var_decl
      type_expression(IntResult)
      scalar($result)
    given_statement
      condition(scalar($result))
      given_block
        when_clause
          condition(type_pattern(Success))
          block
        when_clause
          condition(type_pattern(Error))
          block
    type_declaration
      type_name(MixedResult)
      union_type
        literal('success')
        literal(42)
        literal(true)
        array_constructor
    var_decl
      type_expression(MixedResult)
      scalar($mixed)
    given_statement
      condition(scalar($mixed))
      given_block
        when_clause
          condition(literal('success'))
          block
        when_clause
          condition(literal(42))
          block
        when_clause
          condition(literal(true))
          block
        when_clause
          condition(array_constructor)
          block
    given_statement
      condition(scalar($mixed))
      given_block
        when_clause
          condition(type_pattern(Str))
          block
        when_clause
          condition(type_pattern(Int))
          block
        when_clause
          condition(type_pattern(Bool))
          block
        when_clause
          condition(type_pattern(Array))
          block
    type_declaration
      type_name(FileType)
      union_type
        literal('.txt')
        literal('.md')
        literal('.pl')
        literal('.pm')
    var_decl
      type_expression(FileType)
      scalar($ext)
    given_statement
      condition(scalar($ext))
      given_block
        when_clause
          condition(regex_pattern(qr/\.txt$/))
          block
        when_clause
          condition(regex_pattern(qr/\.md$/))
          block
        when_clause
          condition(regex_pattern(qr/\.p[lm]$/))
          block
    given_statement
      condition(scalar($ext))
      given_block
        when_clause
          condition(regex_pattern(qr/\.p/))
          block
        when_clause
          condition(literal('.pl'))
          block
        when_clause
          condition(literal('.pm'))
          block
        when_clause
          condition(literal('.txt'))
          block
        when_clause
          condition(literal('.md'))
          block
    type_declaration
      type_name(SmallInt)
      union_type
        literal(1)
        literal(2)
        literal(3)
    var_decl
      type_expression(SmallInt)
      scalar($num)
    var_decl
      scalar($in_range)
      smartmatch_expression
        scalar($num)
        array_constructor(literal(1), literal(2), literal(3))
    var_decl
      scalar($partial)
      smartmatch_expression
        scalar($num)
        array_constructor(literal(1), literal(2))
    given_statement
      condition(scalar($num))
      given_block
        when_clause
          condition(comparison_expression($_ <= 2))
          block
        when_clause
          condition(comparison_expression($_ == 3))
          block
    given_statement
      condition(scalar($num))
      given_block
        when_clause
          condition(range_expression(1..2))
          block
        when_clause
          condition(literal(3))
          block
    type_declaration
      type_name(Coordinate)
      union_type
        array_constructor(literal(0), literal(0))
        array_constructor(literal(0), literal(1))
        array_constructor(literal(1), literal(0))
        array_constructor(literal(1), literal(1))
    var_decl
      type_expression(Coordinate)
      scalar($coord)
    given_statement
      condition(scalar($coord))
      given_block
        when_clause
          condition(array_constructor(literal(0), literal(0)))
          block
        when_clause
          condition(array_constructor(literal(0), literal(1)))
          block
        when_clause
          condition(array_constructor(literal(1), literal(0)))
          block
        when_clause
          condition(array_constructor(literal(1), literal(1)))
          block
}
```

## After Type Inference

### Text Format

```
# Same as before type inference - exhaustiveness checking not yet implemented
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $color = 'red';

given ($color) {
    when ('red')   { say 'warm color' }
    when ('green') { say 'cool color' }
    when ('blue')  { say 'cool color' }
}

given ($color) {
    when ('red')   { say 'warm' }
    when ('green') { say 'cool' }
    default { say 'unknown color' }
}

given ($color) {
    when ('red')   { say 'red' }
    when ('green') { say 'green' }
    when ('blue')  { say 'blue' }
    default { say 'impossible' }
}

given ($color) {
    when ('red')   { say 'first red' }
    when ('green') { say 'green' }
    when ('red')   { say 'second red' }
    when ('blue')  { say 'blue' }
}

my $status = '200';

given ($status) {
    when ('200') { say 'OK' }
    when ('201') { say 'Created' }
    when ('404') { say 'Not Found' }
    when ('500') { say 'Server Error' }
}

given ($status) {
    when ('200') { say 'Success' }
    when ('201') { say 'Success' }
    default { say 'Error' }
}

my $pri = 3;

given ($pri) {
    when ([1, 2])    { say 'high priority' }
    when (3)         { say 'medium priority' }
    when ([4, 5])    { say 'low priority' }
}

given ($pri) {
    when (1) { say 'critical' }
    when (2) { say 'high' }
    when (3) { say 'medium' }
    when (4) { say 'low' }
    when (5) { say 'minimal' }
}

given ($pri) {
    when (1) { say 'critical' }
    when (2) { say 'high' }
    when (4) { say 'low' }
    when (5) { say 'minimal' }
    default { say 'unknown priority' }
}

my $choice = true;

given ($choice) {
    when (true)  { say 'yes' }
    when (false) { say 'no' }
}

given ($choice) {
    when (true)  { say 'affirmative' }
    when (1)     { say 'numeric true' }
    when (false) { say 'negative' }
}

my $result = Success[42];

given ($result) {
    when (Success) { say 'got success' }
    when (Error)   { say 'got error' }
}

my $mixed = 'success';

given ($mixed) {
    when ('success') { say 'string success' }
    when (42)        { say 'numeric success' }
    when (true)      { say 'boolean success' }
    when ([])        { say 'array success' }
}

given ($mixed) {
    when (Str)    { say 'string type' }
    when (Int)    { say 'integer type' }
    when (Bool)   { say 'boolean type' }
    when (Array)  { say 'array type' }
}

my $ext = '.txt';

given ($ext) {
    when (qr/\.txt$/) { say 'text file' }
    when (qr/\.md$/)  { say 'markdown file' }
    when (qr/\.p[lm]$/) { say 'perl file' }
}

given ($ext) {
    when (qr/\.p/) { say 'perl-related' }
    when ('.pl')   { say 'perl script' }
    when ('.pm')   { say 'perl module' }
    when ('.txt')  { say 'text file' }
    when ('.md')   { say 'markdown file' }
}

my $num = 2;

my $in_range = $num ~~ [1, 2, 3];
my $partial = $num ~~ [1, 2];

given ($num) {
    when ($_ <= 2) { say 'small' }
    when ($_ == 3) { say 'medium' }
}

given ($num) {
    when (1..2) { say 'low range' }
    when (3)    { say 'high value' }
}

my $coord = [0,0];

given ($coord) {
    when ([0,0]) { say 'origin' }
    when ([0,1]) { say 'north' }
    when ([1,0]) { say 'east' }
    when ([1,1]) { say 'northeast' }
}
```

## Typed Perl Output

```perl
type Color = 'red' | 'green' | 'blue';
my Color $color = 'red';

given ($color) {
    when ('red')   { say 'warm color' }
    when ('green') { say 'cool color' }
    when ('blue')  { say 'cool color' }
}

given ($color) {
    when ('red')   { say 'warm' }
    when ('green') { say 'cool' }
    default { say 'unknown color' }
}

given ($color) {
    when ('red')   { say 'red' }
    when ('green') { say 'green' }
    when ('blue')  { say 'blue' }
    default { say 'impossible' }
}

given ($color) {
    when ('red')   { say 'first red' }
    when ('green') { say 'green' }
    when ('red')   { say 'second red' }
    when ('blue')  { say 'blue' }
}

type HttpStatus = '200' | '201' | '404' | '500';
my HttpStatus $status = '200';

given ($status) {
    when ('200') { say 'OK' }
    when ('201') { say 'Created' }
    when ('404') { say 'Not Found' }
    when ('500') { say 'Server Error' }
}

given ($status) {
    when ('200') { say 'Success' }
    when ('201') { say 'Success' }
    default { say 'Error' }
}

type Priority = 1 | 2 | 3 | 4 | 5;
my Priority $pri = 3;

given ($pri) {
    when ([1, 2])    { say 'high priority' }
    when (3)         { say 'medium priority' }
    when ([4, 5])    { say 'low priority' }
}

given ($pri) {
    when (1) { say 'critical' }
    when (2) { say 'high' }
    when (3) { say 'medium' }
    when (4) { say 'low' }
    when (5) { say 'minimal' }
}

given ($pri) {
    when (1) { say 'critical' }
    when (2) { say 'high' }
    when (4) { say 'low' }
    when (5) { say 'minimal' }
    default { say 'unknown priority' }
}

type BoolChoice = true | false;
my BoolChoice $choice = true;

given ($choice) {
    when (true)  { say 'yes' }
    when (false) { say 'no' }
}

given ($choice) {
    when (true)  { say 'affirmative' }
    when (1)     { say 'numeric true' }
    when (false) { say 'negative' }
}

type Result[T] = Success[T] | Error[Str];
type IntResult = Result[Int];
my IntResult $result = Success[42];

given ($result) {
    when (Success) { say 'got success' }
    when (Error)   { say 'got error' }
}

type MixedResult = 'success' | 42 | true | [];
my MixedResult $mixed = 'success';

given ($mixed) {
    when ('success') { say 'string success' }
    when (42)        { say 'numeric success' }
    when (true)      { say 'boolean success' }
    when ([])        { say 'array success' }
}

given ($mixed) {
    when (Str)    { say 'string type' }
    when (Int)    { say 'integer type' }
    when (Bool)   { say 'boolean type' }
    when (Array)  { say 'array type' }
}

type FileType = '.txt' | '.md' | '.pl' | '.pm';
my FileType $ext = '.txt';

given ($ext) {
    when (qr/\.txt$/) { say 'text file' }
    when (qr/\.md$/)  { say 'markdown file' }
    when (qr/\.p[lm]$/) { say 'perl file' }
}

given ($ext) {
    when (qr/\.p/) { say 'perl-related' }
    when ('.pl')   { say 'perl script' }
    when ('.pm')   { say 'perl module' }
    when ('.txt')  { say 'text file' }
    when ('.md')   { say 'markdown file' }
}

type SmallInt = 1 | 2 | 3;
my SmallInt $num = 2;

my $in_range = $num ~~ [1, 2, 3];
my $partial = $num ~~ [1, 2];

given ($num) {
    when ($_ <= 2) { say 'small' }
    when ($_ == 3) { say 'medium' }
}

given ($num) {
    when (1..2) { say 'low range' }
    when (3)    { say 'high value' }
}

type Coordinate = [0,0] | [0,1] | [1,0] | [1,1];
my Coordinate $coord = [0,0];

given ($coord) {
    when ([0,0]) { say 'origin' }
    when ([0,1]) { say 'north' }
    when ([1,0]) { say 'east' }
    when ([1,1]) { say 'northeast' }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
Warning: Incomplete pattern coverage in given statement at line 12 - missing case 'blue'
Warning: Unreachable default clause at line 22 - all union cases covered
Warning: Duplicate pattern 'red' at line 29 - unreachable code
Warning: Unreachable pattern '1' at line 77 - true already covers truthy values
Warning: Pattern precedence issue at line 127 - regex /\.p/ makes literal '.pl' and '.pm' unreachable
```
