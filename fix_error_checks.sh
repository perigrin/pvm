#!/bin/bash

# Fix calls in function literals and defer statements
find . -name "*.go" -exec sed -i '' 's/\(\s\+\)defer os.RemoveAll(\([^)]\+\))/\1defer func() { _ = os.RemoveAll(\2) }()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\(\s\+\)defer os.Remove(\([^)]\+\))/\1defer func() { _ = os.Remove(\2) }()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\(\s\+\)defer os.Chdir(\([^)]\+\))/\1defer func() { _ = os.Chdir(\2) }()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\(\s\+\)defer f.Close()/\1defer func() { _ = f.Close() }()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\(\s\+\)defer file.Close()/\1defer func() { _ = file.Close() }()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\(\s\+\)defer sourceFile.Close()/\1defer func() { _ = sourceFile.Close() }()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\(\s\+\)defer destFile.Close()/\1defer func() { _ = destFile.Close() }()/g' {} \;

# Fix function literals with cleanup in them
find . -name "*.go" -exec sed -i '' 's/func() { os.RemoveAll(\([^)]\+\)) }/func() { _ = os.RemoveAll(\1) }/g' {} \;

# Fix direct calls
find . -name "*.go" -exec sed -i '' 's/\([^_]\)= os.RemoveAll/\1= _ = os.RemoveAll/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\([^_]\)= os.Remove/\1= _ = os.Remove/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\([^_]\)= os.Chdir/\1= _ = os.Chdir/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\([^_]\)= os.Unsetenv/\1= _ = os.Unsetenv/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\([^_]\)= os.Setenv/\1= _ = os.Setenv/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\([^_]\)= f.Close/\1= _ = f.Close/g' {} \;
find . -name "*.go" -exec sed -i '' 's/\([^_]\)= file.Close/\1= _ = file.Close/g' {} \;

# Fix remaining cases not caught by the above
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)os.RemoveAll(\([^)]\+\))/\1_ = os.RemoveAll(\2)/g' {} \;
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)os.Remove(\([^)]\+\))/\1_ = os.Remove(\2)/g' {} \;
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)os.Chdir(\([^)]\+\))/\1_ = os.Chdir(\2)/g' {} \;
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)os.Unsetenv(\([^)]\+\))/\1_ = os.Unsetenv(\2)/g' {} \;
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)os.Setenv(\([^)]\+\))/\1_ = os.Setenv(\2)/g' {} \;
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)f.Close()/\1_ = f.Close()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)file.Close()/\1_ = file.Close()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)sourceFile.Close()/\1_ = sourceFile.Close()/g' {} \;
find . -name "*.go" -exec sed -i '' 's/^\(\s\+\)destFile.Close()/\1_ = destFile.Close()/g' {} \;
