# Homebrew Automation Setup

This guide explains how to set up automated Homebrew formula updates when you create a new release.

## Prerequisites

1. You've already created the Homebrew tap repository (livelabs-ventures/homebrew-tap)
2. You have admin access to your GitHub repository

## Setup Steps

### 1. Create a GitHub Personal Access Token

1. Go to GitHub Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Give it a descriptive name like "Homebrew Tap Updates"
4. Select the following scopes:
   - `repo` (Full control of private repositories) - Required to push to your tap repository
5. Click "Generate token"
6. **Copy the token immediately** (you won't be able to see it again)

### 2. Add the Token to Repository Secrets

1. Go to your repository: https://github.com/livelabs-ventures/go-bot
2. Navigate to Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Name: `HOMEBREW_TAP_GITHUB_TOKEN`
5. Value: Paste the token you created
6. Click "Add secret"

## How It Works

Once configured, the automation works as follows:

1. When you push a version tag (e.g., `v0.2.0`), it triggers the release workflow
2. GoReleaser builds binaries for all platforms
3. GoReleaser automatically:
   - Creates a GitHub release with the binaries
   - Calculates SHA256 checksums for each binary
   - Generates a Homebrew formula with the correct URLs and checksums
   - Commits and pushes the formula to your homebrew-tap repository

## Creating a New Release

With automation set up, releasing is now simple:

```bash
# Tag and push a new version
git tag -a v0.2.0 -m "Release v0.2.0 with new features"
git push origin v0.2.0
```

That's it! The automation handles everything else.

## Verifying the Release

After pushing a tag:

1. Check the Actions tab: https://github.com/livelabs-ventures/go-bot/actions
2. Once the workflow completes, verify:
   - Release is created: https://github.com/livelabs-ventures/go-bot/releases
   - Formula is updated: https://github.com/livelabs-ventures/homebrew-tap/tree/main/Formula

## Users Can Install

Once the automation completes, users can immediately install:

```bash
brew tap livelabs-ventures/tap
brew install standup-bot
```

Or upgrade if already installed:

```bash
brew upgrade standup-bot
```

## Troubleshooting

### Token Permission Issues

If you see "Permission denied" errors in the workflow:
- Ensure the token has `repo` scope
- Check that the token hasn't expired
- Verify the secret name is exactly `HOMEBREW_TAP_GITHUB_TOKEN`

### Formula Not Updating

If the formula doesn't update:
- Check the GoReleaser logs in the GitHub Actions output
- Ensure the tap repository exists and is accessible
- Verify the token has write access to the tap repository

### Testing Locally

You can test the release process locally without pushing:

```bash
# Dry run (doesn't push anything)
goreleaser release --snapshot --clean --skip=publish
```