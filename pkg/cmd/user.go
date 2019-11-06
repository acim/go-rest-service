package cmd

import (
	"context"

	"github.com/acim/go-rest-server/pkg/model"
	"github.com/acim/go-rest-server/pkg/store"
	"gopkg.in/alecthomas/kingpin.v2"
)

// NewUserCmd create new user command that can manipulate with the users in database.
func NewUserCmd(app *kingpin.Application, users store.Users) *kingpin.CmdClause {
	cmd := app.Command("user", "manipulate users in the database")
	add := cmd.Command("create", "create new user")
	email := add.Arg("email", "user's email address").Required().String()
	password := add.Arg("password", "user's password").Required().String()
	add.Action(func(c *kingpin.ParseContext) error {
		user, err := model.NewUser(*email, *password)
		if err != nil {
			return err
		}
		return users.Insert(context.Background(), user)
	})

	return cmd
}
