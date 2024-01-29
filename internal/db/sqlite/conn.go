package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
	"slices"

	"github.com/ocontest/backend/pkg"
	"github.com/pkg/errors"
)

var supportedDBs = []string{"pgx", "sqlite3", "mysql"}

func NewDBConn(ctx context.Context, dbType, connStr string) (*sql.DB, error) {

	if !slices.Contains(supportedDBs, dbType) {
		return nil, errors.Wrap(pkg.ErrBadRequest, fmt.Sprintf("database type: %v is not supported!", dbType))
	}

	db, err := sql.Open(dbType, connStr)
	if err != nil {
		return nil, err
	}
	err = db.PingContext(ctx)
	return db, err
}
