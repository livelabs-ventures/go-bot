# Standup Bot

A simple and efficient CLI tool for managing daily standup updates via GitHub. The bot streamlines team communication by collecting standup information and committing it to a shared repository, where GitHub-Slack integration broadcasts updates to your team channel.

## Features

- 🚀 **Simple CLI interface** - Quick daily standup entries with minimal friction
- 🔄 **GitHub integration** - Uses GitHub CLI (`gh`) for all Git operations
- 👥 **Team collaboration** - Shared daily branches for all team standups
- 📢 **Slack visibility** - Full standup content visible in Slack via PR descriptions
- 🎯 **Smart workflows** - Automatic branch management and PR creation
- 💾 **Local configuration** - Stores preferences for quick daily use
- 🛡️ **Error recovery** - Saves standups locally if push fails

## Prerequisites

- Go 1.21+ (for building from source)
- [GitHub CLI](https://cli.github.com/) (`gh`) installed and authenticated
- Git repository for storing standups
- GitHub-Slack integration configured for your repository

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/your-org/go-bot.git
cd go-bot

# Build the binary
make build

# Optional: Install to $GOPATH/bin
make install
```

### Binary Distribution

Download the pre-built binary for your platform from the releases page.

## Quick Start

### 1. Initial Setup

Run the bot for the first time to configure:

```bash
standup-bot
```

You'll be prompted for:
- **GitHub Repository**: The repo where standups will be stored (e.g., `org/standup-repo`)
- **Your Name**: Used for your standup file and commit attribution

Configuration is saved to `~/.standup-bot/config.json`.

### 2. Daily Standup

Run the bot each day to record your standup:

```bash
standup-bot
```

You'll answer three questions:
1. **What did you do yesterday?** (multi-line, empty line to finish)
2. **What will you do today?** (multi-line, empty line to finish)
3. **Any blockers?** (single line, can be empty)

### 3. Merge Daily Standups

At the end of the day, anyone can merge all standups:

```bash
standup-bot --merge
```

## Workflows

### Default: Pull Request Workflow

The default workflow creates a shared daily PR containing all team standups:

1. First person creates branch `standup/YYYY-MM-DD` and opens a PR
2. Subsequent team members add their standups to the same branch
3. PR description automatically updates with all standups
4. Single merge notification in Slack when PR is merged

**Benefits:**
- Full standup content visible in Slack
- One PR per day instead of one per person
- Reduced notification noise
- Clear daily boundaries

### Alternative: Direct Commit Workflow

For simpler setups, use direct commits with multi-line messages:

```bash
standup-bot --direct
```

This creates individual commits with the full standup in the commit message.

## Commands

| Command | Description |
|---------|-------------|
| `standup-bot` | Record your daily standup (uses PR workflow) |
| `standup-bot --direct` | Record standup using direct commit workflow |
| `standup-bot --merge` | Merge today's standup pull request |
| `standup-bot --config` | Reconfigure the bot (repository, name) |
| `standup-bot --help` | Show help information |

## File Structure

### Standup Repository

```
stand-ups/
├── alice.md      # Alice's standup history
├── bob.md        # Bob's standup history
└── charlie.md    # Charlie's standup history
```

### Individual Standup File

Each person's standups are appended to their markdown file:

```markdown
# Alice's Standups

## 2025-07-31

**Yesterday:**
- Completed user authentication API endpoints
- Fixed bug in password reset flow

**Today:**
- Start frontend integration for auth
- Write unit tests for auth endpoints

**Blockers:**
None

---

## 2025-07-30

**Yesterday:**
- Set up project structure
- Created database schema

**Today:**
- Work on authentication endpoints

**Blockers:**
Waiting for API design approval

---
```

## Configuration

### Config File Location

`~/.standup-bot/config.json`

### Config Format

```json
{
  "repository": "org/standup-repo",
  "name": "Alice",
  "localRepoPath": "~/.standup-bot/repo"
}
```

### Environment Variables

Currently, no environment variables are used. All configuration is file-based.

## Development

### Project Structure

```
go-bot/
├── cmd/standup-bot/      # Main application entry point
├── internal/cli/         # CLI implementation
├── pkg/                  # Public packages
│   ├── config/          # Configuration management
│   ├── git/             # Git operations wrapper
│   └── standup/         # Standup business logic
├── Makefile             # Build automation
├── go.mod               # Go module definition
└── README.md            # This file
```

### Building

```bash
# Build binary
make build

# Run tests
make test

# Check test coverage
make coverage

# Clean build artifacts
make clean
```

### Testing

The project includes comprehensive unit tests with mocked external dependencies:

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run tests for a specific package
go test ./pkg/git/...
```

## Troubleshooting

### Common Issues

**GitHub CLI not found**
```
Error: GitHub CLI not found. Please install it from https://cli.github.com/
```
Solution: Install GitHub CLI and ensure it's in your PATH.

**Not authenticated**
```
Error: not authenticated with GitHub. Please run 'gh auth login'
```
Solution: Run `gh auth login` to authenticate with GitHub.

**Repository not found**
```
Error: repository not found at ~/.standup-bot/repo. Please run 'standup-bot --config' to set up
```
Solution: Run `standup-bot --config` to reconfigure.

**Push failed**
If pushing fails, your standup is saved locally to `/tmp/standup-{name}-{date}.txt`. You can manually add it later or retry when network is available.

### Reset Configuration

To start fresh:

```bash
rm -rf ~/.standup-bot
standup-bot --config
```

## Best Practices

1. **Run daily**: Make it part of your morning routine
2. **Be concise**: Bullet points work well
3. **Merge regularly**: Designate someone to merge at day's end
4. **Review together**: Use merged PRs for team standup meetings
5. **Keep history**: The markdown files serve as a searchable archive

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with [Cobra](https://github.com/spf13/cobra) for CLI management
- Integrates with [GitHub CLI](https://cli.github.com/) for Git operations
- Designed for teams using GitHub-Slack integration