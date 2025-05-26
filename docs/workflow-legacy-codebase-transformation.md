# Legacy Codebase Transformation Workflow

This workflow provides comprehensive guidance for systematically transforming large, established Perl codebases using PSC analysis, enabling teams to modernize legacy systems while minimizing risk and maintaining operational stability.

## Executive Summary

This document addresses the unique challenges of transforming legacy Perl codebases with thousands of lines of code, complex interdependencies, and active production use. It provides systematic strategies for incremental typing, refactoring approaches, team coordination, and risk mitigation specifically designed for large-scale legacy transformation projects.

## Prerequisites

- Large existing Perl codebase (>10,000 lines)
- Understanding of [typed-perl-specification.md](typed-perl-specification.md)
- Familiarity with [workflow-existing-project-migration.md](workflow-existing-project-migration.md)
- PSC (Perl Script Compiler) installed and configured
- Team coordination and change management processes

## Legacy Codebase Challenges

### Common Legacy Patterns

Legacy Perl codebases often exhibit these challenging patterns:

```perl
# Anti-pattern: Dynamic variable creation
for my $field (@fields) {
    no strict 'refs';
    $self->{$field} = $self->$field() if $self->can($field);
}

# Anti-pattern: Mixed blessed/unblessed hash access
sub process_data {
    my $self = shift;
    my $data = shift;  # Could be hashref, object, or string

    if (ref($data)) {
        # Sometimes treated as object
        return $data->process() if $data->can('process');
        # Sometimes as hashref
        return $data->{value} if exists $data->{value};
    }
    # Sometimes as scalar
    return "processed: $data";
}

# Anti-pattern: Context-dependent returns
sub get_results {
    my @data = fetch_from_db();
    return wantarray ? @data : scalar(@data);
}

# Anti-pattern: Deep inheritance hierarchies
package Legacy::Base::Extended::Specialized::Module;
use parent qw(Legacy::Base::Extended::Specialized);
```

### Assessment Framework

#### Codebase Complexity Analysis

```bash
#!/bin/bash
# scripts/legacy_assessment.sh

echo "=== Legacy Codebase Assessment ==="
echo "Date: $(date)"
echo

# Basic metrics
echo "## Basic Metrics"
echo "Total Perl files: $(find . -name '*.pl' -o -name '*.pm' -o -name '*.t' | wc -l)"
echo "Lines of code: $(find . -name '*.pl' -o -name '*.pm' -o -name '*.t' -exec wc -l {} + | tail -1 | awk '{print $1}')"
echo "Packages: $(grep -r '^package ' --include='*.pm' --include='*.pl' . | wc -l)"
echo "Subroutines: $(grep -r '^sub ' --include='*.pm' --include='*.pl' . | wc -l)"
echo

# Complexity indicators
echo "## Complexity Indicators"
echo "Symbolic references: $(grep -r 'no strict.*refs' --include='*.pm' --include='*.pl' . | wc -l)"
echo "Dynamic method calls: $(grep -r '->\$' --include='*.pm' --include='*.pl' . | wc -l)"
echo "Eval usage: $(grep -r '\beval\b' --include='*.pm' --include='*.pl' . | wc -l)"
echo "Typeglob usage: $(grep -r '\*[a-zA-Z]' --include='*.pm' --include='*.pl' . | wc -l)"
echo "Global variables: $(grep -r '^our ' --include='*.pm' --include='*.pl' . | wc -l)"
echo

# Dependency analysis
echo "## Dependencies"
echo "Use statements: $(grep -r '^use ' --include='*.pm' --include='*.pl' . | wc -l)"
echo "Require statements: $(grep -r '^require ' --include='*.pm' --include='*.pl' . | wc -l)"
echo "CPAN modules: $(grep -r '^use [A-Z]' --include='*.pm' --include='*.pl' . | cut -d' ' -f2 | cut -d';' -f1 | sort -u | wc -l)"
echo

# Testing coverage
echo "## Testing"
echo "Test files: $(find . -name '*.t' | wc -l)"
echo "Test lines: $(find . -name '*.t' -exec wc -l {} + | tail -1 | awk '{print $1}')"
echo "Test coverage ratio: $(echo "scale=2; $(find . -name '*.t' -exec wc -l {} + | tail -1 | awk '{print $1}') / $(find . -name '*.pl' -o -name '*.pm' -exec wc -l {} + | tail -1 | awk '{print $1}') * 100" | bc)%"
echo

# Risk indicators
echo "## Risk Indicators"
echo "Files >500 lines: $(find . -name '*.pl' -o -name '*.pm' -exec wc -l {} + | awk '$1 > 500 {count++} END {print count+0}')"
echo "Functions >50 lines: $(awk '/^sub / {start=NR} /^}/ && start {if(NR-start > 50) count++} END {print count+0}' $(find . -name '*.pl' -o -name '*.pm'))"
echo "Deeply nested files: $(grep -r '{' --include='*.pm' --include='*.pl' . | awk -F: '{print $1}' | sort | uniq -c | awk '$1 > 20 {count++} END {print count+0}')"
```

