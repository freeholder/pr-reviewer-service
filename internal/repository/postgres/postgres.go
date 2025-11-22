package postgres

import (
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgconn"
)

func isUnique(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	msg := err.Error()
	if strings.Contains(msg, "SQLSTATE 23505") ||
		strings.Contains(msg, "duplicate key value violates unique constraint") {
		return true
	}

	return false
}

type DB struct {
	Conn *sql.DB
}

func NewDB(conn *sql.DB) *DB {
	return &DB{Conn: conn}
}
