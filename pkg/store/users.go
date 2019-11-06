package store

import (
	"context"

	"github.com/acim/go-rest-server/pkg/model"
)

// Users ...
type Users interface {
	FindByID(ctx context.Context, id string) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	Insert(ctx context.Context, user *model.User) error
}
