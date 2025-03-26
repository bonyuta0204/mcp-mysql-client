package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/bonyuta0204/mcp-mysql-client/pkg/datastore"
	_ "github.com/go-sql-driver/mysql"
)

// TestConfig holds configuration for integration tests
type TestConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

// GetTestConfig returns MySQL connection config for tests
func GetTestConfig() TestConfig {
	// Default values for local testing with docker-compose
	config := TestConfig{
		Host:     "localhost",
		Port:     "3306",
		Username: "root",
		Password: "test",
		Database: "testdb",
	}

	// Override with environment variables if present
	if host := os.Getenv("TEST_MYSQL_HOST"); host != "" {
		config.Host = host
	}
	if port := os.Getenv("TEST_MYSQL_PORT"); port != "" {
		config.Port = port
	}
	if username := os.Getenv("TEST_MYSQL_USERNAME"); username != "" {
		config.Username = username
	}
	if password := os.Getenv("TEST_MYSQL_PASSWORD"); password != "" {
		config.Password = password
	}
	if database := os.Getenv("TEST_MYSQL_DATABASE"); database != "" {
		config.Database = database
	}

	return config
}

// SetupTestDatastore creates a new MySQL datastore for testing
func SetupTestDatastore(t *testing.T) *datastore.MySQLDatastore {
	t.Helper()

	config := GetTestConfig()
	ds := &datastore.MySQLDatastore{}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Explicitly set allowNativePasswords=true in the connection
	err := ds.Connect(ctx, config.Host, config.Port, config.Username, config.Password, config.Database)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	return ds
}

// CleanupTestDatastore closes the database connection
func CleanupTestDatastore(t *testing.T, ds *datastore.MySQLDatastore) {
	t.Helper()

	if ds != nil && ds.DB != nil {
		if err := ds.Close(); err != nil {
			t.Logf("Error closing test database connection: %v", err)
		}
	}
}

// SkipIfNoMySQL skips the test if MySQL is not available
func SkipIfNoMySQL(t *testing.T) {
	t.Helper()

	config := GetTestConfig()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?allowNativePasswords=true",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Skip("Skipping test: could not connect to MySQL")
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		t.Skip("Skipping test: MySQL is not available")
	}
}
