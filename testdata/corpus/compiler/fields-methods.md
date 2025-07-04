---
category: compiler
subcategory: fields-methods
tags:
    - field-declarations
    - method-signatures
    - attributes
    - multiline-signatures
type_check: false
---

# Field Declarations

Basic field declarations with types

```perl
field Int $count;
field ArrayRef[Str] $items;
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
field $count;
field $items;
```

# Field with Initialization

Field declarations with initial values

```perl
field Int $counter = 0;
field HashRef[Str] $config = {};
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
field $counter = 0;
field $config = {};
```

# Method Declaration

Method with typed parameters

```perl
method Bool add_user(UserId $id, HashRef[Str] $data) {
    return 1;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
method add_user($id, $data) {
    return 1;
}
```

# Multiline Signature

Function with multiline signature

```perl
sub Bool multiline (
    Int $first,
    Str $second,
    ArrayRef[Int] $third
) {
    return 1;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub multiline (
    $first,
    $second,
    $third
) {
    return 1;
}
```

# Attributes with Signature

Function with attributes and typed signature

```perl
sub Int tagged :lvalue :const (Int $value) {
    return $value;
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
sub tagged :lvalue :const ($value) {
    return $value;
}
```
