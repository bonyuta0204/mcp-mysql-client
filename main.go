package main

import (
	"fmt"

	"github.com/bonyuta0204/mcp-mysql-client/pkg/handlers"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

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
	s.AddTool(connectTool, handlers.ConnectHandler)
	s.AddTool(queryTool, handlers.QueryHandler)
	s.AddTool(listDatabasesTool, handlers.ListDatabasesHandler)
	s.AddTool(listTablesTool, handlers.ListTablesHandler)
	s.AddTool(describeTableTool, handlers.DescribeTableHandler)

	// Start the stdio server
	if err := server.ServeStdio(s); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}
