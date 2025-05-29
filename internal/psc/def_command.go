// ABOUTME: PSC type definition command implementation
// ABOUTME: Manages type definitions for static type checking

package psc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"tamarou.com/pvm/internal/cli"
	"tamarou.com/pvm/internal/errors"
	"tamarou.com/pvm/internal/parser"
	"tamarou.com/pvm/internal/typedef"
)

// newDefCommand creates a command to manage type definitions
func newDefCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "def",
		Short: "Manage type definitions",
		Long:  "Generate and manage type definitions for Perl modules",
	}

	// Add subcommands
	cmd.AddCommand(
		newDefListCommand(),
		newDefGenerateCommand(),
		newDefImportCommand(),
		newDefExportCommand(),
		newDefInstallCommand(),
	)

	return cmd
}

// newDefListCommand creates a command to list available type definitions
func newDefListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List available type definitions",
		Long:  "List all available type definitions for Perl modules",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Create a new storage instance
			storage, err := typedef.NewStorage()
			if err != nil {
				return err
			}

			// Get the list of type definitions
			modules, err := storage.List()
			if err != nil {
				return err
			}

			// Print the list of modules
			if len(modules) == 0 {
				fmt.Println("No type definitions available.")
				fmt.Println("Use 'psc def generate' to generate type definitions for modules.")
				return nil
			}

			fmt.Println("Available type definitions:")
			for _, module := range modules {
				fmt.Printf("  %s\n", module)
			}

			fmt.Println("\nUse 'psc def generate [module]' to generate new type definitions.")
			fmt.Println("Use 'psc check [file]' to type-check Perl files.")

			return nil
		},
	}
}

// newDefGenerateCommand creates a command to generate type definitions
func newDefGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate [module]",
		Short: "Generate type definitions",
		Long:  "Generate type definitions for a Perl module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]
			version, _ := cmd.Flags().GetString("version")
			outputFile, _ := cmd.Flags().GetString("output")
			saveTypeDef, _ := cmd.Flags().GetBool("save")
			verbose, _ := cmd.Flags().GetBool("verbose")

			// Create a minimal type definition
			typeDef := &typedef.TypeDefinition{
				Module:     moduleName,
				Version:    version,
				Generated:  time.Now(),
				Maintainer: "PSC type generator",
				Source:     "generated",
				Types:      []typedef.TypeInfo{},
				Packages:   []typedef.PackageInfo{},
				Subs:       []typedef.SubInfo{},
				Methods:    []typedef.MethodInfo{},
			}

			// Extract module type information using enhanced introspection
			if verbose {
				fmt.Printf("Analyzing module %s using enhanced introspection...\n", moduleName)
			}

			moduleTypes, err := analyzeModuleTypes(moduleName)
			if err != nil {
				if verbose {
					fmt.Printf("Warning: Could not fully analyze module: %v\n", err)
					fmt.Println("Generating a basic type definition instead.")
				}
			} else if moduleTypes != nil {
				// Successfully analyzed - use the enhanced type definition
				typeDef = moduleTypes

				if verbose {
					methodCount := len(typeDef.Methods)
					typeCount := len(typeDef.Types)
					pkgCount := len(typeDef.Packages)
					fmt.Printf("Successfully analyzed module %s:\n", moduleName)
					fmt.Printf("  - Found %d packages\n", pkgCount)
					fmt.Printf("  - Found %d types\n", typeCount)
					fmt.Printf("  - Found %d methods\n", methodCount)
				}
			}

			// Add some placeholder types if none provided
			if len(typeDef.Types) == 0 {
				typeDef.Types = append(typeDef.Types, typedef.TypeInfo{
					Name:        moduleName,
					Description: "Auto-generated type information for " + moduleName,
					Kind:        "class",
					Methods:     []typedef.MethodInfo{},
					Properties:  []typedef.PropInfo{},
				})
			}

			// Marshal to JSON with indentation
			data, err := json.MarshalIndent(typeDef, "", "  ")
			if err != nil {
				return errors.NewTypeError(
					"801",
					fmt.Sprintf("Failed to marshal type definition for %s", moduleName),
					err,
				)
			}

			// If an output file is specified, write to it
			if outputFile != "" {
				// Create the directory if it doesn't exist
				dir := filepath.Dir(outputFile)
				if dir != "." {
					if err := os.MkdirAll(dir, 0755); err != nil {
						return errors.NewSystemError("001",
							"Failed to create directory", err).
							WithLocation(dir)
					}
				}

				// Write the file
				if err := os.WriteFile(outputFile, data, 0644); err != nil {
					return errors.NewSystemError("002",
						"Failed to write file", err).
						WithLocation(outputFile)
				}

				fmt.Printf("Generated type definition for %s to %s\n", moduleName, outputFile)
			} else {
				// Print to stdout
				fmt.Println(string(data))
			}

			// If save flag is set, save the type definition
			if saveTypeDef {
				storage, err := typedef.NewStorage()
				if err != nil {
					return err
				}

				if err := storage.Save(typeDef); err != nil {
					return err
				}

				fmt.Printf("Saved type definition for %s to type registry\n", moduleName)
			}

			return nil
		},
	}

	// Add flags
	cmd.Flags().StringP("version", "V", "0.0.1", "Module version")
	cmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	cmd.Flags().BoolP("save", "s", false, "Save the type definition to the registry")

	return cmd
}

