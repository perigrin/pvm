---
category: error-cases
subcategory: syntax-errors
tags: [arrow-syntax, method-syntax, subroutine-syntax, deprecated-syntax]
---

# Arrow Syntax Errors

## Method with Arrow Syntax

<!-- should_error: true -->
<!-- expected_error: error[TSP011] -->
<!-- expected_suggestion: Use 'method Type name()' instead of 'method name() -> Type' -->

```perl
method calculate() -> Int {
    return 42;
}
```

### Expected AST

#### Text Format

```
Parse error: SYS-007: error[TSP001]: parse error (ERROR nodes detected)
Arrow syntax is not supported in the current grammar.
```

#### JSON Format

```json
{
  "error": "SYS-007: error[TSP001]: parse error (ERROR nodes detected)",
  "message": "Arrow syntax is not supported in the current grammar",
  "ast": null
}
```

### Expected Compilation Outcomes

#### Clean Perl Output

```perl
# Error: Arrow syntax not supported - parsing should fail
```

#### Typed Perl Output

```perl
# Error: Arrow syntax not supported - parsing should fail
```

#### Inferred Perl Output

```perl
# Error: Arrow syntax not supported - parsing should fail
```

## Subroutine with Arrow Syntax

<!-- should_error: true -->
<!-- expected_error: error[TSP011] -->
<!-- expected_suggestion: Use 'sub name()' or 'method Type name()' instead of 'sub name() -> Type' -->

```perl
sub process() -> Str {
    return "processed";
}
```

### Expected AST

#### Text Format

```
Parse error: SYS-007: error[TSP001]: parse error (ERROR nodes detected)
Arrow syntax is not supported in the current grammar.
```

#### JSON Format

```json
{
  "error": "SYS-007: error[TSP001]: parse error (ERROR nodes detected)",
  "message": "Arrow syntax is not supported in the current grammar",
  "ast": null
}
```

### Expected Compilation Outcomes

#### Clean Perl Output

```perl
# Error: Arrow syntax not supported - parsing should fail
```

#### Typed Perl Output

```perl
# Error: Arrow syntax not supported - parsing should fail
```

#### Inferred Perl Output

```perl
# Error: Arrow syntax not supported - parsing should fail
```

## Method with Parameters and Arrow Syntax

<!-- should_error: true -->
<!-- expected_error: error[TSP011] -->
<!-- expected_suggestion: Use 'method Type name(ParamType $param)' instead of 'method name(ParamType $param) -> Type' -->

```perl
method greet(Str $name) -> Str {
    return "Hello, $name!";
}
```

### Expected AST

#### Text Format

```
Parse error: SYS-007: error[TSP001]: parse error (ERROR nodes detected)
Arrow syntax is not supported in the current grammar.
```

#### JSON Format

```json
{
  "error": "SYS-007: error[TSP001]: parse error (ERROR nodes detected)",
  "message": "Arrow syntax is not supported in the current grammar",
  "ast": null
}
```

### Expected Compilation Outcomes

#### Clean Perl Output

```perl
# Error: Arrow syntax not supported - parsing should fail
```

#### Typed Perl Output

```perl
# Error: Arrow syntax not supported - parsing should fail
```

#### Inferred Perl Output

```perl
# Error: Arrow syntax not supported - parsing should fail
```

## Complex Type with Arrow Syntax

<!-- should_error: true -->
<!-- expected_error: error[TSP011] -->
<!-- expected_suggestion: Use 'method Type name()' instead of 'method name() -> Type' -->

```perl
method get_data() -> ArrayRef[HashRef[Str, Int]] {
    return [];
}
```

### Expected AST

#### Text Format

```
Parse error: SYS-007: error[TSP001]: parse error (ERROR nodes detected)
Arrow syntax is not supported in the current grammar.
```

#### JSON Format

```json
{
  "error": "SYS-007: error[TSP001]: parse error (ERROR nodes detected)",
  "message": "Arrow syntax is not supported in the current grammar",
  "ast": null
}
```

### Expected Compilation Outcomes

#### Clean Perl Output

```perl
# Error: Arrow syntax not supported - parsing should fail
```

#### Typed Perl Output

```perl
# Error: Arrow syntax not supported - parsing should fail
```

#### Inferred Perl Output

```perl
# Error: Arrow syntax not supported - parsing should fail
```
