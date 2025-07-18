---
category: untyped-perl
subcategory: control-flow
tags:
    - control_flow
    - for
    - list
    - qw
---

# For Loop List

For loop over literal list

```perl
for my $item (qw(apple banana cherry)) {
    print "$item\n";
}
```

# Expected Compilation Outcomes

## Clean Perl Output

```perl
use v5.36;
for my $item (qw(apple banana cherry)) {
    print "$item\n";
}
```

## Typed Perl Output

```perl
for my $item (qw(apple banana cherry)) {
    print "$item\n";
}
```

## Inferred Perl Output

```perl
# Type inference not yet fully implemented
```

## Text AST

```
source_file
  for_statement
    expression_stmt
      literal
    token
    scalar
      token
      token
    token
    quoted_word_list
      expression_stmt
        literal
      expression_stmt
        literal
      expression_stmt
        literal
      expression_stmt
        literal
    token
    block_stmt
      token
      expression_stmt
        literal
      token
      token
```

## JSON AST

```json
{
  "type": "source_file",
  "children": [
    {
      "type": "for_statement",
      "children": [
        {
          "type": "expression_stmt",
          "children": [
            {
              "type": "literal",
              "value": "for",
              "kind": "string"
            }
          ]
        },
        {
          "type": "token",
          "text": "my"
        },
        {
          "type": "scalar",
          "children": [
            {
              "type": "token",
              "text": "$"
            },
            {
              "type": "token",
              "text": "item"
            }
          ]
        },
        {
          "type": "token",
          "text": "("
        },
        {
          "type": "quoted_word_list",
          "children": [
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "qw",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "(",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "apple banana cherry",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": ")",
                  "kind": "string"
                }
              ]
            }
          ]
        },
        {
          "type": "token",
          "text": ")"
        },
        {
          "type": "block_stmt",
          "children": [
            {
              "type": "token",
              "text": "{"
            },
            {
              "type": "expression_stmt",
              "children": [
                {
                  "type": "literal",
                  "value": "print \"$item\\n\"",
                  "kind": "string"
                }
              ]
            },
            {
              "type": "token",
              "text": ";"
            },
            {
              "type": "token",
              "text": "}"
            }
          ]
        }
      ]
    }
  ]
}
```