#### Type Safety Analysis

```perl
#!/usr/bin/env perl
# scripts/type_safety_analysis.pl

use strict;
use warnings;
use File::Find;
use Data::Dumper;

my %analysis = (
    untyped_variables => 0,
    dynamic_calls => 0,
    ref_checks => 0,
    blessing_operations => 0,
    type_coercions => 0,
    context_dependencies => 0
);

my @files;
find(sub { push @files, $File::Find::name if /\.(pl|pm)$/ }, '.');

for my $file (@files) {
    open my $fh, '<', $file or next;
    my $content = do { local $/; <$fh> };
    close $fh;

    # Analyze type safety patterns
    $analysis{untyped_variables} += () = $content =~ /my \$\w+(?![^;]*:)/g;
    $analysis{dynamic_calls} += () = $content =~ /->\$\w+\(/g;
    $analysis{ref_checks} += () = $content =~ /ref\(\$\w+\)/g;
    $analysis{blessing_operations} += () = $content =~ /bless\s*\(/g;
    $analysis{type_coercions} += () = $content =~ /\$\w+\s*=\s*\"\$\w+\"/g;
    $analysis{context_dependencies} += () = $content =~ /wantarray/g;
}

print "=== Type Safety Analysis ===\n";
for my $metric (sort keys %analysis) {
    printf "%-25s: %d\n", $metric, $analysis{$metric};
}

# Risk scoring
my $risk_score =
    ($analysis{dynamic_calls} * 3) +
    ($analysis{untyped_variables} * 1) +
    ($analysis{type_coercions} * 2) +
    ($analysis{context_dependencies} * 2);

printf "\nRisk Score: %d\n", $risk_score;
printf "Risk Level: %s\n",
    $risk_score > 1000 ? "VERY HIGH" :
    $risk_score > 500  ? "HIGH" :
    $risk_score > 100  ? "MEDIUM" : "LOW";
```

## Transformation Strategies

### Strategy 1: Bottom-Up Typing

Start with leaf modules and utility functions that have minimal dependencies:

```perl
# Phase 1: Transform utility modules first
package Utils::StringHelper;

# Before transformation
sub trim {
    my $str = shift;
    $str =~ s/^\s+|\s+$//g;
    return $str;
}

sub truncate {
    my ($str, $length) = @_;
    return length($str) > $length ? substr($str, 0, $length) . '...' : $str;
}

# After transformation
sub trim(Str $str) -> Str {
    $str =~ s/^\s+|\s+$//g;
    return $str;
}

sub truncate(Str $str, Int $length) -> Str {
    return length($str) > $length ? substr($str, 0, $length) . '...' : $str;
}

# Gradual introduction of Maybe types
sub safe_trim(Maybe[Str] $str) -> Maybe[Str] {
    return undef unless defined($str);
    $str =~ s/^\s+|\s+$//g;
    return $str;
}
```

### Strategy 2: Interface-First Typing

Define interfaces for major components before internal implementation:

