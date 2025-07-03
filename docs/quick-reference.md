# PVM Quick Reference

**Fast lookup for common PVM commands and workflows.**

## Most Common Commands

### Getting Started
```bash
# Install PVM (build from source)
git clone https://github.com/your-username/pvm.git && cd pvm && make

# Install a Perl version
pvm install 5.38.0

# Set global default
pvm global 5.38.0

# Initialize new project
pvm project init my-app
cd my-app
```

### Configuration
```bash
# Show current config
pvm config show

# Set project Perl version
pvm config set project.perl_version 5.38.0 --project

# Get a config value
pvm config get pvm.default_perl

# Initialize project config
pvm config init --project
```

### Module Management
```bash
# Add module to project
pvm module add DBI

# Install all dependencies
pvm module install

# Install with dev dependencies
pvm module install --dev

# Generate lockfile
pvm module sync
```

### Development Workflow
```bash
# Start development environment
pvm dev

# Type check code
psc check lib/

# Run with type checking
psc run script.pl

# Build project
pvm build

# Run tests
pvm test
```

### Build System
```bash
# Development build (.pmc files)
pvm build --inline

# Distribution build (CPAN-ready)
pvm build

# Continuous building
pvm build --watch

# Clean build
pvm build --clean
```

## Project Lifecycle Cheat Sheet

### 1. New Project
```bash
pvm project init my-app
cd my-app
pvm config set project.perl_version 5.38.0 --project
pvm module add Test::More --dev
```

### 2. Existing Project
```bash
cd existing-project
pvm project init .
pvm project doctor --fix
pvm module install
```

### 3. Development
```bash
pvm dev                    # Start development environment
# Edit code in another terminal
pvm test                   # Run tests
pvm build --inline        # Quick build
```

### 4. Release
```bash
pvm test                   # Ensure tests pass
pvm build                  # Create distribution
# Build artifacts in build/ directory
```

## Type Annotation Quick Reference

### Variables
```perl
my Str $name = "Alice";
my Int $age = 30;
my Bool $active = 1;
my ArrayRef[Str] $names = ["Alice", "Bob"];
my HashRef[Int] $scores = { alice => 95, bob => 87 };
```

### Functions
```perl
sub Int add(Int $a, Int $b) {
    return $a + $b;
}

sub Maybe[Str] maybe_find(Str $key) {
    # Can return Str or undef
}
```

### Classes
```perl
class User {
    field Str $name :param;
    field Int $age :param;

    method Str display() {
        return "$name (age $age)";
    }
}
```

### Union Types
```perl
my Str|Int $id = get_id();           # Either string or int
my User|Undef $user = find_user();   # User object or undef
```

## Error Troubleshooting

### Type Errors
```bash
# Check for type errors
psc check lib/

# Verbose type checking
psc check --verbose lib/

# Fix and recheck
psc check --fix lib/
```

### Module Issues
```bash
# Check project health
pvm project doctor

# Fix common issues
pvm project doctor --fix

# Reinstall dependencies
rm -rf lib/ && pvm module install
```

### Build Issues
```bash
# Clean build
pvm build --clean

# Check build configuration
pvm config get build

# Debug build
pvm build --verbose
```

### Configuration Issues
```bash
# Validate configuration
pvm config validate

# Show config sources
pvm config sources

# Reset to defaults
pvm config backup
pvm config init --force
```

## File Structure Reference

### Project Layout
```
my-project/
├── .perl-version          # Perl version for this project
├── cpanfile               # Dependencies
├── cpanfile.snapshot      # Lockfile (generated)
├── pvm.toml              # Project configuration
├── lib/                  # Project modules
│   └── MyProject.pm
├── script/               # Executable scripts
├── t/                    # Tests
└── build/               # Build output (generated)
```

### Configuration Files
```
~/.config/pvm/
├── pvm.toml              # User configuration
├── templates/            # Project templates
└── backups/             # Configuration backups

/etc/pvm/
└── pvm.toml             # System configuration

.pvm/
└── pvm.toml             # Project configuration (alternative location)
```

## Integration Quick Setups

### Shell Integration
```bash
# Add to ~/.bashrc or ~/.zshrc
eval "$(pvm init)"

# Or generate init files
pvm shell init
```

### CI/CD Integration
```yaml
# GitHub Actions example
steps:
  - uses: actions/checkout@v3
  - name: Install PVM
    run: |
      git clone https://github.com/your-username/pvm.git
      cd pvm && make && sudo make install
  - name: Install dependencies
    run: pvm module install
  - name: Run tests
    run: pvm test
  - name: Build
    run: pvm build
```

### Editor Integration
```bash
# For LSP support (if available)
pvm mcp-server &  # Start language server
```

## Performance Tips

### Development
- Use `pvm dev` for integrated development environment
- Use `pvm build --inline` for fast development builds
- Use `--watch` flags for continuous building

### Production
- Use `pvm build` for CPAN-ready distributions
- Use `pvm module sync` to generate lockfiles
- Use `--parallel` flags for faster operations

### Debugging
- Use `--verbose` flags for detailed output
- Use `pvm project doctor` for health checks
- Use `pvm config validate` for configuration issues

---

*For complete command documentation, see [command-reference.md](command-reference.md).*
*For detailed workflows, see the workflow guides in the main documentation.*
