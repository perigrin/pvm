# PVM Quickstart Guide

**Get up and running with PVM (Perl Version Manager) and typed-Perl in 15 minutes.**

PVM brings modern type safety to Perl through gradual typing while preserving full backward compatibility. This guide gets you started immediately with working examples.

## Installation

### Option 1: Build from Source (Recommended)
```bash
git clone https://github.com/perigrin/pvm.git
cd pvm
make
```

### Option 2: Download Binary
Download the latest release from the releases page and add to your PATH.

## Hello World Example

Create `hello.pl`:

```perl
#!/usr/bin/perl
use v5.36;

# Simple variable type annotations
my Int $count = 42;
my Str $name = "Hello World";
my Bool $is_active = 1;
my ArrayRef[Int] $numbers = [1, 2, 3, 4, 5];

# Function with type annotations
sub Int add_numbers(Int $a, Int $b) {
    return $a + $b;
}

# Usage
my Int $sum = add_numbers($count, 58);
say "Sum: $sum";
say "Name: $name";
```

**Run it:**
```bash
# Type-check and run
psc run hello.pl

# Or check for errors first
psc check hello.pl
```

## Core Type System

PVM supports intuitive type annotations:

```perl
# Simple types
my Str $message = "Hello";
my Int $count = 42;
my Bool $flag = 1;

# Container types
my ArrayRef[Str] $names = ["Alice", "Bob"];
my HashRef[Str, Int] $ages = { alice => 30, bob => 25 };

# Optional values
my Maybe[Str] $optional = undef;
```

## Object-Oriented Example

Create `user.pl`:

```perl
#!/usr/bin/perl
use v5.36;
use experimental 'class';

class UserAccount {
    field Str $username;
    field Str $email;
    field Int $user_id;

    method Str get_username() {
        return $username;
    }

    method Bool is_valid_email() {
        return $email =~ /\A[^@\s]+@[^@\s]+\z/;
    }
}

# Usage
my $user = UserAccount->new(
    username => "alice",
    email => "alice@example.com",
    user_id => 123
);

say "Username: " . $user->get_username();
say "Valid email: " . ($user->is_valid_email() ? "Yes" : "No");
```

**Run it:**
```bash
psc run user.pl
```

## Essential Commands

```bash
# Set Perl version for project
pvm local 5.36.0

# Install modules with type checking
pvi install JSON::XS DBI

# Type check files
psc check script.pl
psc check --recursive lib/

# Run with type checking
psc run script.pl

# Remove types for compatibility
psc strip script.pl

# Watch files for changes
psc watch lib/
```

## Project Setup (2 minutes)

```bash
# Create new project
mkdir my-typed-project && cd my-typed-project

# Set Perl version
pvm local 5.36.0

# Install dependencies
pvi install Moose JSON::XS

# Create main script with types
cat > main.pl << 'EOF'
#!/usr/bin/perl
use v5.36;
use JSON::XS;

my Str $json_text = '{"name": "John", "age": 30}';
my HashRef $data = decode_json($json_text);

sub Str greet(Str $name, Int $age) {
    return "Hello $name, you are $age years old!";
}

say greet($data->{name}, $data->{age});
EOF

# Type check and run
psc run main.pl
```

## Type Stripping for Production

PVM preserves compatibility by allowing type stripping:

```bash
# Remove all type annotations
psc strip main.pl > main_untyped.pl

# Now runs on any Perl interpreter
perl main_untyped.pl
```

## Integration with Existing Code

PVM works alongside existing tools:

```bash
# Works with existing version managers
pvm use 5.36.0  # Switches Perl version like perlbrew/plenv

# Type-check existing Perl code (gradually add types)
psc check existing_script.pl  # Reports type opportunities

# Install modules normally
pvi install Module::Name  # Like cpanm but with type awareness
```

## Common Patterns

### Error Handling
```perl
use Result qw(Ok Err);

sub Result[Num, Str] divide(Num $a, Num $b) {
    return Err("Division by zero") if $b == 0;
    return Ok($a / $b);
}

my $result = divide(10, 2);
$result->match(
    Ok => sub($value) { say "Result: $value" },
    Err => sub($error) { say "Error: $error" }
);
```

### Optional Values
```perl
sub Maybe[UserAccount] find_user(Str $username) {
    my $user = $db->find_user($username);
    return $user ? Some($user) : None();
}

my $maybe_user = find_user("alice");
if ($maybe_user->is_some()) {
    say "Found: " . $maybe_user->unwrap()->get_username();
}
```

## Next Steps

**For new projects:**
→ Read [New Development Workflow](workflow-new-development.md)

**For existing projects:**
→ Read [Existing Project Migration Workflow](workflow-existing-project-migration.md)

**For CI/CD setup:**
→ Read [CI/CD Integration Workflow](workflow-ci-cd-integration.md)

**For advanced patterns:**
→ Read [Typed-Perl Coding Patterns](workflow-typed-perl-coding-patterns.md)

**For large codebases:**
→ Read [Legacy Codebase Transformation](workflow-legacy-codebase-transformation.md)

**For complete reference:**
→ Read [Typed-Perl Specification](typed-perl-specification.md)

---

**Need help?** The PVM ecosystem includes 4 integrated tools:
- **PVM**: Version management
- **PSC**: Static type checking
- **PVI**: Package installation
- **PVX**: Isolated execution

All commands work together seamlessly. Start with `psc check` on any Perl file to see type opportunities, then gradually add annotations where they provide value.
