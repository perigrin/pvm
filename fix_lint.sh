#\!/bin/bash
# Fix the specific issue on line 560
cd test/e2e
sed -i '' 's/stdout :=/_ =/' pvx_isolation_test.go
# Fix if-else chains
gofmt -w pvx_isolation_test.go
