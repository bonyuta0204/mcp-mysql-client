package e2e

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJSONRPCInterface tests the JSON-RPC interface of the MySQL client
func TestJSONRPCInterface(t *testing.T) {
	client, err := client.NewStdioMCPClient("../bin/mcp-mysql-client", []string{})
	require.NoError(t, err)

	// Ensure cleanup
	defer func() {
		client.Close()
	}()

	ctx := context.Background()

	res, err := client.Initialize(ctx, mcp.InitializeRequest{})

	require.NoError(t, err)
	logJsonResponse(t, res)

	listToolsRes, err := client.ListTools(ctx, mcp.ListToolsRequest{})

	require.NoError(t, err)
	logJsonResponse(t, listToolsRes)

	connectRequest := &mcp.CallToolRequest{}
	connectRequest.Params.Name = "connect"
	connectRequest.Params.Arguments = map[string]interface{}{
		"host":     "localhost",
		"port":     3306,
		"username": "root",
		"password": "test",
	}

	connectRes, err := client.CallTool(ctx, *connectRequest)
	require.NoError(t, err)
	logJsonResponse(t, connectRes)

	assert.Equal(t, connectRes.Content[0].(mcp.TextContent).Text, "Successfully connected to MySQL at localhost:3306")

}

func logJsonResponse(t *testing.T, res interface{}) {
	// Marshal with indentation for better readability
	jsonRes, err := json.MarshalIndent(res, "", "  ")
	require.NoError(t, err)

	// Use a more visually distinct format for the output
	t.Logf("\n┌─────────────── RESPONSE ───────────────┐\n%s\n└──────────────────────────────────────┘", string(jsonRes))
}
