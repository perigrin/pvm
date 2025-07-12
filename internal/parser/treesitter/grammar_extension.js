// ABOUTME: Grammar extension for tree-sitter-perl to handle type annotations
// ABOUTME: Extends the base Perl grammar with typed Perl syntax

// This file contains grammar extensions for tree-sitter-perl to handle typed Perl syntax
// It would be integrated with the main tree-sitter-perl grammar during the build process

module.exports.rules = {
    // Extend variable declaration to include type annotations
    // Example: my Int $var
    _variable_declaration: ($, original) => choice(
      seq(
        field('declarator', choice('my', 'our', 'state')),
        field('type', $.identifier),
        field('name', $._variable),
        optional(seq('=', field('value', $._expression))),
        $._terminator
      ),
      original
    ),

    // Extend subroutine declaration to include parameter and return type annotations
    // Example: sub name(Type $param) -> ReturnType { ... }
    _subroutine_declaration: ($, original) => choice(
      seq(
        'sub',
        field('name', $.identifier),
        field('parameters', $.typed_parameter_list),
        optional(seq(
          '->',
          field('return_type', $.return_type_annotation)
        )),
        choice(
          field('body', $.block),
          $._terminator
        )
      ),
      original
    ),

    // Add a rule for method declarations with type annotations
    // Example: method ReturnType name(Type $param) { ... }
    _method_declaration: ($, original) => choice(
      seq(
        'method',
        field('name', $.identifier),
        field('parameters', $.typed_parameter_list),
        optional(seq(
          '->',
          field('return_type', $.return_type_annotation)
        )),
        choice(
          field('body', $.block),
          $._terminator
        )
      ),
      original
    ),

    // Add a rule for field declarations with type annotations
    // Example: field Type $attr;
    field_declaration: $ => seq(
      'field',
      field('type', $.identifier),
      field('name', $._variable),
      optional(seq('=', field('default', $._expression))),
      $._terminator
    ),

    // Add a rule for type declarations
    // Example: type MyType = OtherType;
    type_declaration: $ => seq(
      'type',
      field('name', $.identifier),
      '=',
      field('definition', $.type_expression),
      $._terminator
    ),

    // Add rules for type expressions
    type_expression: $ => choice(
      // Simple type
      $.identifier,
      // Parameterized type
      $.parameterized_type,
      // Union type
      $.union_type,
      // Intersection type
      $.intersection_type,
      // Negation type
      $.negation_type
    ),

    // Parameterized type
    // Example: ArrayRef[Int], HashRef[Str, Int]
    parameterized_type: $ => seq(
      field('base', $.identifier),
      '[',
      field('parameters', commaSep1($.type_expression)),
      ']'
    ),

    // Union type
    // Example: Int|Str
    union_type: $ => seq(
      field('left', $.type_expression),
      '|',
      field('right', $.type_expression)
    ),

    // Intersection type
    // Example: Object&Serializable
    intersection_type: $ => seq(
      field('left', $.type_expression),
      '&',
      field('right', $.type_expression)
    ),

    // Negation type
    // Example: !Int
    negation_type: $ => seq(
      '!',
      field('type', $.type_expression)
    ),

    // Typed parameter list
    // Example: (Type $param, AnotherType @array)
    typed_parameter_list: $ => seq(
      '(',
      optional(commaSep1($.typed_parameter)),
      ')'
    ),

    // Typed parameter
    // Example: Int $count, Str $name
    typed_parameter: $ => seq(
      field('type', $.type_expression),
      field('name', $._variable)
    ),

    // Return type annotation
    // Example: -> Int
    return_type_annotation: $ => $.type_expression,

    // Type assertion
    // Example: $var as Type
    type_assertion: $ => seq(
      field('expression', $._expression),
      'as',
      field('type', $.type_expression)
    )
};

// Helper functions for grammar definition

// commaSep1 parses a comma-separated list of items with at least one item
function commaSep1(rule) {
  return seq(rule, repeat(seq(',', rule)));
}

// commaSep parses a comma-separated list of items that may be empty
function commaSep(rule) {
  return optional(commaSep1(rule));
}
