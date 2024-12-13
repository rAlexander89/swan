#!/bin/bash

# check if go is installed
if ! command -v go &> /dev/null; then
    echo "go is not installed"
    exit 1
fi

# create .swan directory in home
mkdir -p ~/.swan

# get the code and build
cd $(mktemp -d)
go install github.com/rAlexander89/swan@latest

# ensure GOBIN is in PATH for zsh
SHELL_RC="$HOME/.zshrc"
if ! grep -q 'export PATH="$HOME/.swan:$PATH"' "$SHELL_RC"; then
    echo 'export PATH="$HOME/.swan:$PATH"' >> "$SHELL_RC"
fi

if ! grep -q 'export PATH="$HOME/go/bin:$PATH"' "$SHELL_RC"; then
    echo 'export PATH="$HOME/go/bin:$PATH"' >> "$SHELL_RC"
fi

# move binary to .swan
mv ~/go/bin/swan ~/.swan/

echo "swan CLI installed successfully"
echo "restart your terminal or run: source $SHELL_RC"


