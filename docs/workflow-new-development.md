# New Development Workflow with PVM

## Document Information
- **Title**: Workflow for Starting New Typed-Perl Projects
- **Purpose**: Complete guide for developers starting fresh Perl projects with PVM's typed-Perl features
- **Target Audience**: Developers creating new Perl projects who want to leverage type checking
- **Prerequisites**: Basic Perl knowledge, familiarity with modern development practices

## Table of Contents

1. [Project Setup and Initialization](#1-project-setup-and-initialization)
2. [Development Environment Configuration](#2-development-environment-configuration)
3. [Type Annotation Best Practices](#3-type-annotation-best-practices)
4. [PSC Integration Workflow](#4-psc-integration-workflow)
5. [Testing Strategies for Typed Code](#5-testing-strategies-for-typed-code)
6. [Build and Deployment Pipeline](#6-build-and-deployment-pipeline)
7. [Common Patterns and Examples](#7-common-patterns-and-examples)

---

## 1. Project Setup and Initialization

### 1.1 Installation

**Option 1: Download Binary (Recommended)**
See the [Quickstart Guide](quickstart.md#installation) for platform-specific binary installation instructions.

**Option 2: Build from Source**
For development or unsupported platforms, see [BUILD.md](../BUILD.md).

### 1.2 Create New Project Structure

```bash
# Create project directory with standard structure
mkdir my-typed-project && cd my-typed-project

# Set up standard Perl project structure
mkdir -p lib bin t docs
mkdir -p .pvm

# Initialize project with PVM
pvm local 5.38.0  # Use latest Perl with type support
```

### 1.3 Project Configuration

Create `.pvm/pvm.toml` for project-specific settings:

```toml
[pvm]
default_perl = "5.38.0"
build_jobs = 4

[psc]
# Enable all type checking features for new projects
enable_flow_sensitive = true
strict_mode = true
show_context_lines = 3
enable_colors = true

[pm]
# Automate dependency management
auto_install_deps = true
test_during_install = true
cache_modules = true

[pvx]
# Use medium isolation for development
isolation_level = "medium"
always_install_deps = true
timeout = 300
```

### 1.4 Initialize Version Control

```bash
# Initialize git and create .gitignore
git init
cat > .gitignore << 'EOF'
# PVM cache and build artifacts
.pvm/cache/
.pvm/build/

# Perl artifacts
blib/
pm_to_blib
MYMETA.*
Makefile.old
nytprof.out
cover_db/

# Editor files
.vscode/
.idea/
*.swp
*.swo
*~
EOF

git add .
git commit -m "Initial project setup with PVM configuration"
```

## 2. Development Environment Configuration

### 2.1 Editor Integration

**VS Code Setup**

Create `.vscode/settings.json`:
```json
{
  "perl.languageServer": {
    "enable": true,
    "command": "psc",
    "args": ["lsp", "--stdio"],
    "filetypes": ["perl", "pl", "pm"]
  },
  "perl.typecheck": {
    "enable": true,
    "onSave": true,
    "showDiagnostics": true
  },
  "files.associations": {
    "*.pl": "perl",
    "*.pm": "perl",
    "*.t": "perl"
  }
}
```

**Neovim Setup (with nvim-lspconfig)**

Add to your `init.lua`:
```lua
local lspconfig = require('lspconfig')
local configs = require('lspconfig.configs')

if not configs.psc then
  configs.psc = {
    default_config = {
      cmd = {'psc', 'lsp', '--stdio'},
      filetypes = {'perl'},
      root_dir = function(fname)
        return lspconfig.util.find_git_ancestor(fname) or vim.fn.getcwd()
      end,
      settings = {
        psc = {
          typecheck = true,
          flowSensitive = true,
          diagnostics = "lsp"
        }
      }
    }
  }
end

lspconfig.psc.setup({})
```

### 2.2 Development Scripts

Create `bin/dev-setup.sh`:
```bash
#!/bin/bash
# Development environment setup script

set -e

echo "Setting up development environment..."

# Ensure correct Perl version
pvm current

# Install core dependencies
echo "Installing core dependencies..."
pm install Moose Test2::V0 JSON::XS DBI

# Generate basic project files
if [ ! -f "cpanfile" ]; then
    cat > cpanfile << 'EOF'
requires 'perl', '5.38.0';
requires 'Moose';
requires 'JSON::XS';

on 'test' => sub {
    requires 'Test2::V0';
};
EOF
fi

# Validate setup
echo "Validating setup..."
psc check --recursive lib/ || echo "No lib files to check yet"

echo "Development environment ready!"
```

```bash
chmod +x bin/dev-setup.sh
./bin/dev-setup.sh
```

## 3. Type Annotation Best Practices

### 3.1 Start with Core Types

Create your first typed module - `lib/MyApp/Core.pm`:

```perl
package MyApp::Core;
use v5.38;
use experimental 'class';

# Basic type annotations for clarity
class UserAccount {
    field Str $username;
    field Str $email;
    field Int $user_id;
    field Bool $is_active = 1;

    method new(Str $username, Str $email) {
        $self->{username} = $username;
        $self->{email} = $email;
        $self->{user_id} = $self->_generate_user_id();
        return $self;
    }

    method Str get_username() {
        return $username;
    }

    method Bool is_valid_email() {
        return $email =~ /\A[^@\s]+@[^@\s]+\z/;
    }

    method Int _generate_user_id() {
        return int(rand(1000000));
    }
}

1;
```

### 3.2 Progressive Type Addition

**Level 1: Core Data Types**
```perl
my Str $name = "Alice";
my Int $age = 30;
my Bool $is_admin = 0;
```

**Level 2: Container Types**
```perl
my ArrayRef[Str] $tags = ["perl", "types", "pvm"];
my HashRef[Str, Int] %scores = (alice => 95, bob => 87);
my Maybe[Str] $middle_name = undef;
```

**Level 3: Complex Types and Union Types**
```perl
my Str|Int $flexible_id = 42;  # Can be string or number
my ArrayRef[HashRef[Str, Any]] $records = [
    { name => "Alice", age => 30 },
    { name => "Bob", age => 25 }
];
```

### 3.3 Function Signature Patterns

**Simple Functions**
```perl
sub Num calculate_tax(Num $amount, Num $rate) {
    return $amount * $rate;
}
```

**Context-Aware Functions**
```perl
sub HashRef[Str, Any]|ArrayRef[HashRef] get_user_data(Int $user_id) {
    my $data = fetch_user($user_id);

    # Return single hash in scalar context, array in list context
    return wantarray ? [$data] : $data;
}
```

**Error Handling Patterns**
```perl
sub Num|Undef safe_divide(Num $a, Num $b) {
    return undef if $b == 0;
    return $a / $b;
}
```

## 4. PSC Integration Workflow

### 4.1 Basic Type Checking

```bash
# Check a single file
psc check lib/MyApp/Core.pm

# Check entire project
psc check --recursive lib/

# Continuous checking during development
psc watch lib/
```

### 4.2 Development Workflow with PSC

**Daily Development Cycle:**

1. **Write code with type annotations**
2. **Check types continuously**
   ```bash
   psc check --verbose lib/MyApp/NewFeature.pm
   ```
3. **Run tests with type checking**
   ```bash
   psc run t/basic.t
   ```
4. **Strip types for deployment**
   ```bash
   psc strip lib/MyApp/Core.pm > clean/MyApp/Core.pm
   ```

### 4.3 Type Definition Generation

For external modules without type definitions:

```bash
# Generate type definitions for CPAN modules
psc def generate Moose --save
psc def generate DBI --save

# List available type definitions
psc def list

# Use in your code - types are automatically applied
use Moose;  # Type checker applies Moose.ptd definitions
```

### 4.4 Flow-Sensitive Analysis

Leverage PVM's intelligent type refinement:

```perl
sub Str process_input(Maybe[Str] $input) {
    # Flow-sensitive analysis understands this pattern
    if (defined($input)) {
        # $input is now typed as Str, not Maybe[Str]
        return uc($input);
    }

    return "DEFAULT";
}

sub ArrayRef[Str] handle_data(Any $data) {
    if (ref($data) eq 'ARRAY') {
        # $data is now typed as ArrayRef
        return $data;
    }

    # Convert to array if not already
    return [$data];
}
```

## 5. Testing Strategies for Typed Code

### 5.1 Test Structure

Create `t/01-basic.t`:
```perl
use v5.38;
use Test2::V0;

# Import your typed modules
use lib 'lib';
use MyApp::Core;

subtest 'UserAccount creation' => sub {
    my UserAccount $user = UserAccount->new("alice", "alice@example.com");

    isa_ok($user, 'UserAccount');
    is($user->get_username(), "alice", "Username correct");
    ok($user->is_valid_email(), "Email validation works");

    # Type checker ensures these are the correct types
    my Str $username = $user->get_username();
    my Bool $is_valid = $user->is_valid_email();

    pass("Type annotations verified");
};

done_testing;
```

### 5.2 Type-Aware Testing

```perl
# Test type constraints explicitly
subtest 'type constraints' => sub {
    # This should pass type checking
    my ArrayRef[Int] $numbers = [1, 2, 3];
    is(scalar @$numbers, 3, "Array has correct length");

    # Test that types are enforced (this would fail type checking)
    # my ArrayRef[Int] $bad = ["not", "numbers"];  # Uncomment to see error

    pass("Type constraints working");
};
```

### 5.3 Integration with Test Framework

Create `t/test-helper.pl`:
```perl
# Common test setup with type checking
use v5.38;
use Test2::V0;

# Ensure PSC is available for testing
sub check_psc_available {
    my $has_psc = system("which psc > /dev/null 2>&1") == 0;
    skip_all("PSC not available") unless $has_psc;
}

# Type check test files before running
sub typecheck_before_test {
    my $file = shift;
    my $result = system("psc check $file >/dev/null 2>&1");
    BAIL_OUT("Type errors in $file") if $result != 0;
}

1;
```

## 6. Build and Deployment Pipeline

### 6.1 Makefile for Development

Create `Makefile`:
```makefile
PERL_VERSION = 5.38.0
LIB_DIR = lib
TEST_DIR = t
BUILD_DIR = build
CLEAN_DIR = $(BUILD_DIR)/clean

.PHONY: setup test typecheck clean build deploy

setup:
	pvm local $(PERL_VERSION)
	./bin/dev-setup.sh

typecheck:
	@echo "Type checking all Perl files..."
	psc check --recursive $(LIB_DIR)/
	psc check --recursive $(TEST_DIR)/

test: typecheck
	@echo "Running tests with type checking..."
	find $(TEST_DIR) -name "*.t" -exec psc run {} \;

build: typecheck test
	@echo "Building clean Perl distribution..."
	mkdir -p $(CLEAN_DIR)
	find $(LIB_DIR) -name "*.pm" | while read file; do \
		mkdir -p $(CLEAN_DIR)/$$(dirname $$file); \
		psc strip $$file > $(CLEAN_DIR)/$$file; \
	done

clean:
	rm -rf $(BUILD_DIR)
	find . -name "*.bak" -delete
	find . -name "*~" -delete

deploy: build
	@echo "Deploying clean Perl code..."
	# Add your deployment commands here
	rsync -av $(CLEAN_DIR)/ production-server:/app/
```

### 6.2 Continuous Integration

Create `.github/workflows/test.yml`:
```yaml
name: Test Typed Perl

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
    - uses: actions/checkout@v3

    - name: Install PVM
      run: |
        # Add PVM installation steps
        make setup

    - name: Type Check
      run: make typecheck

    - name: Run Tests
      run: make test

    - name: Build Clean Distribution
      run: make build
```

## 7. Common Patterns and Examples

### 7.1 Database Interaction Pattern

Create `lib/MyApp/Database.pm`:
```perl
package MyApp::Database;
use v5.38;
use DBI;

# Database connection with type safety
class DatabaseManager {
    field Maybe[DBI] $dbh = undef;
    field Str $dsn;
    field Str $username;
    field Str $password;

    method Bool connect(Str $dsn, Str $username, Str $password) {
        $self->{dsn} = $dsn;
        $self->{username} = $username;
        $self->{password} = $password;

        $dbh = DBI->connect($dsn, $username, $password);
        return defined($dbh);
    }

    method ArrayRef[HashRef] execute_query(Str $sql, ArrayRef[Any] $params = []) {
        die "Not connected" unless defined($dbh);

        my $sth = $dbh->prepare($sql);
        $sth->execute(@$params);

        my ArrayRef[HashRef] $results = [];
        while (my $row = $sth->fetchrow_hashref()) {
            push @$results, $row;
        }

        return $results;
    }
}

1;
```

### 7.2 API Client Pattern

Create `lib/MyApp/APIClient.pm`:
```perl
package MyApp::APIClient;
use v5.38;
use JSON::XS;
use HTTP::Tiny;

class APIClient {
    field Str $base_url;
    field HTTP::Tiny $http;
    field JSON::XS $json;

    method new(Str $base_url) {
        $self->{base_url} = $base_url;
        $self->{http} = HTTP::Tiny->new();
        $self->{json} = JSON::XS->new();
        return $self;
    }

    method HashRef|ArrayRef get(Str $endpoint) {
        my Str $url = $base_url . $endpoint;
        my $response = $http->get($url);

        die "HTTP error: $response->{status}" unless $response->{success};

        return $json->decode($response->{content});
    }

    method HashRef post(Str $endpoint, HashRef $data) {
        my Str $url = $base_url . $endpoint;
        my Str $json_data = $json->encode($data);

        my $response = $http->post($url, {
            content => $json_data,
            headers => { 'Content-Type' => 'application/json' }
        });

        die "HTTP error: $response->{status}" unless $response->{success};

        return $json->decode($response->{content});
    }
}

1;
```

### 7.3 Configuration Management Pattern

Create `lib/MyApp/Config.pm`:
```perl
package MyApp::Config;
use v5.38;

class Config {
    field HashRef[Any] $config = {};
    field Bool $loaded = 0;

    method Bool load_from_file(Str $filename) {
        return 0 unless -f $filename;

        open my $fh, '<', $filename or return 0;
        my Str $content = do { local $/; <$fh> };
        close $fh;

        # Simple key=value parser
        for my $line (split /\n/, $content) {
            next if $line =~ /^\s*#/ || $line =~ /^\s*$/;

            if ($line =~ /^(\w+)\s*=\s*(.+)$/) {
                $config->{$1} = $2;
            }
        }

        $loaded = 1;
        return 1;
    }

    method Any get(Str $key, Any $default = undef) {
        return $config->{$key} // $default;
    }

    method Str get_string(Str $key, Str $default = "") {
        my $value = $self->get($key, $default);
        return "$value";  # Ensure string context
    }

    method Int get_int(Str $key, Int $default = 0) {
        my $value = $self->get($key, $default);
        return int($value);
    }

    method Bool is_loaded() {
        return $loaded;
    }
}

1;
```

### 7.4 Error Handling Patterns

Create `lib/MyApp/Result.pm`:
```perl
package MyApp::Result;
use v5.38;

# Result type for error handling
class Result {
    field Bool $success;
    field Any $value = undef;
    field Str $error = "";

    method Result success(Any $value) {
        return Result->new(success => 1, value => $value);
    }

    method Result error(Str $error) {
        return Result->new(success => 0, error => $error);
    }

    method Bool is_success() {
        return $success;
    }

    method Bool is_error() {
        return !$success;
    }

    method Any unwrap() {
        die "Result is error: $error" if $self->is_error();
        return $value;
    }

    method Any unwrap_or(Any $default) {
        return $self->is_success() ? $value : $default;
    }
}

# Usage example
sub Result divide(Num $a, Num $b) {
    return Result->error("Division by zero") if $b == 0;
    return Result->success($a / $b);
}

1;
```

## Next Steps

Once your project is set up with this workflow:

1. **Explore Advanced Features**: See [workflow-typed-perl-new-code.md](workflow-typed-perl-new-code.md) for advanced type system patterns
2. **Set Up CI/CD**: See [workflow-ci-cd-integration.md](workflow-ci-cd-integration.md) for production deployment pipelines
3. **Performance Optimization**: Learn about type-driven optimizations and best practices

## Related Documentation

- **[typed-perl-specification.md](typed-perl-specification.md)** - Complete type system reference
- **[workflow-typed-perl-new-code.md](workflow-typed-perl-new-code.md)** - Advanced coding patterns
- **[workflow-ci-cd-integration.md](workflow-ci-cd-integration.md)** - Production deployment
- **[quickstart.md](quickstart.md)** - Quick 15-minute introduction

This workflow provides a complete foundation for developing new Perl projects with PVM's type system, from initial setup through production deployment.