// newDefImportCommand creates a command to import type definitions from a file
func newDefImportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "import [file]",
		Short: "Import type definitions",
		Long:  "Import type definitions from a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]

			// Check if the file exists
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				return errors.NewUserInputError(cli.PrefixPSC, "001",
					"File not found", err).
					WithLocation(filePath)
			}

			// Read the file
			data, err := os.ReadFile(filePath)
			if err != nil {
				return errors.NewSystemError("003",
					"Failed to read file", err).
					WithLocation(filePath)
			}

			// Unmarshal the type definition
			var typeDef typedef.TypeDefinition
			if err := json.Unmarshal(data, &typeDef); err != nil {
				return errors.NewTypeError(
					"802",
					"Failed to parse type definition",
					err,
				).WithLocation(filePath)
			}

			// Create a new storage instance
			storage, err := typedef.NewStorage()
			if err != nil {
				return err
			}

			// Save the type definition
			if err := storage.Save(&typeDef); err != nil {
				return err
			}

			fmt.Printf("Imported type definition for %s\n", typeDef.Module)
			return nil
		},
	}
}

// newDefExportCommand creates a command to export type definitions to a file
func newDefExportCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "export [module] [file]",
		Short: "Export type definitions",
		Long:  "Export type definitions to a file",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]
			outputFile := args[1]

			// Create a new storage instance
			storage, err := typedef.NewStorage()
			if err != nil {
				return err
			}

			// Load the type definition
			typeDef, err := storage.Load(moduleName)
			if err != nil {
				return err
			}

			// Marshal to JSON with indentation
			data, err := json.MarshalIndent(typeDef, "", "  ")
			if err != nil {
				return errors.NewTypeError(
					"803",
					fmt.Sprintf("Failed to marshal type definition for %s", moduleName),
					err,
				)
			}

			// Create the directory if it doesn't exist
			dir := filepath.Dir(outputFile)
			if dir != "." {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return errors.NewSystemError("004",
						"Failed to create directory", err).
						WithLocation(dir)
				}
			}

			// Write the file
			if err := os.WriteFile(outputFile, data, 0644); err != nil {
				return errors.NewSystemError("005",
					"Failed to write file", err).
					WithLocation(outputFile)
			}

			fmt.Printf("Exported type definition for %s to %s\n", moduleName, outputFile)
			return nil
		},
	}
}

