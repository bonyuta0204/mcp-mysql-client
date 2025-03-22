package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Global database connection pool
var db *sql.DB

func main() {
	// Create MCP server
	s := server.NewMCPServer(
		"MySQL Client",
		"1.0.0",
		server.WithLogging(),
	)

	// Add connection tool
	connectTool := mcp.NewTool("connect",
		mcp.WithDescription("Connect to a MySQL database"),
		mcp.WithString("host",
			mcp.Required(),
			mcp.Description("MySQL host address"),
		),
		mcp.WithString("port",
			mcp.Description("MySQL port"),
			mcp.DefaultString("3306"),
		),
		mcp.WithString("username",
			mcp.Required(),
			mcp.Description("MySQL username"),
		),
		mcp.WithString("password",
			mcp.Required(),
			mcp.Description("MySQL password"),
		),
		mcp.WithString("database",
			mcp.Description("MySQL database name"),
			mcp.DefaultString(""),
		),
	)

	// Add query tool
	queryTool := mcp.NewTool("query",
		mcp.WithDescription("Execute a SQL query"),
		mcp.WithString("sql",
			mcp.Required(),
			mcp.Description("SQL query to execute"),
		),
	)

	// Add list databases tool
	listDatabasesTool := mcp.NewTool("list_databases",
		mcp.WithDescription("List all databases"),
	)

	// Add list tables tool
	listTablesTool := mcp.NewTool("list_tables",
		mcp.WithDescription("List all tables in the current database"),
		mcp.WithString("database",
			mcp.Description("Database name (optional, uses current connection if not specified)"),
		),
	)

	// Add describe table tool
	describeTableTool := mcp.NewTool("describe_table",
		mcp.WithDescription("Describe a table structure"),
		mcp.WithString("table",
			mcp.Required(),
			mcp.Description("Table name"),
		),
	)

	// Add tool handlers
	s.AddTool(connectTool, connectHandler)
	s.AddTool(queryTool, queryHandler)
	s.AddTool(listDatabasesTool, listDatabasesHandler)
	s.AddTool(listTablesTool, listTablesHandler)
	s.AddTool(describeTableTool, describeTableHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

// connectHandler establishes a connection to the MySQL database
func connectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract connection parameters
	host := request.Params.Arguments["host"].(string)
	port := request.Params.Arguments["port"].(string)
	username := request.Params.Arguments["username"].(string)
	password := request.Params.Arguments["password"].(string)
	database := request.Params.Arguments["database"].(string)

	// Close existing connection if any
	if db != nil {
		db.Close()
	}

	// Create DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, database)

	// Open database connection
	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %v", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Minute * 5)

	// Test connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully connected to MySQL at %s:%s", host, port)), nil
}

// queryHandler executes a SQL query
func queryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if db == nil {
		return nil, errors.New("not connected to a database, use connect tool first")
	}

	// Extract query
	sql := request.Params.Arguments["sql"].(string)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Execute query
	rows, err := db.QueryContext(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %v", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get column names: %v", err)
	}

	// Check if this is a SELECT query (has columns)
	if len(columns) == 0 {
		// This is likely a non-SELECT query (INSERT, UPDATE, DELETE, etc.)
		return mcp.NewToolResultText("Query executed successfully (no results to display)"), nil
	}

	// Prepare values for scanning
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	// Build result table
	var result string

	// Add header row
	result += "| " + columns[0]
	for i := 1; i < len(columns); i++ {
		result += " | " + columns[i]
	}
	result += " |\n"

	// Add separator row
	result += "|---"
	for i := 1; i < len(columns); i++ {
		result += "|---"
	}
	result += "|\n"

	// Add data rows
	rowCount := 0
	for rows.Next() {
		// Scan the row into values
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		// Convert values to strings and add to result
		result += "| "
		for i, val := range values {
			if i > 0 {
				result += " | "
			}

			// Handle NULL values and convert to string
			if val == nil {
				result += "NULL"
			} else {
				switch v := val.(type) {
				case []byte:
					result += string(v)
				default:
					result += fmt.Sprintf("%v", v)
				}
			}
		}
		result += " |\n"
		rowCount++
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	// Add summary
	result += fmt.Sprintf("\n%d row(s) returned", rowCount)

	return mcp.NewToolResultText(result), nil
}

// listDatabasesHandler lists all databases
func listDatabasesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if db == nil {
		return nil, errors.New("not connected to a database, use connect tool first")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Execute query to list databases
	rows, err := db.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %v", err)
	}
	defer rows.Close()

	// Build result
	var result string
	result += "| Database |\n"
	result += "|----------|\n"

	var dbName string
	count := 0
	for rows.Next() {
		err := rows.Scan(&dbName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		result += fmt.Sprintf("| %s |\n", dbName)
		count++
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	// Add summary
	result += fmt.Sprintf("\n%d database(s) found", count)

	return mcp.NewToolResultText(result), nil
}

// listTablesHandler lists all tables in a database
func listTablesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if db == nil {
		return nil, errors.New("not connected to a database, use connect tool first")
	}

	// Extract database name if provided
	database, ok := request.Params.Arguments["database"].(string)
	if ok && database != "" {
		// Use the specified database
		_, err := db.ExecContext(ctx, fmt.Sprintf("USE %s", database))
		if err != nil {
			return nil, fmt.Errorf("failed to switch to database %s: %v", database, err)
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Execute query to list tables
	rows, err := db.QueryContext(ctx, "SHOW TABLES")
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %v", err)
	}
	defer rows.Close()

	// Build result
	var result string
	result += "| Table |\n"
	result += "|-------|\n"

	var tableName string
	count := 0
	for rows.Next() {
		err := rows.Scan(&tableName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		result += fmt.Sprintf("| %s |\n", tableName)
		count++
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	// Add summary
	result += fmt.Sprintf("\n%d table(s) found", count)

	return mcp.NewToolResultText(result), nil
}

// describeTableHandler describes a table structure
func describeTableHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	if db == nil {
		return nil, errors.New("not connected to a database, use connect tool first")
	}

	// Extract table name
	table := request.Params.Arguments["table"].(string)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Execute query to describe table
	rows, err := db.QueryContext(ctx, fmt.Sprintf("DESCRIBE %s", table))
	if err != nil {
		return nil, fmt.Errorf("failed to describe table %s: %v", table, err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get column names: %v", err)
	}

	// Build result table
	var result string

	// Add header row
	result += "| " + columns[0]
	for i := 1; i < len(columns); i++ {
		result += " | " + columns[i]
	}
	result += " |\n"

	// Add separator row
	result += "|---"
	for i := 1; i < len(columns); i++ {
		result += "|---"
	}
	result += "|\n"

	// Prepare values for scanning
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	// Add data rows
	rowCount := 0
	for rows.Next() {
		// Scan the row into values
		err := rows.Scan(valuePtrs...)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		// Convert values to strings and add to result
		result += "| "
		for i, val := range values {
			if i > 0 {
				result += " | "
			}

			// Handle NULL values and convert to string
			if val == nil {
				result += "NULL"
			} else {
				switch v := val.(type) {
				case []byte:
					result += string(v)
				default:
					result += fmt.Sprintf("%v", v)
				}
			}
		}
		result += " |\n"
		rowCount++
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %v", err)
	}

	// Add summary
	result += fmt.Sprintf("\n%d column(s) in table %s", rowCount, table)

	return mcp.NewToolResultText(result), nil
}
