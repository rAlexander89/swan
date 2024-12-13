#!/bin/bash

# check if go is installed
if ! command -v go &> /dev/null; then
    echo "go is not installed"
    exit 1
fi

# remove existing installation if present
if [ -f ~/.swan/swan ]; then
    rm ~/.swan/swan
    echo "removing existing swan installation"
fi

# create .swan directory in home
mkdir -p ~/.swan

# store original directory
ORIGINAL_DIR=$(pwd)

# create temp dir but store its path
TEMP_DIR=$(mktemp -d)

# do installation work in temp dir
cd "$TEMP_DIR"
GOBIN=~/.swan go install github.com/rAlexander89/swan@latest

# go back to original directory
cd "$ORIGINAL_DIR"

# ensure paths are set in zshrc
SHELL_RC="$HOME/.zshrc"

if ! grep -q 'export PATH="$HOME/.swan:$PATH"' "$SHELL_RC"; then
    echo 'export PATH="$HOME/.swan:$PATH"' >> "$SHELL_RC"
fi

# export path for current session
export PATH="$HOME/.swan:$PATH"

echo "swan CLI installed successfully"
echo "try running: swan"

# cleanup temp directory
rm -rf "$TEMP_DIR"
