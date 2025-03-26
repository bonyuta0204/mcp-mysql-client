package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJSONRPCInterface tests the JSON-RPC interface of the MySQL client
// This is an end-to-end test that starts the binary and communicates with it
// through stdin/stdout, similar to how the e2e.sh script works but with Go's
// testing capabilities.
func TestJSONRPCInterface(t *testing.T) {
	// Start the server process
	cmd := exec.Command("../bin/mcp-mysql-client")
	
	// Create pipes for stdin and stdout
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	
	// Start the process
	err = cmd.Start()
	require.NoError(t, err)
	
	// Ensure cleanup
	defer func() {
		stdin.Close()
		cmd.Process.Kill()
		cmd.Wait()
	}()

	// Initialize request
	initializeReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"roots":    map[string]interface{}{"listChanged": true},
				"sampling": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "ExampleClient",
				"version": "1.0.0",
			},
		},
	}

	// Send initialize request
	sendRequest(t, stdin, initializeReq)

	// Read initialize response
	initResponse := readResponse(t, stdout)
	t.Logf("Initialize response: %v", initResponse)

	// Validate initialize response
	assert.Equal(t, float64(1), initResponse["id"])
	assert.Contains(t, initResponse, "result")

	// Send initialized notification
	notificationMsg := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	sendRequest(t, stdin, notificationMsg)

	// Send tools/list request
	toolsListReq := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/list",
		"params": map[string]interface{}{
			"cursor": "optional-cursor-value",
		},
	}
	sendRequest(t, stdin, toolsListReq)

	// Read tools/list response
	toolsResponse := readResponse(t, stdout)
	t.Logf("Tools list response: %v", toolsResponse)

	// Validate tools/list response
	assert.Equal(t, float64(1), toolsResponse["id"])
	assert.Contains(t, toolsResponse, "result")

	// Check if tools are present in the response
	result, ok := toolsResponse["result"].(map[string]interface{})
	assert.True(t, ok, "Result should be a map")
	
	tools, ok := result["tools"].([]interface{})
	assert.True(t, ok, "Tools should be an array")
	assert.NotEmpty(t, tools, "Tools array should not be empty")

	// Optional: Check for specific tools
	foundConnectTool := false
	foundQueryTool := false

	for _, tool := range tools {
		toolMap, ok := tool.(map[string]interface{})
		assert.True(t, ok, "Tool should be a map")

		if name, ok := toolMap["name"].(string); ok {
			if name == "connect" {
				foundConnectTool = true
			} else if name == "query" {
				foundQueryTool = true
			}
		}
	}

	assert.True(t, foundConnectTool, "Connect tool should be present")
	assert.True(t, foundQueryTool, "Query tool should be present")
}

// sendRequest marshals and sends a request to stdin
func sendRequest(t *testing.T, stdin io.WriteCloser, request map[string]interface{}) {
	reqBytes, err := json.Marshal(request)
	require.NoError(t, err)
	
	// Add newline to the request
	reqBytes = append(reqBytes, '\n')
	
	// Send request
	_, err = stdin.Write(reqBytes)
	require.NoError(t, err)
	t.Logf("Sent request: %s", reqBytes)
}

// readResponse reads and unmarshals a response from stdout
func readResponse(t *testing.T, stdout io.Reader) map[string]interface{} {
	// Read response with timeout
	buf := new(bytes.Buffer)
	readCh := make(chan struct{})
	errCh := make(chan error)
	
	go func() {
		_, err := io.Copy(buf, io.LimitReader(stdout, 4096))
		if err != nil && err != io.EOF {
			errCh <- err
			return
		}
		readCh <- struct{}{}
	}()
	
	// Wait for response with timeout
	select {
	case <-readCh:
		// Got response
	case err := <-errCh:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for response")
	}
	
	// Parse response
	responseBytes := buf.Bytes()
	t.Logf("Received response: %s", responseBytes)
	
	var response map[string]interface{}
	err := json.Unmarshal(responseBytes, &response)
	require.NoError(t, err)
	
	return response
}