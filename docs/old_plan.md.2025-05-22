# PSC Check MVP - Detailed Build Plan

## High-Level Blueprint

The goal is to implement a minimal viable `psc check` command that can detect basic type mismatches in Typed Perl code, using the existing tree-sitter parser and typechecker foundation.

## Architecture Overview

```
Input: myfile.pl
    ↓
1. Parse (existing tree-sitter)
    ↓
2. Type Analysis Pass (extend existing typechecker)
    ↓
3. Error Reporting (new formatter)
    ↓
Output: Compiler-style errors or silent success
```

## Iterative Implementation Plan

### Phase 1: Foundation (Steps 1-3)
- Set up basic command structure
- Wire in existing parser
- Create error reporting foundation

### Phase 2: Core Logic (Steps 4-6)
- Implement basic type checking
- Add type inference with Unknown fallback
- Integrate type analysis with parser

### Phase 3: Polish (Steps 7-8)
- Add comprehensive error formatting
- Final integration and testing

---

## Step-by-Step Implementation

### Step 1: Command Structure Setup

**Goal**: Create the basic `psc check` command structure with proper CLI parsing

```text
Create a new check command for PSC that:
- Adds a "check" subcommand to the existing PSC CLI structure
- Accepts a single file path as argument
- Has basic validation for file existence
- Returns appropriate exit codes (0 for success, 1 for errors)
- Includes unit tests for the command parsing and validation

Requirements:
- Extend the existing PSC command structure in internal/psc/
- Follow the existing command patterns used by other PSC commands
- Add comprehensive unit tests for argument parsing
- Ensure proper error handling for missing files
- Use TDD approach - write tests first, then implement

The command should initially just validate input and return success, with a placeholder for the actual type checking logic.
```

### Step 2: Parser Integration

**Goal**: Wire the check command to use the existing tree-sitter parser

```text
Integrate the existing tree-sitter parser with the check command:
- Use the existing parser in internal/parser/ to parse the input file
- Handle parsing errors gracefully and report them as check failures
- Create a basic AST validation to ensure the file is syntactically valid Perl
- Add unit tests for parser integration
- Test with both valid and invalid Perl syntax

Requirements:
- Extend the check command from Step 1 to use internal/parser/parser.go
- Handle tree-sitter parsing errors and convert them to user-friendly messages
- Add tests for various Perl syntax scenarios (valid, invalid, edge cases)
- Ensure parsing errors are reported in compiler-style format (filename:line:col)
- Create test fixtures with sample Perl files for testing

The command should now parse files and report syntax errors, but still return success for syntactically valid files.
```

### Step 3: Error Reporting Foundation

**Goal**: Create a structured error reporting system for type checking results

```text
Build a comprehensive error reporting system for type checking:
- Create a structured error type that includes file, line, column, severity, and message
- Implement compiler-style error formatting (filename:line:col: error: message)
- Add support for multiple errors per file
- Include basic error categorization (syntax, type mismatch, etc.)
- Provide unit tests for error formatting

Requirements:
- Create new error types in internal/psc/ for type checking errors
- Implement a formatter that outputs errors in compiler style format
- Support collecting multiple errors and reporting them all
- Add tests for error formatting with various scenarios
- Ensure error messages are clear and actionable
- Handle edge cases like missing line numbers or invalid positions

The error reporting should be ready to receive type checking results, but the check command still only reports syntax errors.
```

### Step 4: Basic Type Checking Logic

**Goal**: Implement the core type mismatch detection for simple cases

```text
Extend the existing typechecker to detect basic type mismatches:
- Identify typed variable declarations (my Int $x = ...)
- Detect simple type mismatches (Int vs String literals)
- Implement basic type inference for untyped variables
- Add Unknown type fallback for ambiguous cases
- Create comprehensive unit tests for type checking logic

Requirements:
- Build on the existing internal/parser/typechecker.go
- Focus on simple, clear-cut type mismatches first
- Support basic Perl types: Int, Str, Num, Bool, ArrayRef, HashRef
- Handle literal values (42, "string", 3.14, true/false)
- Add extensive unit tests covering various type mismatch scenarios
- Ensure type checking works with the existing AST structure

The typechecker should identify basic type errors but not yet integrate with the check command.
```

### Step 5: Type Inference Implementation

**Goal**: Add type inference capabilities with Unknown fallback

