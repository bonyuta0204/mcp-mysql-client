package db

import (
	"context"
	"database/sql"
)

type DBInterface interface {
	Connect(ctx context.Context, host, port, username, password, database string) error
	Connection() *sql.DB
	CheckConnection() error
	IsConnected() bool
}
