package cmd

import (
	"context"
	"errors"

	"github.com/acim/go-rest-server/pkg/model"
	"github.com/acim/go-rest-server/pkg/store"
	"github.com/asaskevich/govalidator"
	"gopkg.in/alecthomas/kingpin.v2"
)

// NewUserCmd create new user command that can manipulate with the users in database.
func NewUserCmd(app *kingpin.Application, users store.Users) *kingpin.CmdClause {
	cmd := app.Command("user", "manipulate users in the database")
	add := cmd.Command("create", "create new user")
	email := add.Arg("email", "user's email address").Required().String()
	password := add.Arg("password", "user's password").Required().String()
	add.Action(func(c *kingpin.ParseContext) error {
		if !govalidator.IsEmail(*email) {
			return errors.New("email not valid")
		}

		if len(*password) < 8 {
			return errors.New("password should contain minimum 8 characters")
		}

		user, err := model.NewUser(*email, *password)
		if err != nil {
			return err
		}

		return users.Insert(context.Background(), user)
	})

	return cmd
}
