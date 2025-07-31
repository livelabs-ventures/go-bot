#!/bin/bash

# Example script showing how to use standup-bot programmatically

# Method 1: Direct JSON string
echo "Method 1: Direct JSON string"
standup-bot --json '{"yesterday": ["Completed task A"], "today": ["Work on task B"], "blockers": "None"}' --output json
echo ""

# Method 2: From a file
echo "Method 2: From a JSON file"
standup-bot --json standup-template.json --output json
echo ""

# Method 3: Using stdin with dynamic content
echo "Method 3: Dynamic content via stdin"
cat << EOF | standup-bot --json - --output json
{
  "yesterday": [
    "$(git log --oneline -1 --pretty=format:'%s')",
    "Attended team meeting"
  ],
  "today": [
    "Continue current work",
    "Code reviews"
  ],
  "blockers": "None"
}
EOF
echo ""

# Method 4: With error handling
echo "Method 4: With error handling"
RESULT=$(standup-bot --json '{"yesterday": ["Task A"], "today": ["Task B"]}' --output json 2>&1)

if echo "$RESULT" | grep -q '"success": true'; then
    PR_URL=$(echo "$RESULT" | grep -o '"pr_url": "[^"]*"' | cut -d'"' -f4)
    echo "✅ Standup submitted successfully!"
    echo "📎 Pull Request: $PR_URL"
else
    ERROR=$(echo "$RESULT" | grep -o '"error": "[^"]*"' | cut -d'"' -f4)
    echo "❌ Failed to submit standup: $ERROR"
fi