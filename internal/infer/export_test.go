// ABOUTME: Exposes the lowered path's unexported helpers to the infer_test package's tests.
// ABOUTME: It is compiled only with the tests, so the package's API does not grow.

package infer

// The lowered pass 1 and the helpers the guard and parameter rules read
// a sub with, under the names the tests call them by.
var (
	CollectLoweredDeclarations = collectLoweredDeclarations
	FileSubs                   = fileSubs
	SubParam                   = subParam
	SubParams                  = subParams
	SubReturnExpr              = subReturnExpr
	ArgVarName                 = argVarName
	Finite                     = finite
	SpaceshipResult            = spaceshipResult
	Numified                   = numified
)
