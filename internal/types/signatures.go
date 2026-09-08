// ABOUTME: Builtin function, binary operator, and unary operator signatures for PSC type inference.
// ABOUTME: Provides typed arities and return types ported from Chalk's TypeLibrary.

package types

// BuiltinSig describes the type signature of a Perl builtin function.
// ArgTypes holds the expected argument types in order; the last entry is
// considered variadic (may be repeated). MinArity is the minimum number
// of arguments required.
type BuiltinSig struct {
	MinArity   int
	ArgTypes   []Type
	ReturnType Type
}

// BinaryOpSig describes the type signature of a binary operator: the types
// of the left and right operands and the result type.
type BinaryOpSig struct {
	Left, Right, Result Type
}

// UnaryOpSig describes the type signature of a unary operator: the type of
// the operand and the result type.
type UnaryOpSig struct {
	Operand, Result Type
}

// builtins maps Perl builtin function names to their type signatures.
// The last element in ArgTypes is variadic — it may appear more than once.
var builtins = map[string]BuiltinSig{
	// push/unshift/splice take a LIST, not Any. Everything flattens into it —
	// measured, scalars, arrays, hashes and refs all append — so List accepts
	// every value Any did, while still excluding Code and Glob. Any is the
	// annotation escape hatch and has no business standing in for an
	// unexamined slot.
	"push":    {MinArity: 2, ArgTypes: []Type{Array, List}, ReturnType: Int},
	"pop":     {MinArity: 0, ArgTypes: []Type{Array}, ReturnType: Scalar},
	"shift":   {MinArity: 0, ArgTypes: []Type{Array}, ReturnType: Scalar},
	"unshift": {MinArity: 2, ArgTypes: []Type{Array, List}, ReturnType: Int},
	"splice":  {MinArity: 1, ArgTypes: []Type{Array, Int, Int, List}, ReturnType: List},

	"keys":   {MinArity: 1, ArgTypes: []Type{Hash | Array}, ReturnType: List},
	"values": {MinArity: 1, ArgTypes: []Type{Hash | Array}, ReturnType: List},
	"delete": {MinArity: 1, ArgTypes: []Type{Scalar}, ReturnType: Scalar},
	"exists": {MinArity: 1, ArgTypes: []Type{Scalar}, ReturnType: Bool},
	"each":   {MinArity: 1, ArgTypes: []Type{Hash | Array}, ReturnType: List},

	"length": {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: Int},
	"chomp":  {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: Int},
	"chop":   {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: Str},
	"chr":    {MinArity: 0, ArgTypes: []Type{Int}, ReturnType: Str},
	"ord":    {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: Int},

	// join takes a separator and then a LIST, not a series of strings: the
	// variadic tail repeats the last ArgTypes entry, so Str there rejected
	// `join ":", @foo`. Measured: join(":", @f) flattens the array.
	"join": {MinArity: 2, ArgTypes: []Type{Str, List}, ReturnType: Str},
	// split's pattern may be a compiled Regex or a plain Str: perl compiles a
	// string into a pattern, so `split ":", $x` is ordinary Perl. Measured on
	// 5.42, /:/ and ":" and $sep and qr/:/ all behave identically. Regex
	// alone rejected the string form, which was the largest single class of
	// false positives on perl5/lib.
	"split": {MinArity: 0, ArgTypes: []Type{Regex | Str, Str, Int}, ReturnType: List},
	// sprintf takes a format and then the LIST of values it interpolates.
	"sprintf": {MinArity: 1, ArgTypes: []Type{Str, List}, ReturnType: Str},
	"substr":  {MinArity: 2, ArgTypes: []Type{Str, Num, Num}, ReturnType: Str},

	"defined": {MinArity: 0, ArgTypes: []Type{Scalar}, ReturnType: Bool},
	"ref":     {MinArity: 0, ArgTypes: []Type{Scalar}, ReturnType: Str},
	// scalar() imposes scalar context on anything, so List rather than Any:
	// every value can be evaluated in scalar context, and List still excludes
	// Code and Glob. The RETURN type depends on the argument and is computed
	// by contextualReturnType — a signature can only name one type, and this
	// builtin's answer is "whatever the argument becomes".
	"scalar": {MinArity: 1, ArgTypes: []Type{List}, ReturnType: Scalar},

	"die":  {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: None},
	"warn": {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: Bool},

	"bless": {MinArity: 1, ArgTypes: []Type{Ref, Str}, ReturnType: Object},

	"print":  {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: Bool},
	"say":    {MinArity: 0, ArgTypes: []Type{Str}, ReturnType: Bool},
	"return": {MinArity: 0, ArgTypes: []Type{Any}, ReturnType: Any},

	// Numeric transforms. abs follows its argument and int always truncates
	// to an integer; both are narrowed further by contextualReturnType, which
	// a signature cannot express.
	"abs": {MinArity: 1, ArgTypes: []Type{Num}, ReturnType: Num},
	"int": {MinArity: 1, ArgTypes: []Type{Num}, ReturnType: Int},

	// String transforms. Each takes a string and yields one; return types
	// measured on 5.42 rather than transcribed.
	"uc":      {MinArity: 1, ArgTypes: []Type{Str}, ReturnType: Str},
	"lc":      {MinArity: 1, ArgTypes: []Type{Str}, ReturnType: Str},
	"ucfirst": {MinArity: 1, ArgTypes: []Type{Str}, ReturnType: Str},
	"lcfirst": {MinArity: 1, ArgTypes: []Type{Str}, ReturnType: Str},

	// index/rindex search a string for a substring and yield a position, or
	// -1 when absent — an Int either way.
	"index":  {MinArity: 2, ArgTypes: []Type{Str, Str, Int}, ReturnType: Int},
	"rindex": {MinArity: 2, ArgTypes: []Type{Str, Str, Int}, ReturnType: Int},

	// reverse is context-dependent: a reversed list in list context, a
	// reversed string in scalar context. List covers both under the arity
	// ordering, since Scalar <: List.
	"reverse": {MinArity: 1, ArgTypes: []Type{List}, ReturnType: List},

	"map":  {MinArity: 2, ArgTypes: []Type{Code, List}, ReturnType: List},
	"grep": {MinArity: 2, ArgTypes: []Type{Code, List}, ReturnType: List},
	"sort": {MinArity: 1, ArgTypes: []Type{List}, ReturnType: List},
}

