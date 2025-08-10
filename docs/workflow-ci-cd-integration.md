# CI/CD Integration Workflow

This workflow provides comprehensive guidance for integrating PVM into continuous integration and deployment pipelines, ensuring reliable automated builds, testing, and releases of typed-Perl projects.

## Executive Summary

This document covers production-ready CI/CD pipeline configurations for PVM-enabled projects, addressing cross-platform builds, automated testing strategies, isolation-aware deployments, type checking integration, and release automation. It provides practical examples for GitHub Actions, GitLab CI, Jenkins, and custom deployment scenarios.

## Prerequisites

- PVM ecosystem installed (pvm, pvx, pm, psc)
- Git repository with typed-Perl code
- Understanding of [workflow-new-development.md](workflow-new-development.md) concepts
- Basic familiarity with CI/CD platforms

## Core CI/CD Integration Concepts

### Automated Type Checking

PSC integration ensures type safety throughout the development pipeline:

```yaml
# .github/workflows/type-check.yml
name: Type Check
on: [push, pull_request]

jobs:
  type-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install PVM
        run: |
          curl -sSL https://raw.githubusercontent.com/perigrin/pvm/main/scripts/install.sh | bash
          echo "$HOME/.pvm/bin" >> $GITHUB_PATH
      - name: Install Perl Version
        run: pm 5.38.0
      - name: Type Check
        run: |
          psc check --verbose --warnings --format json lib/ > type-check-results.json
          psc check --verbose --warnings src/
```

### Isolation-Aware Testing

PVX isolation levels ensure consistent test environments:

```yaml
# Test with different isolation levels
test-isolation:
  strategy:
    matrix:
      isolation: [none, low, medium, high]
  runs-on: ubuntu-latest
  steps:
    - name: Test with isolation ${{ matrix.isolation }}
      run: |
        pvx --isolation ${{ matrix.isolation }} \
            --isolated-output \
            --save-output-dir test-results/${{ matrix.isolation }} \
            t/run_tests.pl
```

### Cross-Platform Compatibility

Build matrices ensure compatibility across environments:

```yaml
# Cross-platform testing
cross-platform:
  strategy:
    matrix:
      os: [ubuntu-latest, macos-latest, windows-latest]
      perl: [5.32.0, 5.36.0, 5.38.0]
  runs-on: ${{ matrix.os }}
  steps:
    - name: Setup Perl ${{ matrix.perl }}
      run: pm ${{ matrix.perl }}
    - name: Run tests
      run: pvx --isolation medium t/all_tests.pl
```

## GitHub Actions Integration

### Complete Workflow Example

