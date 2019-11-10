package pgstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/acim/go-rest-server/pkg/model"
	"github.com/acim/go-rest-server/pkg/store"
	"github.com/jmoiron/sqlx"
)

var _ store.Users = (*Users)(nil)

// Users implements store.Users interface.
type Users struct {
	db              *sqlx.DB
	tableName       string
	prepFindByID    *sqlx.NamedStmt
	prepFindByEmail *sqlx.NamedStmt
	prepInsert      *sqlx.NamedStmt
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

// FindByID finds user by id (UUID).
func (s *Users) FindByID(ctx context.Context, id string) (*model.User, error) {
	var err error

	if s.prepFindByID == nil {
		sql := "SELECT id, email, password FROM table WHERE id=:id"
		sql = strings.Replace(sql, "table", s.tableName, 1)

		s.prepFindByID, err = s.db.PrepareNamedContext(ctx, sql)
		if err != nil {
			return nil, fmt.Errorf("prepare: %w", err)
		}
	}

	u := &model.User{}
	err = s.prepFindByID.QueryRowxContext(ctx, map[string]interface{}{"id": id}).StructScan(u)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, store.ErrNotFound
	}

	return u, err
}

// FindByEmail finds user by email address.
func (s *Users) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var err error

	if s.prepFindByEmail == nil {
		sql := "SELECT id, email, password FROM table WHERE email=:email"
		sql = strings.Replace(sql, "table", s.tableName, 1)

		s.prepFindByEmail, err = s.db.PrepareNamedContext(ctx, sql)
		if err != nil {
			return nil, fmt.Errorf("prepare: %w", err)
		}
	}

	u := &model.User{}
	err = s.prepFindByEmail.QueryRowxContext(ctx, map[string]interface{}{"email": email}).StructScan(u)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, store.ErrNotFound
	}

	return u, err
}

// Insert implements store.Users interface.
func (s *Users) Insert(ctx context.Context, u *model.User) error {
	var err error

	if s.prepInsert == nil {
		sql := "INSERT INTO table (id, email, password) VALUES (:id, :email, :password)"
		sql = strings.Replace(sql, "table", s.tableName, 1)

		s.prepInsert, err = s.db.PrepareNamedContext(ctx, sql)
		if err != nil {
			return fmt.Errorf("prepare: %w", err)
		}
	}

	_, err = s.prepInsert.ExecContext(ctx, u)

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
