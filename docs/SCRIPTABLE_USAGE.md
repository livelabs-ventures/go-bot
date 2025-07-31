# Scriptable Usage Guide for Standup Bot

This guide explains how to use standup-bot programmatically from scripts, CI/CD pipelines, or other tools like Claude Code.

## Prerequisites

Before using standup-bot in scripts:

1. **Ensure the standup repository exists**:
   - The repository specified in your config must be created beforehand
   - You must have write access to this repository

2. **GitHub CLI must be authenticated**:
   ```bash
   gh auth status
   ```

3. **Standup-bot must be configured**:
   ```bash
   standup-bot --config
   ```

## JSON Input Format

The standup bot accepts JSON input with the following structure:

```json
{
  "yesterday": ["Task 1", "Task 2", "..."],
  "today": ["Task 1", "Task 2", "..."],
  "blockers": "Description of blockers or 'None'"
}
```

### Required Fields
- At least one of `yesterday` or `today` must contain entries
- If `blockers` is omitted or empty, it defaults to "None"

## Input Methods

### 1. Direct JSON String
```bash
standup-bot --json '{"yesterday": ["Fixed authentication bug", "Updated documentation"], "today": ["Write unit tests", "Code review"], "blockers": "None"}'
```

### 2. Read from stdin (Piping)
```bash
# From echo
echo '{"yesterday": ["Fixed bug"], "today": ["Write tests"], "blockers": "None"}' | standup-bot --json -

# From a file using cat
cat standup.json | standup-bot --json -

# From another command
generate-standup | standup-bot --json -
```

### 3. Read from File
```bash
# Directly specify file path
standup-bot --json standup.json

# With full path
standup-bot --json /path/to/standup.json
```

## JSON Output Format

When using `--output json`, the tool returns structured JSON output:

### Success Response
```json
{
  "success": true,
  "message": "Standup recorded successfully",
  "date": "2025-07-31",
  "user": "john.doe",
  "yesterday": ["Task 1", "Task 2"],
  "today": ["Task 3", "Task 4"],
  "blockers": "None",
  "file_path": "/path/to/repo/stand-ups/john.doe.md",
  "pr_number": "42",
  "pr_url": "https://github.com/org/repo/pull/42"
}
```

### Error Response
```json
{
  "success": false,
  "error": "Failed to parse JSON input: invalid character...",
  "date": "2025-07-31"
}
```

## Usage Examples

### Basic Scriptable Usage
```bash
# Submit standup and capture result
RESULT=$(standup-bot --json '{"yesterday": ["Completed API"], "today": ["Testing"], "blockers": "None"}' --output json)

# Check if successful
if [ $(echo $RESULT | jq -r '.success') = "true" ]; then
    PR_URL=$(echo $RESULT | jq -r '.pr_url')
    echo "Standup submitted! PR: $PR_URL"
else
    ERROR=$(echo $RESULT | jq -r '.error')
    echo "Failed: $ERROR"
fi
```

### Python Integration
```python
import json
import subprocess

# Prepare standup data
standup_data = {
    "yesterday": ["Implemented feature X", "Fixed bug Y"],
    "today": ["Write tests for feature X", "Start feature Z"],
    "blockers": "Waiting for API documentation"
}

# Submit standup
result = subprocess.run(
    ["standup-bot", "--json", json.dumps(standup_data), "--output", "json"],
    capture_output=True,
    text=True
)

# Parse response
response = json.loads(result.stdout)
if response["success"]:
    print(f"Standup submitted! PR: {response['pr_url']}")
else:
    print(f"Error: {response['error']}")
```

### Shell Script with File Input
```bash
#!/bin/bash

# Generate standup content
cat > /tmp/standup.json << EOF
{
  "yesterday": [
    "$(git log --oneline --since=yesterday --author=$(git config user.name) | head -5)"
  ],
  "today": ["Continue development", "Code reviews"],
  "blockers": "None"
}
EOF

# Submit standup
standup-bot --json /tmp/standup.json --output json

# Clean up
rm /tmp/standup.json
```

### CI/CD Pipeline Example
```yaml
# GitHub Actions example
- name: Submit Daily Standup
  run: |
    STANDUP_JSON=$(cat <<EOF
    {
      "yesterday": ["${{ steps.get-commits.outputs.commits }}"],
      "today": ["Continue CI/CD improvements"],
      "blockers": "None"
    }
    EOF
    )
    echo "$STANDUP_JSON" | standup-bot --json - --output json
```

## Direct Commit Mode

For workflows that don't require pull requests:

```bash
# Direct commit with JSON input
standup-bot --direct --json '{"yesterday": ["Task A"], "today": ["Task B"]}' --output json
```

## Error Handling

The tool provides clear error messages in JSON format when `--output json` is used:

```bash
# Invalid JSON
$ echo 'invalid json' | standup-bot --json - --output json
{
  "success": false,
  "error": "Failed to parse JSON input: invalid character 'i' looking for beginning of value",
  "date": "2025-07-31"
}

# Missing required fields
$ echo '{"blockers": "None"}' | standup-bot --json - --output json
{
  "success": false,
  "error": "Failed to parse JSON input: at least one of 'yesterday' or 'today' must have entries",
  "date": "2025-07-31"
}
```

## Best Practices

1. **Always validate JSON** before sending to standup-bot
2. **Use --output json** for scriptable workflows to ensure consistent parsing
3. **Check the success field** in the response before proceeding
4. **Handle errors gracefully** - the tool will return error details in the JSON response
5. **Use stdin (-) for dynamic content** generated by other tools
6. **Store templates** in JSON files for consistent standup formats

## Integration with LLMs/AI Tools

The JSON interface makes it easy for AI tools to generate and submit standups:

```bash
# Example: AI generates standup content
AI_STANDUP=$(ai-tool generate-standup --format json)
echo "$AI_STANDUP" | standup-bot --json - --output json
```