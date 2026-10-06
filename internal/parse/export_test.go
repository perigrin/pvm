// ABOUTME: Exposes package internals to this package's external tests, and only to them.
// ABOUTME: Compiled only under go test.
package parse

// CoreTable is coreTable, for TestCoreDeclarationsArePerls (coretable_test.go): its test asks
// perl through conformance, which imports this package.
var CoreTable = coreTable
