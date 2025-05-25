#!/usr/bin/env perl

package TestModule;
use strict;
use warnings;

=head1 NAME

TestModule - A test module for introspection

=head1 SYNOPSIS

    use TestModule;

    my $obj = TestModule->new(name => 'test');
    my $result = $obj->process($data);

=head1 METHODS

=head2 new

    my $obj = TestModule->new(%args);

Constructor. Accepts the following parameters:

=over 4

=item * C<name> - I<Str> - The name of the object (required)

=item * C<debug> - I<Bool> - Enable debug mode (optional, default: 0)

=back

=cut

sub new {
    my ($class, %args) = @_;

    my $self = {
        name => $args{name} || die "name is required",
        debug => $args{debug} || 0,
        _internal => {},
    };

    return bless $self, $class;
}

=head2 process

    my $result = $obj->process($data);

Process the given data.

=over 4

=item * C<$data> - I<HashRef> - The data to process

=back

Returns: I<ArrayRef[Str]> - The processed results

=cut

sub process {
    my ($self, $data) = @_;

    my @results;
    foreach my $key (keys %$data) {
        push @results, "$key: " . $data->{$key};
    }

    return \@results;
}

=head2 get_name

    my $name = $obj->get_name();

Returns the object's name.

Returns: I<Str>

=cut

sub get_name {
    my $self = shift;
    return $self->{name};
}

# Dynamic method generation
__PACKAGE__->mk_accessors(qw(status priority));

# Class::Accessor style
sub mk_accessors {
    my $class = shift;
    foreach my $field (@_) {
        no strict 'refs';
        *{"${class}::${field}"} = sub {
            my $self = shift;
            $self->{$field} = shift if @_;
            return $self->{$field};
        };
    }
}

=head1 ATTRIBUTES

=item status

The current status. Type: Str

=item priority

The priority level. Type: Int

=back

=cut

1;
