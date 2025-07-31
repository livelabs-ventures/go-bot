#!/usr/bin/env bash

# Release script for standup-bot
# This script helps create a new release using goreleaser

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if goreleaser is installed
if ! command -v goreleaser &> /dev/null; then
    print_error "goreleaser is not installed. Please install it first:"
    echo "  brew install goreleaser"
    echo "  or visit: https://goreleaser.com/install/"
    exit 1
fi

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "This script must be run from the project root directory"
    exit 1
fi

# Get the current version from git tags
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
print_info "Current version: $CURRENT_VERSION"

# Prompt for new version
read -p "Enter new version (e.g., v0.2.0): " NEW_VERSION

# Validate version format
if [[ ! "$NEW_VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-.*)?$ ]]; then
    print_error "Invalid version format. Please use semantic versioning (e.g., v0.2.0)"
    exit 1
fi

# Check if tag already exists
if git rev-parse "$NEW_VERSION" >/dev/null 2>&1; then
    print_error "Tag $NEW_VERSION already exists"
    exit 1
fi

# Run tests
print_info "Running tests..."
if ! make test; then
    print_error "Tests failed. Please fix them before releasing."
    exit 1
fi

# Create git tag
print_info "Creating git tag $NEW_VERSION..."
git tag -a "$NEW_VERSION" -m "Release $NEW_VERSION"

# Ask if we should do a dry run first
read -p "Do you want to do a dry run first? (y/n): " DRY_RUN
if [[ "$DRY_RUN" == "y" || "$DRY_RUN" == "Y" ]]; then
    print_info "Running goreleaser in snapshot mode (dry run)..."
    goreleaser release --snapshot --clean --skip=publish
    print_info "Dry run complete. Check the dist/ directory for artifacts."
    
    read -p "Do you want to proceed with the actual release? (y/n): " PROCEED
    if [[ "$PROCEED" != "y" && "$PROCEED" != "Y" ]]; then
        print_warn "Release cancelled. Removing tag..."
        git tag -d "$NEW_VERSION"
        exit 0
    fi
fi

# Push the tag
print_info "Pushing tag to GitHub..."
git push origin "$NEW_VERSION"

print_info "Release tag pushed. GitHub Actions will now build and publish the release."
print_info "You can monitor the progress at: https://github.com/livelabs-ventures/go-bot/actions"

# Optional: Run goreleaser locally
read -p "Do you want to run goreleaser locally as well? (y/n): " LOCAL_RELEASE
if [[ "$LOCAL_RELEASE" == "y" || "$LOCAL_RELEASE" == "Y" ]]; then
    print_warn "Note: This will create a draft release on GitHub. Make sure you have GITHUB_TOKEN set."
    goreleaser release --clean
fi

print_info "Release process complete!"