// newDefInstallCommand creates a command to install type definitions for a CPAN module
func newDefInstallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [module]",
		Short: "Install type definitions",
		Long:  "Install type definitions for a CPAN module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			moduleName := args[0]
			forceGenerate, _ := cmd.Flags().GetBool("force")
			verbose, _ := cmd.Flags().GetBool("verbose")

			// Create a new storage instance
			storage, err := typedef.NewStorage()
			if err != nil {
				return err
			}

			// Check if the type definition already exists
			_, err = storage.Load(moduleName)
			if err == nil && !forceGenerate {
				fmt.Printf("Type definition for %s already exists. Use --force to regenerate.\n", moduleName)
				return nil
			}

			// Generate and save the type definition
			typeDef := &typedef.TypeDefinition{
				Module:     moduleName,
				Version:    "0.0.1",
				Generated:  time.Now(),
				Maintainer: "PSC type generator",
				Source:     "installed",
				Types:      []typedef.TypeInfo{},
				Packages:   []typedef.PackageInfo{},
				Subs:       []typedef.SubInfo{},
				Methods:    []typedef.MethodInfo{},
			}

			// Extract module type information using enhanced introspection
			if verbose {
				fmt.Printf("Analyzing module %s using enhanced introspection...\n", moduleName)
			}

			moduleTypes, err := analyzeModuleTypes(moduleName)
			if err != nil {
				if verbose {
					fmt.Printf("Warning: Could not fully analyze module: %v\n", err)
					fmt.Println("Generating a basic type definition instead.")
				}
			} else if moduleTypes != nil {
				// Successfully analyzed - use the enhanced type definition
				typeDef = moduleTypes

				if verbose {
					methodCount := len(typeDef.Methods)
					typeCount := len(typeDef.Types)
					pkgCount := len(typeDef.Packages)
					fmt.Printf("Successfully analyzed module %s:\n", moduleName)
					fmt.Printf("  - Found %d packages\n", pkgCount)
					fmt.Printf("  - Found %d types\n", typeCount)
					fmt.Printf("  - Found %d methods\n", methodCount)
				}
			}

			// Add a placeholder type if none were found
			if len(typeDef.Types) == 0 {
				typeDef.Types = append(typeDef.Types, typedef.TypeInfo{
					Name:        moduleName,
					Description: "Auto-generated type information for " + moduleName,
					Kind:        "class",
					Methods:     []typedef.MethodInfo{},
					Properties:  []typedef.PropInfo{},
				})
			}

			// Save the type definition
			if err := storage.Save(typeDef); err != nil {
				return err
			}

			fmt.Printf("Installed type definition for %s\n", moduleName)

			// Add instructions for using the type definition
			fmt.Printf("\nYou can now use the type definition in your Perl code:\n\n")
			fmt.Printf("use %s : typed;\n\n", moduleName)
			fmt.Printf("Or check your code with:\n\n")
			fmt.Printf("psc check your_script.pl\n")

			return nil
		},
	}

	// Add flags
	cmd.Flags().BoolP("force", "f", false, "Force generation of type definition even if it already exists")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	return cmd
}

// analyzeModuleTypes performs enhanced type analysis of a Perl module using introspection
func analyzeModuleTypes(moduleName string) (*typedef.TypeDefinition, error) {
	// Create the enhanced introspector
	introspector, err := parser.NewEnhancedIntrospector()
	if err != nil {
		// Fall back to the Perl introspection method
		return analyzeModuleTypesWithPerl(moduleName)
	}

	// Perform comprehensive analysis
	result, err := introspector.AnalyzeModule(moduleName)
	if err != nil {
		// Fall back to Perl introspection if enhanced introspection fails
		return analyzeModuleTypesWithPerl(moduleName)
	}

	// Check if we have a valid type definition
	if result.TypeDefinition == nil {
		// Fall back to Perl introspection
		return analyzeModuleTypesWithPerl(moduleName)
	}

	// Log warnings if any
	if len(result.Warnings) > 0 {
		for _, warning := range result.Warnings {
			fmt.Fprintf(os.Stderr, "Warning: %s\n", warning)
		}
	}

	// Log confidence scores if verbose
	if result.Confidence.Overall < 0.7 {
		fmt.Fprintf(os.Stderr, "Note: Type inference confidence is %.0f%% (Methods: %.0f%%, Attributes: %.0f%%)\n",
			result.Confidence.Overall*100,
			result.Confidence.Methods*100,
			result.Confidence.Attributes*100)
	}

	return result.TypeDefinition, nil
}

