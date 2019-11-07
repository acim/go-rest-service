package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/acim/go-rest-server/pkg/cmd"
	"github.com/acim/go-rest-server/pkg/controller"
	"github.com/acim/go-rest-server/pkg/rest"
	"github.com/acim/go-rest-server/pkg/store/pgstore"
	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
	"github.com/go-chi/jwtauth"
	"github.com/jmoiron/sqlx"
	"github.com/kelseyhightower/envconfig"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
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
}

func main() {
	c := &config{}
	if err := envconfig.Process("", c); err != nil {
		log.Fatalf("failed parsing environment variables: %v", err)
	}

	logger, err := logger(c.Environment)
	if err != nil {
		log.Fatalf("failed initializing logger: %v", err)
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Database.Hostname, c.Database.Username, c.Database.Password, c.Database.Name)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	users := pgstore.NewUsers(db, pgstore.UsersTableName("admin"))

	jwtAuth := jwtauth.New("HS256", []byte(c.JWT.Secret), nil)

	authController := controller.NewAuth(users, jwtAuth, logger)

	router := rest.DefaultRouter(c.ServiceName, logger)
	router.Use(getCors().Handler)

	router.Post("/auth/login", authController.Login)

	router.Group(func(r chi.Router) {
		r.Use(jwtauth.Verifier(jwtAuth))
		r.Use(jwtauth.Authenticator)

		r.Get("/auth/user", authController.User)
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

func logger(env string) (*zap.Logger, error) {
	var logger *zap.Logger

	var err error

	switch env {
	case "prod":
		config := zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.WarnLevel)
		logger, err = config.Build()
	case "dev":
		logger, err = zap.NewDevelopment()
	default:
		return nil, fmt.Errorf("logger: unknown environment: '%s'", env)
	}

	if err != nil {
		return nil, fmt.Errorf("logger: %w", err)
	}

	return logger, nil
}

func getCors() *cors.Cors {
	return cors.New(cors.Options{
		// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
		AllowedOrigins: []string{"*"},
		// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		// AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})
}
