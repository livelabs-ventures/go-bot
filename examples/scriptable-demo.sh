#!/bin/bash
# Demo script showing scriptable features of standup-bot

echo "=== Standup Bot Scriptable Features Demo ==="
echo

# 1. Get suggestions from git history
echo "1. Getting suggestions from recent commits..."
echo "Command: standup-bot --suggest --output json"
echo
./standup-bot --suggest --output json 2>/dev/null | jq '.' || echo "No commits found or not in a git repository"
echo

# 2. Create standup with JSON input
echo "2. Creating standup with JSON input..."
echo 'Command: standup-bot --json '"'"'{"yesterday":["Implemented feature X","Fixed bug Y"],"today":["Write tests","Review PRs"],"blockers":"None"}'"'"' --output json'
echo
STANDUP_JSON='{
  "yesterday": ["Implemented feature X", "Fixed bug Y"],
  "today": ["Write tests", "Review PRs"],
  "blockers": "None"
}'
echo "$STANDUP_JSON" | jq '.'
echo

# 3. Pipe JSON through standup-bot
echo "3. Example of piping JSON through standup-bot..."
echo 'Command: echo "$STANDUP_JSON" | standup-bot --json "$(cat)" --output json'
echo

# 4. Error handling example
echo "4. Error handling with malformed JSON..."
echo 'Command: standup-bot --json '"'"'{invalid json}'"'"' --output json'
./standup-bot --json '{invalid json}' --output json 2>/dev/null | jq '.' || echo "Error output would appear here"
echo

# 5. Integration example
echo "5. Example integration script:"
cat << 'EOF'
#!/bin/bash
# Auto-standup from git commits
SUGGESTIONS=$(standup-bot --suggest --output json)
if [ $? -eq 0 ]; then
  YESTERDAY=$(echo "$SUGGESTIONS" | jq -r '.yesterday')
  if [ "$YESTERDAY" != "[]" ]; then
    STANDUP_DATA=$(echo "$SUGGESTIONS" | jq '{
      yesterday: .yesterday,
      today: ["Continue work from yesterday", "Code reviews"],
      blockers: "None"
    }')
    standup-bot --json "$STANDUP_DATA" --output json
  fi
fi
EOF

echo
echo "=== Demo Complete ==="
echo
echo "Note: This demo shows the commands without actually executing them"
echo "      (to avoid modifying your standup repository)"