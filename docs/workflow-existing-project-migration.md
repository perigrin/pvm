# Existing Project Migration Workflow

## Document Information
- **Title**: Workflow for Adding PVM to Existing Perl Codebases
- **Purpose**: Guide for migrating established Perl projects to use PVM's typed-Perl features gradually
- **Target Audience**: Developers with existing Perl projects who want to adopt typed-Perl incrementally
- **Prerequisites**: Existing Perl codebase, basic understanding of type systems

## Table of Contents

1. [Codebase Assessment and Planning](#1-codebase-assessment-and-planning)
2. [PVM Setup Alongside Existing Tools](#2-pvm-setup-alongside-existing-tools)
3. [Gradual Type Annotation Strategy](#3-gradual-type-annotation-strategy)
4. [PSC Integration for Legacy Analysis](#4-psc-integration-for-legacy-analysis)
5. [Migration Patterns for Common Code Structures](#5-migration-patterns-for-common-code-structures)
6. [Coexistence with Existing Tooling](#6-coexistence-with-existing-tooling)
7. [Risk Mitigation and Rollback Procedures](#7-risk-mitigation-and-rollback-procedures)

---

## 1. Codebase Assessment and Planning

### 1.1 Pre-Migration Analysis

Before introducing PVM, assess your existing codebase to plan an effective migration strategy.

**Create Assessment Script - `bin/assess-codebase.sh`:**
```bash
#!/bin/bash
# Codebase assessment for PVM migration

echo "=== PVM Migration Assessment ==="

# Basic project metrics
echo "## Project Structure"
echo "Perl files: $(find . -name "*.pl" -o -name "*.pm" -o -name "*.t" | wc -l)"
echo "Lines of code: $(find . -name "*.pl" -o -name "*.pm" -o -name "*.t" -exec wc -l {} + | tail -1)"
echo "Modules: $(find . -name "*.pm" | wc -l)"
echo "Scripts: $(find . -name "*.pl" | wc -l)"
echo "Tests: $(find . -name "*.t" | wc -l)"

# Perl version detection
echo -e "\n## Current Perl Setup"
echo "System Perl: $(perl -v | grep version)"
if command -v plenv >/dev/null 2>&1; then
    echo "plenv detected: $(plenv version)"
fi
if command -v perlbrew >/dev/null 2>&1; then
    echo "perlbrew detected: $(perlbrew list | grep '*' || echo 'No active version')"
fi

# Dependencies analysis
echo -e "\n## Dependencies"
if [ -f "cpanfile" ]; then
    echo "cpanfile found - $(grep -c 'requires' cpanfile) dependencies"
elif [ -f "Makefile.PL" ]; then
    echo "Makefile.PL found - traditional CPAN module"
elif [ -f "Build.PL" ]; then
    echo "Build.PL found - Module::Build setup"
else
    echo "No dependency file detected"
fi

# Code complexity indicators
echo -e "\n## Complexity Indicators"
echo "use statements: $(grep -r "^use " --include="*.pl" --include="*.pm" . | wc -l)"
echo "package declarations: $(grep -r "^package " --include="*.pm" . | wc -l)"
echo "subroutines: $(grep -r "^sub " --include="*.pl" --include="*.pm" . | wc -l)"

# Type annotation readiness
echo -e "\n## Type Annotation Candidates"
echo "Variable declarations: $(grep -r "my \\\$" --include="*.pl" --include="*.pm" . | wc -l)"
echo "Function definitions: $(grep -r "sub [a-zA-Z_]" --include="*.pl" --include="*.pm" . | wc -l)"

echo -e "\n## Recommendation"
if [ $(find . -name "*.pm" | wc -l) -lt 5 ]; then
    echo "Small project - suitable for rapid migration"
elif [ $(find . -name "*.pm" | wc -l) -lt 20 ]; then
    echo "Medium project - plan phased migration"
else
    echo "Large project - start with pilot module"
fi
```

```bash
chmod +x bin/assess-codebase.sh
./bin/assess-codebase.sh
```

### 1.2 Migration Planning Matrix

Create a migration plan based on your assessment:

**Small Projects (< 5 modules)**
- **Timeline**: 1-2 weeks
- **Strategy**: Direct migration module by module
- **Risk**: Low - quick rollback possible

**Medium Projects (5-20 modules)**
- **Timeline**: 1-2 months
- **Strategy**: Core modules first, then periphery
- **Risk**: Medium - staged approach

**Large Projects (20+ modules)**
- **Timeline**: 3-6 months
- **Strategy**: Pilot module, then gradual expansion
- **Risk**: High - requires careful planning

### 1.3 Module Priority Classification

Classify modules by migration priority:

```perl
# Create assessment/module-priority.pl
use v5.20;
use strict;
use warnings;
use File::Find;
use File::Slurp;

my %modules;
my %dependencies;

# Analyze all Perl modules
find(sub {
    return unless /\.pm$/;
    my $file = $File::Find::name;
    my $content = read_file($file);

    # Extract module name
    my ($module) = $content =~ /package\s+([^;\s]+)/;
    return unless $module;

    # Count lines and subroutines
    my @lines = split /\n/, $content;
    my $sub_count = () = $content =~ /^\s*sub\s+/gm;

    # Check for complex patterns
    my $has_oo = $content =~ /bless|Moose|Moo|Class::/;
    my $has_exports = $content =~ /Exporter|EXPORT/;
    my $use_count = () = $content =~ /^\s*use\s+/gm;

    $modules{$module} = {
        file => $file,
        lines => scalar @lines,
        subs => $sub_count,
        has_oo => $has_oo,
        has_exports => $has_exports,
        dependencies => $use_count,
        complexity => _calculate_complexity($content)
    };

}, '.');

# Output priority recommendations
say "=== Module Migration Priority ===\n";

say "HIGH PRIORITY (Start Here):";
for my $module (sort_by_priority('high')) {
    say "  $module - " . priority_reason($module, 'high');
}

say "\nMEDIUM PRIORITY (Second Wave):";
for my $module (sort_by_priority('medium')) {
    say "  $module - " . priority_reason($module, 'medium');
}

say "\nLOW PRIORITY (Final Wave):";
for my $module (sort_by_priority('low')) {
    say "  $module - " . priority_reason($module, 'low');
}

sub _calculate_complexity {
    my $content = shift;
    my $score = 0;
    $score += () = $content =~ /if\s*\(/g;  # Conditionals
    $score += () = $content =~ /while|for|foreach/g;  # Loops
    $score += () = $content =~ /eval\s*[{"]/g;  # Dynamic code
    return $score;
}

sub sort_by_priority {
    my $priority = shift;

    return grep {
        my $m = $modules{$_};
        if ($priority eq 'high') {
            $m->{lines} < 200 && $m->{complexity} < 10 && !$m->{has_oo}
        } elsif ($priority eq 'medium') {
            $m->{lines} < 500 && $m->{complexity} < 25
        } else {
            1  # Everything else
        }
    } keys %modules;
}

sub priority_reason {
    my ($module, $priority) = @_;
    my $m = $modules{$module};

    if ($priority eq 'high') {
        return "Simple, small module ($m->{lines} lines, $m->{subs} subs)";
    } elsif ($priority eq 'medium') {
        return "Moderate complexity ($m->{lines} lines, complexity: $m->{complexity})";
    } else {
        return "Complex module - handle after experience gained";
    }
}
```

## 2. PVM Setup Alongside Existing Tools

### 2.1 Non-Destructive Installation

PVM is designed to coexist with existing Perl version managers.

**Install PVM Without Disrupting Existing Setup:**
```bash
# Install PVM (doesn't interfere with plenv/perlbrew)
git clone https://github.com/your-username/pvm.git
cd pvm
make build

# Add to PATH without replacing existing tools
echo 'export PATH="$HOME/pvm/bin:$PATH"' >> ~/.bashrc_pvm
echo 'source ~/.bashrc_pvm' >> ~/.bashrc

# Verify existing tools still work
perl -v
plenv version 2>/dev/null || perlbrew list 2>/dev/null || echo "No existing version manager"
```

**Import Existing Perl Installations:**
```bash
# Register existing plenv installations
pvm import-from plenv

# Register existing perlbrew installations
pvm import-from perlbrew

# Verify import
pvm versions
```

### 2.2 Project-Specific PVM Configuration

Create PVM configuration that doesn't interfere with existing workflows:

**Create `.pvm/pvm.toml` in your project:**
```toml
[pvm]
# Use existing Perl version initially
default_perl = "system"  # or your current version

[psc]
# Start with relaxed mode for gradual adoption
strict_mode = false
enable_flow_sensitive = true
show_context_lines = 3

# Only check files we're actively migrating
watch_include = ["lib/MyApp/Migrated/*.pm"]
watch_exclude = ["**"]

[pvi]
# Don't auto-install deps to avoid conflicts
auto_install_deps = false
test_during_install = false

[pvx]
# Use minimal isolation to avoid disrupting existing workflows
isolation_level = "low"
```

### 2.3 Gradual Toolchain Introduction

Introduce PVM tools gradually without disrupting existing development:

**Phase 1: Analysis Only**
```bash
# Use PSC for analysis without changing code
psc check lib/MyApp/Simple.pm  # Start with one simple module
psc check --format json lib/ > type-analysis.json  # Generate baseline
```

**Phase 2: Development Integration**
```bash
# Add type checking to development workflow
alias perl-check="psc check"
alias perl-run="psc run"

# Use alongside existing tools
make test  # Your existing test command
perl-check lib/  # Additional type checking
```

**Phase 3: Build Integration**
```bash
# Add to Makefile without disrupting existing targets
typecheck:
	psc check --recursive lib/

test-with-types: typecheck test

# Optional target that extends existing workflow
```

## 3. Gradual Type Annotation Strategy

### 3.1 Start with High-Value, Low-Risk Modules

Begin with modules identified as high-priority in your assessment.

**Example: Migrating a Simple Utility Module**

Original `lib/MyApp/Utils.pm`:
```perl
package MyApp::Utils;
use strict;
use warnings;
use Exporter 'import';

our @EXPORT_OK = qw(trim format_currency);

sub trim {
    my $str = shift;
    $str =~ s/^\s+|\s+$//g;
    return $str;
}

sub format_currency {
    my ($amount, $symbol) = @_;
    $symbol ||= '$';
    return sprintf("%s%.2f", $symbol, $amount);
}

1;
```

**Step 1: Add Basic Type Annotations**
```perl
package MyApp::Utils;
use v5.38;  # Update Perl version for better type support
use Exporter 'import';

our @EXPORT_OK = qw(trim format_currency);

sub Str trim(Str $str) {
    $str =~ s/^\s+|\s+$//g;
    return $str;
}

sub Str format_currency(Num $amount, Str $symbol = '$') {
    return sprintf("%s%.2f", $symbol, $amount);
}

1;
```

**Step 2: Validate and Test**
```bash
# Check types
psc check lib/MyApp/Utils.pm

# Test with existing test suite
perl t/utils.t

# Run type-checked tests
psc run t/utils.t
```

### 3.2 Progressive Type Enhancement

Gradually enhance type coverage:

**Level 1: Function Signatures Only**
```perl
sub HashRef process_data($data) {
    # Implementation unchanged
}
```

**Level 2: Add Parameter Types**
```perl
sub HashRef[Str, Any] process_data(ArrayRef[HashRef] $data) {
    # Implementation unchanged
}
```

**Level 3: Add Variable Types**
```perl
sub HashRef[Str, Any] process_data(ArrayRef[HashRef] $data) {
    my Int $count = 0;
    my ArrayRef[Str] $errors = [];

    # Rest of implementation
}
```

**Level 4: Leverage Flow-Sensitive Analysis**
```perl
sub HashRef[Str, Any] process_data(Maybe[ArrayRef[HashRef]] $data) {
    return { error => "No data provided" } unless defined($data);

    # Here $data is automatically refined to ArrayRef[HashRef]
    my Int $count = scalar @$data;
    # ...
}
```

### 3.3 Incremental Adoption Pattern

**Week 1-2: Foundation**
```bash
# Choose 1-2 simple modules
psc check lib/MyApp/Utils.pm
psc check lib/MyApp/Config.pm

# Add basic function signatures
# Test thoroughly
```

**Week 3-4: Expansion**
```bash
# Add 2-3 more modules
# Enhance existing modules with variable types
# Start using flow-sensitive features
```

**Week 5-8: Integration**
```bash
# Begin using PVM tools in daily development
# Add editor integration
# Extend to test files
```

## 4. PSC Integration for Legacy Analysis

### 4.1 Understanding Your Existing Code

Use PSC to analyze existing code before adding type annotations:

**Generate Type Inference Report:**
```bash
# Analyze existing code without annotations
psc check --format json lib/ > baseline-analysis.json

# Generate readable report
cat baseline-analysis.json | jq '.inferred_types[] | select(.confidence > 0.8)' > high-confidence-types.json
```

**Identify Type Annotation Opportunities:**
```bash
# Find functions that would benefit from type annotations
grep -r "sub " lib/ | while read line; do
    file=$(echo "$line" | cut -d: -f1)
    echo "Analyzing $file"
    psc check --verbose "$file" 2>&1 | grep -A 5 -B 5 "inferred"
done > type-opportunities.txt
```

### 4.2 Automated Type Discovery

Create scripts to help identify type patterns:

**Create `bin/discover-types.pl`:**
```perl
#!/usr/bin/env perl
use v5.20;
use strict;
use warnings;
use File::Find;
use JSON;

my %type_patterns;

find(sub {
    return unless /\.pm$/;
    analyze_file($File::Find::name);
}, 'lib/');

# Output discovered patterns
say encode_json(\%type_patterns);

sub analyze_file {
    my $file = shift;
    open my $fh, '<', $file or return;

    while (my $line = <$fh>) {
        # Look for variable patterns that suggest types
        if ($line =~ /my\s+\$(\w+)\s*=\s*(.+);/) {
            my ($var, $value) = ($1, $2);

            if ($value =~ /^\d+$/) {
                $type_patterns{$var}{Int}++;
            } elsif ($value =~ /^["']/) {
                $type_patterns{$var}{Str}++;
            } elsif ($value =~ /^\[/) {
                $type_patterns{$var}{ArrayRef}++;
            } elsif ($value =~ /^\{/) {
                $type_patterns{$var}{HashRef}++;
            }
        }

        # Look for function return patterns
        if ($line =~ /return\s+(.+);/) {
            my $return_value = $1;
            # Analyze return patterns...
        }
    }

    close $fh;
}
```

### 4.3 Legacy Code Type Safety

Improve existing code safety without major refactoring:

**Add Type Assertions for Critical Functions:**
```perl
# Original function
sub calculate_price {
    my ($base, $tax_rate, $discount) = @_;
    return $base * (1 + $tax_rate) * (1 - $discount);
}

# Enhanced with type safety
sub calculate_price {
    my ($base, $tax_rate, $discount) = @_;

    # Type assertions for safety
    die "Base price must be numeric" unless looks_like_number($base);
    die "Tax rate must be numeric" unless looks_like_number($tax_rate);
    die "Discount must be numeric" unless looks_like_number($discount);

    return $base * (1 + $tax_rate) * (1 - $discount);
}

# Eventually migrate to typed version
sub Num calculate_price(Num $base, Num $tax_rate, Num $discount) {
    return $base * (1 + $tax_rate) * (1 - $discount);
}
```

## 5. Migration Patterns for Common Code Structures

### 5.1 Object-Oriented Code Migration

**Traditional Perl OO → Typed Perl**

Original:
```perl
package MyApp::User;
use strict;
use warnings;

sub new {
    my ($class, %args) = @_;
    my $self = {
        id => $args{id},
        name => $args{name},
        email => $args{email},
        active => $args{active} // 1,
    };
    return bless $self, $class;
}

sub get_name { return $_[0]->{name}; }
sub set_name { $_[0]->{name} = $_[1]; }
sub is_active { return $_[0]->{active}; }
```

Migrated (Step 1 - Add type annotations):
```perl
package MyApp::User;
use v5.38;

sub MyApp::User new(Str $class, Str :$name, Str :$email, Int :$id, Bool :$active = 1) {
    my $self = {
        id => $id,
        name => $name,
        email => $email,
        active => $active,
    };
    return bless $self, $class;
}

sub Str get_name(MyApp::User $self) {
    return $self->{name};
}

sub Void set_name(MyApp::User $self, Str $name) {
    $self->{name} = $name;
}

sub Bool is_active(MyApp::User $self) {
    return $self->{active};
}

1;
```

Migrated (Step 2 - Modern class syntax):
```perl
package MyApp::User;
use v5.38;
use experimental 'class';

class MyApp::User {
    field Int $id :param;
    field Str $name :param;
    field Str $email :param;
    field Bool $active :param = 1;

    method Str get_name() { return $name; }
    method Void set_name(Str $new_name) { $name = $new_name; }
    method Bool is_active() { return $active; }
}

1;
```

### 5.2 Moose/Moo Migration

**Moose → Typed Perl Class**

Original Moose:
```perl
package MyApp::Person;
use Moose;

has 'name' => (is => 'rw', isa => 'Str', required => 1);
has 'age' => (is => 'rw', isa => 'Int', default => 0);
has 'skills' => (is => 'rw', isa => 'ArrayRef[Str]', default => sub { [] });

sub introduce {
    my $self = shift;
    return "Hi, I'm " . $self->name . " and I'm " . $self->age . " years old.";
}

__PACKAGE__->meta->make_immutable;
1;
```

Migrated to Typed Perl:
```perl
package MyApp::Person;
use v5.38;
use experimental 'class';

class MyApp::Person {
    field Str $name :param;
    field Int $age :param = 0;
    field ArrayRef[Str] $skills :param = [];

    method Str introduce() {
        return "Hi, I'm $name and I'm $age years old.";
    }

    method Void add_skill(Str $skill) {
        push @$skills, $skill;
    }

    method ArrayRef[Str] get_skills() {
        return $skills;
    }
}

1;
```

### 5.3 Functional Code Migration

**Procedural → Typed Functions**

Original:
```perl
sub process_orders {
    my @orders = @_;
    my @processed = ();

    for my $order (@orders) {
        next unless $order->{status} eq 'pending';

        my $total = 0;
        for my $item (@{$order->{items}}) {
            $total += $item->{price} * $item->{quantity};
        }

        push @processed, {
            id => $order->{id},
            total => $total,
            processed_at => time(),
        };
    }

    return @processed;
}
```

Migrated with types:
```perl
# Define types first
type OrderItem = {
    price: Num,
    quantity: Int,
    name: Str
};

type Order = {
    id: Int,
    status: Str,
    items: ArrayRef[OrderItem]
};

type ProcessedOrder = {
    id: Int,
    total: Num,
    processed_at: Int
};

sub ArrayRef[ProcessedOrder] process_orders(ArrayRef[Order] $orders) {
    my ArrayRef[ProcessedOrder] $processed = [];

    for my Order $order (@$orders) {
        next unless $order->{status} eq 'pending';

        my Num $total = 0;
        for my OrderItem $item (@{$order->{items}}) {
            $total += $item->{price} * $item->{quantity};
        }

        push @$processed, {
            id => $order->{id},
            total => $total,
            processed_at => time(),
        };
    }

    return $processed;
}
```

### 5.4 Database Code Migration

**DBI Code with Type Safety**

Original:
```perl
sub get_user_by_id {
    my ($dbh, $user_id) = @_;

    my $sth = $dbh->prepare("SELECT id, name, email FROM users WHERE id = ?");
    $sth->execute($user_id);

    return $sth->fetchrow_hashref();
}
```

Migrated:
```perl
type UserRecord = {
    id: Int,
    name: Str,
    email: Str
};

sub Maybe[UserRecord] get_user_by_id(DBI::db $dbh, Int $user_id) {
    my $sth = $dbh->prepare("SELECT id, name, email FROM users WHERE id = ?");
    $sth->execute($user_id);

    my Maybe[HashRef] $row = $sth->fetchrow_hashref();
    return undef unless defined($row);

    # Type-safe construction
    return {
        id => int($row->{id}),
        name => "$row->{name}",  # Ensure string
        email => "$row->{email}"
    };
}
```

## 6. Coexistence with Existing Tooling

### 6.1 Build System Integration

**Makefile.PL Integration:**
```perl
# Add to Makefile.PL
use ExtUtils::MakeMaker;

WriteMakefile(
    NAME => 'MyApp',
    VERSION_FROM => 'lib/MyApp.pm',
    # ... existing configuration ...

    # Add type checking targets
    postamble => {
        typecheck => 'psc check --recursive lib/',
        'test-types' => 'psc check --recursive t/',
        'build-clean' => 'psc strip lib/ build/lib/',
    }
);
```

**cpanfile Integration:**
```perl
# cpanfile - add development dependencies
requires 'perl', '5.38.0';

# ... existing dependencies ...

on 'develop' => sub {
    # PVM is a development tool, not a runtime dependency
    # Include in documentation but not as a hard requirement
};
```

### 6.2 Testing Framework Integration

**Maintain Existing Test Suite:**
```perl
# t/01-basic.t - Enhanced existing test
use strict;
use warnings;
use Test::More;

# Run existing tests
use_ok('MyApp::Utils');

# Add type-aware tests
SKIP: {
    skip "PSC not available", 3 unless system("which psc >/dev/null 2>&1") == 0;

    # Test that type checking passes
    is(system("psc check lib/MyApp/Utils.pm >/dev/null 2>&1"), 0, "Type checking passes");

    # Test specific type behaviors
    my $result = MyApp::Utils::trim("  hello  ");
    is($result, "hello", "trim function works");

    # This would fail type checking if uncommented:
    # my $bad_result = MyApp::Utils::trim(123);
}

done_testing;
```

**Create Type-Specific Tests:**
```perl
# t/types.t - New test file for type-specific behavior
use v5.38;
use Test2::V0;

SKIP: {
    skip_all "PSC not available" unless system("which psc >/dev/null 2>&1") == 0;

    # Test that our typed modules load correctly
    use_ok('MyApp::Utils');

    # Test type inference
    subtest 'type inference' => sub {
        # These should pass type checking
        my $trimmed = MyApp::Utils::trim("  test  ");
        is($trimmed, "test", "trim works with strings");

        my $currency = MyApp::Utils::format_currency(19.99);
        is($currency, '$19.99', "currency formatting works");

        pass("All type-safe operations completed");
    };
}

done_testing;
```

### 6.3 Continuous Integration Compatibility

**Enhance Existing CI without Breaking It:**

```yaml
# .github/workflows/test.yml - Enhanced CI
name: Test

on: [push, pull_request]

jobs:
  test-compatibility:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Setup Perl
      uses: shogo82148/actions-setup-perl@v1
      with:
        perl-version: '5.38'

    - name: Install Dependencies
      run: |
        cpanm --installdeps .

    - name: Run Existing Tests
      run: |
        perl Makefile.PL
        make test

    # Add type checking as separate step
    - name: Install PVM (Optional)
      run: |
        # Install PVM if available, don't fail if not
        git clone https://github.com/your-username/pvm.git || true
        if [ -d "pvm" ]; then
          cd pvm && make build && echo "$PWD/bin" >> $GITHUB_PATH
        fi
      continue-on-error: true

    - name: Type Check (Optional)
      run: |
        if command -v psc >/dev/null 2>&1; then
          echo "Running type checks..."
          psc check --recursive lib/
        else
          echo "PSC not available, skipping type checks"
        fi
      continue-on-error: true
```

### 6.4 Development Environment Coexistence

**Editor Configuration for Mixed Codebases:**

VS Code `settings.json`:
```json
{
  "files.associations": {
    "*.pm": "perl",
    "*.pl": "perl"
  },

  "perl.languageServer": {
    "enable": true,
    "command": "psc",
    "args": ["lsp", "--stdio"]
  },

  "perl.typecheck": {
    "enable": true,
    "onSave": false,  // Don't auto-check all files
    "showDiagnostics": true
  },

  // Only enable type checking for migrated files
  "perl.typecheck.includeGlob": [
    "**/lib/MyApp/Migrated/**/*.pm",
    "**/lib/MyApp/Utils.pm"
  ]
}
```

## 7. Risk Mitigation and Rollback Procedures

### 7.1 Backup and Version Control Strategy

**Create Migration Branches:**
```bash
# Create dedicated migration branch
git checkout -b feature/pvm-migration

# Create backup points for each module
git checkout -b backup/pre-migration-utils before migrating Utils.pm
git checkout feature/pvm-migration

# Make incremental commits
git add lib/MyApp/Utils.pm
git commit -m "Add type annotations to Utils.pm

- Add function parameter and return types
- Maintain backward compatibility
- All existing tests pass"
```

**Maintain Parallel Versions:**
```bash
# Keep clean versions for deployment
mkdir -p deployment/lib/MyApp/
psc strip lib/MyApp/Utils.pm > deployment/lib/MyApp/Utils.pm

# Automated backup during development
cat > bin/backup-before-types.sh << 'EOF'
#!/bin/bash
# Backup script before adding types

BACKUP_DIR="backup/$(date +%Y%m%d-%H%M%S)"
mkdir -p "$BACKUP_DIR"

# Copy current state
rsync -av lib/ "$BACKUP_DIR/lib/"
rsync -av t/ "$BACKUP_DIR/t/"

echo "Backup created in $BACKUP_DIR"
EOF

chmod +x bin/backup-before-types.sh
```

### 7.2 Testing and Validation Strategy

**Comprehensive Validation Pipeline:**
```bash
# Create validation script
cat > bin/validate-migration.sh << 'EOF'
#!/bin/bash
# Comprehensive validation for PVM migration

echo "=== Migration Validation ==="

# 1. Existing functionality
echo "## Testing existing functionality..."
if ! make test; then
    echo "ERROR: Existing tests failed!"
    exit 1
fi

# 2. Type checking
echo "## Type checking migrated modules..."
if command -v psc >/dev/null 2>&1; then
    if ! psc check --recursive lib/; then
        echo "ERROR: Type checking failed!"
        exit 1
    fi
else
    echo "WARNING: PSC not available, skipping type checks"
fi

# 3. Performance comparison
echo "## Performance validation..."
if [ -f "bench/performance.pl" ]; then
    perl bench/performance.pl
fi

# 4. Integration tests
echo "## Integration testing..."
if [ -f "t/integration.t" ]; then
    perl t/integration.t
fi

echo "✓ All validations passed"
EOF

chmod +x bin/validate-migration.sh
```

### 7.3 Rollback Procedures

**Quick Rollback Strategy:**
```bash
# Create rollback script
cat > bin/rollback-migration.sh << 'EOF'
#!/bin/bash
# Quick rollback from PVM migration

BACKUP_DIR=${1:-"backup/latest"}

if [ ! -d "$BACKUP_DIR" ]; then
    echo "ERROR: Backup directory $BACKUP_DIR not found"
    echo "Available backups:"
    ls -la backup/
    exit 1
fi

echo "Rolling back to $BACKUP_DIR..."

# Restore library files
rsync -av "$BACKUP_DIR/lib/" lib/
rsync -av "$BACKUP_DIR/t/" t/

# Remove PVM configuration
rm -rf .pvm/

# Restore original configuration files
if [ -f "$BACKUP_DIR/Makefile.PL" ]; then
    cp "$BACKUP_DIR/Makefile.PL" .
fi

if [ -f "$BACKUP_DIR/cpanfile" ]; then
    cp "$BACKUP_DIR/cpanfile" .
fi

echo "Rollback complete. Testing..."
make test

echo "✓ Rollback successful"
EOF

chmod +x bin/rollback-migration.sh
```

### 7.4 Gradual Deployment Strategy

**Staged Production Deployment:**

```bash
# Production deployment with staged rollout
cat > bin/deploy-with-types.sh << 'EOF'
#!/bin/bash
# Deploy typed code to production with staged rollout

ENVIRONMENT=${1:-"staging"}
PERCENT=${2:-"10"}

echo "Deploying to $ENVIRONMENT with $PERCENT% rollout..."

# Generate clean code for production
mkdir -p build/production/lib/
find lib/ -name "*.pm" | while read file; do
    output_file="build/production/$file"
    mkdir -p "$(dirname "$output_file")"

    if command -v psc >/dev/null 2>&1; then
        # Strip types for production
        psc strip "$file" > "$output_file"
    else
        # Fallback: copy as-is
        cp "$file" "$output_file"
    fi
done

# Deploy with canary strategy
if [ "$ENVIRONMENT" = "production" ]; then
    # Canary deployment
    echo "Deploying to $PERCENT% of production servers..."
    # Your deployment logic here
else
    # Full staging deployment
    echo "Deploying to staging..."
    # Your staging deployment logic here
fi

echo "✓ Deployment complete"
EOF

chmod +x bin/deploy-with-types.sh
```

### 7.5 Monitoring and Health Checks

**Post-Migration Monitoring:**
```perl
# monitoring/health-check.pl
use v5.20;
use strict;
use warnings;
use JSON;
use Time::HiRes qw(time);

my %health_status;

# Check if typed modules load correctly
eval {
    require MyApp::Utils;
    $health_status{modules}{MyApp::Utils} = 'ok';
} or do {
    $health_status{modules}{MyApp::Utils} = "error: $@";
};

# Performance check
my $start_time = time;
for (1..1000) {
    MyApp::Utils::trim("  test  ");
}
my $duration = time - $start_time;
$health_status{performance}{trim_1000_calls} = "${duration}s";

# Type checker availability
$health_status{tools}{psc} = system("which psc >/dev/null 2>&1") == 0 ? 'available' : 'not_available';

say encode_json(\%health_status);
```

## Migration Success Metrics

Track these metrics to ensure successful migration:

### Quantitative Metrics
- **Type Coverage**: Percentage of functions with type annotations
- **Error Reduction**: Decrease in runtime type-related errors
- **Development Velocity**: Time to implement new features
- **Bug Discovery**: Earlier detection of type-related issues

### Qualitative Metrics
- **Developer Confidence**: Team comfort with typed code
- **Code Maintainability**: Ease of understanding and modifying code
- **Integration Smoothness**: Compatibility with existing workflows

### Monitoring Dashboard
```bash
# Create metrics collection script
cat > bin/collect-metrics.sh << 'EOF'
#!/bin/bash
echo "=== PVM Migration Metrics ==="

echo "Type Coverage:"
echo "  Typed functions: $(grep -r " returns " lib/ --include="*.pm" | wc -l)"
echo "  Total functions: $(grep -r "^sub " lib/ --include="*.pm" | wc -l)"

echo "Code Quality:"
echo "  PSC errors: $(psc check lib/ 2>&1 | grep -c "error:" || echo "0")"
echo "  PSC warnings: $(psc check lib/ 2>&1 | grep -c "warning:" || echo "0")"

echo "Test Coverage:"
echo "  Tests passing: $(make test 2>&1 | grep -c "ok" || echo "unknown")"
echo "  Type tests: $(find t/ -name "*type*" | wc -l)"
EOF

chmod +x bin/collect-metrics.sh
```

## Next Steps

After successfully migrating your existing project:

1. **Expand Coverage**: See [workflow-typed-perl-new-code.md](workflow-typed-perl-new-code.md) for advanced type patterns
2. **Optimize Performance**: Learn about type-driven optimizations
3. **Team Training**: Develop team expertise with typed-Perl development
4. **Legacy Transformation**: See [workflow-psc-legacy-codebases.md](workflow-psc-legacy-codebases.md) for systematic large-scale transformations

## Related Documentation

- **[typed-perl-specification.md](typed-perl-specification.md)** - Complete type system reference
- **[workflow-new-development.md](workflow-new-development.md)** - New project patterns for inspiration
- **[workflow-psc-legacy-codebases.md](workflow-psc-legacy-codebases.md)** - Large-scale transformation strategies
- **[workflow-ci-cd-integration.md](workflow-ci-cd-integration.md)** - Production deployment with types
- **[quickstart.md](quickstart.md)** - Quick evaluation setup

This migration workflow provides a safe, incremental path to adopting PVM's typed-Perl features in existing projects while maintaining stability and minimizing risk.
