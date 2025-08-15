#!/usr/bin/env python3
"""
Systematically fix parser test expectations.
"""

import os
import subprocess
import re

def run_test(test_name):
    """Run a specific test and capture output."""
    cmd = f'go test ./internal/parser -run "{test_name}" -v 2>&1'
    result = subprocess.run(cmd, shell=True, capture_output=True, text=True)
    return result.stdout

def extract_diff(output):
    """Extract the diff from test output."""
    lines = output.split('\n')
    diff_lines = []
    in_diff = False

    for line in lines:
        if '(-want +got)' in line or '(-expected +actual)' in line:
            in_diff = True
            continue
        if in_diff:
            if line.strip().startswith('---') and 'FAIL' in line:
                break
            diff_lines.append(line)

    return '\n'.join(diff_lines)

def get_failing_tests():
    """Get list of failing typed-perl tests."""
    output = run_test("TestParserByCategory/typed-perl")
    failing = []

    for line in output.split('\n'):
        if '--- FAIL' in line and 'typed-perl/' in line:
            # Extract test name
            match = re.search(r'typed-perl/([^)]+)', line)
            if match:
                failing.append(match.group(1))

    return failing

def find_test_file(test_name):
    """Find the markdown file for a test."""
    # Convert test name to likely file name
    # e.g., "method-signatures_method_signatures" -> "method-signatures.md"
    base_name = test_name.split('_')[0] if '_' in test_name else test_name

    for root, dirs, files in os.walk('testdata/corpus/parser/typed-perl'):
        for file in files:
            if file.endswith('.md') and base_name in file:
                return os.path.join(root, file)

    return None

def main():
    """Main function."""
    print("Finding failing tests...")
    failing_tests = get_failing_tests()
    print(f"Found {len(failing_tests)} failing tests")

    for test_name in failing_tests:
        print(f"\n=== Analyzing: {test_name} ===")

        # Run test to get diff
        output = run_test(f"TestParserByCategory/typed-perl/{test_name}")
        diff = extract_diff(output)

        if diff:
            print("Test output diff:")
            for line in diff.split('\n')[:20]:  # Show first 20 lines
                print(f"  {line}")

        # Find test file
        test_file = find_test_file(test_name)
        if test_file:
            print(f"Test file: {test_file}")
        else:
            print("Could not find test file")

    print("\n=== Summary ===")
    print(f"Total failing tests: {len(failing_tests)}")
    print("\nThese tests need manual review to fix expectations.")
    print("Common issues:")
    print("  - return_stmt vs expression_stmt mismatches")
    print("  - literal vs variable mismatches")
    print("  - Missing or extra tokens")

if __name__ == '__main__':
    main()
