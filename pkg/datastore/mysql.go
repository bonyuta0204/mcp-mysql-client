package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLDatastore struct {
	DB *sql.DB
}

func (d *MySQLDatastore) IsConnected() bool {
	return d.DB != nil
}

func (d *MySQLDatastore) CheckConnection() error {
	if !d.IsConnected() {
		return fmt.Errorf("not connected to a database, use connect tool first")
	}
	return nil
}

func (d *MySQLDatastore) Close() error {
	if d.DB != nil {
		return d.DB.Close()
	}
	return nil
}

func (d *MySQLDatastore) Connection() *sql.DB {
	return d.DB
}

func (d *MySQLDatastore) Connect(ctx context.Context, host, port, username, password, database string) error {
	// Close existing connection if any
	if d.IsConnected() {
		d.Close()
	}

	// Create DSN (Data Source Name)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", username, password, host, port, database)

	// Open database connection
	var err error
	d.DB, err = sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	d.DB.SetMaxOpenConns(10)
	d.DB.SetMaxIdleConns(5)
	d.DB.SetConnMaxLifetime(time.Minute * 5)

	// Test connection
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = d.DB.PingContext(ctxTimeout)
	if err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// Global instance of MySQLDatastore
var DB *MySQLDatastore

func init() {
	DB = &MySQLDatastore{}
}
