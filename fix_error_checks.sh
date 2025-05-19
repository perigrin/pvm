#!/bin/bash

# Fix all defer statements that return errors
find . -name "*.go" -exec sed -i '' 's/defer \([a-zA-Z0-9_.]\+\)\.Close()/defer func() { _ = \1.Close() }()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/defer os\.RemoveAll(\([^)]\+\))/defer func() { _ = os.RemoveAll(\1) }()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/defer os\.Remove(\([^)]\+\))/defer func() { _ = os.Remove(\1) }()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/defer os\.Chdir(\([^)]\+\))/defer func() { _ = os.Chdir(\1) }()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/defer os\.Unsetenv(\([^)]\+\))/defer func() { _ = os.Unsetenv(\1) }()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/defer os\.Setenv(\([^)]\+\))/defer func() { _ = os.Setenv(\1) }()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/defer resp\.Body\.Close()/defer func() { _ = resp.Body.Close() }()/g' {} \;

# Fix direct calls to functions that return errors
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)os\.RemoveAll(/\1_ = os.RemoveAll(/g' {} \;
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)os\.Remove(/\1_ = os.Remove(/g' {} \;
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)os\.Chdir(/\1_ = os.Chdir(/g' {} \;
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)os\.Unsetenv(/\1_ = os.Unsetenv(/g' {} \;
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)os\.Setenv(/\1_ = os.Setenv(/g' {} \;

# Fix in cleanup functions
find . -name "*.go" -exec sed -i '' 's/func() { os\.RemoveAll(/func() { _ = os.RemoveAll(/g' {} \;
find . -name "*.go" -exec sed -i '' 's/env\.cleanup = append(env\.cleanup, func() { os\.RemoveAll(/env.cleanup = append(env.cleanup, func() { _ = os.RemoveAll(/g' {} \;

# Fix patterns in test files
find ./test -name "*.go" -exec sed -i '' 's/os\.Setenv(/_ = os.Setenv(/g' {} \;
find ./test -name "*.go" -exec sed -i '' 's/os\.Unsetenv(/_ = os.Unsetenv(/g' {} \;
find ./test -name "*.go" -exec sed -i '' 's/os\.RemoveAll(/_ = os.RemoveAll(/g' {} \;
find ./test -name "*.go" -exec sed -i '' 's/os\.Remove(/_ = os.Remove(/g' {} \;

# Fix specific cases in downloaded
find . -name "download_test.go" -exec sed -i '' 's/defer os\.Remove(tmpFile\.Name())/defer func() { _ = os.Remove(tmpFile.Name()) }()/g' {} \;
find . -name "legacy.go" -exec sed -i '' 's/defer file\.Close()/defer func() { _ = file.Close() }()/g' {} \;
find . -name "legacy_test.go" -exec sed -i '' 's/defer os\.RemoveAll(tempDir)/defer func() { _ = os.RemoveAll(tempDir) }()/g' {} \;
find . -name "shell_test.go" -exec sed -i '' 's/defer os\.Chdir(/defer func() { _ = os.Chdir(/g' {} \;
find . -name "*.go" -exec sed -i '' 's/os\.Chdir(tmpDir)/_ = os.Chdir(tmpDir)/g' {} \;
