package e2e

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/joho/godotenv"
)

// TestJSONRPCInterface tests the JSON-RPC interface of the MySQL client
func TestJSONRPCInterface(t *testing.T) {
	err := setupDB()
	require.NoError(t, err)
	client, err := client.NewStdioMCPClient("../bin/mcp-mysql-client", []string{})
	require.NoError(t, err)

	// Ensure cleanup
	defer func() {
		client.Close()
	}()

	ctx := context.Background()

	// Test initialize
	res, err := client.Initialize(ctx, mcp.InitializeRequest{})
	require.NoError(t, err)
	logJsonResponse(t, res)

	// Test list tools
	listToolsRes, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	require.NoError(t, err)
	logJsonResponse(t, listToolsRes)

	// Test connect to MySQL
	connectRequest := &mcp.CallToolRequest{}
	connectRequest.Params.Name = "connect"
	connectRequest.Params.Arguments = map[string]interface{}{
		"host":     "localhost",
		"port":     3306,
		"username": "root",
		"password": "test",
		"database": "testdb",
	}

	connectRes, err := client.CallTool(ctx, *connectRequest)
	require.NoError(t, err)
	logJsonResponse(t, connectRes)
	assert.Equal(t, "Successfully connected to MySQL at localhost:3306", connectRes.Content[0].(mcp.TextContent).Text)

	// Test list databases
	listDbRequest := &mcp.CallToolRequest{}
	listDbRequest.Params.Name = "list_databases"
	listDbRequest.Params.Arguments = map[string]interface{}{}

	listDbRes, err := client.CallTool(ctx, *listDbRequest)
	require.NoError(t, err)
	logJsonResponse(t, listDbRes)
	assert.Contains(t, listDbRes.Content[0].(mcp.TextContent).Text, "testdb")
	assert.Contains(t, listDbRes.Content[0].(mcp.TextContent).Text, "seconddb")

	// Test list tables in default database
	listTablesRequest := &mcp.CallToolRequest{}
	listTablesRequest.Params.Name = "list_tables"
	listTablesRequest.Params.Arguments = map[string]interface{}{}

	listTablesRes, err := client.CallTool(ctx, *listTablesRequest)
	require.NoError(t, err)
	logJsonResponse(t, listTablesRes)
	assert.Contains(t, listTablesRes.Content[0].(mcp.TextContent).Text, "users")
	assert.Contains(t, listTablesRes.Content[0].(mcp.TextContent).Text, "products")

	// Test list tables in second database
	listTablesSecondDbRequest := &mcp.CallToolRequest{}
	listTablesSecondDbRequest.Params.Name = "list_tables"
	listTablesSecondDbRequest.Params.Arguments = map[string]interface{}{
		"database": "seconddb",
	}

	listTablesSecondDbRes, err := client.CallTool(ctx, *listTablesSecondDbRequest)
	require.NoError(t, err)
	logJsonResponse(t, listTablesSecondDbRes)
	assert.Contains(t, listTablesSecondDbRes.Content[0].(mcp.TextContent).Text, "items")

	// Switch back to testdb explicitly before continuing
	switchDbRequest := &mcp.CallToolRequest{}
	switchDbRequest.Params.Name = "connect"
	switchDbRequest.Params.Arguments = map[string]interface{}{
		"host":     "localhost",
		"port":     3306,
		"username": "root",
		"password": "test",
		"database": "testdb",
	}

	switchDbRes, err := client.CallTool(ctx, *switchDbRequest)
	require.NoError(t, err)
	logJsonResponse(t, switchDbRes)

	// Test describe table
	describeTableRequest := &mcp.CallToolRequest{}
	describeTableRequest.Params.Name = "describe_table"
	describeTableRequest.Params.Arguments = map[string]interface{}{
		"table": "users",
	}

	describeTableRes, err := client.CallTool(ctx, *describeTableRequest)
	require.NoError(t, err)
	logJsonResponse(t, describeTableRes)
	assert.Contains(t, describeTableRes.Content[0].(mcp.TextContent).Text, "id")
	assert.Contains(t, describeTableRes.Content[0].(mcp.TextContent).Text, "username")
	assert.Contains(t, describeTableRes.Content[0].(mcp.TextContent).Text, "email")

	// Test query - select users
	queryUsersRequest := &mcp.CallToolRequest{}
	queryUsersRequest.Params.Name = "query"
	queryUsersRequest.Params.Arguments = map[string]interface{}{
		"sql": "SELECT * FROM users",
	}

	queryUsersRes, err := client.CallTool(ctx, *queryUsersRequest)
	require.NoError(t, err)
	logJsonResponse(t, queryUsersRes)
	assert.Contains(t, queryUsersRes.Content[0].(mcp.TextContent).Text, "username")
	assert.Contains(t, queryUsersRes.Content[0].(mcp.TextContent).Text, "email")
	assert.Contains(t, queryUsersRes.Content[0].(mcp.TextContent).Text, "user1@example.com")
	assert.Contains(t, queryUsersRes.Content[0].(mcp.TextContent).Text, "user2@example.com")
	assert.Contains(t, queryUsersRes.Content[0].(mcp.TextContent).Text, "user3@example.com")

	// Test query - select products
	queryProductsRequest := &mcp.CallToolRequest{}
	queryProductsRequest.Params.Name = "query"
	queryProductsRequest.Params.Arguments = map[string]interface{}{
		"sql": "SELECT name, price FROM products ORDER BY price ASC",
	}

	queryProductsRes, err := client.CallTool(ctx, *queryProductsRequest)
	require.NoError(t, err)
	logJsonResponse(t, queryProductsRes)
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "name")
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "price")
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "Product A")
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "19.99")
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "Product B")
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "29.99")
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "Product C")
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "39.99")

	// Test query - count query
	queryCountRequest := &mcp.CallToolRequest{}
	queryCountRequest.Params.Name = "query"
	queryCountRequest.Params.Arguments = map[string]interface{}{
		"sql": "SELECT COUNT(*) as count FROM users",
	}

	queryCountRes, err := client.CallTool(ctx, *queryCountRequest)
	require.NoError(t, err)
	logJsonResponse(t, queryCountRes)
	assert.Contains(t, queryCountRes.Content[0].(mcp.TextContent).Text, "count")
	assert.Contains(t, queryCountRes.Content[0].(mcp.TextContent).Text, "3")
}

func logJsonResponse(t *testing.T, res interface{}) {
	// Marshal with indentation for better readability
	jsonRes, err := json.MarshalIndent(res, "", "  ")
	require.NoError(t, err)

	// Use a more visually distinct format for the output
	t.Logf("\n┌─────────────── RESPONSE ───────────────┐\n%s\n└──────────────────────────────────────┘", string(jsonRes))
}

func setupDB() error {
	err := godotenv.Load("../.env")

	if err != nil {
		return err
	}

	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	port := os.Getenv("DB_PORT")
	password := os.Getenv("DB_PASSWORD")

	c := mysql.Config{
		User:   user,
		Passwd: password,
		Addr:   fmt.Sprintf("%s:%s", host, port),
		Net:    "tcp",
	}

	dsn := c.FormatDSN()

	db, err := sql.Open("mysql", dsn)

	initSQL, err := os.ReadFile("./testdata/init.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(initSQL))

	if err != nil {
		return err
	}

	return nil

}