```yaml
# .github/workflows/ci.yml
name: CI Pipeline
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

env:
  PVM_VERSION: "latest"

jobs:
  setup:
    name: Setup and Cache
    runs-on: ubuntu-latest
    outputs:
      cache-key: ${{ steps.cache-key.outputs.key }}
    steps:
      - uses: actions/checkout@v4
      - id: cache-key
        run: echo "key=pvm-${{ hashFiles('cpanfile', 'pvm.toml') }}" >> $GITHUB_OUTPUT

      - name: Cache PVM Installation
        uses: actions/cache@v4
        with:
          path: ~/.pvm
          key: ${{ steps.cache-key.outputs.key }}

      - name: Install PVM
        run: |
          if [ ! -d ~/.pvm ]; then
            curl -sSL https://install.pvm.dev | bash
            echo "$HOME/.pvm/bin" >> $GITHUB_PATH
          fi

  lint-and-format:
    needs: setup
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Restore PVM Cache
        uses: actions/cache@v4
        with:
          path: ~/.pvm
          key: ${{ needs.setup.outputs.cache-key }}

      - name: Setup Environment
        run: echo "$HOME/.pvm/bin" >> $GITHUB_PATH

      - name: Check Code Formatting
        run: |
          # Custom formatter for typed-Perl
          psc strip --check-only lib/ src/

      - name: Lint Configuration
        run: |
          psc def validate
          pvx --check-config

  type-check:
    needs: setup
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Restore PVM Cache
        uses: actions/cache@v4
        with:
          path: ~/.pvm
          key: ${{ needs.setup.outputs.cache-key }}

      - name: Setup Environment
        run: echo "$HOME/.pvm/bin" >> $GITHUB_PATH

      - name: Install Dependencies
        run: pm --install-deps

      - name: Type Check with Flow Analysis
        run: |
          psc check --verbose --warnings \
                    --format json \
                    --flow-sensitive \
                    --show-refinements \
                    lib/ src/ > type-check-results.json

      - name: Upload Type Check Results
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: type-check-results
          path: type-check-results.json

  test:
    needs: setup
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
        perl: [5.36.0, 5.38.0]
        isolation: [medium, high]
    runs-on: ${{ matrix.os }}
    steps:
      - uses: actions/checkout@v4
      - name: Restore PVM Cache
        uses: actions/cache@v4
        with:
          path: ~/.pvm
          key: ${{ needs.setup.outputs.cache-key }}

      - name: Setup Environment
        run: echo "$HOME/.pvm/bin" >> $GITHUB_PATH

      - name: Install Perl ${{ matrix.perl }}
        run: pm ${{ matrix.perl }}

      - name: Install Test Dependencies
        run: |
          pvx --isolation low --module-path ~/.pvm/test-modules \
              scripts/install_test_deps.pl

      - name: Run Unit Tests
        run: |
          pvx --isolation ${{ matrix.isolation }} \
              --isolated-output \
              --save-output-dir test-results \
              --preserve-env GITHUB_ACTIONS \
              --preserve-env CI \
              t/unit/run_all.pl

      - name: Run Integration Tests
        run: |
          pvx --isolation ${{ matrix.isolation }} \
              --ro-path /usr/share \
              --rw-path $PWD/test-data \
              t/integration/run_all.pl

      - name: Upload Test Results
        uses: actions/upload-artifact@v4
        if: always()
        with:
          name: test-results-${{ matrix.os }}-${{ matrix.perl }}-${{ matrix.isolation }}
          path: test-results/

  security-scan:
    needs: setup
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Restore PVM Cache
        uses: actions/cache@v4
        with:
          path: ~/.pvm
          key: ${{ needs.setup.outputs.cache-key }}

      - name: Setup Environment
        run: echo "$HOME/.pvm/bin" >> $GITHUB_PATH

      - name: Security Scan
        run: |
          # Scan for hardcoded secrets
          pvx --isolation high \
              --ro-path $PWD \
              scripts/security_scan.pl

          # Check for unsafe patterns in typed code
          psc check --warnings --exclude-pattern "test_*" \
                    --security-checks lib/ src/

  build:
    needs: [lint-and-format, type-check, test]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Restore PVM Cache
        uses: actions/cache@v4
        with:
          path: ~/.pvm
          key: ${{ needs.setup.outputs.cache-key }}

      - name: Setup Environment
        run: echo "$HOME/.pvm/bin" >> $GITHUB_PATH

      - name: Build Distribution
        run: |
          # Strip types for CPAN distribution
          mkdir -p dist/lib
          find lib -name "*.pm" -exec bash -c '
            psc strip "$1" "dist/$1"
          ' _ {} \;

          # Build executable scripts
          mkdir -p dist/bin
          find bin -name "*.pl" -exec bash -c '
            psc strip "$1" "dist/bin/$(basename "$1" .pl)"
            chmod +x "dist/bin/$(basename "$1" .pl)"
          ' _ {} \;

      - name: Create Archive
        run: |
          tar czf distribution.tar.gz -C dist .

      - name: Upload Build Artifacts
        uses: actions/upload-artifact@v4
        with:
          name: distribution
          path: distribution.tar.gz

  deploy-staging:
    needs: build
    if: github.ref == 'refs/heads/develop'
    runs-on: ubuntu-latest
    environment: staging
    steps:
      - uses: actions/checkout@v4
      - name: Download Build Artifacts
        uses: actions/download-artifact@v4
        with:
          name: distribution

      - name: Deploy to Staging
        run: |
          # Extract and deploy
          tar xzf distribution.tar.gz

          # Deploy with PVX isolation
          pvx --isolation high \
              --preserve-env DEPLOY_KEY \
              --preserve-env STAGING_HOST \
              scripts/deploy.pl staging

  deploy-production:
    needs: build
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/checkout@v4
      - name: Download Build Artifacts
        uses: actions/download-artifact@v4
        with:
          name: distribution

      - name: Deploy to Production
        run: |
          tar xzf distribution.tar.gz

          # Production deployment with strict isolation
          pvx --isolation high \
              --preserve-env DEPLOY_KEY \
              --preserve-env PROD_HOST \
              --ro-path /etc/ssl \
              scripts/deploy.pl production
```

