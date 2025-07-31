#!/bin/bash
set -e

# Script to update Homebrew formula after a release
# This should be run after goreleaser creates a release

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}[INFO]${NC} Updating Homebrew formula for standup-bot"

# Check if version is provided
if [ -z "$1" ]; then
    echo -e "${RED}[ERROR]${NC} Please provide the version tag (e.g., v0.1.0)"
    echo "Usage: $0 <version-tag>"
    exit 1
fi

VERSION="$1"
VERSION_NO_V="${VERSION#v}" # Remove the 'v' prefix

# Clone the tap repository
TEMP_DIR=$(mktemp -d)
echo -e "${BLUE}[INFO]${NC} Cloning homebrew-tap repository..."
git clone git@github.com:livelabs-ventures/homebrew-tap.git "$TEMP_DIR"
cd "$TEMP_DIR"

# Create the formula
echo -e "${BLUE}[INFO]${NC} Creating formula..."
cat > Formula/standup-bot.rb << EOF
class StandupBot < Formula
  desc "CLI tool for managing daily standup updates via GitHub"
  homepage "https://github.com/livelabs-ventures/go-bot"
  version "${VERSION_NO_V}"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/livelabs-ventures/go-bot/releases/download/${VERSION}/standup-bot_Darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"
    elsif Hardware::CPU.intel?
      url "https://github.com/livelabs-ventures/go-bot/releases/download/${VERSION}/standup-bot_Darwin_x86_64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_X86_64"
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/livelabs-ventures/go-bot/releases/download/${VERSION}/standup-bot_Linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"
    elsif Hardware::CPU.intel?
      url "https://github.com/livelabs-ventures/go-bot/releases/download/${VERSION}/standup-bot_Linux_x86_64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_X86_64"
    end
  end

  depends_on "gh" => :optional

  def install
    bin.install "standup-bot"
  end

  def caveats
    <<~EOS
      The standup-bot requires GitHub CLI (gh) to be installed and authenticated.
      
      Install GitHub CLI:
        brew install gh
      
      Then authenticate:
        gh auth login
      
      Configure standup-bot:
        standup-bot --config
    EOS
  end

  test do
    assert_match "standup-bot version", shell_output("#{bin}/standup-bot --version")
  end
end
EOF

echo -e "${YELLOW}[WARN]${NC} Formula created with placeholder SHA256 values"
echo -e "${YELLOW}[WARN]${NC} You need to:"
echo -e "${YELLOW}[WARN]${NC} 1. Download the release artifacts from GitHub"
echo -e "${YELLOW}[WARN]${NC} 2. Calculate SHA256 for each platform:"
echo -e "${YELLOW}[WARN]${NC}    shasum -a 256 standup-bot_Darwin_arm64.tar.gz"
echo -e "${YELLOW}[WARN]${NC}    shasum -a 256 standup-bot_Darwin_x86_64.tar.gz"
echo -e "${YELLOW}[WARN]${NC}    shasum -a 256 standup-bot_Linux_arm64.tar.gz"
echo -e "${YELLOW}[WARN]${NC}    shasum -a 256 standup-bot_Linux_x86_64.tar.gz"
echo -e "${YELLOW}[WARN]${NC} 3. Replace the PLACEHOLDER_SHA256_* values in Formula/standup-bot.rb"
echo -e "${YELLOW}[WARN]${NC} 4. Commit and push the changes"
echo ""
echo -e "${BLUE}[INFO]${NC} Formula location: $TEMP_DIR/Formula/standup-bot.rb"
echo -e "${BLUE}[INFO]${NC} When ready, run:"
echo -e "${GREEN}  cd $TEMP_DIR${NC}"
echo -e "${GREEN}  # Edit Formula/standup-bot.rb to add SHA256 values${NC}"
echo -e "${GREEN}  git add Formula/standup-bot.rb${NC}"
echo -e "${GREEN}  git commit -m \"Update standup-bot to ${VERSION}\"${NC}"
echo -e "${GREEN}  git push origin main${NC}"