// findModuleFile attempts to find the file path for a Perl module
func findModuleFile(moduleName string) (string, error) {
	// Convert module name to file path
	moduleFile := strings.ReplaceAll(moduleName, "::", "/") + ".pm"

	// Search in @INC paths using perl
	cmd := exec.Command("perl", "-e", fmt.Sprintf(`
		foreach my $inc (@INC) {
			my $file = "$inc/%s";
			if (-f $file) {
				print $file;
				exit 0;
			}
		}
		exit 1;
	`, moduleFile))

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("module not found in @INC: %s", moduleName)
	}

	return strings.TrimSpace(string(output)), nil
}

// convertIntrospectionToTypeDef converts introspection results to TypeDefinition
func convertIntrospectionToTypeDef(moduleName string, result *parser.ModuleIntrospectionResult) *typedef.TypeDefinition {
	typeDef := &typedef.TypeDefinition{
		Module:     moduleName,
		Version:    "0.0.1",
		Generated:  time.Now(),
		Maintainer: "PSC enhanced type generator",
		Source:     "introspection",
		Types:      []typedef.TypeInfo{},
		Packages:   []typedef.PackageInfo{},
		Subs:       []typedef.SubInfo{},
		Methods:    []typedef.MethodInfo{},
	}

	// Convert packages to TypeDefinition format
	for pkgName, pkgInfo := range result.Packages {
		// Add package info
		typeDef.Packages = append(typeDef.Packages, typedef.PackageInfo{
			Name:        pkgName,
			Description: fmt.Sprintf("Package %s", pkgName),
			Exports:     []typedef.ExportInfo{},
		})

		// Create type info for the package
		typeInfo := typedef.TypeInfo{
			Name:        pkgName,
			Description: fmt.Sprintf("Type information for %s", pkgName),
			Kind:        "class",
			Methods:     []typedef.MethodInfo{},
			Properties:  []typedef.PropInfo{},
		}

		// Convert methods
		for methodName, methodSig := range pkgInfo.Methods {
			// Convert parameters
			params := []typedef.ParamInfo{}
			for _, param := range methodSig.Parameters {
				params = append(params, typedef.ParamInfo{
					Name:        param.Name,
					Type:        param.Type,
					Description: param.Documentation,
					Optional:    param.IsOptional,
					Default:     param.DefaultValue,
				})
			}

			// Convert return type to ReturnInfo
			returns := []typedef.ReturnInfo{}
			if methodSig.ReturnType != "" && methodSig.ReturnType != "Any" {
				returns = append(returns, typedef.ReturnInfo{
					Type:        methodSig.ReturnType,
					Description: "Return value",
				})
			}

			methodInfo := typedef.MethodInfo{
				Name:        methodName,
				Description: methodSig.Documentation,
				Parameters:  params,
				Returns:     returns,
			}

			typeInfo.Methods = append(typeInfo.Methods, methodInfo)

			// Also add to global methods list
			typeDef.Methods = append(typeDef.Methods, typedef.MethodInfo{
				Name:        methodName,
				Description: methodSig.Documentation,
				Parameters:  params,
				Returns:     returns,
			})
		}

		// Convert attributes to properties
		for attrName, attrInfo := range pkgInfo.Attributes {
			typeInfo.Properties = append(typeInfo.Properties, typedef.PropInfo{
				Name:        attrName,
				Type:        attrInfo.Type,
				Description: attrInfo.Documentation,
			})
		}

		typeDef.Types = append(typeDef.Types, typeInfo)
	}

	// Add detected frameworks as metadata
	if len(result.DetectedFrameworks) > 0 {
		// This information could be added to a metadata field if TypeDefinition supported it
		// For now, we'll add it to the description of the main type
		if len(typeDef.Types) > 0 {
			typeDef.Types[0].Description += fmt.Sprintf(" (Frameworks: %s)",
				strings.Join(result.DetectedFrameworks, ", "))
		}
	}

	return typeDef
}

