package pgstore

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// NewDB creates new Postgres database handle.
func NewDB(hostname, username, password, databaseName string) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		hostname, username, password, databaseName)

	return sqlx.Connect("postgres", dsn)
}
