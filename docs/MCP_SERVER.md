# MCP Server Documentation

The standup-bot includes a Model Context Protocol (MCP) server that allows AI assistants to interact with standup functionality programmatically.

## Running the MCP Server

```bash
standup-bot mcp-server
```

The server uses stdio transport for communication with MCP clients.

## Available Tools

### 1. submit_standup

Submit a daily standup with yesterday's accomplishments, today's plans, and blockers.

**Parameters:**
- `yesterday` (array of strings, required): List of tasks completed yesterday
- `today` (array of strings, required): List of tasks planned for today
- `blockers` (string, optional): Any blockers or impediments (default: "None")
- `direct` (boolean, optional): Use direct commit workflow instead of PR workflow (default: false)

**Example:**
```json
{
  "yesterday": ["Fixed bug in authentication", "Reviewed 3 PRs"],
  "today": ["Implement MCP server", "Write documentation"],
  "blockers": "None",
  "direct": false
}
```

### 2. create_standup_pr

Create or manage a pull request with standup entries for the day.

**Parameters:**
- `merge` (boolean, optional): Whether to merge the PR after creation (default: false)

**Example:**
```json
{
  "merge": true
}
```

### 3. get_standup_status

Check if today's standup has been completed.

**Parameters:** None

**Response:** Returns the status (complete/incomplete) and additional information about existing PRs.

## Integration with AI Assistants

### Claude Desktop Configuration

To use with Claude Desktop, add the following to your Claude configuration:

```json
{
  "mcpServers": {
    "standup-bot": {
      "command": "/path/to/standup-bot",
      "args": ["mcp-server"]
    }
  }
}
```

### Usage Example

Once configured, you can interact with the standup-bot through natural language:

```
"Submit my standup: Yesterday I worked on the authentication bug and reviewed PRs. Today I'll implement the MCP server. No blockers."
```

The AI assistant will use the appropriate MCP tools to submit your standup.

## Technical Details

- Transport: stdio (standard input/output)
- Protocol: Model Context Protocol
- Server Name: standup-bot-mcp
- Version: 1.0.0

## Prerequisites

Before using the MCP server, ensure:
1. The standup-bot is configured (`standup-bot --config`)
2. GitHub CLI is installed and authenticated
3. You have access to the standup repository

## Error Handling

The MCP server provides detailed error messages for common issues:
- Configuration not found
- GitHub authentication failures
- Repository access issues
- Network connectivity problems

All errors are returned in a structured format that AI assistants can parse and communicate effectively.