```perl
# Phase 1: Define clear interfaces
package DataProcessor;

# Interface definition with types
method process_records(ArrayRef[HashRef[Str, Any]] $records) -> ProcessingResult {
    # Implementation comes later
}

method validate_record(HashRef[Str, Any] $record) -> ValidationResult {
    # Implementation comes later
}

method format_output(ProcessingResult $result, Str $format) -> Str {
    # Implementation comes later
}

# Phase 2: Implement with type safety
method process_records(ArrayRef[HashRef[Str, Any]] $records) -> ProcessingResult {
    my @processed;
    my @errors;

    for my $record (@$records) {
        my ValidationResult $validation = $self->validate_record($record);

        if ($validation->is_valid) {
            push @processed, $self->transform_record($record);
        } else {
            push @errors, $validation->get_errors();
        }
    }

    return ProcessingResult->new(
        processed_records => \@processed,
        errors => \@errors,
        total_count => scalar(@$records)
    );
}
```

### Strategy 3: Facade Pattern for Legacy Integration

Wrap legacy components with typed facades:

```perl
# Typed facade for legacy database layer
package Database::TypedFacade;

use parent 'Database::LegacyConnection';

# Legacy method has unclear interface
# sub get_user { ... } # Returns various types based on context

# New typed interface
method find_user_by_id(Int $user_id) -> Maybe[User] {
    # Call legacy method with proper error handling
    my $result = eval { $self->SUPER::get_user($user_id) };

    if ($@ || !defined($result)) {
        return undef;
    }

    # Ensure result conforms to User type
    if (ref($result) eq 'HASH' && exists $result->{id}) {
        return User->new($result);
    }

    return undef;
}

method find_users_by_criteria(SearchCriteria $criteria) -> ArrayRef[User] {
    my $legacy_results = eval { $self->SUPER::search_users($criteria->to_legacy_format()) };

    return [] if $@ || !defined($legacy_results);

    my @users;
    for my $result (@$legacy_results) {
        if (my $user = $self->_convert_legacy_user($result)) {
            push @users, $user;
        }
    }

    return \@users;
}

# Private conversion method
method _convert_legacy_user(Any $legacy_data) -> Maybe[User] {
    # Handle various legacy formats
    if (ref($legacy_data) eq 'HASH') {
        return User->new($legacy_data) if $legacy_data->{id};
    } elsif (blessed($legacy_data) && $legacy_data->can('to_hash')) {
        return User->new($legacy_data->to_hash());
    }

    return undef;
}
```

### Strategy 4: Gradual State Machine Conversion

Transform complex state-dependent code into typed state machines:

```perl
# Before: Legacy state handling
package OrderProcessor;

sub process_order {
    my ($self, $order) = @_;

    # Complex state logic embedded throughout
    if ($order->{status} eq 'pending') {
        if ($self->validate_payment($order)) {
            $order->{status} = 'confirmed';
            $self->send_confirmation($order);
            if ($self->can_ship($order)) {
                $order->{status} = 'shipped';
                $self->ship_order($order);
            }
        } else {
            $order->{status} = 'failed';
        }
    }
    # ... more complex logic
}

# After: Typed state machine
class OrderState {
    # Abstract base for all states
}

class PendingOrderState isa OrderState {
    method can_transition_to(Str $new_state) -> Bool {
        return $new_state eq 'confirmed' || $new_state eq 'failed';
    }

    method process(Order $order, OrderProcessor $processor) -> OrderState {
        if ($processor->validate_payment($order)) {
            $processor->send_confirmation($order);
            return ConfirmedOrderState->new();
        } else {
            return FailedOrderState->new();
        }
    }
}

class ConfirmedOrderState isa OrderState {
    method can_transition_to(Str $new_state) -> Bool {
        return $new_state eq 'shipped' || $new_state eq 'cancelled';
    }

    method process(Order $order, OrderProcessor $processor) -> OrderState {
        if ($processor->can_ship($order)) {
            $processor->ship_order($order);
            return ShippedOrderState->new();
        }
        return $self;  # Stay in current state
    }
}

class Order {
    field OrderState $state ;
    field HashRef[Any] $data ;

    method transition_to(Str $new_state, OrderProcessor $processor) -> Bool {
        if ($state->can_transition_to($new_state)) {
            $state = $state->process($self, $processor);
            return 1;
        }
        return 0;
    }
}
```

## Advanced Transformation Patterns

### Pattern 1: Legacy Data Structure Modernization