### Release Automation

```yaml
# .github/workflows/release.yml
name: Release
on:
  workflow_dispatch:
    inputs:
      version_type:
        description: 'Version bump type'
        required: true
        default: 'patch'
        type: choice
        options:
          - patch
          - minor
          - major
      custom_version:
        description: 'Custom version (optional)'
        required: false

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          token: ${{ secrets.RELEASE_TOKEN }}

      - name: Setup PVM
        run: |
          curl -sSL https://install.pvm.dev | bash
          echo "$HOME/.pvm/bin" >> $GITHUB_PATH

      - name: Determine Version
        id: version
        run: |
          if [ -n "${{ github.event.inputs.custom_version }}" ]; then
            echo "version=${{ github.event.inputs.custom_version }}" >> $GITHUB_OUTPUT
          else
            current=$(git describe --tags --abbrev=0 2>/dev/null || echo "0.0.0")
            new=$(scripts/bump_version.pl $current ${{ github.event.inputs.version_type }})
            echo "version=$new" >> $GITHUB_OUTPUT
          fi

      - name: Update Version Files
        run: |
          echo "${{ steps.version.outputs.version }}" > VERSION
          scripts/update_version_refs.pl ${{ steps.version.outputs.version }}

      - name: Run Final Type Check
        run: |
          psc check --verbose lib/ src/

      - name: Build Release Assets
        run: |
          mkdir -p release

          # Create source distribution
          psc strip --recursive lib/ release/lib/
          psc strip --recursive src/ release/src/
          cp -r bin/ release/

          # Create documentation
          psc extract-docs lib/ > release/API.md

          tar czf release/typed-perl-${{ steps.version.outputs.version }}.tar.gz -C release .

      - name: Commit and Tag
        run: |
          git config user.name "Release Bot"
          git config user.email "release@pvm.dev"
          git add .
          git commit -m "Release ${{ steps.version.outputs.version }}"
          git tag -a "v${{ steps.version.outputs.version }}" -m "Release ${{ steps.version.outputs.version }}"
          git push origin main --tags

      - name: Create GitHub Release
        uses: actions/create-release@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: v${{ steps.version.outputs.version }}
          release_name: Release ${{ steps.version.outputs.version }}
          draft: false
          prerelease: false
```

## GitLab CI Integration

### GitLab CI Configuration

