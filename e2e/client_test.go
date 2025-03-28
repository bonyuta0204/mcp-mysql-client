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
	// Setup test environment
	err := setupDB()
	require.NoError(t, err)

	// Initialize client
	client, err := client.NewStdioMCPClient("../bin/mcp-mysql-client", []string{})
	require.NoError(t, err)

	// Ensure cleanup
	defer client.Close()

	ctx := context.Background()

	// Test client initialization and capabilities
	testClientInitialization(t, ctx, client)

	// Test database connection and basic operations
	testDatabaseConnection(t, ctx, client)

	// Test table operations
	testTableOperations(t, ctx, client)

	// Test query operations
	testQueryOperations(t, ctx, client)

	// Test special data types (NULL, integer, float)
	testSpecialDataTypes(t, ctx, client)
}

// testClientInitialization tests the initialization of the client and listing of available tools
func testClientInitialization(t *testing.T, ctx context.Context, client *client.StdioMCPClient) {
	// Test initialize
	res, err := client.Initialize(ctx, mcp.InitializeRequest{})
	require.NoError(t, err)
	logJsonResponse(t, res)

	// Test list tools
	listToolsRes, err := client.ListTools(ctx, mcp.ListToolsRequest{})
	require.NoError(t, err)
	logJsonResponse(t, listToolsRes)

	// Verify expected tools are available
	assertToolsAvailable(t, listToolsRes)
}

// assertToolsAvailable verifies that all expected tools are available in the response
func assertToolsAvailable(t *testing.T, listToolsRes *mcp.ListToolsResult) bool {
	expectedTools := []string{"connect", "list_databases", "list_tables", "describe_table", "query"}

	for _, tool := range expectedTools {
		found := false
		for _, responseTool := range listToolsRes.Tools {
			if responseTool.Name == tool {
				found = true
				break
			}
		}
		assert.True(t, found, "Tool %s should be available", tool)
	}
	return true
}

// testDatabaseConnection tests connecting to the database and listing databases
func testDatabaseConnection(t *testing.T, ctx context.Context, client *client.StdioMCPClient) {
	// Test connect to MySQL
	connectRes := connectToDatabase(t, ctx, client, "testdb")
	assert.Equal(t, "Successfully connected to MySQL at localhost:3306", connectRes.Content[0].(mcp.TextContent).Text)

	// Test list databases
	listDbRes := callTool(t, ctx, client, "list_databases", map[string]interface{}{})
	logToolCallResult(t, listDbRes)
	assert.Contains(t, listDbRes.Content[0].(mcp.TextContent).Text, "testdb")
	assert.Contains(t, listDbRes.Content[0].(mcp.TextContent).Text, "seconddb")
}

// testTableOperations tests operations related to tables
func testTableOperations(t *testing.T, ctx context.Context, client *client.StdioMCPClient) {
	// Test list tables in default database
	listTablesRes := callTool(t, ctx, client, "list_tables", map[string]interface{}{})
	logToolCallResult(t, listTablesRes)
	assert.Contains(t, listTablesRes.Content[0].(mcp.TextContent).Text, "users")
	assert.Contains(t, listTablesRes.Content[0].(mcp.TextContent).Text, "products")

	// Test list tables in second database
	listTablesSecondDbRes := callTool(t, ctx, client, "list_tables", map[string]interface{}{
		"database": "seconddb",
	})
	logToolCallResult(t, listTablesSecondDbRes)
	assert.Contains(t, listTablesSecondDbRes.Content[0].(mcp.TextContent).Text, "items")

	// Switch back to testdb explicitly before continuing
	connectToDatabase(t, ctx, client, "testdb")

	// Test describe table
	describeTableRes := callTool(t, ctx, client, "describe_table", map[string]interface{}{
		"table": "users",
	})
	logToolCallResult(t, describeTableRes)
	assert.Contains(t, describeTableRes.Content[0].(mcp.TextContent).Text, "id")
	assert.Contains(t, describeTableRes.Content[0].(mcp.TextContent).Text, "username")
	assert.Contains(t, describeTableRes.Content[0].(mcp.TextContent).Text, "email")
}

