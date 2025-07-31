#!/bin/bash
# Demo script showing scriptable features of standup-bot

echo "=== Standup Bot Scriptable Features Demo ==="
echo

# 1. Create standup with JSON input
echo "1. Creating standup with JSON input..."
echo 'Command: standup-bot --json '"'"'{"yesterday":["Implemented feature X","Fixed bug Y"],"today":["Write tests","Review PRs"],"blockers":"None"}'"'"' --output json'
echo
STANDUP_JSON='{
  "yesterday": ["Implemented feature X", "Fixed bug Y"],
  "today": ["Write tests", "Review PRs"],
  "blockers": "None"
}'
echo "$STANDUP_JSON" | jq '.'
echo

# 2. Pipe JSON through standup-bot
echo "2. Example of piping JSON through standup-bot..."
echo 'Command: echo "$STANDUP_JSON" | standup-bot --json "$(cat)" --output json'
echo

# 3. Error handling example
echo "3. Error handling with malformed JSON..."
echo 'Command: standup-bot --json '"'"'{invalid json}'"'"' --output json'
./standup-bot --json '{invalid json}' --output json 2>/dev/null | jq '.' || echo "Error output would appear here"
echo

# 4. Integration example
echo "4. Example integration script:"
cat << 'EOF'
#!/bin/bash
# Auto-standup with pre-defined content
STANDUP_DATA='{
  "yesterday": ["Completed feature X", "Fixed critical bug"],
  "today": ["Continue work from yesterday", "Code reviews"],
  "blockers": "None"
}'
standup-bot --json "$STANDUP_DATA" --output json
EOF

echo
echo "=== Demo Complete ==="
echo
echo "Note: This demo shows the commands without actually executing them"
echo "      (to avoid modifying your standup repository)"