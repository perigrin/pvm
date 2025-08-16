/**
 * Tree-sitter grammar tests for variable declarations with initializers
 *
 * These tests verify that the grammar correctly parses Perl variable
 * declarations with assignment expressions.
 *
 * CRITICAL ISSUE: Currently the grammar does not capture initializers
 * in variable_declaration nodes, which breaks safety analysis.
 */

describe('Variable Declaration Parsing', () => {

  describe('Simple variable declarations', () => {
    it('should parse basic variable declaration without initializer', () => {
      const code = 'my $name;';
      const tree = parser.parse(code);

      expect(tree.rootNode.hasError()).toBe(false);

      const varDecl = tree.rootNode.descendantForTypeString('variable_declaration');
      expect(varDecl).not.toBeNull();
      expect(varDecl.text).toBe('my $name');
    });

    it('should parse variable declaration with string literal initializer', () => {
      const code = 'my $name = "John";';
      const tree = parser.parse(code);

      expect(tree.rootNode.hasError()).toBe(false);

      const varDecl = tree.rootNode.descendantForTypeString('variable_declaration');
      expect(varDecl).not.toBeNull();

      // CRITICAL TEST: This should include the entire declaration with initializer
      // Currently FAILING: varDecl.text only contains "my $name"
      expect(varDecl.text).toBe('my $name = "John"');

      // The variable_declaration node should have an initializer field
      const initializer = varDecl.childForFieldName('initializer');
      expect(initializer).not.toBeNull();
      expect(initializer.text).toBe('"John"');
    });
  });

  describe('Variable declarations with hash access', () => {
    it('should parse variable declaration with simple hash access', () => {
      const code = 'my $name = $input->{name};';
      const tree = parser.parse(code);

      expect(tree.rootNode.hasError()).toBe(false);

      const varDecl = tree.rootNode.descendantForTypeString('variable_declaration');
      expect(varDecl).not.toBeNull();

      // CRITICAL TEST: This is the core issue causing safety analysis failures
      // The variable_declaration should include the hash access expression
      expect(varDecl.text).toBe('my $name = $input->{name}');

      // The initializer should be a hash access expression
      const initializer = varDecl.childForFieldName('initializer');
      expect(initializer).not.toBeNull();
      expect(initializer.type).toBe('hash_access'); // or similar
      expect(initializer.text).toBe('$input->{name}');
    });

    it('should parse variable declaration with nested hash access', () => {
      const code = 'my $host = $config->{database}->{host};';
      const tree = parser.parse(code);

      expect(tree.rootNode.hasError()).toBe(false);

      const varDecl = tree.rootNode.descendantForTypeString('variable_declaration');
      expect(varDecl).not.toBeNull();
      expect(varDecl.text).toBe('my $host = $config->{database}->{host}');

      const initializer = varDecl.childForFieldName('initializer');
      expect(initializer).not.toBeNull();
      expect(initializer.text).toBe('$config->{database}->{host}');
    });
  });

  describe('Variable declarations with other expressions', () => {
    it('should parse variable declaration with array access', () => {
      const code = 'my $first = $data->[0];';
      const tree = parser.parse(code);

      expect(tree.rootNode.hasError()).toBe(false);

      const varDecl = tree.rootNode.descendantForTypeString('variable_declaration');
      expect(varDecl).not.toBeNull();
      expect(varDecl.text).toBe('my $first = $data->[0]');

      const initializer = varDecl.childForFieldName('initializer');
      expect(initializer).not.toBeNull();
      expect(initializer.text).toBe('$data->[0]');
    });

    it('should parse variable declaration with method call', () => {
      const code = 'my $result = $obj->method();';
      const tree = parser.parse(code);

      expect(tree.rootNode.hasError()).toBe(false);

      const varDecl = tree.rootNode.descendantForTypeString('variable_declaration');
      expect(varDecl).not.toBeNull();
      expect(varDecl.text).toBe('my $result = $obj->method()');

      const initializer = varDecl.childForFieldName('initializer');
      expect(initializer).not.toBeNull();
      expect(initializer.text).toBe('$obj->method()');
    });

    it('should parse variable declaration with function call', () => {
      const code = 'my $length = length($string);';
      const tree = parser.parse(code);

      expect(tree.rootNode.hasError()).toBe(false);

      const varDecl = tree.rootNode.descendantForTypeString('variable_declaration');
      expect(varDecl).not.toBeNull();
      expect(varDecl.text).toBe('my $length = length($string)');

      const initializer = varDecl.childForFieldName('initializer');
      expect(initializer).not.toBeNull();
      expect(initializer.text).toBe('length($string)');
    });

    it('should parse variable declaration with defined-or operator', () => {
      const code = 'my $timeout = $config->{timeout} // 30;';
      const tree = parser.parse(code);

      expect(tree.rootNode.hasError()).toBe(false);

      const varDecl = tree.rootNode.descendantForTypeString('variable_declaration');
      expect(varDecl).not.toBeNull();
      expect(varDecl.text).toBe('my $timeout = $config->{timeout} // 30');

      const initializer = varDecl.childForFieldName('initializer');
      expect(initializer).not.toBeNull();
      expect(initializer.text).toBe('$config->{timeout} // 30');
    });
  });

  describe('Multiple variable declarations', () => {
    it('should parse multiple variable declarations with initializers', () => {
      const code = `
        my $name = $input->{name};
        my $id = $input->{user_id};
      `;
      const tree = parser.parse(code);

      expect(tree.rootNode.hasError()).toBe(false);

      const varDecls = tree.rootNode.descendantsOfType('variable_declaration');
      expect(varDecls).toHaveLength(2);

      // Check first declaration
      expect(varDecls[0].text).toBe('my $name = $input->{name}');
      const init1 = varDecls[0].childForFieldName('initializer');
      expect(init1).not.toBeNull();
      expect(init1.text).toBe('$input->{name}');

      // Check second declaration
      expect(varDecls[1].text).toBe('my $id = $input->{user_id}');
      const init2 = varDecls[1].childForFieldName('initializer');
      expect(init2).not.toBeNull();
      expect(init2.text).toBe('$input->{user_id}');
    });
  });

  describe('Different declaration types', () => {
    it('should parse state variable with initializer', () => {
      const code = 'state $counter = 0;';
      const tree = parser.parse(code);

      expect(tree.rootNode.hasError()).toBe(false);

      const varDecl = tree.rootNode.descendantForTypeString('variable_declaration');
      expect(varDecl).not.toBeNull();
      expect(varDecl.text).toBe('state $counter = 0');

      const initializer = varDecl.childForFieldName('initializer');
      expect(initializer).not.toBeNull();
      expect(initializer.text).toBe('0');
    });

    it('should parse our variable with initializer', () => {
      const code = "our $VERSION = '1.0';";
      const tree = parser.parse(code);

      expect(tree.rootNode.hasError()).toBe(false);

      const varDecl = tree.rootNode.descendantForTypeString('variable_declaration');
      expect(varDecl).not.toBeNull();
      expect(varDecl.text).toBe("our $VERSION = '1.0'");

      const initializer = varDecl.childForFieldName('initializer');
      expect(initializer).not.toBeNull();
      expect(initializer.text).toBe("'1.0'");
    });
  });

  describe('Grammar structure verification', () => {
    it('should have proper AST structure for variable declaration with initializer', () => {
      const code = 'my $name = $input->{name};';
      const tree = parser.parse(code);

      expect(tree.rootNode.hasError()).toBe(false);

      const varDecl = tree.rootNode.descendantForTypeString('variable_declaration');
      expect(varDecl).not.toBeNull();

      // Verify expected fields are present
      const declType = varDecl.childForFieldName('lexical'); // or similar field name
      expect(declType).not.toBeNull();
      expect(declType.text).toBe('my');

      const variable = varDecl.childForFieldName('variable'); // or similar field name
      expect(variable).not.toBeNull();
      expect(variable.text).toBe('$name');

      // CRITICAL: The initializer field should exist
      const initializer = varDecl.childForFieldName('initializer');
      expect(initializer).not.toBeNull();
      expect(initializer.text).toBe('$input->{name}');

      // Log the actual structure for debugging if test fails
      if (!initializer) {
        console.log('Variable declaration children:');
        for (let i = 0; i < varDecl.childCount; i++) {
          const child = varDecl.child(i);
          console.log(`  Child ${i}: ${child.type} = "${child.text}"`);
        }

        console.log('Variable declaration named children:');
        for (let i = 0; i < varDecl.namedChildCount; i++) {
          const child = varDecl.namedChild(i);
          console.log(`  Named child ${i}: ${child.type} = "${child.text}"`);
        }
      }
    });
  });
});

