# Standup Bot Specification

## Overview

A simple CLI tool written in Go that facilitates daily standup updates via GitHub. The bot collects standup information from team members and commits it to a shared repository, where GitHub-Slack integration broadcasts updates to the team channel.

## Key Features

- **Simple CLI interface** - Minimal interaction required for daily standups
- **GitHub CLI integration** - Uses `gh` for all Git operations
- **Local configuration** - Stores user preferences locally
- **Automated Git workflow** - Handles pull, rebase, commit, and push automatically
- **Slack-friendly commits** - Structured commit messages for clean Slack notifications

## Architecture

### Technology Stack
- **Language**: Go (for easy binary distribution)
- **Dependencies**: GitHub CLI (`gh`)
- **Storage**: Local filesystem for config and repo cache

### File Structure
```
~/.standup-bot/
   config.json          # User configuration
   repo/               # Local clone of standup repository
       stand-ups/
           alice.md
           bob.md
           charlie.md
```

## CLI Interface

### Initial Setup
```bash
standup-bot
# First run prompts for:
# - GitHub repository (e.g., org/standup-repo)
# - Your name (e.g., Alice)
```

### Reconfiguration
```bash
standup-bot --config
# Re-runs the configuration prompts
```

### Daily Standup
```bash
standup-bot
# If already configured, jumps straight to standup questions
```

## Configuration Management

### Config File Location
`~/.standup-bot/config.json`

### Config Structure
```json
{
  "repository": "org/standup-repo",
  "name": "Alice",
  "localRepoPath": "~/.standup-bot/repo"
}
```

### Configuration Flow
1. Check if config exists
2. If not, prompt for repository and name
3. Clone/update repository to local cache
4. Save configuration

## Standup Workflow

### Questions
1. **What did you do yesterday?**
   - Multi-line input supported
   - Empty line to finish

2. **What will you do today?**
   - Multi-line input supported
   - Empty line to finish

3. **Any blockers?**
   - Single line input
   - Can be left empty

### Git Operations

#### Option A: Direct Commit (Simple, Limited Slack Display)
1. **Pull latest changes**: `gh repo sync --force`
2. **Update standup file**: Append new entry to `stand-ups/{name}.md`
3. **Commit changes**: With structured commit message
4. **Push to remote**: `git push`

#### Option B: Pull Request Workflow (Full Slack Display)
1. **Pull latest changes**: `gh repo sync --force`
2. **Create feature branch**: `git checkout -b standup/{name}-{date}`
3. **Update standup file**: Append new entry to `stand-ups/{name}.md`
4. **Commit changes**: Simple commit message
5. **Push branch**: `git push -u origin standup/{name}-{date}`
6. **Create PR**: `gh pr create` with full standup in description
7. **Auto-merge PR**: `gh pr merge --auto --squash`

### Commit Message Format (Direct Commit)
```
[Standup] Alice - 2024-01-31

Yesterday: Completed API endpoints
Today: Working on frontend integration
Blockers: None
```

### Pull Request Format (PR Workflow)
**Title**: `[Standup] Alice - 2024-01-31`

**Body**:
```markdown
## Daily Standup - Alice

**Date**: 2024-01-31

### Yesterday
- Completed user authentication API endpoints
- Fixed bug in password reset flow

### Today
- Start frontend integration for auth
- Write unit tests for auth endpoints

### Blockers
None
```

## Standup File Format

### Individual Standup File (`stand-ups/alice.md`)
```markdown
# Alice's Standups

## 2024-01-31

**Yesterday:**
- Completed user authentication API endpoints
- Fixed bug in password reset flow

**Today:**
- Start frontend integration for auth
- Write unit tests for auth endpoints

**Blockers:**
None

---

## 2024-01-30

**Yesterday:**
- Set up project structure
- Created database schema

**Today:**
- Work on authentication endpoints

**Blockers:**
Waiting for API design approval

---
```

## Error Handling

### Common Scenarios
1. **No GitHub CLI**: Prompt to install `gh` with instructions
2. **Auth issues**: Guide user to run `gh auth login`
3. **Network failures**: Retry with exponential backoff
4. **Merge conflicts**: Attempt auto-resolution, fail gracefully with instructions
5. **Missing repo**: Clear error message with setup instructions

### User-Friendly Messages
- Clear, actionable error messages
- Suggest fixes for common issues
- Never lose user's standup input (save to temp file if push fails)

## Implementation Considerations

### Binary Distribution
- Single static binary for each platform
- No runtime dependencies except `gh`
- Version checking and auto-update capability

### Performance
- Cache repo locally to minimize network calls
- Only pull/push when necessary
- Fast startup time (<100ms)

### Security
- Rely on GitHub CLI for authentication
- No credentials stored locally
- Respect system proxy settings

## Future Enhancements (Out of Scope v1)

- Team standup summary generation
- Slack direct integration (bypass GitHub)
- Web interface
- Mobile app
- Analytics and reporting
- Custom question templates
- Multiple repository support

## Success Criteria

1. **Ease of Use**: Complete standup in under 60 seconds
2. **Reliability**: 99%+ success rate for Git operations
3. **Adoption**: Zero training required for new users
4. **Integration**: Seamless Slack notifications via existing GitHub integration

## Example Usage Flow

```bash
$ standup-bot
Welcome to Standup Bot!
GitHub Repository (e.g., org/standup-repo): myteam/standups
Your Name: Alice
Configuration saved!

What did you do yesterday?
> Completed user authentication API endpoints
> Fixed bug in password reset flow
>

What will you do today?
> Start frontend integration for auth
> Write unit tests for auth endpoints
>

Any blockers?
> None

Syncing repository...
Recording standup...
Pushing changes...
 Standup recorded successfully!
```

## Technical Requirements

### Minimum Viable Product (MVP)
- [ ] Go CLI application with single binary output
- [ ] Configuration management (create, read, update)
- [ ] Interactive standup questionnaire
- [ ] Git operations via GitHub CLI
- [ ] Markdown file generation/updating
- [ ] Structured commit messages
- [ ] Basic error handling

### Dependencies
- Go 1.21+
- GitHub CLI (user-installed)
- Git (via GitHub CLI)

### Platform Support
- macOS (Apple Silicon & Intel)
- Linux (amd64, arm64)
- Windows (amd64)