// ensureRequiredPerlModules checks for required Perl modules and provides helpful error messages
func ensureRequiredPerlModules() error {
	requiredModules := []string{"JSON", "Module::Load", "Class::Inspector"}
	missingModules := []string{}
	
	for _, module := range requiredModules {
		// Check if module is available, including local::lib paths
		checkScript := fmt.Sprintf(`perl -I ~/perl5/lib/perl5/ -M%s -e 'print "OK"' 2>/dev/null || perl -M%s -e 'print "OK"' 2>/dev/null`, module, module)
		cmd := exec.Command("sh", "-c", checkScript)
		if err := cmd.Run(); err != nil {
			missingModules = append(missingModules, module)
		}
	}
	
	if len(missingModules) > 0 {
		suggestion := fmt.Sprintf(
			"Missing required Perl modules for PSC type analysis: %s\n"+
			"Please install them using one of these methods:\n"+
			"1. Using PVI (recommended): pvi install %s\n"+
			"2. Using cpanminus: cpanm -l ~/perl5 %s && eval $(perl -I ~/perl5/lib/perl5/ -Mlocal::lib)\n"+
			"3. Using system package manager: sudo apt-get install lib%s-perl (on Ubuntu/Debian)\n"+
			"\nNote: PSC's advanced type analysis features require these modules for Perl introspection.\n"+
			"Basic type checking will still work without them.",
			strings.Join(missingModules, ", "),
			strings.Join(missingModules, " "),
			strings.Join(missingModules, " "),
			strings.ToLower(strings.ReplaceAll(missingModules[0], "::", "-")))
		
		return errors.NewSystemError("010",
			"Required Perl modules not available for type analysis", 
			fmt.Errorf("%s", suggestion))
	}
	return nil
}