```perl
# Legacy: Mixed data structures
sub process_user_data {
    my $data = shift;  # Could be anything

    # Handle multiple formats
    if (ref($data) eq 'ARRAY') {
        return map { process_single_user($_) } @$data;
    } elsif (ref($data) eq 'HASH') {
        return process_single_user($data);
    } elsif (!ref($data)) {
        # Assume it's a user ID
        my $user_data = get_user_by_id($data);
        return process_single_user($user_data);
    }
}

# Modern: Typed data structures with clear contracts
class UserDataProcessor {
    method process_user_list(ArrayRef[UserData] $users) -> ArrayRef[ProcessedUser] {
        return [map { $self->process_single_user($_) } @$users];
    }

    method process_single_user(UserData $user) -> ProcessedUser {
        return ProcessedUser->new(
            id => $user->get_id(),
            processed_data => $self->transform_data($user->get_data()),
            timestamp => time()
        );
    }

    method process_user_by_id(UserId $id) -> Maybe[ProcessedUser] {
        my Maybe[UserData] $user_data = $self->user_repository->find_by_id($id);

        if (defined($user_data)) {
            return $self->process_single_user($user_data);
        }

        return undef;
    }
}
```

### Pattern 2: Error Handling Modernization

```perl
# Legacy: Mixed error handling
sub legacy_operation {
    my ($self, $input) = @_;

    # Various error signaling methods
    return undef if !defined($input);
    return "ERROR: Invalid input" if $input eq '';
    die "Critical error" if $input =~ /bad/;

    # Success case
    return process_input($input);
}

# Modern: Result type pattern
class OperationResult[T] {
    field Bool $success ;
    field Maybe[T] $value = undef;
    field Maybe[Str] $error_message = undef;
    field Maybe[ErrorCode] $error_code = undef;

    method success(T $val) -> OperationResult[T] {
        return $class->new(
            success => 1,
            value => $val
        );
    }

    method failure(Str $message, Maybe[ErrorCode] $code = undef) -> OperationResult[T] {
        return $class->new(
            success => 0,
            error_message => $message,
            error_code => $code
        );
    }

    method map(CodeRef $transform) -> OperationResult[Any] {
        return $self unless $success;

        my $transformed_value;
        eval {
            $transformed_value = $transform->($value);
        };

        if ($@) {
            return OperationResult->failure("Transform failed: $@");
        }

        return OperationResult->success($transformed_value);
    }
}

class ModernOperations {
    method safe_operation(Str $input) -> OperationResult[ProcessedData] {
        # Validation
        return OperationResult->failure("Input is required", ErrorCode::INVALID_INPUT)
            if !defined($input) || $input eq '';

        return OperationResult->failure("Input contains prohibited content", ErrorCode::VALIDATION_ERROR)
            if $input =~ /bad/;

        # Processing
        my $result = eval { $self->process_input($input) };
        if ($@) {
            return OperationResult->failure("Processing failed: $@", ErrorCode::PROCESSING_ERROR);
        }

        return OperationResult->success($result);
    }
}
```

### Pattern 3: Configuration Modernization

```perl
# Legacy: Global configuration with unclear types
our $CONFIG = {
    database => {
        host => 'localhost',
        port => 5432,
        timeout => 30
    },
    cache => {
        enabled => 1,
        size => '100MB',
        ttl => 3600
    }
};

sub get_config {
    my $key = shift;
    my @parts = split /\./, $key;
    my $value = $CONFIG;
    for my $part (@parts) {
        return undef unless ref($value) eq 'HASH' && exists $value->{$part};
        $value = $value->{$part};
    }
    return $value;
}

# Modern: Typed configuration classes
class DatabaseConfig {
    field Str $host ;
    field Int $port ;
    field Int $timeout ;

    method BUILD($args) {
        die "Port must be between 1 and 65535" unless $port >= 1 && $port <= 65535;
        die "Timeout must be positive" unless $timeout > 0;
    }
}

class CacheConfig {
    field Bool $enabled ;
    field Str $size ;  # e.g., "100MB"
    field Int $ttl ;

    method get_size_bytes() -> Int {
        if ($size =~ /^(\d+)(MB|KB|GB)$/) {
            my ($num, $unit) = ($1, $2);
            return $num * 1024 * 1024 if $unit eq 'MB';
            return $num * 1024 if $unit eq 'KB';
            return $num * 1024 * 1024 * 1024 if $unit eq 'GB';
        }
        die "Invalid size format: $size";
    }
}

class ApplicationConfig {
    field DatabaseConfig $database ;
    field CacheConfig $cache ;

    method from_hash(HashRef[Any] $config_data) -> ApplicationConfig {
        return $class->new(
            database => DatabaseConfig->new($config_data->{database}),
            cache => CacheConfig->new($config_data->{cache})
        );
    }

    method validate() -> ArrayRef[Str] {
        my @errors;

        eval { $database->BUILD({}) };
        push @errors, "Database config: $@" if $@;

        eval { $cache->get_size_bytes() };
        push @errors, "Cache config: $@" if $@;

        return \@errors;
    }
}
```

