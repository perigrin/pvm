#!/bin/bash

# Test script to simulate shell integration behavior
echo "=== Testing PATH Management Fix ==="

# Simulate starting with a polluted PATH (5.42.0 in PATH)
export PATH="/Users/perigrin/.local/share/pvm/versions/5.42.0/bin:/usr/bin:/bin"
echo "Initial PATH (polluted): $PATH"

# Simulate the _pvm_update_perl_path function logic
xdg_data_home="${XDG_DATA_HOME:-$HOME/.local/share}"

echo ""
echo "=== Step 1: Clean PATH ==="
clean_path=""
remaining_path="$PATH"
while [ -n "$remaining_path" ]; do
    if echo "$remaining_path" | grep -q ':'; then
        path_entry="${remaining_path%%:*}"
        remaining_path="${remaining_path#*:}"
    else
        path_entry="$remaining_path"
        remaining_path=""
    fi
    
    [ -z "$path_entry" ] && continue
    
    case "$path_entry" in
        */pvm/versions/*/bin)
            echo "Removing PVM path: $path_entry"
            ;;
        *)
            if [ -z "$clean_path" ]; then
                clean_path="$path_entry"
            else
                clean_path="$clean_path:$path_entry"
            fi
            ;;
    esac
done

echo "Clean PATH: $clean_path"

echo ""
echo "=== Step 2: Set clean PATH temporarily ==="
export PATH="$clean_path"

echo ""
echo "=== Step 3: Query PVM for current version ==="
current_version="$(pvm current --bare 2>/dev/null)"
echo "PVM says current version should be: $current_version"

echo ""
echo "=== Step 4: Add version to PATH if needed ==="
if [ -n "$current_version" ] && [ "$current_version" != "system" ]; then
    new_perl_bin="$xdg_data_home/pvm/versions/$current_version/bin"
    if [ -d "$new_perl_bin" ]; then
        export PATH="$new_perl_bin:$clean_path"
        echo "Added $new_perl_bin to PATH"
    else
        echo "Version directory doesn't exist: $new_perl_bin"
        export PATH="$clean_path"
    fi
else
    echo "Using system perl or no version - keeping clean PATH"
    export PATH="$clean_path"
fi

echo ""
echo "Final PATH: $PATH"
echo ""
echo "=== Verification ==="
echo "which perl: $(which perl 2>/dev/null || echo 'not found')"
echo "perl version: $(perl -v 2>/dev/null | head -2 | tail -1 || echo 'not available')"