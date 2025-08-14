---
category: typed-perl
subcategory: smartmatch
tags:
    - union-types
    - pattern-matching
    - given-when
    - type-constraints
type_check: true
should_error: true
expected_error: "parse error"
---

# Union Type Pattern Matching with Given/When

Tests comprehensive union type pattern matching using given/when statements with type constraints.
This validates that PSC's type system makes given/when behavior predictable and safe.

```perl
# String literal union - Currency example from issue #328
type Currency = 'USD' | 'EUR' | 'GBP' | 'JPY';
my Currency $currency = 'USD';

given ($currency) {
    when ('USD') { say '$' }
    when ('EUR') { say '€' }
    when ('GBP') { say '£' }
    when ('JPY') { say '¥' }
}

# Status enum pattern
type Status = 'active' | 'inactive' | 'pending';
my Status $status = 'active';

given ($status) {
    when ('active')   { say 'running' }
    when ('inactive') { say 'stopped' }
    when ('pending')  { say 'waiting' }
}

# HTTP status codes
type HttpCode = '200' | '404' | '500';
my HttpCode $code = '404';

given ($code) {
    when ('200') { say 'OK' }
    when ('404') { say 'Not Found' }
    when ('500') { say 'Server Error' }
}

# Numeric priority union
type Priority = 1 | 2 | 3 | 4 | 5;
my Priority $pri = 3;

given ($pri) {
    when (1) { say 'critical' }
    when (2) { say 'high' }
    when (3) { say 'medium' }
    when (4) { say 'low' }
    when (5) { say 'minimal' }
}

# Boolean-like toggle union
type Toggle = 'on' | 'off' | 0 | 1;
my Toggle $setting = 'on';

given ($setting) {
    when ('on')  { say 'enabled' }
    when ('off') { say 'disabled' }
    when (1)     { say 'enabled' }
    when (0)     { say 'disabled' }
}
```

# Expected AST

## Before Type Inference

### Text Format

```
AST {
  Path:
  Source length: 1327 characters
  Type Annotations:
    TypeAnnotation: Currency = 'USD' | 'EUR' | 'GBP' | 'JPY' at 1:1
    VarAnnotation: $currency :: Currency at 2:1
    TypeAnnotation: Status = 'active' | 'inactive' | 'pending' at 10:1
    VarAnnotation: $status :: Status at 11:1
    TypeAnnotation: HttpCode = '200' | '404' | '500' at 18:1
    VarAnnotation: $code :: HttpCode at 19:1
    TypeAnnotation: Priority = 1 | 2 | 3 | 4 | 5 at 26:1
    VarAnnotation: $pri :: Priority at 27:1
    TypeAnnotation: Toggle = 'on' | 'off' | 0 | 1 at 36:1
    VarAnnotation: $setting :: Toggle at 37:1
  Root: source_file
  Tree Structure:
  source_file
    type_declaration
      type_name
      union_type
        literal('USD')
        literal('EUR')
        literal('GBP')
        literal('JPY')
    var_decl
      type_expression(Currency)
      scalar($currency)
    given_statement
      condition(scalar($currency))
      given_block
        when_clause
          condition(literal('USD'))
          block
        when_clause
          condition(literal('EUR'))
          block
        when_clause
          condition(literal('GBP'))
          block
        when_clause
          condition(literal('JPY'))
          block
    type_declaration
      type_name
      union_type
        literal('active')
        literal('inactive')
        literal('pending')
    var_decl
      type_expression(Status)
      scalar($status)
    given_statement
      condition(scalar($status))
      given_block
        when_clause
          condition(literal('active'))
          block
        when_clause
          condition(literal('inactive'))
          block
        when_clause
          condition(literal('pending'))
          block
    type_declaration
      type_name
      union_type
        literal('200')
        literal('404')
        literal('500')
    var_decl
      type_expression(HttpCode)
      scalar($code)
    given_statement
      condition(scalar($code))
      given_block
        when_clause
          condition(literal('200'))
          block
        when_clause
          condition(literal('404'))
          block
        when_clause
          condition(literal('500'))
          block
    type_declaration
      type_name
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
    type_declaration
      type_name
      union_type
        literal('on')
        literal('off')
        literal(0)
        literal(1)
    var_decl
      type_expression(Toggle)
      scalar($setting)
    given_statement
      condition(scalar($setting))
      given_block
        when_clause
          condition(literal('on'))
          block
        when_clause
          condition(literal('off'))
          block
        when_clause
          condition(literal(1))
          block
        when_clause
          condition(literal(0))
          block
}
```

## After Type Inference

### Text Format

