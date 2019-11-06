package cmd

import (
	"github.com/acim/go-rest-server/pkg/rest"
	"gopkg.in/alecthomas/kingpin.v2"
)

// NewServerCmd create new REST server command.
func NewServerCmd(app *kingpin.Application, server *rest.Server) *kingpin.CmdClause {
	cmd := app.Command("server", "starts REST server").Default()
	cmd.Action(func(c *kingpin.ParseContext) error {
		server.Run()
		return nil
	})

	return cmd
}