```yaml
# .gitlab-ci.yml
stages:
  - setup
  - validate
  - test
  - build
  - deploy

variables:
  PVM_CACHE_DIR: "${CI_PROJECT_DIR}/.pvm-cache"

cache:
  key: pvm-$CI_COMMIT_REF_SLUG
  paths:
    - .pvm-cache/

before_script:
  - |
    if [ ! -d "$PVM_CACHE_DIR" ]; then
      mkdir -p "$PVM_CACHE_DIR"
      curl -sSL https://install.pvm.dev | PVM_DIR="$PVM_CACHE_DIR" bash
    fi
    export PATH="$PVM_CACHE_DIR/bin:$PATH"

setup:
  stage: setup
  script:
    - pvm --version
    - pm --install-deps
  artifacts:
    reports:
      dotenv: build.env

type-check:
  stage: validate
  script:
    - psc check --format json --verbose lib/ src/ > type-check.json
  artifacts:
    reports:
      junit: type-check.json
    paths:
      - type-check.json

test:unit:
  stage: test
  parallel:
    matrix:
      - PERL_VERSION: ["5.36.0", "5.38.0"]
        ISOLATION: ["medium", "high"]
  script:
    - pm $PERL_VERSION
    - |
      pvx --isolation $ISOLATION \
          --isolated-output \
          --save-output-dir test-results \
          t/unit/run_all.pl
  artifacts:
    reports:
      junit: test-results/junit.xml
    paths:
      - test-results/

test:integration:
  stage: test
  script:
    - |
      pvx --isolation high \
          --ro-path /usr/share \
          --rw-path $CI_PROJECT_DIR/test-data \
          t/integration/run_all.pl
  artifacts:
    paths:
      - test-data/results/

build:distribution:
  stage: build
  script:
    - mkdir -p dist
    - find lib -name "*.pm" -exec psc strip {} dist/{} \;
    - find bin -name "*.pl" -exec psc strip {} dist/bin/{} \;
    - tar czf distribution.tar.gz -C dist .
  artifacts:
    paths:
      - distribution.tar.gz

deploy:staging:
  stage: deploy
  environment:
    name: staging
    url: https://staging.example.com
  script:
    - tar xzf distribution.tar.gz
    - |
      pvx --isolation high \
          --preserve-env DEPLOY_TOKEN \
          scripts/deploy.pl staging
  only:
    - develop

deploy:production:
  stage: deploy
  environment:
    name: production
    url: https://production.example.com
  script:
    - tar xzf distribution.tar.gz
    - |
      pvx --isolation high \
          --preserve-env DEPLOY_TOKEN \
          --ro-path /etc/ssl \
          scripts/deploy.pl production
  only:
    - main
  when: manual
```

## Jenkins Integration

### Jenkinsfile Configuration

