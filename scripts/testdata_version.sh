#!/usr/bin/env bash

# Check if a version argument is provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <version_argument>"
    exit 1
fi

# Check if the git status is clean in the ./testdata directory
if ! git diff --quiet -- ./testdata; then
    echo "Error: Your git status is dirty in the ./testdata directory. Please commit or stash your changes before running this script."
    exit 1
fi

VERSION_ARGUMENT=$1

# Find and replace the string in all .tf files in the ./testdata/ directory
find ./testdata/ -type f -name "*.tf" -exec sed -i.bak "/coder = {/{
    # Capture the entire coder block
    :a
    N
    /}/!ba
    # Remove any existing version lines
    s|[ ]*version[ ]*=[^\\n]*\\n||g
    # Replace the coder block with the new format
    s|    coder = {\n.*}|    coder = {\n      source = \"coder/coder\"\n      version = \"$VERSION_ARGUMENT\"\n    }|
}" {} +

# Remove backup files
find ./testdata/ -type f -name "*.bak" -exec rm {} +

echo "Replacement complete."