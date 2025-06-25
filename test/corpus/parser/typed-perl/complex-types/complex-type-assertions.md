---
category: typed-perl
subcategory: complex-types
type_check: true
tags:
    - type-assertions
    - method-calls
    - complex-types
    - parameterized-types
---

# Complex Type Assertions

Type assertions with complex type expressions

```perl
my $result = $data as ArrayRef[HashRef[Int|Bool]];
my $complex = ($input->process()) as Map[Str, ArrayRef[MyType]];
my $transformed = $obj->convert() as Result[Data[User], Error[String]];
```

## Expected AST

### Before Type Inference

```
source_file
  expression_statement
    var_decl
      variable
  token
  expression_statement
    var_decl
      variable
  token
  expression_statement
    var_decl
      variable
  token
```

### After Type Inference

```
source_file
  expression_statement
    var_decl
      variable
  token
  expression_statement
    var_decl
      variable
  token
  expression_statement
    var_decl
      variable
  token
```

## Expected Type Errors

(none)
