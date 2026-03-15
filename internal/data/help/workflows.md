# Common PVM Workflows

💡 For command list: pvm -h
💡 For contextual help: pvm help

## New Project
Creating a New Perl Project:

1. Initialize the project:
   ```
   pvm workspace init my-app
   cd my-app
   ```

2. Add dependencies:
   ```
   pvm module add DBI
   pvm module add Test::More --dev
   ```

3. Install dependencies:
   ```
   pvm module install
   ```

4. Start development:
   ```
   pvm dev
   ```

This sets up a complete project with dependency management and development tools.

## Existing Project
Working with an Existing Project:

1. Check workspace status:
   ```
   pvm workspace status
   ```

2. Install dependencies:
   ```
   pvm module install
   ```

3. Run tests:
   ```
   pvm test
   ```

4. Start development mode:
   ```
   pvm dev
   ```

The dev command provides file watching, automatic builds, and test running.

## Module Development
Developing a CPAN Module:

1. Initialize with module template:
   ```
   pvm workspace init --template=module My::Module
   ```

2. Set up distribution metadata:
   ```
   # Edit pvm.toml to configure author, license, etc.
   ```

3. Develop with type checking:
   ```
   pvm dev
   ```

4. Build for distribution:
   ```
   pvm build
   ```

5. Test the distribution:
   ```
   cd build && perl Makefile.PL && make test
   ```

The build command creates a CPAN-ready distribution in the build/ directory.

## Testing
Testing Workflows:

1. Run all tests:
   ```
   pvm test
   ```

2. Run specific test file:
   ```
   pvm test t/basic.t
   ```

3. Run tests with coverage:
   ```
   pvm test --coverage
   ```

4. Continuous testing:
   ```
   pvm dev  # Includes automatic test running
   ```

5. Debug test failures:
   ```
   pvm test --verbose
   ```

Tests are automatically discovered in the t/ directory.

## Building
Build Workflows:

1. Development build (fast, with .pmc files):
   ```
   pvm build --inline
   ```

2. Distribution build (for CPAN):
   ```
   pvm build
   ```

3. Continuous building:
   ```
   pvm build --watch
   ```

4. Type-check only:
   ```
   pvm build --check-only
   ```

5. Clean build:
   ```
   pvm build --clean
   ```

Build outputs go to the build/ directory and include all necessary CPAN metadata.
