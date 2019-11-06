package store

import (
	"context"

	"github.com/acim/go-rest-server/pkg/model"
)

// Users ...
type Users interface {
	Insert(context.Context, *model.User) error
}
