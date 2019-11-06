package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/acim/go-rest-server/pkg/cmd"
	abmiddleware "github.com/acim/go-rest-server/pkg/middleware"
	"github.com/acim/go-rest-server/pkg/rest"
	"github.com/acim/go-rest-server/pkg/store/pgstore"
	"github.com/go-chi/valve"
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
	Database    struct {
		Hostname string `envconfig:"DB_HOST"`
		Username string `envconfig:"DB_USER"`
		Password string `envconfig:"DB_PASS"`
		Name     string `envconfig:"DB_NAME"`
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

	router := rest.DefaultRouter(c.ServiceName, logger)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		res := abmiddleware.ResponseFromContext(r.Context())
		res.SetPayload("hello world")
	})

	router.Get("/heavy", func(w http.ResponseWriter, r *http.Request) {
		err := valve.Lever(r.Context()).Open()
		if err != nil {
			logger.Error("open valve lever", zap.Error(err))
		}
		defer valve.Lever(r.Context()).Close()

		select {
		case <-valve.Lever(r.Context()).Stop():
			logger.Info("valve closed, finishing")
		case <-time.After(2 * time.Second):
			// Do some heave lifting
			time.Sleep(2 * time.Second)
		}

		res := abmiddleware.ResponseFromContext(r.Context())
		res.SetPayload("all done")
	})

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable",
		c.Database.Hostname, c.Database.Username, c.Database.Password, c.Database.Name)
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	users := pgstore.NewUsers(db, pgstore.UsersTableName("admin"))

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
