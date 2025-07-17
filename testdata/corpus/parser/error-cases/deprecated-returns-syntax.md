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

### Expected AST

#### Text Format

```
AST {
  Path: deprecated_test1.pl
  Source length: 50 characters
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      anonymous_method_expression
        token
        type_expression
          expression_stmt
            literal
        expression_stmt
          literal
        expression_stmt
          literal
        block_stmt
          token
          expression_stmt
            literal
          token
          token
}
```

#### JSON Format

```json
{
  "path": "deprecated_test1.pl",
  "root": {
    "type": "source_file",
    "children": [
      {
        "type": "expression_statement",
        "children": [
          {
            "type": "anonymous_method_expression",
            "children": [
              {"type": "token", "text": "method"},
              {
                "type": "type_expression",
                "children": [
                  {
                    "type": "expression_stmt",
                    "children": [
                      {"type": "literal", "value": "calculate", "kind": "string"}
                    ]
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "children": [
                  {"type": "literal", "value": "()", "kind": "string"}
                ]
              },
              {
                "type": "expression_stmt",
                "children": [
                  {"type": "literal", "value": "returns Int", "kind": "string"}
                ]
              },
              {
                "type": "block_stmt",
                "children": [
                  {"type": "token", "text": "{"},
                  {
                    "type": "expression_stmt",
                    "children": [
                      {"type": "literal", "value": "return 42", "kind": "string"}
                    ]
                  },
                  {"type": "token", "text": ";"},
                  {"type": "token", "text": "}"}
                ]
              }
            ]
          }
        ]
      }
    ]
  },
  "type_annotations": [],
  "errors": []
}
```

## Method with Parameters and Returns Syntax

<!-- This syntax is no longer supported as of July 12, 2025 -->

```perl
method greet(Str $name) returns Str {
    return "Hello, $name!";
}
```

### Expected AST

#### Text Format

```
AST {
  Path: deprecated_test2.pl
  Source length: 64 characters
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      anonymous_method_expression
        token
        type_expression
          expression_stmt
            literal
        expression_stmt
          literal
        expression_stmt
          literal
        block_stmt
          token
          expression_stmt
            literal
          token
          token
}
```

#### JSON Format

```json
{
  "path": "deprecated_test2.pl",
  "root": {
    "type": "source_file",
    "children": [
      {
        "type": "expression_statement",
        "children": [
          {
            "type": "anonymous_method_expression",
            "children": [
              {"type": "token", "text": "method"},
              {
                "type": "type_expression",
                "children": [
                  {
                    "type": "expression_stmt",
                    "children": [
                      {"type": "literal", "value": "greet", "kind": "string"}
                    ]
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "children": [
                  {"type": "literal", "value": "(Str $name)", "kind": "string"}
                ]
              },
              {
                "type": "expression_stmt",
                "children": [
                  {"type": "literal", "value": "returns Str", "kind": "string"}
                ]
              },
              {
                "type": "block_stmt",
                "children": [
                  {"type": "token", "text": "{"},
                  {
                    "type": "expression_stmt",
                    "children": [
                      {"type": "literal", "value": "return \"Hello, $name!\"", "kind": "string"}
                    ]
                  },
                  {"type": "token", "text": ";"},
                  {"type": "token", "text": "}"}
                ]
              }
            ]
          }
        ]
      }
    ]
  },
  "type_annotations": [],
  "errors": []
}
```

## Method with Complex Return Type and Returns Syntax

<!-- This syntax is no longer supported as of July 12, 2025 -->

```perl
method get_data() returns ArrayRef[HashRef[Str, Int]] {
    return [];
}
```

### Expected AST

#### Text Format

```
AST {
  Path: deprecated_test3.pl
  Source length: 74 characters
  Root: source_file
  Tree Structure:
  source_file
    expression_statement
      anonymous_method_expression
        token
        type_expression
          expression_stmt
            literal
        expression_stmt
          literal
        expression_stmt
          literal
        block_stmt
          token
          expression_stmt
            literal
          token
          token
}
```

#### JSON Format

```json
{
  "path": "deprecated_test3.pl",
  "root": {
    "type": "source_file",
    "children": [
      {
        "type": "expression_statement",
        "children": [
          {
            "type": "anonymous_method_expression",
            "children": [
              {"type": "token", "text": "method"},
              {
                "type": "type_expression",
                "children": [
                  {
                    "type": "expression_stmt",
                    "children": [
                      {"type": "literal", "value": "get_data", "kind": "string"}
                    ]
                  }
                ]
              },
              {
                "type": "expression_stmt",
                "children": [
                  {"type": "literal", "value": "()", "kind": "string"}
                ]
              },
              {
                "type": "expression_stmt",
                "children": [
                  {"type": "literal", "value": "returns ArrayRef[HashRef[Str, Int]]", "kind": "string"}
                ]
              },
              {
                "type": "block_stmt",
                "children": [
                  {"type": "token", "text": "{"},
                  {
                    "type": "expression_stmt",
                    "children": [
                      {"type": "literal", "value": "return []", "kind": "string"}
                    ]
                  },
                  {"type": "token", "text": ";"},
                  {"type": "token", "text": "}"}
                ]
              }
            ]
          }
        ]
      }
    ]
  },
  "type_annotations": [],
  "errors": []
}
```
