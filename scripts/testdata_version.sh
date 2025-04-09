#!/usr/bin/env bash

# Check if a version argument is provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <version_argument>"
    exit 1
fi

VERSION_ARGUMENT=$1

# Find and replace the string in all .tf files in the ./testdata/ directory
find ./testdata/ -type f -name "*.tf" -exec sed "s|    coder = {|    coder = {\n      version = \"$VERSION_ARGUMENT\"|" {} +

echo "Replacement complete. Backup files have been removed."