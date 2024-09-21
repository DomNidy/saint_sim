#!/bin/bash

#* Important: This script relies on the 'github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest' binary being installed
# This script generates types that are defined in /apps/{app_name}/openapi.yaml files

# path to the openapi.yaml file in the /apps/api application
API_OPENAPI_SPEC="./apps/api/openapi.yaml"

# Where to place generate type files
OUTPUT_DIR="./pkg/interfaces"

# Path to the codegen binary
# First we check if we can execute it without specifying .exe (for linux)
# If we cant, just try to use oapi-codegen.exe instead
# This is mostly meant for WSL users, as installing the codegen tool inside wsl
# may be a bit tedious.
CODEGEN_BINARY_PATH=""

# check if we can find oapi-codegen binary without the .exe extension
if which oapi-codegen &>/dev/null; then
    CODEGEN_BINARY_PATH="oapi-codegen"
# check if we can find oapi-codegen binary with the .exe extension
elif which oapi-codegen.exe &>/dev/null; then
    CODEGEN_BINARY_PATH="oapi-codegen.exe"
fi

# if we couldn't find the codegen binary ($CODEGEN_BINARY_PATH is still an empty str), exit out
if [ -z "$CODEGEN_BINARY_PATH" ]; then
    echo "Error: oapi-codegen not found. Please ensure it is installed and accessible in your PATH."
    echo "hint: maybe you want to run 'go install github.com/deepmap/oapi-codegen/cmd/oapi-codegen@latest'"
    exit 1
fi

# use command substitution to capture exit code
codegen_output=$($CODEGEN_BINARY_PATH --generate types,skip-prune -o $OUTPUT_DIR/api_interfaces.gen.go -package interfaces $API_OPENAPI_SPEC 2>&1)

# check exit code of oapi-codegen command
if [[ $? -eq 0 ]]; then
    echo "Successfully generated API interfaces, placed them in: $OUTPUT_DIR/api_interfaces.gen.go"
else
    echo "Error generating interfaces:"
    echo "$codegen_output"
    exit 1
fi
