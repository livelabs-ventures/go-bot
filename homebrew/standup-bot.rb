# typed: false
# frozen_string_literal: true

# This is a template Homebrew formula for standup-bot
# It will be automatically updated by goreleaser when creating releases
class StandupBot < Formula
  desc "CLI tool for managing daily standup updates via GitHub"
  homepage "https://github.com/livelabs-ventures/go-bot"
  version "0.1.0"
  license "MIT"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/livelabs-ventures/go-bot/releases/download/v0.1.0/standup-bot_Darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_ARM64"

      def install
        bin.install "standup-bot"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/livelabs-ventures/go-bot/releases/download/v0.1.0/standup-bot_Darwin_x86_64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_DARWIN_X86_64"

      def install
        bin.install "standup-bot"
      end
    end
  end

  on_linux do
    if Hardware::CPU.arm? && Hardware::CPU.is_64_bit?
      url "https://github.com/livelabs-ventures/go-bot/releases/download/v0.1.0/standup-bot_Linux_arm64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_ARM64"

      def install
        bin.install "standup-bot"
      end
    end
    if Hardware::CPU.intel?
      url "https://github.com/livelabs-ventures/go-bot/releases/download/v0.1.0/standup-bot_Linux_x86_64.tar.gz"
      sha256 "PLACEHOLDER_SHA256_LINUX_X86_64"

      def install
        bin.install "standup-bot"
      end
    end
  end

  # GitHub CLI is an optional dependency
  depends_on "gh" => :optional

  def caveats
    <<~EOS
      The standup-bot requires GitHub CLI (gh) to be installed and authenticated.
      
      If you haven't installed GitHub CLI:
        brew install gh
      
      Then authenticate:
        gh auth login
    EOS
  end

  test do
    system "#{bin}/standup-bot", "--version"
  end
end