// testQueryOperations tests SQL query operations
func testQueryOperations(t *testing.T, ctx context.Context, client *client.StdioMCPClient) {
	// Test query - select users
	queryUsersRes := callTool(t, ctx, client, "query", map[string]interface{}{
		"sql": "SELECT * FROM users",
	})
	logToolCallResult(t, queryUsersRes)
	assert.Contains(t, queryUsersRes.Content[0].(mcp.TextContent).Text, "username")
	assert.Contains(t, queryUsersRes.Content[0].(mcp.TextContent).Text, "email")
	assert.Contains(t, queryUsersRes.Content[0].(mcp.TextContent).Text, "user1@example.com")
	assert.Contains(t, queryUsersRes.Content[0].(mcp.TextContent).Text, "user2@example.com")
	assert.Contains(t, queryUsersRes.Content[0].(mcp.TextContent).Text, "user3@example.com")

	// Test query - select products
	queryProductsRes := callTool(t, ctx, client, "query", map[string]interface{}{
		"sql": "SELECT name, price FROM products ORDER BY price ASC",
	})
	logToolCallResult(t, queryProductsRes)
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "name")
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "price")
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "Product A")
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "19.99")
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "Product C")
	assert.Contains(t, queryProductsRes.Content[0].(mcp.TextContent).Text, "39.99")

	// Test query - count query
	queryCountRes := callTool(t, ctx, client, "query", map[string]interface{}{
		"sql": "SELECT COUNT(*) as count FROM users",
	})
	logToolCallResult(t, queryCountRes)
	assert.Contains(t, queryCountRes.Content[0].(mcp.TextContent).Text, "count")
	assert.Contains(t, queryCountRes.Content[0].(mcp.TextContent).Text, "3")
}

// testSpecialDataTypes tests special data types (NULL, integer, float)
func testSpecialDataTypes(t *testing.T, ctx context.Context, client *client.StdioMCPClient) {
	// Test query - select from data_types table with NULL values
	queryNullValuesRes := callTool(t, ctx, client, "query", map[string]interface{}{
		"sql": "SELECT id, nullable_column FROM data_types WHERE nullable_column IS NULL",
	})
	logToolCallResult(t, queryNullValuesRes)
	// Verify NULL values are properly displayed
	assert.Contains(t, queryNullValuesRes.Content[0].(mcp.TextContent).Text, "NULL")
	// Should have 2 rows with NULL values
	assert.Contains(t, queryNullValuesRes.Content[0].(mcp.TextContent).Text, "1")
	assert.Contains(t, queryNullValuesRes.Content[0].(mcp.TextContent).Text, "4")

	// Test query - select integer values
	queryIntegerValuesRes := callTool(t, ctx, client, "query", map[string]interface{}{
		"sql": "SELECT id, integer_column FROM data_types ORDER BY integer_column",
	})
	logToolCallResult(t, queryIntegerValuesRes)
	// Verify integer values are properly displayed
	assert.Contains(t, queryIntegerValuesRes.Content[0].(mcp.TextContent).Text, "-100")
	assert.Contains(t, queryIntegerValuesRes.Content[0].(mcp.TextContent).Text, "0")
	assert.Contains(t, queryIntegerValuesRes.Content[0].(mcp.TextContent).Text, "42")
	assert.Contains(t, queryIntegerValuesRes.Content[0].(mcp.TextContent).Text, "9999")

	// Test query - select float values
	queryFloatValuesRes := callTool(t, ctx, client, "query", map[string]interface{}{
		"sql": "SELECT id, float_column, decimal_column FROM data_types ORDER BY float_column",
	})
	logToolCallResult(t, queryFloatValuesRes)
	// Verify float values are properly displayed
	assert.Contains(t, queryFloatValuesRes.Content[0].(mcp.TextContent).Text, "-1.5")
	assert.Contains(t, queryFloatValuesRes.Content[0].(mcp.TextContent).Text, "0")
	assert.Contains(t, queryFloatValuesRes.Content[0].(mcp.TextContent).Text, "3.14")
	assert.Contains(t, queryFloatValuesRes.Content[0].(mcp.TextContent).Text, "123.456")
	// Verify decimal values are properly displayed
	assert.Contains(t, queryFloatValuesRes.Content[0].(mcp.TextContent).Text, "-199.99")
	assert.Contains(t, queryFloatValuesRes.Content[0].(mcp.TextContent).Text, "0.00")
	assert.Contains(t, queryFloatValuesRes.Content[0].(mcp.TextContent).Text, "99.99")
	assert.Contains(t, queryFloatValuesRes.Content[0].(mcp.TextContent).Text, "1000.01")

	// Test query - direct NULL, integer and float literals
	queryLiteralsRes := callTool(t, ctx, client, "query", map[string]interface{}{
		"sql": "SELECT NULL as null_value, 123 as int_value, 45.67 as float_value",
	})
	logToolCallResult(t, queryLiteralsRes)
	assert.Contains(t, queryLiteralsRes.Content[0].(mcp.TextContent).Text, "null_value")
	assert.Contains(t, queryLiteralsRes.Content[0].(mcp.TextContent).Text, "NULL")
	assert.Contains(t, queryLiteralsRes.Content[0].(mcp.TextContent).Text, "int_value")
	assert.Contains(t, queryLiteralsRes.Content[0].(mcp.TextContent).Text, "123")
	assert.Contains(t, queryLiteralsRes.Content[0].(mcp.TextContent).Text, "float_value")
	assert.Contains(t, queryLiteralsRes.Content[0].(mcp.TextContent).Text, "45.67")
}

