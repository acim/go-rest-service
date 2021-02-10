package main

import (
	"fmt"
	"os"
	"time"

	"github.com/acim/arc/pkg/controller"
	"github.com/acim/arc/pkg/mail"
	"github.com/acim/arc/pkg/rest"
	"github.com/acim/arc/pkg/store/pgstore"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/mailgun/mailgun-go/v4"
	"go.ectobit.com/act"
	"go.uber.org/zap"
)

type config struct {
	ServiceName string `def:"arc"`
	ServerPort  int    `def:"3000"`
	MetricsPort int    `def:"3001"`
	Environment string `def:"dev"`
	JWT         struct {
		Secret                 string
		AuthTokenExpiration    time.Duration `env:"JWT_AUTH_TOKEN_EXP" def:"15m"`
		RefreshTokenExpiration time.Duration `env:"JWT_REFRESH_TOKEN_EXP" def:"168h"`
	}
	Database struct {
		Hostname string `env:"DB_HOST" def:"postgres"`
		Username string `env:"DB_USER" def:"postgres"`
		Password string `env:"DB_PASS"`
		Name     string `env:"DB_NAME" defa:"postgres"`
	}
	Mailgun struct {
		Domain    string `env:"MG_DOMAIN"`
		APIKey    string `env:"MG_API_KEY"`
		Recipient string `env:"MG_EMAIL_TO"`
	}
}

func main() { //nolint:funlen
	c := &config{}

	cmd := os.Args[1]
	switch cmd {
	case "serve":
		serveCmd := act.New("serve")

		if err := serveCmd.Parse(c, os.Args[2:]); err != nil {
			exit("parse arguments", err)
		}

		logger, err := rest.NewLogger(c.Environment)
		if err != nil {
			exit("logger", err)
		}

		db, err := pgstore.NewDB(c.Database.Hostname, c.Database.Username, c.Database.Password, c.Database.Name)
		if err != nil {
			logger.Error("pg connect", zap.Error(err))
		}

		users := pgstore.NewUsers(db, pgstore.UsersTableName("admin"))
		jwtAuth := jwtauth.New("HS256", []byte(c.JWT.Secret), nil)
		authController := controller.NewAuth(users, jwtAuth, logger)

		mailSender := mail.NewMailgun(mailgun.NewMailgun(c.Mailgun.Domain, c.Mailgun.APIKey))
		mailController := controller.NewMail(mailSender, c.Mailgun.Recipient, logger)

		router := rest.DefaultRouter(c.ServiceName, nil, logger)
		router.Post("/auth", authController.Login)
		router.Post("/mail", mailController.Send)

		router.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(jwtAuth))
			r.Use(jwtauth.Authenticator)

			r.Get("/auth", authController.User)
			r.Delete("/auth", authController.Logout)
		})

		// router.Get("/heavy", func(w http.ResponseWriter, r *http.Request) {
		// 	err := valve.Lever(r.Context()).Open()
		// 	if err != nil {
		// 		logger.Error("open valve lever", zap.Error(err))
		// 	}
		// 	defer valve.Lever(r.Context()).Close()

		// 	select {
		// 	case <-valve.Lever(r.Context()).Stop():
		// 		logger.Info("valve closed, finishing")
		// 	case <-time.After(2 * time.Second):
		// 		// Do some heave lifting
		// 		time.Sleep(2 * time.Second)
		// 	}

		// 	res := abmiddleware.ResponseFromContext(r.Context())
		// 	res.SetPayload("all done")
		// })

		app := rest.NewServer(c.ServiceName, c.ServerPort, c.MetricsPort, router, logger)
		app.Run()

	case "user":
		config := &struct{}{}
		userCmd := act.New("user")

		if err := userCmd.Parse(config, os.Args[2:]); err != nil {
			exit("parse arguments", err)
		}

	// 		email := add.Arg("email", "user's email address").Required().String()
	// password := add.Arg("password", "user's password").Required().String()
	// add.Action(func(c *kingpin.ParseContext) error {
	// 	if !govalidator.IsEmail(*email) {
	// 		return errors.New("email not valid")
	// 	}

	// 	if len(*password) < 8 {
	// 		return errors.New("password should contain minimum 8 characters")
	// 	}

	// 	user, err := model.NewUser(*email, *password)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	return users.Insert(context.Background(), user)
	// })

	default:
		usage()
	}
}

func usage() {
	usage := `Usage of arc:
  arc <command>

  Available commands:
    serve	start rest server
    user	create new user`

	fmt.Println(usage) //nolint:forbidigo
	os.Exit(2)         //nolint:gomnd
}

func exit(message string, err error) {
	fmt.Printf("%s: %v\n", message, err) //nolint:forbidigo
	os.Exit(2)                           //nolint:gomnd
}