## Risk Management and Testing

### Automated Regression Testing

```perl
#!/usr/bin/env perl
# scripts/regression_test_generator.pl

use strict;
use warnings;
use File::Find;
use Test::More;
use Test::Deep;
use Capture::Tiny qw(capture);

# Generate regression tests for legacy behavior
sub generate_regression_tests {
    my ($module_path) = @_;

    # Extract all public subroutines
    my @subroutines = extract_subroutines($module_path);

    for my $sub (@subroutines) {
        generate_behavior_test($module_path, $sub);
    }
}

sub generate_behavior_test {
    my ($module, $subroutine) = @_;

    # Create test data matrix
    my @test_cases = generate_test_cases($subroutine);

    my $test_file = "t/regression_" . lc($module) . "_" . lc($subroutine) . ".t";

    open my $fh, '>', $test_file or die "Cannot create $test_file: $!";

    print $fh <<"TEST_HEADER";
use strict;
use warnings;
use Test::More;
use Test::Deep;
use $module;

# Regression tests for $module\::$subroutine
# Generated on: @{[scalar localtime]}

TEST_HEADER

    for my $case (@test_cases) {
        print $fh generate_test_case($module, $subroutine, $case);
    }

    print $fh "\ndone_testing();\n";
    close $fh;
}

sub generate_test_case {
    my ($module, $sub, $case) = @_;

    return <<"TEST_CASE";

# Test case: $case->{description}
subtest '$case->{description}' => sub {
    my \$result = eval { $module\::$sub($case->{input}) };

    if ($case->{should_die}) {
        ok(\$@, "Should die: $case->{expected_error}");
        like(\$@, qr/$case->{error_pattern}/, "Error message matches pattern");
    } else {
        ok(!\$@, "Should not die") or diag(\$@);
        cmp_deeply(\$result, $case->{expected_output}, "Output matches expected");
    }
};
TEST_CASE
}
```

### Type Safety Validation

```bash
#!/bin/bash
# scripts/type_safety_validation.sh

echo "=== Type Safety Validation ==="

# Run PSC on all modified files
echo "Running PSC type checker..."
find . -name '*.pl' -o -name '*.pm' | while read file; do
    if psc check "$file" 2>/dev/null; then
        echo "✓ $file - Type safe"
    else
        echo "✗ $file - Type errors found"
        psc check "$file" 2>&1 | head -5
        echo
    fi
done

# Check for dangerous patterns
echo "Checking for dangerous patterns..."

dangerous_patterns=(
    "no strict.*refs"     "Symbolic references found"
    "eval.*\$"            "String eval found"
    "->\$\w+\("           "Dynamic method calls found"
    "\*\w+\s*=.*\\&"     "Typeglob assignment found"
)

for ((i=0; i<${#dangerous_patterns[@]}; i+=2)); do
    pattern="${dangerous_patterns[i]}"
    message="${dangerous_patterns[i+1]}"

    matches=$(grep -r "$pattern" --include='*.pl' --include='*.pm' . | wc -l)
    if [ "$matches" -gt 0 ]; then
        echo "⚠ $message ($matches occurrences)"
    fi
done

echo "Validation complete."
```

### Incremental Rollout Strategy

