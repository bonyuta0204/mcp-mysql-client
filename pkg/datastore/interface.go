package datastore

import (
	"context"
	"database/sql"
)

type DatastoreInterface interface {
	Connect(ctx context.Context, host, port, username, password, database string) error
	CheckConnection() error
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	IsConnected() bool
}
