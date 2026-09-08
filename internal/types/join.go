// ABOUTME: Join — the type of a value arriving from two branches of a control-flow merge.
// ABOUTME: Least upper bound in the lattice, with Unknown deliberately excluded from absorbing.

package types

// Join returns the least type that both a and b are subtypes of.
//
// Every control-flow merge — `$c ? $a : $b`, if/else, `//`, `&&` — has to give
// one type to a value that arrived from either arm. The paper calls that the
// join: the least T with A <: T and B <: T.
//
// In a bitset lattice the least upper bound is the union of the bits, because
// a type IS the set of leaves it admits. That falls out for the named masks
// too: Bool|Int is not a named type, but every named type containing both
// bits contains Scalar's, and Scalar is the smallest such — so
// Join(Bool, Int) is Bool|Int, which prints and behaves as the subset of
// Scalar it is. The paper's `Boolean ⊔ Int = Scalar` is the same claim at the
// granularity of named types.
//
// Two cases are not the plain union:
//
// None is the identity. It is the empty set, so it adds no leaves — but it
// also carries a sentinel bit that must not survive into the result. The
// paper makes this load-bearing: a recursive function types from its base
// case only if T ⊔ None = T, since the recursive arm is None on the first
// pass.
//
// Unknown is dropped rather than absorbed. It is not a claim about a value —
// it is the statement that inference has not determined one — and as the top
// of the lattice a naive least-upper-bound would let it swallow every arm
// that WAS inferred. `Int ⊔ Unknown` is Int: the un-inferred arm contributes
// no constraint, and erasing the constraint the other arm supplied would be
// strictly worse than keeping it. This is epistemic, and distinct from
// dropping Undef in a `//` merge, which is semantic.
func Join(a, b Type) Type {
	// None is the join identity: the empty set adds nothing, and its
	// sentinel bit must not leak into the result.
	if a == None {
		return b
	}
	if b == None {
		return a
	}

	// An un-inferred arm contributes no constraint. Only when BOTH arms are
	// un-inferred is the merge itself un-inferred.
	if a == Unknown {
		return b
	}
	if b == Unknown {
		return a
	}

	return a | b
}

// JoinAll returns the join of every type in the slice, or Unknown when the
// slice is empty. It is the n-ary form of Join, for merges with more than two
// arms — an if/elsif/else chain, or a subroutine with several return paths.
func JoinAll(ts ...Type) Type {
	result := Unknown
	for _, t := range ts {
		result = Join(result, t)
	}
	return result
}
