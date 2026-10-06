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
	// slot, '@' or '%' for a slurpy.
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
}
