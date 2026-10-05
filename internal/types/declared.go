// ABOUTME: A sub's typed signature as a .pmt declaration states it: each parameter's type and default.
// ABOUTME: Read by internal/parse; held here so its consumers need not import the parser.

package types

// Signature is the typed parameter list of a .pmt declaration, RFC 0001
// "Typed Perl, in `.pmt` only": `sub bless :prototype($;$) (Ref $ref, Str
// $class = __PACKAGE__) Object;`.
type Signature struct {
	Params []Param

	// Returns is the lattice type the declaration states after its closing
	// paren, Unknown when it states none.
	Returns Type
}

// Param is one typed parameter, `Str $class = __PACKAGE__`.
type Param struct {
	// Name is the variable's name without its sigil: "class".
	Name string

	// Sigil is '$' for a scalar, '@' or '%' for a slurpy.
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