```perl
# Feature flag system for gradual rollout
class FeatureFlags {
    field HashRef[Bool] $_flags = {};
    field Str $environment ;

    method is_enabled(Str $feature_name) -> Bool {
        # Check environment-specific overrides
        my $env_key = "${feature_name}_${environment}";
        return $_flags->{$env_key} if exists $_flags->{$env_key};

        # Check global flag
        return $_flags->{$feature_name} // 0;
    }

    method enable_for_percentage(Str $feature_name, Int $percentage, Str $user_id) -> Bool {
        # Consistent percentage rollout based on user ID
        my $hash = substr(sha256_hex($user_id . $feature_name), 0, 8);
        my $user_percentage = hex($hash) % 100;

        return $user_percentage < $percentage;
    }
}

# Usage in legacy system
class UserService {
    field FeatureFlags $feature_flags ;

    method process_user(User $user) -> ProcessedUser {
        if ($feature_flags->is_enabled('typed_user_processing')) {
            return $self->typed_user_processing($user);
        } else {
            return $self->legacy_user_processing($user->to_legacy_hash());
        }
    }

    method typed_user_processing(User $user) -> ProcessedUser {
        # New typed implementation
        return ProcessedUser->new(
            id => $user->get_id(),
            data => $self->transform_user_data($user->get_data()),
            processed_at => DateTime->now()
        );
    }

    method legacy_user_processing(HashRef[Any] $user_data) -> ProcessedUser {
        # Wrapped legacy implementation
        my $legacy_result = $self->old_process_user($user_data);

        # Convert back to typed result
        return ProcessedUser->from_legacy_hash($legacy_result);
    }
}
```

## Team Coordination Strategies

### Code Review Guidelines

```markdown
# Typed Perl Code Review Checklist

## Type Safety
- [ ] All new functions have type signatures
- [ ] Type annotations are specific (avoid `Any` unless necessary)
- [ ] Maybe types used for optional values
- [ ] Union types used appropriately for multiple possible types
- [ ] No unsafe type coercions

## Legacy Integration
- [ ] Legacy interfaces preserved for existing callers
- [ ] Typed facades created for legacy components
- [ ] Error handling improved from legacy patterns
- [ ] Performance impact assessed

## Testing
- [ ] Regression tests generated for modified legacy behavior
- [ ] Type-specific test cases added
- [ ] Edge cases covered (null, empty, malformed data)
- [ ] Integration tests updated

## Documentation
- [ ] Type definitions documented
- [ ] Migration notes added for breaking changes
- [ ] Examples provided for new typed interfaces
- [ ] Performance characteristics documented
```

### Migration Tracking

```perl
#!/usr/bin/env perl
# scripts/migration_tracker.pl

use strict;
use warnings;
use DBI;
use DateTime;

# Track migration progress
class MigrationTracker {
    field DBI $dbh ;

    method record_file_migration(Str $file_path, Str $migration_type, Int $lines_converted) -> Void {
        $dbh->do(qq{
            INSERT INTO migration_log (file_path, migration_type, lines_converted, migrated_at)
            VALUES (?, ?, ?, ?)
        }, undef, $file_path, $migration_type, $lines_converted, DateTime->now());
    }

    method get_migration_stats() -> HashRef[Any] {
        my $stats = $dbh->selectall_hashref(qq{
            SELECT
                migration_type,
                COUNT(*) as files_count,
                SUM(lines_converted) as total_lines,
                AVG(lines_converted) as avg_lines_per_file
            FROM migration_log
            GROUP BY migration_type
        }, 'migration_type');

        return $stats;
    }

    method get_recent_migrations(Int $days = 7) -> ArrayRef[HashRef[Any]] {
        my $cutoff = DateTime->now()->subtract(days => $days);

        return $dbh->selectall_arrayref(qq{
            SELECT file_path, migration_type, lines_converted, migrated_at
            FROM migration_log
            WHERE migrated_at >= ?
            ORDER BY migrated_at DESC
        }, {Slice => {}}, $cutoff);
    }
}
```

### Performance Impact Assessment