```groovy
// Jenkinsfile
pipeline {
    agent any

    environment {
        PVM_DIR = "${WORKSPACE}/.pvm"
        PATH = "${PVM_DIR}/bin:${PATH}"
    }

    stages {
        stage('Setup') {
            steps {
                script {
                    if (!fileExists("${PVM_DIR}")) {
                        sh '''
                            curl -sSL https://install.pvm.dev | bash
                            echo "PVM installed to ${PVM_DIR}"
                        '''
                    }
                }
                sh 'pvm --version'
            }
        }

        stage('Dependencies') {
            steps {
                sh 'pm --install-deps'
            }
        }

        stage('Type Check') {
            parallel {
                stage('Static Analysis') {
                    steps {
                        sh '''
                            psc check --format json --verbose \
                                      --flow-sensitive \
                                      lib/ src/ > type-check-results.json
                        '''
                        publishTestResults testResultsPattern: 'type-check-results.json'
                    }
                }
                stage('Definition Validation') {
                    steps {
                        sh 'psc def validate'
                    }
                }
            }
        }

        stage('Test') {
            parallel {
                stage('Unit Tests') {
                    steps {
                        sh '''
                            pvx --isolation medium \
                                --isolated-output \
                                --save-output-dir test-results/unit \
                                t/unit/run_all.pl
                        '''
                        publishTestResults testResultsPattern: 'test-results/unit/junit.xml'
                    }
                }
                stage('Integration Tests') {
                    steps {
                        sh '''
                            pvx --isolation high \
                                --ro-path /usr/share \
                                --rw-path ${WORKSPACE}/test-data \
                                t/integration/run_all.pl
                        '''
                        publishTestResults testResultsPattern: 'test-data/junit.xml'
                    }
                }
                stage('Security Tests') {
                    steps {
                        sh '''
                            pvx --isolation high \
                                --ro-path ${WORKSPACE} \
                                scripts/security_tests.pl
                        '''
                    }
                }
            }
        }

        stage('Build') {
            steps {
                sh '''
                    mkdir -p dist/{lib,bin}

                    # Strip types for distribution
                    find lib -name "*.pm" -exec bash -c '
                        psc strip "$1" "dist/$1"
                    ' _ {} \\;

                    find bin -name "*.pl" -exec bash -c '
                        psc strip "$1" "dist/bin/$(basename "$1" .pl)"
                        chmod +x "dist/bin/$(basename "$1" .pl)"
                    ' _ {} \\;

                    tar czf distribution.tar.gz -C dist .
                '''
                archiveArtifacts artifacts: 'distribution.tar.gz'
            }
        }

        stage('Deploy') {
            when {
                anyOf {
                    branch 'main'
                    branch 'develop'
                }
            }
            steps {
                script {
                    def environment = env.BRANCH_NAME == 'main' ? 'production' : 'staging'
                    sh """
                        tar xzf distribution.tar.gz
                        pvx --isolation high \\
                            --preserve-env DEPLOY_TOKEN \\
                            --preserve-env ${environment.toUpperCase()}_HOST \\
                            scripts/deploy.pl ${environment}
                    """
                }
            }
        }
    }

    post {
        always {
            publishTestResults testResultsPattern: '**/junit.xml'
            publishHTML([
                allowMissing: false,
                alwaysLinkToLastBuild: true,
                keepAll: true,
                reportDir: 'test-results',
                reportFiles: 'coverage.html',
                reportName: 'Coverage Report'
            ])
        }
        failure {
            emailext (
                subject: "Build Failed: ${env.JOB_NAME} - ${env.BUILD_NUMBER}",
                body: "Build failed. Check console output at ${env.BUILD_URL}",
                to: "${env.CHANGE_AUTHOR_EMAIL}"
            )
        }
    }
}
```

## Docker Integration

### Multi-Stage Dockerfile

```dockerfile
# Dockerfile
FROM perl:5.38-slim as base

# Install system dependencies
RUN apt-get update && apt-get install -y \
    curl \
    git \
    build-essential \
    && rm -rf /var/lib/apt/lists/*

# Install PVM
RUN curl -sSL https://install.pvm.dev | bash
ENV PATH="/root/.pvm/bin:$PATH"

# Development stage
FROM base as development
WORKDIR /app
COPY cpanfile pvm.toml ./
RUN pm --install-deps

COPY . .
RUN psc check lib/ src/

# Test stage
FROM development as test
RUN pvx --isolation medium t/run_all.pl

# Build stage
FROM development as build
RUN mkdir -p dist
RUN find lib -name "*.pm" -exec psc strip {} dist/{} \;
RUN find bin -name "*.pl" -exec psc strip {} dist/bin/{} \;

# Production stage
FROM perl:5.38-slim as production
RUN apt-get update && apt-get install -y \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app
COPY --from=build /app/dist .
COPY --from=build /app/cpanfile .

# Install runtime dependencies only
RUN cpanm --installdeps .

EXPOSE 8080
CMD ["perl", "bin/app"]
```

### Docker Compose for Development

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build:
      context: .
      target: development
    volumes:
      - .:/app
      - pvm-cache:/root/.pvm
    environment:
      - PVM_ENVIRONMENT=development
    ports:
      - "8080:8080"
    command: |
      bash -c "
        psc watch lib/ &
        pvx --isolation low --preserve-env PVM_ENVIRONMENT bin/app.pl
      "

  test:
    build:
      context: .
      target: test
    volumes:
      - .:/app
      - pvm-cache:/root/.pvm
    environment:
      - PVM_ENVIRONMENT=test
    command: |
      pvx --isolation medium \
          --isolated-output \
          --save-output-dir /app/test-results \
          t/run_all.pl

volumes:
  pvm-cache:
