#!/usr/bin/env bash

# Script to help set up a Homebrew tap repository
# This is a one-time setup process

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

GITHUB_ORG="livelabs-ventures"
TAP_REPO="homebrew-tap"

print_info "Homebrew Tap Setup for $GITHUB_ORG/$TAP_REPO"
print_info "================================================"

# Check if gh is installed
if ! command -v gh &> /dev/null; then
    print_error "GitHub CLI (gh) is not installed. Please install it first:"
    echo "  brew install gh"
    exit 1
fi

# Check if user is authenticated
if ! gh auth status &> /dev/null; then
    print_error "You are not authenticated with GitHub CLI. Please run:"
    echo "  gh auth login"
    exit 1
fi

print_info "This script will help you create a Homebrew tap repository."
print_warn "Make sure you have admin access to the $GITHUB_ORG organization."

read -p "Do you want to create the tap repository? (y/n): " CREATE_REPO
if [[ "$CREATE_REPO" != "y" && "$CREATE_REPO" != "Y" ]]; then
    print_info "Skipping repository creation."
else
    print_info "Creating repository $GITHUB_ORG/$TAP_REPO..."
    
    # Create the repository
    if gh repo create "$GITHUB_ORG/$TAP_REPO" --public --description "Homebrew tap for $GITHUB_ORG tools" --confirm; then
        print_info "Repository created successfully!"
    else
        print_warn "Repository might already exist or creation failed."
    fi
fi

# Clone the repository
TEMP_DIR=$(mktemp -d)
print_info "Cloning repository to $TEMP_DIR..."
cd "$TEMP_DIR"

if ! gh repo clone "$GITHUB_ORG/$TAP_REPO"; then
    print_error "Failed to clone repository. Make sure it exists and you have access."
    exit 1
fi

cd "$TAP_REPO"

# Create the basic structure
print_info "Setting up tap structure..."

# Create Formula directory
mkdir -p Formula

# Create README
cat > README.md << 'EOF'
# Homebrew Tap for LiveLabs Ventures

This is a [Homebrew](https://brew.sh) tap for LiveLabs Ventures tools.

## Installation

```bash
brew tap livelabs-ventures/tap
```

## Available Formulas

### standup-bot

A CLI tool for managing daily standup updates via GitHub.

```bash
brew install standup-bot
```

## Development

This tap is automatically updated when new releases are created in the respective repositories.

## License

See individual formula files for license information.
EOF

# Create .gitignore
cat > .gitignore << 'EOF'
*.bottle.*
.DS_Store
EOF

# Create a placeholder formula (will be replaced by goreleaser)
cat > Formula/.gitkeep << 'EOF'
# This directory contains Homebrew formulas
EOF

# Commit and push
print_info "Committing initial structure..."
git add .
git commit -m "Initial tap structure"
git push origin main

print_info "Homebrew tap setup complete!"
print_info ""
print_info "Next steps:"
print_info "1. Update .goreleaser.yml in your project to use this tap"
print_info "2. Create a GitHub Personal Access Token with 'repo' scope"
print_info "3. Add the token as HOMEBREW_TAP_GITHUB_TOKEN in your GitHub Actions secrets"
print_info "4. When you create a new release, the formula will be automatically updated"
print_info ""
print_info "Users can now install your tools with:"
print_info "  brew tap $GITHUB_ORG/tap"
print_info "  brew install standup-bot"

# Clean up
cd /
rm -rf "$TEMP_DIR"