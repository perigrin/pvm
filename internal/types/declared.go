// ABOUTME: A sub's typed signature as a .pmt declaration states it: each parameter's type and default.
// ABOUTME: Read by internal/parse; held here so its consumers need not import the parser.

package types

// Signature is the typed parameter list of a .pmt declaration, RFC 0001
// "Typed Perl, in `.pmt` only": `sub bless :prototype($;$) (Ref $ref, Str
// $class = __PACKAGE__) Object;`.
type Signature struct {
	// Invocant is the leading slot a call fills with no comma after it,
	// RFC 0001 "Builtins that keep their own parse": `print STDERR LIST`'s
	// handle, written `(FileHandle $fh = select(): List[Str] @args)`. nil
	// when the declaration has none. It is not one of Params, so it never
	// binds an argument by position.
	Invocant *Param

	Params []Param

	// Returns is the lattice type the declaration states after its closing
	// paren, Unknown when it states none.
	Returns Type

	// Context is the set of calling contexts the declaration answers for,
	// its `:context(...)`.
	Context Contexts

	// Unary is whether the declaration states `:unary`, RFC 0001
	// "Builtins with no prototype": a builtin with no prototype that perl
	// reads as a named unary, so `defined $a, $b` is `(defined $a), $b`.
	// Without it such a builtin is a list operator, as an ordinary sub is.
	Unary bool

	// ListOp is whether the declaration states `:listop`, RFC 0001
	// "Builtins with no prototype": a builtin with no prototype whose
	// parameters are typed and positional, so its types derive none. split
	// is one: `split $a, $b` is `split(/$a/, $b, 0)`, a list operator, yet
	// it takes its string in scalar context.
	ListOp bool
}

// Contexts is a set of calling contexts, RFC 0001 "`:context(...)`":
// `:context($@)` is ContextSet(ScalarCtx, ListCtx) and `:context()` is
// ContextSet(VoidCtx). As with `:prototype`, absent and empty differ: a
// declaration stating no `:context` is EveryContext.
type Contexts uint8

// EveryContext is the set of a declaration that states no `:context`.
const EveryContext Contexts = 0

// ContextSet is the set holding each of cs.
func ContextSet(cs ...Context) Contexts {
	var s Contexts
	for _, c := range cs {
		s |= 1 << c
	}
	return s
}

// Param is one typed parameter, `Str $class = __PACKAGE__`.
type Param struct {
	// Name is the variable's name without its sigil: "class".
	Name string

	// Sigil is '$' for a scalar, '&' for a code slot, '*' for a glob
	// slot, '@' or '%' for an array or hash: a slurpy, unless Alias.
	Sigil byte

	// Type is the lattice type the parameter's type name resolves to.
	Type Type

	// Element is the element type of a container type, Str for `List[Str]
	// @args`, Unknown when the type names no element. What a container
	// means is provisional until the paper defines it (RFC 0001, "A slurpy
	// takes no bare element type"), so the element is recorded beside the
	// container and the lattice stays flat.
	Element Type

	// Default is the default expression as written, "" when there is none.
	Default string

	// Required is whether a call must pass this argument. A scalar with no
	// default is required, as in a perl signature. A slurpy is not: it
	// accepts zero arguments. A `= die` default makes either required (RFC
	// 0001 "A required argument defaults to `die`"): it runs only when the
	// argument is omitted, so it is recorded here and not as a Default.
	Required bool

	// Alias is whether the parameter is backslashed, `Array \@a`: it
	// aliases the one container the caller writes, as perlref's
	// refaliasing does (RFC 0001, "The scalar container"), and takes
	// exactly one argument whatever its sigil.
	Alias bool

	// AliasEach is whether the parameter is perlref's list form, `List[Str]
	// \(@args)`: it aliases each argument, and takes the rest of the call
	// as an unbackslashed List parameter does.
	AliasEach bool
}

// Variable is the parameter's variable as a declaration writes it, with
// the backslash an aliased one takes: `$x`, `\@a`, `\(@args)`.
func (p Param) Variable() string {
	v := string(p.Sigil) + p.Name
	switch {
	case p.Alias:
		return `\` + v
	case p.AliasEach:
		return `\(` + v + ")"
	}
	return v
}

// Slurpy reports whether the parameter takes the rest of the call: an
// `@` or `%` that does not alias one container.
func (p Param) Slurpy() bool {
	return (p.Sigil == '@' || p.Sigil == '%') && !p.Alias
}
