#!/usr/bin/env bash

# Script to test Homebrew formula locally
# This is useful for testing before publishing to a tap

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "This script must be run from the project root directory"
    exit 1
fi

# Build the binary
print_info "Building standup-bot binary..."
make build

# Create a temporary directory for testing
TEMP_DIR=$(mktemp -d)
print_info "Created temporary directory: $TEMP_DIR"

# Create a tarball (simulating a release artifact)
TAR_NAME="standup-bot_$(uname -s)_$(uname -m).tar.gz"
print_info "Creating tarball: $TAR_NAME"
tar -czf "$TEMP_DIR/$TAR_NAME" standup-bot README.md

# Calculate SHA256
SHA256=$(shasum -a 256 "$TEMP_DIR/$TAR_NAME" | awk '{print $1}')
print_info "SHA256: $SHA256"

# Copy and update the formula
cp homebrew/standup-bot.rb "$TEMP_DIR/standup-bot.rb"

# Update the formula with local path and correct SHA
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s|https://github.com/.*\.tar\.gz|file://$TEMP_DIR/$TAR_NAME|g" "$TEMP_DIR/standup-bot.rb"
    sed -i '' "s|PLACEHOLDER_SHA256_[A-Z_]*|$SHA256|g" "$TEMP_DIR/standup-bot.rb"
else
    # Linux
    sed -i "s|https://github.com/.*\.tar\.gz|file://$TEMP_DIR/$TAR_NAME|g" "$TEMP_DIR/standup-bot.rb"
    sed -i "s|PLACEHOLDER_SHA256_[A-Z_]*|$SHA256|g" "$TEMP_DIR/standup-bot.rb"
fi

print_info "Testing Homebrew formula..."

# Test installation
print_info "Installing formula..."
brew install --build-from-source "$TEMP_DIR/standup-bot.rb"

# Test the installed binary
print_info "Testing installed binary..."
if standup-bot --version; then
    print_info "Binary works correctly!"
else
    print_error "Binary test failed"
    exit 1
fi

# Run brew test
print_info "Running brew test..."
brew test standup-bot

# Uninstall
print_info "Cleaning up..."
brew uninstall standup-bot

# Clean up temporary directory
rm -rf "$TEMP_DIR"

print_info "Local Homebrew formula test completed successfully!"
print_info "You can now proceed with publishing to your Homebrew tap."