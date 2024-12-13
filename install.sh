#!/bin/bash

# check if go is installed
if ! command -v go &> /dev/null; then
    echo "go is not installed"
    exit 1
fi

# store original directory
ORIGINAL_DIR=$(pwd)

# remove existing installation if present
if [ -f ~/.swan/swan ]; then
    rm ~/.swan/swan
    echo "removing existing swan installation"
fi

# create .swan directory in home
mkdir -p ~/.swan

# install the binary
GOBIN=~/.swan go install github.com/rAlexander89/swan@latest

# ensure paths are set in zshrc
SHELL_RC="$HOME/.zshrc"

if ! grep -q 'export PATH="$HOME/.swan:$PATH"' "$SHELL_RC"; then
    echo 'export PATH="$HOME/.swan:$PATH"' >> "$SHELL_RC"
fi

# add path for current session
PATH="$HOME/.swan:$PATH"

# change back to original directory
cd "$ORIGINAL_DIR"

# add swan to current session
export PATH="$HOME/.swan:$PATH"

# make swan executable
chmod +x ~/.swan/swan

echo "swan CLI installed successfully"
echo "reloading shell..."

# reload shell with new configuration
exec zsh -l 

