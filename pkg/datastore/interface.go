package datastore

import (
	"context"
	"database/sql"
)

type DatastoreInterface interface {
	Connect(ctx context.Context, host, port, username, password, database string) error
	Connection() *sql.DB
	CheckConnection() error
	IsConnected() bool
}
