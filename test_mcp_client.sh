#!/bin/bash

# This script tests the MCP server by simulating a client interaction

# Function to send JSON-RPC request and wait for response
send_request() {
    echo "$1"
    sleep 0.1  # Small delay to ensure server processes the request
}

# Start the MCP server in the background
./standup-bot mcp-server 2>mcp_server.log &
MCP_PID=$!

# Give the server time to start
sleep 1

# Clean up on exit
trap "kill $MCP_PID 2>/dev/null; rm -f mcp_server.log" EXIT

# Test sequence
{
    # 1. Initialize
    send_request '{"jsonrpc": "2.0", "method": "initialize", "params": {"protocolVersion": "2024-11-05", "capabilities": {"roots": {"listChanged": true}}, "clientInfo": {"name": "test-client", "version": "1.0.0"}}, "id": 1}'
    
    # 2. Get tool list
    send_request '{"jsonrpc": "2.0", "method": "tools/list", "params": {}, "id": 2}'
    
    # 3. Call get_standup_status
    send_request '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "get_standup_status", "arguments": {}}, "id": 3}'
    
    # 4. Submit a test standup
    send_request '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "submit_standup", "arguments": {"yesterday": ["Debugged MCP server"], "today": ["Test MCP implementation"], "blockers": "None", "direct": true}}, "id": 4}'
    
} | nc -q 1 localhost 0 2>/dev/null || echo "Note: 'nc' test failed. This is expected if netcat is not available."

# Show server logs
echo "=== Server logs ==="
cat mcp_server.log

# Kill the server
kill $MCP_PID 2>/dev/null