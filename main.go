package main

import (
	"log"
	"os"
	"time"

	"github.com/acim/go-rest-server/pkg/cmd"
	"github.com/acim/go-rest-server/pkg/controller"
	"github.com/acim/go-rest-server/pkg/mail"
	"github.com/acim/go-rest-server/pkg/rest"
	"github.com/acim/go-rest-server/pkg/store/pgstore"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"github.com/mailgun/mailgun-go/v3"
	"gopkg.in/alecthomas/kingpin.v2"
)

type config struct {
	ServiceName string `default:"rest-server"`
	ServerPort  int    `default:"3000"`
	MetricsPort int    `default:"3001"`
	Environment string `default:"dev"`
	JWT         struct {
		Secret                 string        `required:"true"`
		AuthTokenExpiration    time.Duration `envconfig:"JWT_AUTH_TOKEN_EXP" default:"15m"`
		RefreshTokenExpiration time.Duration `envconfig:"JWT_REFRESH_TOKEN_EXP" default:"168h"`
	}
	Database struct {
		Hostname string `envconfig:"DB_HOST" required:"true"`
		Username string `envconfig:"DB_USER" required:"true"`
		Password string `envconfig:"DB_PASS" required:"true"`
		Name     string `envconfig:"DB_NAME" required:"true"`
	}
	Mailgun struct {
		Domain    string `envconfig:"MG_DOMAIN"`
		APIKey    string `envconfig:"MG_API_KEY"`
		Recipient string `envconfig:"MG_EMAIL_TO"`
	}
}

func main() {
	c := &config{}
	if err := envconfig.Process("", c); err != nil {
		log.Fatalf("failed parsing environment variables: %v", err)
	}

	logger, err := rest.NewLogger(c.Environment)
	if err != nil {
		log.Fatalf("failed initializing logger: %v", err)
	}

	db, err := pgstore.NewDB(c.Database.Hostname, c.Database.Username, c.Database.Password, c.Database.Name)
	if err != nil {
		log.Fatalln(err)
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

	app := kingpin.New("go-rest-server", "REST API server")
	cmd.NewUserCmd(app, users)
	cmd.NewServerCmd(app, rest.NewServer(c.ServiceName, c.ServerPort, c.MetricsPort, router, logger))
	kingpin.MustParse(app.Parse(os.Args[1:]))
}
