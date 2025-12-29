#!/bin/bash
# EchoNext CLI Installer for Linux and macOS

set -e

INSTALL_DIR="$HOME/.local/bin"
BINARY_NAME="echonext"

echo "EchoNext CLI Installer"
echo "======================"
echo ""

# Check Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is required but not installed."
    echo "Please install Go from https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}')
echo "Found $GO_VERSION"

# Create install directory
mkdir -p "$INSTALL_DIR"

# Install using go install
echo "Installing EchoNext CLI..."
go install github.com/abdussamadbello/echonext/cmd/echonext-cli@latest

# Get GOBIN path (respects GOBIN env var, falls back to GOPATH/bin)
GOBIN=$(go env GOBIN)
if [ -z "$GOBIN" ]; then
    GOBIN=$(go env GOPATH)/bin
fi

# Copy to install directory with renamed binary
if [ -f "$GOBIN/echonext-cli" ]; then
    cp "$GOBIN/echonext-cli" "$INSTALL_DIR/$BINARY_NAME"
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    echo "Installed $BINARY_NAME to $INSTALL_DIR"
else
    echo "Error: Binary not found at $GOBIN/echonext-cli"
    exit 1
fi

# Check if install directory is in PATH
if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    echo ""
    echo "NOTE: $INSTALL_DIR is not in your PATH."
    echo ""
    echo "Add this line to your shell config file:"
    echo ""

    # Detect shell and suggest appropriate config file
    SHELL_NAME=$(basename "$SHELL")
    case "$SHELL_NAME" in
        bash)
            echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.bashrc"
            echo ""
            echo "Then run: source ~/.bashrc"
            ;;
        zsh)
            echo "  echo 'export PATH=\"\$PATH:$INSTALL_DIR\"' >> ~/.zshrc"
            echo ""
            echo "Then run: source ~/.zshrc"
            ;;
        fish)
            echo "  echo 'set -gx PATH \$PATH $INSTALL_DIR' >> ~/.config/fish/config.fish"
            echo ""
            echo "Then restart your terminal or run: source ~/.config/fish/config.fish"
            ;;
        *)
            echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
            echo ""
            echo "Add this to your shell's config file and restart your terminal."
            ;;
    esac
else
    echo ""
    echo "You can now use 'echonext' from anywhere!"
fi

echo ""
echo "Verify installation:"
echo "  echonext version"
echo ""
echo "Get started:"
echo "  echonext init myproject"
