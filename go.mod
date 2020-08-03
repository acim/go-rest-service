module github.com/acim/go-rest-server

go 1.13

replace encoding/json => github.com/json-iterator/go v1.1.8

require (
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/go-chi/cors v1.1.1
	github.com/go-chi/jwtauth v4.0.3+incompatible
	github.com/go-chi/valve v0.0.0-20170920024740-9e45288364f4
	github.com/gobuffalo/envy v1.7.1 // indirect
	github.com/google/uuid v1.1.1
	github.com/jmoiron/sqlx v1.2.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lib/pq v1.0.0
	github.com/mailgun/mailgun-go v2.0.0+incompatible
	github.com/mailgun/mailgun-go/v3 v3.6.1
	github.com/prometheus/client_golang v1.2.1
	go.uber.org/zap v1.12.0
	golang.org/x/crypto v0.0.0-20190510104115-cbcb75029529
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
)