// connectToDatabase is a helper function to connect to a specific database
func connectToDatabase(t *testing.T, ctx context.Context, client *client.StdioMCPClient, database string) *mcp.CallToolResult {
	connectRes := callTool(t, ctx, client, "connect", map[string]interface{}{
		"host":     "localhost",
		"port":     3306,
		"username": "root",
		"password": "test",
		"database": database,
	})
	logToolCallResult(t, connectRes)
	return connectRes
}

// callTool is a helper function to call a tool with the given parameters
func callTool(t *testing.T, ctx context.Context, client *client.StdioMCPClient, toolName string, arguments map[string]interface{}) *mcp.CallToolResult {
	request := &mcp.CallToolRequest{}
	request.Params.Name = toolName
	request.Params.Arguments = arguments

	res, err := client.CallTool(ctx, *request)
	require.NoError(t, err)
	return res
}

// logJsonResponse logs a JSON response in a visually distinct format
func logJsonResponse(t *testing.T, res interface{}) {
	// Marshal with indentation for better readability
	jsonRes, err := json.MarshalIndent(res, "", "  ")
	require.NoError(t, err)

	// Use a more visually distinct format for the output
	t.Logf("\n┌─────────────── RESPONSE ───────────────┐\n%s\n└──────────────────────────────────────┘", string(jsonRes))
}

func logToolCallResult(t *testing.T, res *mcp.CallToolResult) {
	// Use a more visually distinct format for the output
	textContent := res.Content[0].(mcp.TextContent).Text
	t.Logf("\n┌─────────────── RESPONSE ───────────────┐\n%s\n└──────────────────────────────────────┘", textContent)
}

// setupDB initializes the test database with sample data
func setupDB() error {
	// Load environment variables
	err := godotenv.Load("../.env")
	if err != nil {
		return err
	}

	// Get database connection parameters from environment
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	port := os.Getenv("DB_PORT")
	password := os.Getenv("DB_PASSWORD")

	// Configure MySQL connection
	c := mysql.Config{
		User:                 user,
		Passwd:               password,
		Addr:                 fmt.Sprintf("%s:%s", host, port),
		Net:                  "tcp",
		DBName:               "mysql", // Connect to the system database first
		AllowNativePasswords: true,
		Params: map[string]string{
			"multiStatements": "true",
		},
	}

	dsn := c.FormatDSN()

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer db.Close() // Ensure connection is closed when function returns

	// Test the connection
	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Read and execute initialization SQL script
	return executeInitScript(db)
}

// executeInitScript reads and executes the SQL initialization script
func executeInitScript(db *sql.DB) error {
	// Read SQL initialization file
	initSQL, err := os.ReadFile("./testdata/init.sql")
	if err != nil {
		return fmt.Errorf("failed to read init.sql file: %w", err)
	}

	// Execute the SQL statements
	_, err = db.Exec(string(initSQL))
	if err != nil {
		return fmt.Errorf("failed to execute init SQL: %w", err)
	}

	return nil
}
