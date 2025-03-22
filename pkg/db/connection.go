package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// DB is the global database connection pool
var DB *sql.DB

// Connect establishes a connection to the MySQL database
func Connect(ctx context.Context, host, port, username, password, database string) error {
	// Close existing connection if any
	if DB != nil {
		DB.Close()
	}

	// Create DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, database)

	// Open database connection
	var err error
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(time.Minute * 5)

	// Test connection
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = DB.PingContext(ctxTimeout)
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// IsConnected checks if there is an active database connection
func IsConnected() bool {
	return DB != nil
}

// CheckConnection verifies if the database is connected and returns an error if not
func CheckConnection() error {
	if !IsConnected() {
		return fmt.Errorf("not connected to a database, use connect tool first")
	}
	return nil
}