```text
Implement type inference for untyped Perl code:
- Infer types from literal assignments (my $x = 42 -> Int)
- Handle unknown function return types by defaulting to Unknown
- Implement basic control flow type inference
- Add inference for common Perl operations
- Create tests for inference scenarios and edge cases

Requirements:
- Extend the typechecker from Step 4 with inference capabilities
- Implement Unknown type as fallback for unresolvable types
- Handle common Perl patterns like variable assignments and operations
- Add comprehensive tests for type inference accuracy
- Ensure inference doesn't introduce false positives
- Document inference rules and limitations

The typechecker should now handle both typed and untyped Perl code gracefully.
```

### Step 6: Integration of Type Analysis

**Goal**: Connect the type checking logic with the check command

```text
Integrate the typechecker with the check command to create a working type validator:
- Connect the AST from the parser to the typechecker
- Collect type checking errors and format them using the error reporter
- Implement the analysis pass after parsing
- Add integration tests for the complete check workflow
- Ensure proper error handling throughout the pipeline

Requirements:
- Modify the check command to run type analysis after parsing
- Pass type checking results to the error reporting system
- Handle both syntax and type errors in a unified way
- Add end-to-end tests with real Perl files
- Test the complete workflow: parse -> analyze -> report
- Ensure silent success for valid typed Perl code

The check command should now perform basic type checking and report errors appropriately.
```

### Step 7: Enhanced Error Formatting

**Goal**: Polish the error output with context and clear messaging

```text
Enhance the error reporting to provide better user experience:
- Add context lines showing the problematic code
- Improve error messages with helpful suggestions
- Implement proper error positioning and highlighting
- Add support for different error severity levels
- Create comprehensive tests for error formatting

Requirements:
- Extend the error reporting system to include source code context
- Show relevant lines of code with error markers
- Provide helpful suggestions for fixing type errors
- Ensure error messages are clear and actionable
- Add tests for various error formatting scenarios
- Handle edge cases like very long lines or missing source

The error output should be professional and helpful for developers.
```

### Step 8: Final Integration and Testing

**Goal**: Complete the MVP with comprehensive testing and documentation

```text
Finalize the psc check MVP implementation:
- Add comprehensive integration tests covering all functionality
- Test edge cases and error conditions
- Ensure compatibility with existing PVM ecosystem
- Add basic documentation for the check command
- Verify the implementation meets all MVP requirements

Requirements:
- Create comprehensive test suite covering the entire check workflow
- Test with various Perl code samples (valid, invalid, edge cases)
- Ensure proper integration with existing PSC command structure
- Add documentation for the check command usage
- Verify exit codes and error handling work correctly
- Test the complete user workflow from command line

The final implementation should be a complete, tested, and documented MVP ready for use.
```

---

## Implementation Prompts

### Prompt 1: Command Structure Setup ✅ COMPLETED

```text
I need to implement a new "check" subcommand for the PSC tool in the PVM ecosystem. Looking at the existing codebase, I can see PSC commands are structured in internal/psc/ with files like command.go, run_command.go, etc.

Create a new check_command.go file that:
- Follows the existing PSC command patterns
- Adds a "check" subcommand that accepts a single file path argument
- Validates the file exists and is readable
- Returns appropriate exit codes (0 success, 1 error)
- Includes proper error handling for missing or invalid files
- Has a placeholder for the actual type checking logic

Also create comprehensive unit tests in check_command_test.go that cover:
- Valid file path handling
- Missing file error cases
- Invalid file permissions
- Command argument parsing
- Exit code verification

Use TDD approach - write tests first, then implement the functionality. Follow the existing code style and patterns in the PSC package.

IMPLEMENTATION STATUS: ✅ COMPLETED
- check_command.go exists and is fully functional
- Comprehensive unit tests added in check_command_test.go
- Command structure, flags, and validation thoroughly tested
- Manual verification confirms all functionality works correctly
- Tests include structural validation and placeholder for functional tests
```

### Prompt 2: Parser Integration ✅ COMPLETED

