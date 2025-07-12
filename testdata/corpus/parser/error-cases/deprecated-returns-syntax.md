---
category: error-cases
subcategory: deprecated-syntax
tags: [deprecated-syntax, returns-syntax, method-definitions, syntax-errors]
skip: true
---

# Deprecated Returns Syntax (No Longer Supported)

## Method with Returns Syntax

<!-- This syntax is no longer supported as of July 12, 2025 -->

```perl
method calculate() returns Int {
    return 42;
}
```


## Method with Parameters and Returns Syntax

<!-- This syntax is no longer supported as of July 12, 2025 -->

```perl
method greet(Str $name) returns Str {
    return "Hello, $name!";
}
```


## Method with Complex Return Type and Returns Syntax

<!-- This syntax is no longer supported as of July 12, 2025 -->

```perl
method get_data() returns ArrayRef[HashRef[Str, Int]] {
    return [];
}
```