// analyzeModuleTypesWithPerl is the original Perl-based analysis function
func analyzeModuleTypesWithPerl(moduleName string) (*typedef.TypeDefinition, error) {
	// Ensure required Perl modules are available
	if err := ensureRequiredPerlModules(); err != nil {
		return nil, err
	}
	// Create a temporary Perl script to analyze the module
	tempScript, err := os.CreateTemp("", "psc-module-analyzer-*.pl")
	if err != nil {
		return nil, errors.NewSystemError("001",
			"Failed to create temporary file", err)
	}
	defer os.Remove(tempScript.Name())

	// Perl script for module analysis
	// This script uses Perl's introspection capabilities to extract type information
	analyzerScript := fmt.Sprintf(`#!/usr/bin/perl
use strict;
use warnings;
use feature 'say';
use lib "$ENV{HOME}/perl5/lib/perl5";
use JSON;
use Module::Load;
use Class::Inspector;
use B::Deparse;
use Data::Dumper;

# Module to analyze
my $module = '%s';

# Result structure
my $result = {
    module => $module,
    version => '0.0.1',
    types => [],
    subs => [],
    methods => [],
    packages => []
};

# Try to load the module
eval {
    load $module;
};
if ($@) {
    say STDERR "Failed to load module: $@";
    # Still attempt basic analysis even if we can't load it
}

# Analyze the module
analyze_module($module);

# Output the result as JSON
print encode_json($result);

# Main analysis function
sub analyze_module {
    my ($module) = @_;

    # Get the package namespace
    my $package = $module;
    $package =~ s/::/\//g;

    # Add basic module info
    push @{$result->{packages}}, {
        name => $module,
        description => "Package $module",
        version => $module->VERSION // '0.0.1',
        imports => []
    };

    # Get all the classes in the module's namespace
    my @classes = ($module);
    push @classes, find_subclasses($module);

    foreach my $class (@classes) {
        # Extract class information
        my $class_info = {
            name => $class,
            description => "Class $class",
            kind => "class",
            methods => [],
            properties => []
        };

        # Get methods
        my $methods = Class::Inspector->methods($class, 'public') || [];
        foreach my $method (@$methods) {
            # Skip some common Perl internals
            next if $method =~ /^(BEGIN|DESTROY|AUTOLOAD|import|END)$/;

            # Extract method info
            my $method_info = {
                name => $method,
                description => "Method $method",
                params => [],
                returns => "Any"
            };

            # Try to infer parameter types and return type
            if (my $code = $class->can($method)) {
                my $deparse = B::Deparse->new;
                my $src = $deparse->coderef2text($code);

                # Look for parameter assignments
                if ($src =~ /my\s*\(([^)]+)\)\s*=\s*\@_/) {
                    my $params = $1;
                    my @param_vars = split /\s*,\s*/, $params;
                    foreach my $var (@param_vars) {
                        push @{$method_info->{params}}, {
                            name => $var,
                            type => guess_type($var, $src)
                        };
                    }
                }

                # Try to infer return type
                if ($src =~ /return\s+(.+?);/s) {
                    $method_info->{returns} = guess_return_type($1, $src);
                }
            }

            # Add method to class
            push @{$class_info->{methods}}, $method_info;

            # Also add to the global methods list
            push @{$result->{methods}}, {
                name => $method,
                class => $class,
                description => "Method $class->$method",
                params => $method_info->{params},
                returns => $method_info->{returns}
            };
        }

        # Try to find properties
        my $properties = find_properties($class);
        $class_info->{properties} = $properties if @$properties;

        # Add class to types
        push @{$result->{types}}, $class_info;
    }

    # Find standalone functions
    my $subs = Class::Inspector->functions($module) || [];
    foreach my $sub (@$subs) {
        # Skip some common Perl internals
        next if $sub =~ /^(BEGIN|DESTROY|AUTOLOAD|import|END)$/;

        # Extract function info
        my $sub_info = {
            name => $sub,
            description => "Function $module\::$sub",
            params => [],
            returns => "Any"
        };

        # Try to infer parameter types and return type
        if (my $code = $module->can($sub)) {
            my $deparse = B::Deparse->new;
            my $src = $deparse->coderef2text($code);

            # Look for parameter assignments
            if ($src =~ /my\s*\(([^)]+)\)\s*=\s*\@_/) {
                my $params = $1;
                my @param_vars = split /\s*,\s*/, $params;
                foreach my $var (@param_vars) {
                    push @{$sub_info->{params}}, {
                        name => $var,
                        type => guess_type($var, $src)
                    };
                }
            }

            # Try to infer return type
            if ($src =~ /return\s+(.+?);/s) {
                $sub_info->{returns} = guess_return_type($1, $src);
            }
        }

        # Add function to subs
        push @{$result->{subs}}, $sub_info;
    }
}

# Helper to find subclasses in the module's namespace
sub find_subclasses {
    my ($parent) = @_;
    my @classes;

    # This is a simplified approach - in a production system,
    # you would want a more robust method to find all related classes
    foreach my $symbol (keys %%{$parent . '::'}) {
        next if $symbol =~ /^(BEGIN|DESTROY|AUTOLOAD|import|END)$/;
        my $full_name = $parent . '::' . $symbol;
        $full_name =~ s/::$//;

        # Check if it's a package
        if ($full_name->can('can')) {
            push @classes, $full_name unless $full_name eq $parent;
        }
    }

    return @classes;
}

# Helper to find class properties
sub find_properties {
    my ($class) = @_;
    my @properties;

    # Try to find has/field declarations (for Moo/Moose/Mouse classes)
    if ($class->can('meta') && $class->meta->can('get_attribute_list')) {
        my @attrs = $class->meta->get_attribute_list;
        foreach my $attr (@attrs) {
            push @properties, {
                name => $attr,
                type => "Any",
                description => "Property $attr"
            };
        }
    } else {
        # Try to infer from accessor methods
        my $methods = Class::Inspector->methods($class, 'public') || [];
        my %%potential_props;

        foreach my $method (@$methods) {
            if ($method =~ /^(get|set)_(.+)$/ ||
                $method =~ /^(.+?)(?:_accessor)?$/ && $class->can("${1}_accessor")) {
                my $prop = $2 || $1;
                $potential_props{$prop} = 1;
            }
        }

        foreach my $prop (keys %%potential_props) {
            push @properties, {
                name => $prop,
                type => "Any",
                description => "Property $prop"
            };
        }
    }

    return \@properties;
}

# Helper to guess a variable's type from its usage in code
sub guess_type {
    my ($var, $src) = @_;

    # Remove sigil for pattern matching
    my $bare_var = $var;
    $bare_var =~ s/^[\$\@\%%]//;

    # Look for hints in the code
    if ($src =~ /\Q$var\E\s*=~\s*\//) {
        return "Str";
    } elsif ($src =~ /\Q$var\E\s*\+\s*\d+/ ||
             $src =~ /\d+\s*\+\s*\Q$var\E/ ||
             $src =~ /\Q$var\E\s*\*\s*\d+/ ||
             $src =~ /\d+\s*\*\s*\Q$var\E/) {
        return "Num";
    } elsif ($src =~ /\Q$var\E\s*=\s*\d+\s*;/) {
        return "Int";
    } elsif ($src =~ /\Q$var\E\s*=\s*['"]/) {
        return "Str";
    } elsif ($src =~ /\Q$var\E\s*=\s*\[\s*/) {
        return "ArrayRef";
    } elsif ($src =~ /\Q$var\E\s*=\s*\{\s*/) {
        return "HashRef";
    } elsif ($src =~ /\Q$var\E\s*=\s*sub\s*\{/) {
        return "CodeRef";
    } elsif ($src =~ /\Q$var\E\s*=\s*bless/) {
        return "Object";
    } elsif ($var =~ /^\$/) {
        return "Scalar";
    } elsif ($var =~ /^\@/) {
        return "Array";
    } elsif ($var =~ /^\%%/) {
        return "Hash";
    }

    return "Any";
}

# Helper to guess a function's return type
sub guess_return_type {
    my ($return_expr, $src) = @_;

    if ($return_expr =~ /^\d+(\.\d+)?$/) {
        return $return_expr =~ /\./ ? "Num" : "Int";
    } elsif ($return_expr =~ /^['"]/) {
        return "Str";
    } elsif ($return_expr =~ /^\[/) {
        return "ArrayRef";
    } elsif ($return_expr =~ /^\{/) {
        return "HashRef";
    } elsif ($return_expr =~ /^sub\s*\{/) {
        return "CodeRef";
    } elsif ($return_expr =~ /^bless/) {
        return "Object";
    } elsif ($return_expr =~ /^\$/) {
        return "Scalar";
    } elsif ($return_expr =~ /^\@/) {
        return "Array";
    } elsif ($return_expr =~ /^\%%/) {
        return "Hash";
    }

    return "Any";
}
`, moduleName)

	// Write the analyzer script
	if _, err := tempScript.Write([]byte(analyzerScript)); err != nil {
		return nil, errors.NewSystemError("002",
			"Failed to write temporary file", err)
	}

	// Close the file to flush changes
	if err := tempScript.Close(); err != nil {
		return nil, errors.NewSystemError("003",
			"Failed to close temporary file", err)
	}

	// Make the script executable
	if err := os.Chmod(tempScript.Name(), 0755); err != nil {
		return nil, errors.NewSystemError("004",
			"Failed to make temporary file executable", err)
	}

	// Execute the script to analyze the module
	cmd := exec.Command("perl", tempScript.Name())

	// Capture stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	// Run the analysis script
	if err := cmd.Run(); err != nil {
		// Check if we got any error output
		stderrStr := stderrBuf.String()
		if stderrStr != "" {
			return nil, errors.NewTypeError(
				"PSC-810",
				fmt.Sprintf("Module analysis failed: %s", stderrStr),
				err)
		}

		return nil, errors.NewTypeError(
			"PSC-811",
			"Module analysis failed",
			err)
	}

	// Parse the JSON output
	var typeDef typedef.TypeDefinition
	if err := json.Unmarshal(stdoutBuf.Bytes(), &typeDef); err != nil {
		return nil, errors.NewTypeError(
			"PSC-812",
			"Failed to parse module analysis output",
			err)
	}

	return &typeDef, nil
}