```

## Performance Optimization

### Caching Strategies

```bash
#!/bin/bash
# scripts/ci_cache_setup.sh

# PVM binary cache
if [ ! -d "$HOME/.pvm" ]; then
    echo "Installing PVM..."
    curl -sSL https://install.pvm.dev | bash
fi

# Perl version cache
PERL_VERSIONS="5.36.0 5.38.0"
for version in $PERL_VERSIONS; do
    if ! pvm list | grep -q "$version"; then
        echo "Installing Perl $version..."
        pm "$version"
    fi
done

# Module cache
if [ ! -d "$HOME/.pvm/module-cache" ]; then
    mkdir -p "$HOME/.pvm/module-cache"
    pvx --isolation low \
        --module-path "$HOME/.pvm/module-cache" \
        scripts/install_common_deps.pl
fi
```

### Parallel Testing

```perl
#!/usr/bin/env perl
# scripts/parallel_test_runner.pl

use strict;
use warnings;
use Parallel::ForkManager;

my $pm = Parallel::ForkManager->new(4); # 4 parallel processes

my @test_suites = qw(
    unit
    integration
    security
    performance
);

for my $suite (@test_suites) {
    $pm->start and next;

    # Child process
    system(qq{
        pvx --isolation medium \\
            --isolated-output \\
            --save-output-dir test-results/$suite \\
            t/$suite/run_all.pl
    });

    $pm->finish;
}

$pm->wait_all_children;
```

## Monitoring and Alerting

### Health Check Scripts

```perl
#!/usr/bin/env perl
# scripts/health_check.pl

use strict;
use warnings;

# Type system health
my $type_check_result = `psc check --format json lib/ 2>&1`;
if ($? != 0) {
    die "Type check failed: $type_check_result";
}

# Execution environment health
my $exec_result = `pvx --isolation high --check-health 2>&1`;
if ($? != 0) {
    die "Execution environment unhealthy: $exec_result";
}

# Module dependencies health
my $deps_result = `pm --check-deps 2>&1`;
if ($? != 0) {
    die "Dependencies unhealthy: $deps_result";
}

print "All health checks passed\n";
```

### Metrics Collection

```bash
#!/bin/bash
# scripts/collect_metrics.sh

# Type checking metrics
psc check --format json --verbose lib/ src/ | \
    jq '.metrics' > metrics/type-check.json

# Test coverage metrics
pvx --isolation medium \
    --save-output-dir metrics \
    scripts/coverage_runner.pl

# Performance metrics
pvx --isolation low \
    scripts/benchmark_runner.pl > metrics/performance.json
```

## Deployment Strategies

### Blue-Green Deployment

```perl
#!/usr/bin/env perl
# scripts/blue_green_deploy.pl

use strict;
use warnings;

my ($environment, $version) = @ARGV;
die "Usage: $0 <environment> <version>" unless $environment && $version;

# Validate types before deployment
system("psc check lib/ src/") == 0
    or die "Type check failed";

# Deploy to blue environment
my $blue_result = `pvx --isolation high \\
    --preserve-env DEPLOY_TOKEN \\
    --preserve-env ${environment}_BLUE_HOST \\
    scripts/deploy_blue.pl $version`;

die "Blue deployment failed: $blue_result" if $? != 0;

# Health check blue environment
my $health_result = `pvx --isolation high \\
    scripts/health_check_remote.pl ${environment}_blue`;

die "Blue health check failed: $health_result" if $? != 0;

# Switch traffic to blue
system("scripts/switch_traffic.pl", $environment, "blue") == 0
    or die "Traffic switch failed";

print "Deployment successful\n";
```

### Canary Deployment

```perl
#!/usr/bin/env perl
# scripts/canary_deploy.pl

use strict;
use warnings;

my ($environment, $version, $traffic_percent) = @ARGV;
$traffic_percent //= 10;