```perl
#!/usr/bin/env perl
# scripts/performance_impact.pl

use strict;
use warnings;
use Benchmark qw(cmpthese);
use Memory::Usage;

# Compare legacy vs typed implementations
class PerformanceComparator {
    field Any $legacy_implementation ;
    field Any $typed_implementation ;
    field ArrayRef[Any] $test_data ;

    method run_comparison() -> HashRef[Any] {
        my Memory::Usage $mu = Memory::Usage->new();

        # Memory usage comparison
        $mu->record('baseline');

        # Test legacy implementation
        my $legacy_start = time();
        for my $data (@$test_data) {
            $legacy_implementation->process($data);
        }
        my $legacy_time = time() - $legacy_start;
        $mu->record('after_legacy');

        # Test typed implementation
        my $typed_start = time();
        for my $data (@$test_data) {
            $typed_implementation->process($data);
        }
        my $typed_time = time() - $typed_start;
        $mu->record('after_typed');

        # Benchmark detailed comparison
        my $results = cmpthese(-5, {
            'legacy' => sub { $legacy_implementation->process($test_data->[0]) },
            'typed'  => sub { $typed_implementation->process($test_data->[0]) }
        });

        return {
            execution_time => {
                legacy => $legacy_time,
                typed => $typed_time,
                ratio => $typed_time / $legacy_time
            },
            memory_usage => $mu->dump(),
            benchmark_results => $results
        };
    }
}
```

## Success Metrics and Monitoring

### Transformation Quality Metrics

```perl
#!/usr/bin/env perl
# scripts/quality_metrics.pl

use strict;
use warnings;

# Quality metrics for transformation progress
class TransformationMetrics {
    method calculate_type_coverage() -> Num {
        my $total_functions = $self->count_total_functions();
        my $typed_functions = $self->count_typed_functions();

        return $total_functions > 0 ? ($typed_functions / $total_functions) * 100 : 0;
    }

    method calculate_error_reduction() -> HashRef[Any] {
        my $before_errors = $self->get_legacy_error_count();
        my $after_errors = $self->get_current_error_count();

        return {
            before => $before_errors,
            after => $after_errors,
            reduction_percentage => $before_errors > 0 ?
                (($before_errors - $after_errors) / $before_errors) * 100 : 0
        };
    }

    method calculate_maintainability_score() -> Int {
        my $score = 0;

        # Type coverage (40% of score)
        $score += $self->calculate_type_coverage() * 0.4;

        # Test coverage (30% of score)
        $score += $self->calculate_test_coverage() * 0.3;

        # Code complexity reduction (20% of score)
        $score += $self->calculate_complexity_reduction() * 0.2;

        # Documentation quality (10% of score)
        $score += $self->calculate_documentation_score() * 0.1;

        return int($score);
    }
}
```

### Automated Health Monitoring

```bash
#!/bin/bash
# scripts/transformation_health_check.sh

echo "=== Transformation Health Check ==="
echo "Date: $(date)"

# Type safety checks
echo "## Type Safety Status"
type_errors=$(find . -name '*.pl' -o -name '*.pm' | xargs psc check 2>&1 | grep -c "error:" || echo "0")
echo "Type errors found: $type_errors"

if [ "$type_errors" -eq 0 ]; then
    echo "✓ All files pass type checking"
else
    echo "✗ Type errors need resolution"
fi

# Test coverage
echo "## Test Coverage"
coverage=$(prove -l t/ 2>&1 | grep -o '[0-9]*\.[0-9]*%' | tail -1 || echo "0%")
echo "Test coverage: $coverage"

# Performance regression check
echo "## Performance Status"
if [ -f "performance_baseline.json" ]; then
    echo "Running performance regression tests..."
    perl scripts/performance_regression_check.pl
else
    echo "No performance baseline found"
fi

# Migration progress
echo "## Migration Progress"
typed_files=$(grep -r "method\|field" --include='*.pm' --include='*.pl' . | cut -d: -f1 | sort -u | wc -l)
total_files=$(find . -name '*.pl' -o -name '*.pm' | wc -l)
progress=$(echo "scale=1; $typed_files * 100 / $total_files" | bc)
echo "Files with type annotations: $typed_files/$total_files ($progress%)"

echo "Health check complete."
```

## Related Documentation

- [workflow-existing-project-migration.md](workflow-existing-project-migration.md) - Basic migration strategies
- [workflow-typed-perl-coding-patterns.md](workflow-typed-perl-coding-patterns.md) - Modern coding patterns
- [workflow-ci-cd-integration.md](workflow-ci-cd-integration.md) - Automated testing and deployment
- [typed-perl-specification.md](typed-perl-specification.md) - Type system reference

## Advanced Topics

For complex legacy transformation scenarios including mainframe integration, distributed system coordination, and enterprise-scale rollouts, see the [Development Log](development-log.md) for detailed case studies and lessons learned from real-world legacy transformations.
