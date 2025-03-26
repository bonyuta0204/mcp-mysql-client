package handlers

import (
	"context"
	"fmt"
	"time"

	"github.com/bonyuta0204/mcp-mysql-client/pkg/datastore"
	"github.com/bonyuta0204/mcp-mysql-client/pkg/utils"
	"github.com/mark3labs/mcp-go/mcp"
)

// ConnectHandler establishes a connection to the MySQL database
func ConnectHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return withDatastoreInstance(connectHandler, ctx, request, datastore.DB)
}

func QueryHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return withDatastoreInstance(queryHandler, ctx, request, datastore.DB)
}

func ListDatabasesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return withDatastoreInstance(listDatabasesHandler, ctx, request, datastore.DB)
}

func ListTablesHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return withDatastoreInstance(listTablesHandler, ctx, request, datastore.DB)
}

func DescribeTableHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return withDatastoreInstance(describeTableHandler, ctx, request, datastore.DB)
}

func withDatastoreInstance(handler func(ctx context.Context, request mcp.CallToolRequest, ds datastore.DatastoreInterface) (*mcp.CallToolResult, error), ctx context.Context, request mcp.CallToolRequest, ds datastore.DatastoreInterface) (*mcp.CallToolResult, error) {
	return handler(ctx, request, ds)
}

func connectHandler(ctx context.Context, request mcp.CallToolRequest, ds datastore.DatastoreInterface) (*mcp.CallToolResult, error) {
	// Extract connection parameters
	host := request.Params.Arguments["host"].(string)
	port := request.Params.Arguments["port"].(string)
	username := request.Params.Arguments["username"].(string)
	password := request.Params.Arguments["password"].(string)
	database := request.Params.Arguments["database"].(string)

	// Connect to the database
	err := ds.Connect(ctx, host, port, username, password, database)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(fmt.Sprintf("Successfully connected to MySQL at %s:%s", host, port)), nil
}

// QueryHandler executes a SQL query
func queryHandler(ctx context.Context, request mcp.CallToolRequest, ds datastore.DatastoreInterface) (*mcp.CallToolResult, error) {
	// Check if connected to a database
	if err := ds.CheckConnection(); err != nil {
		return nil, err
	}

	// Extract query
	sql := request.Params.Arguments["sql"].(string)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Execute query
	rows, err := ds.QueryContext(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("query execution failed: %w", err)
	}
	defer rows.Close()

	// Format the result
	result, err := utils.FormatQueryResult(rows)
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(result), nil
}

// ListDatabasesHandler lists all databases
func listDatabasesHandler(ctx context.Context, request mcp.CallToolRequest, ds datastore.DatastoreInterface) (*mcp.CallToolResult, error) {
	// Check if connected to a database
	if err := ds.CheckConnection(); err != nil {
		return nil, err
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Execute query to list databases
	rows, err := ds.QueryContext(ctx, "SHOW DATABASES")
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	defer rows.Close()

	// Format the result
	result, err := utils.FormatSimpleTable(rows, "Database")
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(result), nil
}

// ListTablesHandler lists all tables in a database
func listTablesHandler(ctx context.Context, request mcp.CallToolRequest, ds datastore.DatastoreInterface) (*mcp.CallToolResult, error) {
	// Check if connected to a database
	if err := ds.CheckConnection(); err != nil {
		return nil, err
	}

	// Extract database name if provided
	database, ok := request.Params.Arguments["database"].(string)
	if ok && database != "" {
		// Use the specified database
		_, err := ds.ExecContext(ctx, fmt.Sprintf("USE %s", database))
		if err != nil {
			return nil, fmt.Errorf("failed to switch to database %s: %w", database, err)
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Execute query to list tables
	rows, err := ds.QueryContext(ctx, "SHOW TABLES")
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}
	defer rows.Close()

	// Format the result
	result, err := utils.FormatSimpleTable(rows, "Table")
	if err != nil {
		return nil, err
	}

	return mcp.NewToolResultText(result), nil
}

// DescribeTableHandler describes a table structure
func describeTableHandler(ctx context.Context, request mcp.CallToolRequest, ds datastore.DatastoreInterface) (*mcp.CallToolResult, error) {
	// Check if connected to a database
	if err := ds.CheckConnection(); err != nil {
		return nil, err
	}

	// Extract table name
	table := request.Params.Arguments["table"].(string)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Execute query to describe table
	rows, err := ds.QueryContext(ctx, fmt.Sprintf("DESCRIBE %s", table))
	if err != nil {
		return nil, fmt.Errorf("failed to describe table %s: %w", table, err)
	}
	defer rows.Close()

	// Format the result
	result, err := utils.FormatQueryResult(rows)
	if err != nil {
		return nil, err
	}

	// We need to modify the result to show columns instead of rows
	// Extract the row count from the result
	rowCountIndex := len(result) - 20 // Approximate position of the row count
	for i := rowCountIndex; i < len(result); i++ {
		if result[i] == '\n' {
			result = result[:i]
			break
		}
	}

	// Add a table-specific summary
	result += fmt.Sprintf("\n%s table structure described successfully", table)

	return mcp.NewToolResultText(result), nil
}
