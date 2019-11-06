package pgstore

import (
	"context"
	"fmt"
	"strings"

	"github.com/acim/go-rest-server/pkg/model"
	"github.com/acim/go-rest-server/pkg/store"
	"github.com/jmoiron/sqlx"
)

var _ store.Users = (*Users)(nil)

// Users implements store.Users interface.
type Users struct {
	db         *sqlx.DB
	tableName  string
	prepInsert *sqlx.NamedStmt
}

// NewUsers creates new users store.
func NewUsers(db *sqlx.DB, opts ...UsersOption) *Users {
	u := &Users{
		db:        db,
		tableName: "user",
	}

	for _, opt := range opts {
		opt(u)
	}

	return u
}

// Insert implements store.Users interface.
func (us *Users) Insert(ctx context.Context, u *model.User) error {
	var err error

	if us.prepInsert == nil {
		sql := "INSERT INTO table (id, email, password) VALUES (:id, :email, :password)"
		sql = strings.Replace(sql, "table", us.tableName, 1)

		us.prepInsert, err = us.db.PrepareNamedContext(ctx, sql)
		if err != nil {
			return fmt.Errorf("prepare: %w", err)
		}
	}

	_, err = us.prepInsert.ExecContext(ctx, u)

	return err
}

// UsersOption ...
type UsersOption func(*Users)

// UsersTableName ...
func UsersTableName(name string) UsersOption {
	return func(u *Users) {
		u.tableName = name
	}
}
