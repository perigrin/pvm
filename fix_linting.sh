#\!/bin/bash
# Script to fix linting issues

# Fix if statements with tagged switches in version_constraints.go
sed -i '' 's/if parsedConstraint.Operator == OpGreaterThan {/switch parsedConstraint.Operator {/g' internal/pvi/deps/version_constraints.go
sed -i '' 's/lowerSatisfied = compareVersions(normVersion, parsedConstraint.Version) > 0/case OpGreaterThan:\n\t\t\t\tlowerSatisfied = compareVersions(normVersion, parsedConstraint.Version) > 0/g' internal/pvi/deps/version_constraints.go
sed -i '' 's/} else if parsedConstraint.Operator == OpGreaterThanOrEqual {/case OpGreaterThanOrEqual:/g' internal/pvi/deps/version_constraints.go

sed -i '' 's/if parsedConstraint.UpperOperator == OpLessThan {/switch parsedConstraint.UpperOperator {/g' internal/pvi/deps/version_constraints.go
sed -i '' 's/upperSatisfied = compareVersions(normVersion, parsedConstraint.UpperVersion) < 0/case OpLessThan:\n\t\t\t\tupperSatisfied = compareVersions(normVersion, parsedConstraint.UpperVersion) < 0/g' internal/pvi/deps/version_constraints.go
sed -i '' 's/} else if parsedConstraint.UpperOperator == OpLessThanOrEqual {/case OpLessThanOrEqual:/g' internal/pvi/deps/version_constraints.go

# Fix unused disableNetwork field in CPANProvider
# sed -i '' 's/disableNetwork bool/\/\/ disableNetwork bool - unused for now/g' internal/cpan/cpan.go

echo "Linting fixes applied"
