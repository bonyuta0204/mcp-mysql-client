# MySQL MCP Server

A Model Context Protocol (MCP) server for interacting with MySQL databases.

## Overview

This project implements a MySQL client as an MCP server using Go. It allows AI models to interact with MySQL databases through a standardized interface, enabling them to perform database operations like querying, listing databases and tables, and describing table structures.

## Features

- Connect to MySQL databases
- Execute SQL queries
- List available databases
- List tables in a database
- Describe table structure

## Project Structure

```
.
├── main.go              # Main application entry point
├── pkg/
│   ├── datastore/       # Database connection management
│   │   ├── interface.go # Interface for datastore operations
│   │   └── mysql.go     # MySQL implementation
│   ├── handlers/        # MCP tool handlers
│   │   ├── handlers.go
│   │   └── handlers_test.go
│   ├── integration/     # Integration tests with real MySQL
│   │   ├── helper.go
│   │   └── handlers_integration_test.go
│   └── utils/           # Utility functions
│       └── formatter.go
├── docker-compose.yml   # Docker setup for testing
└── README.md
```

## Requirements

- Go 1.24 or later
- MySQL server

## Dependencies

- [github.com/mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - MCP Go library
- [github.com/go-sql-driver/mysql](https://github.com/go-sql-driver/mysql) - MySQL driver for Go

## Building and Running

```bash
# Build the server
go build

# Run the server
./mcp-mysql-client
```

## Testing

### Unit Tests

Run the unit tests with:

```bash
make test-unit
```

## MCP Tools

### Connect

Establishes a connection to a MySQL database.

**Parameters:**
- `host` (required): MySQL host address
- `port` (default: "3306"): MySQL port
- `username` (required): MySQL username
- `password` (required): MySQL password
- `database` (default: ""): MySQL database name

### Query

Executes a SQL query on the connected database.

**Parameters:**
- `sql` (required): SQL query to execute

### List Databases

Lists all databases available on the connected MySQL server.

**Parameters:** None

### List Tables

Lists all tables in the current database or a specified database.

**Parameters:**
- `database` (optional): Database name (uses current connection if not specified)

### Describe Table

Describes the structure of a specified table.

**Parameters:**
- `table` (required): Table name

## License

MIT