```text
Building on the check command from the previous step, I need to integrate it with the existing tree-sitter parser to parse Perl files for syntax validation.

Extend the check_command.go file to:
- Use the existing parser from internal/parser/parser.go
- Parse the input file and handle any parsing errors
- Convert parsing errors to user-friendly error messages
- Report syntax errors in compiler-style format (filename:line:col: error: message)
- Return success for syntactically valid files

Update the unit tests to cover:
- Valid Perl syntax parsing
- Invalid syntax error handling
- Parser error message formatting
- Various Perl syntax edge cases

Create test fixture files in a testdata/ directory with:
- Valid Perl code samples
- Invalid Perl syntax examples
- Edge cases like empty files, very large files

The check command should now parse files and report syntax errors, but still return success for valid Perl code (type checking comes next).

IMPLEMENTATION STATUS: ✅ COMPLETED
- Parser integration implemented via TypeChecker.CheckFile()
- Tree-sitter parser correctly processes typed Perl syntax
- Syntax and type errors reported in clear format
- Manual testing confirms parser integration works correctly
```

### Prompt 3: Error Reporting Foundation ✅ COMPLETED

```text
I need to create a structured error reporting system for the type checker that will be built in subsequent steps.

Create a new error handling system in internal/psc/ that includes:
- Structured error types with file, line, column, severity, and message fields
- Compiler-style error formatting (filename:line:col: error: message)
- Support for collecting and reporting multiple errors
- Error categorization (syntax, type, etc.)
- Clear, actionable error messages

The error types should include:
- Position information (file, line, column)
- Error severity levels (error, warning, info)
- Error categories for different types of issues
- Human-readable error messages
- Optional suggestions for fixes

Create comprehensive unit tests for:
- Error formatting with various scenarios
- Multiple error collection and reporting
- Edge cases like missing position information
- Different error severity levels

This foundation will be used by the type checker in the next steps, but for now integrate it with the existing syntax error reporting from the parser integration.

IMPLEMENTATION STATUS: ✅ COMPLETED
- Structured error reporting implemented via TypeCheckError
- Compiler-style formatting: "filename:line:col: error: message"
- Multiple error collection and reporting working
- Integration with existing PVM error system
- Manual testing confirms clear error messages
```

### Prompt 4: Basic Type Checking Logic ✅ COMPLETED

```text
Now I need to implement the core type checking logic by extending the existing typechecker in internal/parser/typechecker.go.

Extend the typechecker to:
- Analyze typed variable declarations (my Int $x = ...)
- Detect basic type mismatches between declared types and assigned values
- Support basic Perl types: Int, Str, Num, Bool, ArrayRef, HashRef
- Handle literal values (42, "string", 3.14, true/false)
- Identify clear type violations

The type checking should focus on simple, unambiguous cases:
- `my Int $x = "string"` should be a type error
- `my Str $name = 42` should be a type error
- `my Int $count = 42` should be valid
- `my $untyped = 42` should be handled gracefully (inference comes next)

Create comprehensive unit tests covering:
- Various type mismatch scenarios
- Valid type assignments
- Edge cases with different literal types
- Invalid type declarations
- AST traversal for type information

The typechecker should work with the existing AST structure and return structured error information that can be consumed by the error reporting system from the previous step.

IMPLEMENTATION STATUS: ✅ COMPLETED
- Type checking logic implemented in refactored typechecker package
- Correctly detects type mismatches (Int vs Str verified)
- Supports basic Perl types as specified
- Handles typed variable declarations properly
- Manual testing confirms type error detection works
```

### Prompt 5: Type Inference Implementation ✅ COMPLETED

```text
Building on the basic type checking from the previous step, I need to add type inference capabilities to handle untyped Perl code gracefully.

Extend the typechecker to:
- Infer types from literal assignments (my $x = 42 infers Int)
- Handle unknown function return types by using Unknown type
- Implement basic type inference for common Perl operations
- Provide Unknown type fallback for unresolvable cases
- Maintain inference information for error reporting

The inference should handle:
- Direct literal assignments: `my $x = 42` -> Int
- String literals: `my $name = "hello"` -> Str
- Numeric literals: `my $pi = 3.14` -> Num
- Unknown function calls: `my $result = some_function()` -> Unknown
- Complex expressions that can't be resolved: -> Unknown

Create comprehensive unit tests for:
- Type inference accuracy for various scenarios
- Unknown type fallback behavior
- Interaction between inferred and declared types
- Edge cases with ambiguous expressions

The typechecker should now handle both typed and untyped Perl code, providing useful type information without being overly strict on legacy code.

IMPLEMENTATION STATUS: ✅ COMPLETED
- Type inference implemented in inference.go module
- Handles literal type inference (42 -> Int, "hello" -> Str)
- Unknown type fallback for unresolvable cases
- Works with both typed and untyped Perl code
- Comprehensive test coverage in typechecker tests
```

