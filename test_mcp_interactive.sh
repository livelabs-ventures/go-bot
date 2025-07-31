#!/bin/bash

# Create a named pipe for bidirectional communication
PIPE_IN=$(mktemp -u)
PIPE_OUT=$(mktemp -u)
mkfifo "$PIPE_IN" "$PIPE_OUT"

# Clean up on exit
trap "rm -f $PIPE_IN $PIPE_OUT" EXIT

# Start the MCP server with input from pipe and output to pipe
./standup-bot mcp-server <"$PIPE_IN" >"$PIPE_OUT" 2>mcp_test.log &
MCP_PID=$!

# Give server time to start
sleep 0.5

# Function to send request and read response
send_and_read() {
    echo "Sending: $1" >&2
    echo "$1" >"$PIPE_IN"
    
    # Read response with timeout
    if timeout 2 head -n 1 "$PIPE_OUT"; then
        echo ""
    else
        echo "No response received" >&2
    fi
}

# Test the server
echo "=== Testing MCP Server ==="

# 1. Initialize
send_and_read '{"jsonrpc": "2.0", "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {"roots": {"listChanged": true}}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}, "id": 1}'

# 2. List tools
send_and_read '{"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": 2}'

# 3. Get standup status
send_and_read '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "get_standup_status", "arguments": {}}, "id": 3}'

# Kill the server
kill $MCP_PID 2>/dev/null

echo ""
echo "=== Server logs ==="
cat mcp_test.log

rm -f mcp_test.log