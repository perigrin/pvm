// ABOUTME: Tests for enhanced type definition generation
// ABOUTME: Verifies improved Perl module introspection capabilities

package psc

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnhancedTypeGeneration(t *testing.T) {
	// Create a test module
	tmpDir := t.TempDir()
	modulePath := filepath.Join(tmpDir, "TestModule.pm")

	moduleContent := `package TestModule;
use strict;
use warnings;

=head1 NAME

TestModule - A test module for type generation

=head1 SYNOPSIS

  use TestModule;

  my $obj = TestModule->new(name => 'test');
  my $result = $obj->process($data);

=head1 METHODS

=head2 new

  my $obj = TestModule->new(%args);

Constructor. Accepts the following arguments:

=over 4

=item * name (Str) - The object name (required)

=item * count (Int) - Item count (optional, default: 0)

=back

=cut

sub new {
	my ($class, %args) = @_;
	my $self = {
		name => $args{name} || die "name required",
		count => $args{count} || 0,
	};
	return bless $self, $class;
}

=head2 process

  my $result = $obj->process($data);

Process the given data. C<$data> should be a HashRef.

Returns an ArrayRef of processed items.

=cut

sub process {
	my ($self, $data) = @_;
	return [] unless ref($data) eq 'HASH';

	my @results;
	foreach my $key (keys %$data) {
		push @results, "$key: " . $data->{$key};
	}

	return \@results;
}

=head2 get_name

  my $name = $obj->get_name();

Returns the object's name (Str).

=cut

sub get_name {
	my ($self) = @_;
	return $self->{name};
}

=head2 set_count

  $obj->set_count($count);

Sets the count. C<$count> should be an Int.

=cut

sub set_count {
	my ($self, $count) = @_;
	$self->{count} = $count;
}

1;
`

	// Write the test module
	if err := os.WriteFile(modulePath, []byte(moduleContent), 0644); err != nil {
		t.Fatalf("Failed to write test module: %v", err)
	}

	// Test type generation
	typeDef, err := analyzeModuleTypes("TestModule")
	if err != nil {
		// Skip test if required Perl modules are not available
		if strings.Contains(err.Error(), "Can't locate JSON.pm") ||
			strings.Contains(err.Error(), "Module::Load") ||
			strings.Contains(err.Error(), "Class::Inspector") ||
			strings.Contains(err.Error(), "Required Perl modules not available") {
			t.Skip("Skipping test: required Perl modules not available (JSON, Module::Load, Class::Inspector)")
		}
		t.Errorf("Failed to analyze module: %v", err)
		return
	}

	if typeDef == nil {
		t.Fatal("Expected type definition, got nil")
	}

	// Verify basic information
	if typeDef.Module != "TestModule" {
		t.Errorf("Expected module name 'TestModule', got '%s'", typeDef.Module)
	}

	// Check for methods
	expectedMethods := []string{"new", "process", "get_name", "set_count"}
	foundMethods := make(map[string]bool)

	for _, method := range typeDef.Methods {
		foundMethods[method.Name] = true

		// Check specific method details
		switch method.Name {
		case "new":
			if len(method.Parameters) < 1 {
				t.Errorf("Expected parameters for 'new' method")
			}

		case "process":
			if len(method.Parameters) < 2 {
				t.Errorf("Expected at least 2 parameters for 'process' method")
			}
			if len(method.Returns) == 0 {
				t.Errorf("Expected return type for 'process' method")
			}

		case "get_name":
			if len(method.Returns) == 0 {
				t.Errorf("Expected return type for 'get_name' method")
			}
		}
	}

	// Verify all expected methods were found
	for _, expected := range expectedMethods {
		if !foundMethods[expected] {
			t.Errorf("Expected to find method '%s'", expected)
		}
	}

	// Check for type information
	if len(typeDef.Types) == 0 {
		t.Error("Expected at least one type definition")
	}
}

func TestTypeGenerationWithMoose(t *testing.T) {
	// Skip if Moose is not available
	if !isMooseAvailable() {
		t.Skip("Moose not available, skipping test")
	}

	// Create a Moose-based test module
	tmpDir := t.TempDir()
	modulePath := filepath.Join(tmpDir, "MooseTest.pm")

	moduleContent := `package MooseTest;
use Moose;

has 'name' => (
    is       => 'ro',
    isa      => 'Str',
    required => 1,
);

has 'items' => (
    is      => 'rw',
    isa     => 'ArrayRef[Str]',
    default => sub { [] },
);

sub add_item {
    my ($self, $item) = @_;
    push @{$self->items}, $item;
}

sub get_item_count {
    my ($self) = @_;
    return scalar @{$self->items};
}

__PACKAGE__->meta->make_immutable;
1;
`

	// Write the test module
	if err := os.WriteFile(modulePath, []byte(moduleContent), 0644); err != nil {
		t.Fatalf("Failed to write test module: %v", err)
	}

	// Test type generation
	typeDef, err := analyzeModuleTypes("MooseTest")
	if err != nil {
		t.Errorf("Failed to analyze Moose module: %v", err)
	}

	if typeDef == nil {
		t.Fatal("Expected type definition, got nil")
	}

	// Check for attributes
	hasNameAttr := false
	hasItemsAttr := false

	for _, typeInfo := range typeDef.Types {
		if typeInfo.Name == "MooseTest" {
			for _, prop := range typeInfo.Properties {
				switch prop.Name {
				case "name":
					hasNameAttr = true
					if prop.Type != "Str" {
						t.Errorf("Expected 'name' to have type 'Str', got '%s'", prop.Type)
					}
				case "items":
					hasItemsAttr = true
					// Might be detected as ArrayRef[Str] or just ArrayRef
					if !strings.Contains(prop.Type, "ArrayRef") {
						t.Errorf("Expected 'items' to have ArrayRef type, got '%s'", prop.Type)
					}
				}
			}
		}
	}

	if !hasNameAttr {
		t.Error("Expected to find 'name' attribute")
	}
	if !hasItemsAttr {
		t.Error("Expected to find 'items' attribute")
	}
}

func isMooseAvailable() bool {
	// Check if Moose is available by trying to load it
	cmd := exec.Command("perl", "-MMoose", "-e", "1")
	return cmd.Run() == nil
}
