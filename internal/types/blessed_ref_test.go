// ABOUTME: BlessedRef is any blessed reference; Object is the class feature's
// ABOUTME: instance (reftype OBJECT) under it; Regex (qr//) is beside Object.

package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"tamarou.com/pvm/internal/types"
)

// Measured on perl 5.42: a class-feature instance is blessed with reftype
// OBJECT, `bless {}` is HASH, and qr// is blessed Regexp with reftype REGEXP.
// All three are blessed references; only the first is an Object.
func TestBlessedRefIsAboveObjectAndRegex(t *testing.T) {
	assert.True(t, types.IsSubtype(types.BlessedRef, types.Ref), "a BlessedRef is a Ref")
	assert.True(t, types.IsSubtype(types.Object, types.BlessedRef), "an Object is a BlessedRef")
	assert.True(t, types.IsSubtype(types.Regex, types.BlessedRef), "a Regex is a BlessedRef")
	assert.False(t, types.IsSubtype(types.Regex, types.Object), "a Regex is not an Object")
	assert.False(t, types.IsSubtype(types.BlessedRef, types.Object), "not every BlessedRef is an Object")
}
