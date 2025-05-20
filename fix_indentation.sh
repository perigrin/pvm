#\!/bin/bash
# Script to fix indentation issues

# Use awk to fix the indentation
awk '{
    # Replace tabs with proper indentation for switch statement
    gsub(/switch parsedConstraint.Operator {/, "switch parsedConstraint.Operator {")
    gsub(/				case OpGreaterThan:/, "			case OpGreaterThan:")
    gsub(/					lowerSatisfied = compareVersions/, "				lowerSatisfied = compareVersions")

    gsub(/switch parsedConstraint.UpperOperator {/, "switch parsedConstraint.UpperOperator {")
    gsub(/				case OpLessThan:/, "			case OpLessThan:")
    gsub(/					upperSatisfied = compareVersions/, "				upperSatisfied = compareVersions")

    print
}' internal/pvi/deps/version_constraints.go > temp_file && mv temp_file internal/pvi/deps/version_constraints.go

echo "Indentation fixes applied"
