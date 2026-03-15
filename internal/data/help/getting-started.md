# Getting Started with PVM

PVM (Perl Version Manager) provides modern tooling for Perl development with TypeScript-quality developer experience.

💡 For command list: pvm -h

## First Time Setup

1. Verify installation: `pvm version`
2. Install a Perl version:
   ```
   pvm available          # See available versions
   pvm install 5.38.0     # Install modern Perl
   ```
3. Set up shell integration: `pvm init               # Follow the instructions`

## Create Your First Project

1. Initialize a new workspace:
   ```
   pvm workspace init my-app
   cd my-app
   ```
2. Add dependencies:
   ```
   pvm module add DBI
   pvm module add Test::More --dev
   ```
3. Install dependencies: `pvm module install`
4. Start development: `pvm dev`

## Key Concepts

- **Project Context**: PVM automatically detects projects via .perl-version, cpanfile, or pvm.toml
- **Module Management**: Use 'pvm module' commands for dependency management
- **Build System**: 'pvm build' provides type checking and distribution creation
- **Development Mode**: 'pvm dev' watches files and provides instant feedback

## Need Help?

- Check workspace status: `pvm workspace status`
- Get contextual help: `pvm help`
- See workflows: `pvm help workflows`
- Diagnose issues: `pvm workspace doctor`