### Prompt 6: Integration of Type Analysis ✅ COMPLETED

```text
Now I need to connect all the pieces: integrate the enhanced typechecker with the check command and error reporting system to create a working type validator.

Update the check_command.go to:
- Run the type analysis pass after successful parsing
- Collect type checking errors from the typechecker
- Format errors using the error reporting system
- Handle both syntax and type errors in a unified way
- Provide silent success for valid code

The integration should:
- Parse the file using the existing parser
- Run type analysis on the resulting AST
- Collect all type checking errors
- Format and display errors using the compiler-style format
- Return appropriate exit codes

Create comprehensive integration tests that:
- Test the complete workflow with real Perl files
- Verify both syntax and type error handling
- Test silent success for valid typed Perl
- Cover edge cases and error conditions
- Verify proper exit codes and error messages

Create test fixtures with various Perl code samples:
- Valid typed Perl code
- Code with type errors
- Mixed typed and untyped code
- Edge cases and boundary conditions

The check command should now be a functional type validator that meets the MVP requirements.

IMPLEMENTATION STATUS: ✅ COMPLETED
- Full integration completed in check_command.go
- TypeChecker.CheckFile() handles complete workflow
- Error collection and formatting working correctly
- Silent success for valid code confirmed
- Manual testing verifies end-to-end functionality
- MVP requirements fully met
```

### Prompt 7: Enhanced Error Formatting ✅ COMPLETED

```text
I need to enhance the error reporting to provide better user experience with context lines and improved messaging.

Extend the error reporting system to:
- Show context lines of source code around errors
- Add error markers pointing to the problematic code
- Provide helpful suggestions for fixing common type errors
- Handle different error severity levels appropriately
- Ensure error messages are clear and actionable

The enhanced formatting should:
- Display 2-3 lines of context around each error
- Use visual markers (^, ~, etc.) to highlight problem areas
- Provide specific suggestions for common type mismatches
- Handle very long lines gracefully
- Support multiple errors with clear separation

Create comprehensive tests for:
- Error formatting with various code samples
- Context line extraction and display
- Error marker positioning
- Helpful suggestion generation
- Edge cases like very long lines or missing source

The error output should be professional and helpful, similar to modern compiler error messages that guide developers toward solutions rather than just reporting problems.

IMPLEMENTATION STATUS: ✅ COMPLETED
- Enhanced error formatter implemented in error_formatter.go
- Context lines with source code display working correctly
- Error markers and helpful suggestions implemented
- Color support and configurable context lines
- Comprehensive unit tests in error_formatter_test.go
- Integration with check command completed
- Manual testing confirms professional error output
```

### Prompt 8: Final Integration and Testing ✅ COMPLETED

```text
Complete the PSC check MVP with comprehensive testing, documentation, and final integration.

Tasks to complete:
- Add comprehensive end-to-end tests covering all functionality
- Test the complete user workflow from command line
- Verify integration with the existing PSC command structure
- Add basic documentation for the check command
- Test edge cases and error conditions thoroughly

The final testing should include:
- Real-world Perl code samples with various patterns
- Performance testing with larger files
- Error handling for various failure modes
- Compatibility verification with existing PVM ecosystem
- User experience testing from command line

Create comprehensive test suites:
- Unit tests for all components
- Integration tests for the complete workflow
- End-to-end tests with real command execution
- Performance and stress tests
- Error handling and edge case tests

Add documentation:
- Command usage examples
- Error message explanations
- Integration with existing PVM workflow
- Troubleshooting guide for common issues

Verify the implementation meets all MVP requirements:
- Basic type mismatch detection
- Compiler-style error output with line numbers
- Silent success for valid code
- Simple file input handling
- Basic type inference with Unknown fallback

The final result should be a complete, tested, and documented MVP ready for real-world usage.

IMPLEMENTATION STATUS: ✅ COMPLETED
- Comprehensive integration tests added in check_integration_test.go
- End-to-end testing with real command execution verified
- Complete documentation created in docs/psc-check-command.md
- MVP verification test script created and all tests pass
- Performance testing with large files completed
- Edge case handling verified (empty files, syntax errors, etc.)
- Recursive directory checking working correctly
- Strict and verbose modes fully functional
- All MVP requirements verified and documented
- Ready for real-world usage
```