# Deploy canary version
system("pvx --isolation high \\
    scripts/deploy_canary.pl $environment $version") == 0
    or die "Canary deployment failed";

# Gradually increase traffic
for my $percent (10, 25, 50, 75, 100) {
    last if $percent > $traffic_percent;

    system("scripts/adjust_traffic.pl", $environment, "canary", $percent) == 0
        or die "Traffic adjustment failed";

    # Monitor for 5 minutes
    sleep(300);

    # Check error rates
    my $error_rate = `scripts/check_error_rate.pl $environment canary`;
    chomp $error_rate;

    if ($error_rate > 1.0) {
        system("scripts/rollback_canary.pl", $environment);
        die "High error rate detected: $error_rate%";
    }
}

print "Canary deployment successful\n";
```

## Troubleshooting

### Common CI/CD Issues

#### PSC Build Failures
```bash
# Check tree-sitter dependencies
psc --debug check --version

# Rebuild tree-sitter if needed
cd tree-sitter-typed-perl
npm run build

# Test with minimal example
echo 'my Int $x = 42;' | psc check -
```

#### PVX Isolation Issues
```bash
# Debug isolation environment
pvx --isolation high --verbose --dry-run script.pl

# Check isolation directory
pvx --isolation medium --no-cleanup --verbose script.pl
ls -la /tmp/pvx-*

# Test with minimal isolation
pvx --isolation none script.pl
```

#### Dependency Resolution Problems
```bash
# Clear module cache
rm -rf ~/.pvm/module-cache

# Reinstall dependencies
pm --force-reinstall --install-deps

# Check dependency conflicts
pm --check-conflicts
```

### Performance Issues

#### Slow Type Checking
```bash
# Profile type checking
psc check --profile --verbose lib/

# Use incremental checking
psc check --incremental --cache-dir .psc-cache lib/

# Exclude large generated files
psc check --exclude-pattern "*_generated.pl" lib/
```

#### Test Timeout Issues
```bash
# Increase test timeout
pvx --timeout 300 t/long_running_test.pl

# Use parallel testing
pvx --parallel 4 t/run_all.pl

# Debug hanging tests
pvx --isolation high --verbose --debug t/problematic_test.pl
```

## Best Practices

### CI/CD Pipeline Design
1. **Early Type Checking**: Run PSC checks before expensive operations
2. **Isolation by Default**: Use PVX isolation for all test execution
3. **Incremental Builds**: Cache PVM installations and Perl versions
4. **Parallel Execution**: Run independent operations concurrently
5. **Fail Fast**: Stop pipeline on critical errors (type failures, security issues)

### Security Considerations
1. **Isolation Levels**: Use high isolation for untrusted code
2. **Environment Variables**: Carefully control preserved variables
3. **File System Access**: Restrict read/write paths in production
4. **Secrets Management**: Never expose secrets in build logs
5. **Security Scanning**: Include automated security checks

### Monitoring and Observability
1. **Type Check Metrics**: Track type errors over time
2. **Test Coverage**: Monitor coverage trends
3. **Performance Metrics**: Track build and test execution times
4. **Deployment Health**: Monitor post-deployment metrics
5. **Alert Thresholds**: Set appropriate alerts for failures

### Documentation and Maintenance
1. **Pipeline Documentation**: Document all CI/CD processes
2. **Runbook Creation**: Maintain troubleshooting guides
3. **Dependency Updates**: Regularly update PVM and dependencies
4. **Security Updates**: Monitor and apply security patches
5. **Performance Reviews**: Regular pipeline performance analysis

## Related Documentation

- [workflow-new-development.md](workflow-new-development.md) - Development environment setup
- [workflow-existing-project-migration.md](workflow-existing-project-migration.md) - Migration strategies
- [typed-perl-specification.md](typed-perl-specification.md) - Type system reference

## Advanced Topics

For complex CI/CD scenarios including multi-repository builds, custom deployment strategies, and advanced monitoring setups, see the [Development Log](development-log.md) for detailed implementation examples and lessons learned.
