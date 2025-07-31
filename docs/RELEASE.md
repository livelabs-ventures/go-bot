# Release Process

This document describes the release process for standup-bot.

## Prerequisites

1. Install [goreleaser](https://goreleaser.com/install/):
   ```bash
   brew install goreleaser
   ```

2. Set up GitHub Personal Access Token:
   - Create a token with `repo` scope at https://github.com/settings/tokens
   - Export it: `export GITHUB_TOKEN=your_token_here`

3. Set up Homebrew tap (one-time setup):
   - Create a new repository named `homebrew-tap` in the `livelabs-ventures` organization
   - Initialize it with a README and a `Formula` directory

## Release Steps

### Automated Release (Recommended)

1. Run the release script:
   ```bash
   make release
   ```

2. Follow the prompts to:
   - Enter the new version number (e.g., `v0.2.0`)
   - Optionally do a dry run first
   - Push the tag to trigger GitHub Actions

3. GitHub Actions will automatically:
   - Build binaries for all platforms
   - Create a GitHub release with artifacts
   - Update the Homebrew formula (if tap is configured)

### Manual Release

1. Run tests:
   ```bash
   make test
   ```

2. Create and push a tag:
   ```bash
   git tag -a v0.2.0 -m "Release v0.2.0"
   git push origin v0.2.0
   ```

3. Run goreleaser locally:
   ```bash
   goreleaser release --clean
   ```

## Testing

### Test Release Locally

Before creating an actual release, you can test the process:

```bash
# Dry run with goreleaser
make release-snapshot

# Test Homebrew formula locally
make test-homebrew
```

### Verify Release

After releasing:

1. Check the GitHub releases page
2. Test installation via Homebrew:
   ```bash
   brew update
   brew install livelabs-ventures/tap/standup-bot
   standup-bot --version
   ```

## Homebrew Formula Updates

The Homebrew formula is automatically updated by goreleaser when:
- A new tag is pushed
- GitHub Actions workflow runs successfully
- The `homebrew-tap` repository is properly configured

To manually update the formula:
1. Download the release artifacts
2. Calculate SHA256 checksums
3. Update the formula in `homebrew-tap/Formula/standup-bot.rb`
4. Test the formula locally
5. Push to the tap repository

## Troubleshooting

### Goreleaser Fails

- Check configuration: `goreleaser check`
- Ensure GITHUB_TOKEN is set correctly
- Verify you have push access to the repository

### Homebrew Installation Fails

- Ensure the tap is added: `brew tap livelabs-ventures/tap`
- Update Homebrew: `brew update`
- Check formula syntax: `brew audit --strict standup-bot`

### GitHub Actions Fails

- Check the Actions tab for error logs
- Ensure secrets are properly configured
- Verify the workflow file syntax

## Version Numbering

We follow [Semantic Versioning](https://semver.org/):
- MAJOR version for incompatible API changes
- MINOR version for new functionality in a backward compatible manner
- PATCH version for backward compatible bug fixes

Examples:
- `v0.1.0` - Initial release
- `v0.1.1` - Bug fixes
- `v0.2.0` - New features
- `v1.0.0` - First stable release