// GetBuiltin returns the BuiltinSig for the named Perl builtin and true if
// the name is known, or the zero value and false otherwise.
func GetBuiltin(name string) (BuiltinSig, bool) {
	sig, ok := builtins[name]
	return sig, ok
}

// HasBuiltin reports whether the named function is a known Perl builtin.
func HasBuiltin(name string) bool {
	_, ok := builtins[name]
	return ok
}

// binaryOps maps binary operator symbols/keywords to their type signatures.
var binaryOps = map[string]BinaryOpSig{
	// Arithmetic
	"+":  {Left: Num, Right: Num, Result: Num},
	"-":  {Left: Num, Right: Num, Result: Num},
	"*":  {Left: Num, Right: Num, Result: Num},
	"/":  {Left: Num, Right: Num, Result: Num},
	"%":  {Left: Num, Right: Num, Result: Num},
	"**": {Left: Num, Right: Num, Result: Num},

	// String
	".": {Left: Str, Right: Str, Result: Str},
	"x": {Left: Str, Right: Int, Result: Str},

	// Numeric comparison
	"==":  {Left: Num, Right: Num, Result: Bool},
	"!=":  {Left: Num, Right: Num, Result: Bool},
	"<":   {Left: Num, Right: Num, Result: Bool},
	">":   {Left: Num, Right: Num, Result: Bool},
	"<=":  {Left: Num, Right: Num, Result: Bool},
	">=":  {Left: Num, Right: Num, Result: Bool},
	"<=>": {Left: Num, Right: Num, Result: Int},

	// String comparison
	"eq":  {Left: Str, Right: Str, Result: Bool},
	"ne":  {Left: Str, Right: Str, Result: Bool},
	"lt":  {Left: Str, Right: Str, Result: Bool},
	"gt":  {Left: Str, Right: Str, Result: Bool},
	"le":  {Left: Str, Right: Str, Result: Bool},
	"ge":  {Left: Str, Right: Str, Result: Bool},
	"cmp": {Left: Str, Right: Str, Result: Int},

	// Logical
	"&&":  {Left: Any, Right: Any, Result: Any},
	"||":  {Left: Any, Right: Any, Result: Any},
	"//":  {Left: Any, Right: Any, Result: Any},
	"and": {Left: Any, Right: Any, Result: Any},
	"or":  {Left: Any, Right: Any, Result: Any},
	"xor": {Left: Any, Right: Any, Result: Bool},

	// Bitwise
	"&":  {Left: Int, Right: Int, Result: Int},
	"|":  {Left: Int, Right: Int, Result: Int},
	"^":  {Left: Int, Right: Int, Result: Int},
	"<<": {Left: Int, Right: Int, Result: Int},
	">>": {Left: Int, Right: Int, Result: Int},

	// Type test
	"isa": {Left: Scalar, Right: Str, Result: Bool},

	// Regex binding
	"=~": {Left: Str, Right: Regex, Result: Bool},
	"!~": {Left: Str, Right: Regex, Result: Bool},

	// Range
	"..":  {Left: Int, Right: Int, Result: List},
	"...": {Left: Int, Right: Int, Result: List},

	// Assignment
	"=": {Left: Any, Right: Any, Result: Any},
}

// GetBinaryOp returns the BinaryOpSig for the given operator symbol or keyword
// and true if it is known, or the zero value and false otherwise.
func GetBinaryOp(op string) (BinaryOpSig, bool) {
	sig, ok := binaryOps[op]
	return sig, ok
}

// unaryOps maps unary operator symbols/keywords to their type signatures.
var unaryOps = map[string]UnaryOpSig{
	"-":   {Operand: Num, Result: Num},
	"+":   {Operand: Num, Result: Num},
	"!":   {Operand: Any, Result: Bool},
	"not": {Operand: Any, Result: Bool},
	"~":   {Operand: Int, Result: Int},
	`\`:   {Operand: Any, Result: Ref},
}

// GetUnaryOp returns the UnaryOpSig for the given operator symbol or keyword
// and true if it is known, or the zero value and false otherwise.
func GetUnaryOp(op string) (UnaryOpSig, bool) {
	sig, ok := unaryOps[op]
	return sig, ok
}
