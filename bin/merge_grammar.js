#!/usr/bin/env node
// ABOUTME: Script to merge tree-sitter-perl base grammar with our type annotation extensions
// ABOUTME: Safely integrates our custom grammar rules into the tree-sitter-perl grammar

const fs = require('fs');
const path = require('path');

// Read the base grammar file
const grammarPath = './grammar.js';
const grammarContent = fs.readFileSync(grammarPath, 'utf8');

// Simple type annotation rules that work with tree-sitter grammar format
const typeAnnotationRules = `
    // Type annotation extensions
    field_declaration: $ => seq(
      'field',
      field('type', $.identifier),
      field('name', $.scalar),
      optional(seq('=', field('default', $._expr))),
      ';'
    ),

    type_declaration: $ => seq(
      'type',
      field('name', $.identifier),
      '=',
      field('definition', $.type_expression),
      ';'
    ),

    type_expression: $ => choice(
      $.identifier,
      $.parameterized_type,
      $.union_type
    ),

    parameterized_type: $ => seq(
      field('base', $.identifier),
      '[',
      field('parameters', $.identifier),
      ']'
    ),

    union_type: $ => seq(
      field('left', $.type_expression),
      '|',
      field('right', $.type_expression)
    ),
`;

// Check if already modified
if (grammarContent.includes('field_declaration:')) {
  console.log('Grammar already contains type annotations');
  process.exit(0);
}

// Clean up any existing extension attempts and find the ...primitives line to add our rules before it
let cleanedContent = grammarContent;

// Remove any existing extension lines at the end
cleanedContent = cleanedContent.replace(/\/\/ Include type annotation extensions[\s\S]*$/, '');

// Find the ...primitives line and add our rules before it
const modifiedContent = cleanedContent.replace(
  /(\s+)\.\.\.primitives,/,
  `$1${typeAnnotationRules}$1...primitives,`
);

// Write the modified grammar back
fs.writeFileSync(grammarPath, modifiedContent, 'utf8');
console.log('Successfully integrated type annotation extensions into grammar.js');