```
AST {
  Path:
  Source length: 1327 characters
  Type Annotations:
    TypeAnnotation: Currency = 'USD' | 'EUR' | 'GBP' | 'JPY' at 1:1
    VarAnnotation: $currency :: Currency at 2:1
    TypeAnnotation: Status = 'active' | 'inactive' | 'pending' at 10:1
    VarAnnotation: $status :: Status at 11:1
    TypeAnnotation: HttpCode = '200' | '404' | '500' at 18:1
    VarAnnotation: $code :: HttpCode at 19:1
    TypeAnnotation: Priority = 1 | 2 | 3 | 4 | 5 at 26:1
    VarAnnotation: $pri :: Priority at 27:1
    TypeAnnotation: Toggle = 'on' | 'off' | 0 | 1 at 36:1
    VarAnnotation: $setting :: Toggle at 37:1
  Root: source_file
  Tree Structure:
  source_file
    type_declaration
      type_name
      union_type
        literal('USD')
        literal('EUR')
        literal('GBP')
        literal('JPY')
    var_decl
      type_expression(Currency)
      scalar($currency)
    given_statement
      condition(scalar($currency))
      given_block
        when_clause
          condition(literal('USD'))
          block
        when_clause
          condition(literal('EUR'))
          block
        when_clause
          condition(literal('GBP'))
          block
        when_clause
          condition(literal('JPY'))
          block
    type_declaration
      type_name
      union_type
        literal('active')
        literal('inactive')
        literal('pending')
    var_decl
      type_expression(Status)
      scalar($status)
    given_statement
      condition(scalar($status))
      given_block
        when_clause
          condition(literal('active'))
          block
        when_clause
          condition(literal('inactive'))
          block
        when_clause
          condition(literal('pending'))
          block
    type_declaration
      type_name
      union_type
        literal('200')
        literal('404')
        literal('500')
    var_decl
      type_expression(HttpCode)
      scalar($code)
    given_statement
      condition(scalar($code))
      given_block
        when_clause
          condition(literal('200'))
          block
        when_clause
          condition(literal('404'))
          block
        when_clause
          condition(literal('500'))
          block
    type_declaration
      type_name
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
    type_declaration
      type_name
      union_type
        literal('on')
        literal('off')
        literal(0)
        literal(1)
    var_decl
      type_expression(Toggle)
      scalar($setting)
    given_statement
      condition(scalar($setting))
      given_block
        when_clause
          condition(literal('on'))
          block
        when_clause
          condition(literal('off'))
          block
        when_clause
          condition(literal(1))
          block
        when_clause
          condition(literal(0))
          block
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
my $currency = 'USD';

given ($currency) {
    when ('USD') { say '$' }
    when ('EUR') { say '€' }
    when ('GBP') { say '£' }
    when ('JPY') { say '¥' }
}

my $status = 'active';

given ($status) {
    when ('active')   { say 'running' }
    when ('inactive') { say 'stopped' }
    when ('pending')  { say 'waiting' }
}

my $code = '404';

given ($code) {
    when ('200') { say 'OK' }
    when ('404') { say 'Not Found' }
    when ('500') { say 'Server Error' }
}

my $pri = 3;

given ($pri) {
    when (1) { say 'critical' }
    when (2) { say 'high' }
    when (3) { say 'medium' }
    when (4) { say 'low' }
    when (5) { say 'minimal' }
}

my $setting = 'on';

given ($setting) {
    when ('on')  { say 'enabled' }
    when ('off') { say 'disabled' }
    when (1)     { say 'enabled' }
    when (0)     { say 'disabled' }
}
```

## Typed Perl Output

```perl
type Currency = 'USD' | 'EUR' | 'GBP' | 'JPY';
my Currency $currency = 'USD';

given ($currency) {
    when ('USD') { say '$' }
    when ('EUR') { say '€' }
    when ('GBP') { say '£' }
    when ('JPY') { say '¥' }
}

type Status = 'active' | 'inactive' | 'pending';
my Status $status = 'active';

given ($status) {
    when ('active')   { say 'running' }
    when ('inactive') { say 'stopped' }
    when ('pending')  { say 'waiting' }
}

type HttpCode = '200' | '404' | '500';
my HttpCode $code = '404';

given ($code) {
    when ('200') { say 'OK' }
    when ('404') { say 'Not Found' }
    when ('500') { say 'Server Error' }
}

type Priority = 1 | 2 | 3 | 4 | 5;
my Priority $pri = 3;

given ($pri) {
    when (1) { say 'critical' }
    when (2) { say 'high' }
    when (3) { say 'medium' }
    when (4) { say 'low' }
    when (5) { say 'minimal' }
}

type Toggle = 'on' | 'off' | 0 | 1;
my Toggle $setting = 'on';

given ($setting) {
    when ('on')  { say 'enabled' }
    when ('off') { say 'disabled' }
    when (1)     { say 'enabled' }
    when (0)     { say 'disabled' }
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

# Expected Type Errors

```
(none - all patterns are exhaustive and type-safe)
```
