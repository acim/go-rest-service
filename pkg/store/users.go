// Package store contains store abstraction.
package store

import (
	"context"
	"errors"

	"github.com/acim/arc/pkg/model"
)

// Users ...
type Users interface {
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Insert(ctx context.Context, user *model.User) error
}

// ErrNotFound it returned when there are no results in the query.
var ErrNotFound = errors.New("not found")
