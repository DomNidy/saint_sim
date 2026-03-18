#!/bin/bash

# This script runs `go mod tidy` across all modules in the workspace

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
project_root="$(cd -- "$script_dir/.." && pwd)"

# find all directories with a go.mod file, and run go mod tidy in them
find "$project_root" -type f -name "go.mod" -print0 | while IFS= read -r -d '' mod_file; do
    module_dir=$(dirname "$mod_file")
    echo "Tidying $module_dir"
    cd "$module_dir" || exit 1
    go mod tidy
done
