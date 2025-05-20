#\!/bin/bash
cd test/e2e
# Create a fixed copy of the file
cp pvx_isolation_test.go pvx_isolation_test.go.fixed
# Replace all occurrences of "stdout :=" with "stdout =" for existing variables
sed -i '' 's/stdout :=/stdout =/g' pvx_isolation_test.go.fixed
# Format the file
gofmt -w pvx_isolation_test.go.fixed
# Replace the original with the fixed version
mv pvx_isolation_test.go.fixed pvx_isolation_test.go