/**
 * Instructions for fixing the grammar:
 *
 * The variable_declaration rule in grammar.js currently looks like:
 *
 * variable_declaration: $ => prec.left(TERMPREC.QUESTION_MARK + 1,
 *   seq(
 *     choice('my', 'state', 'our', 'field', 'local'),
 *     optional(choice(
 *       field('type', $.type_expression),
 *       seq('(', field('type', $.type_expression), ')')
 *     )),
 *     choice(
 *       field('variable', $._declared_vars),
 *       field('variables', $._decl_variable_list)),
 *     optseq(':', optional(field('attributes', $.attrlist))))
 * ),
 *
 * It should be updated to include an optional initializer:
 *
 * variable_declaration: $ => prec.left(TERMPREC.QUESTION_MARK + 1,
 *   seq(
 *     choice('my', 'state', 'our', 'field', 'local'),
 *     optional(choice(
 *       field('type', $.type_expression),
 *       seq('(', field('type', $.type_expression), ')')
 *     )),
 *     choice(
 *       field('variable', $._declared_vars),
 *       field('variables', $._decl_variable_list)),
 *     optional(seq('=', field('initializer', $._term))),  // ADD THIS LINE
 *     optseq(':', optional(field('attributes', $.attrlist))))
 * ),
 *
 * This would allow the parser to capture the entire variable declaration
 * including the assignment expression, making it available for safety analysis.
 */
