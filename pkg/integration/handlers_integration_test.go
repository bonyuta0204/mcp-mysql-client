package integration

import (
	"context"
	"testing"

	"github.com/bonyuta0204/mcp-mysql-client/pkg/handlers"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConnectHandlerIntegration tests the connect handler with a real MySQL database
func TestConnectHandlerIntegration(t *testing.T) {

	// Get test config
	config := GetTestConfig()

	// Create datastore
	ds := SetupTestDatastore(t)
	defer CleanupTestDatastore(t, ds)

	// Create request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"host":     config.Host,
		"port":     config.Port,
		"username": config.Username,
		"password": config.Password,
		"database": config.Database,
	}

	// Call handler
	result, err := handlers.ConnectHandler(context.Background(), request)

	// Verify results
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "Successfully connected to MySQL")
}

// TestQueryHandlerIntegration tests the query handler with a real MySQL database
func TestQueryHandlerIntegration(t *testing.T) {

	// Create datastore
	ds := SetupTestDatastore(t)
	defer CleanupTestDatastore(t, ds)

	// Test cases
	tests := []struct {
		name           string
		sql            string
		expectedOutput []string
	}{
		{
			name: "select users",
			sql:  "SELECT * FROM users",
			expectedOutput: []string{
				"username",
				"email",
				"user1@example.com",
				"user2@example.com",
				"user3@example.com",
			},
		},
		{
			name: "select products",
			sql:  "SELECT name, price FROM products ORDER BY price ASC",
			expectedOutput: []string{
				"name",
				"price",
				"Product A",
				"19.99",
				"Product B",
				"29.99",
				"Product C",
				"39.99",
			},
		},
		{
			name: "count query",
			sql:  "SELECT COUNT(*) as count FROM users",
			expectedOutput: []string{
				"count",
				"3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			request := mcp.CallToolRequest{}
			request.Params.Arguments = map[string]interface{}{
				"sql": tt.sql,
			}

			// Call handler
			result, err := handlers.QueryHandler(context.Background(), request)

			// Verify results
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Len(t, result.Content, 1)

			// Check that all expected outputs are in the result
			for _, expected := range tt.expectedOutput {
				assert.Contains(t, result.Content[0].(mcp.TextContent).Text, expected)
			}
		})
	}
}

// TestListDatabasesHandlerIntegration tests the list databases handler with a real MySQL database
func TestListDatabasesHandlerIntegration(t *testing.T) {
	// Create datastore
	ds := SetupTestDatastore(t)
	defer CleanupTestDatastore(t, ds)

	// Create request
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{}

	// Call handler
	result, err := handlers.ListDatabasesHandler(context.Background(), request)

	// Verify results
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	// Check that both test databases are in the result
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "testdb")
	assert.Contains(t, result.Content[0].(mcp.TextContent).Text, "seconddb")
}

// TestListTablesHandlerIntegration tests the list tables handler with a real MySQL database
func TestListTablesHandlerIntegration(t *testing.T) {
	// Create datastore
	ds := SetupTestDatastore(t)
	defer CleanupTestDatastore(t, ds)

	// Test cases
	tests := []struct {
		name           string
		database       string
		expectedTables []string
	}{
		{
			name:           "default database",
			database:       "",
			expectedTables: []string{"users", "products"},
		},
		{
			name:           "second database",
			database:       "seconddb",
			expectedTables: []string{"items"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			request := mcp.CallToolRequest{}
			args := map[string]interface{}{}
			if tt.database != "" {
				args["database"] = tt.database
			}
			request.Params.Arguments = args

			// Call handler
			result, err := handlers.ListTablesHandler(context.Background(), request)

			// Verify results
			require.NoError(t, err)
			require.NotNil(t, result)
			require.Len(t, result.Content, 1)

			// Check that all expected tables are in the result
			for _, expectedTable := range tt.expectedTables {
				assert.Contains(t, result.Content[0].(mcp.TextContent).Text, expectedTable)
			}
		})
	}
}

// TestDescribeTableHandlerIntegration tests the describe table handler with a real MySQL database
func TestDescribeTableHandlerIntegration(t *testing.T) {
	// Skip if MySQL is not available
	SkipIfNoMySQL(t)

	// Create datastore
	ds := SetupTestDatastore(t)
	defer CleanupTestDatastore(t, ds)

	// First, explicitly select the testdb database
	ctx := context.Background()
	_, err := ds.ExecContext(ctx, "USE testdb")
	require.NoError(t, err, "Failed to switch to testdb database")

	// Create request with explicit database prefix for the table
	request := mcp.CallToolRequest{}
	request.Params.Arguments = map[string]interface{}{
		"table": "testdb.users", // Explicitly specify the database
	}

	// Call handler
	result, err := handlers.DescribeTableHandler(context.Background(), request)

	// Verify results
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Len(t, result.Content, 1)

	// Check that the result contains expected column information
	output := result.Content[0].(mcp.TextContent).Text
	assert.Contains(t, output, "Field")
	assert.Contains(t, output, "Type")
	assert.Contains(t, output, "Null")
	assert.Contains(t, output, "Key")
	assert.Contains(t, output, "id")
	assert.Contains(t, output, "username")
	assert.Contains(t, output, "email")
	assert.Contains(t, output, "created_at")
	assert.Contains(t, output, "varchar")
	assert.Contains(t, output, "PRI") // Primary key
}
