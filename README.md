# Go skeleton application for REST server with additional metrics endpoint

![check](https://github.com/acim/arc/workflows/check/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/acim/arc.svg)](https://pkg.go.dev/github.com/acim/arc)
[![Go Report](https://goreportcard.com/badge/github.com/acim/arc)](https://goreportcard.com/report/github.com/acim/arc)
[![License](https://img.shields.io/github/license/acim/arc)](LICENSE)

You can run the server by typing **docker-compose up --build**.

Check [main.go](https://github.com/acim/arc/blob/main/cmd/arc/main.go) for example usage.

## This project also includes the following middlewares which can be use independently of the server

- RenderJSON - simplifies implementation of JSON REST API endpoints
- ZapLogger - [chi](https://github.com/go-chi/chi) middleware for logging using [zap](https://github.com/uber-go/zap) logger
- PromMetrics - [chi](https://github.com/go-chi/chi) middleware providing [Prometheus](https://prometheus.io/) metrics to your HTTP server
  Tracks total number of requests and requests duration partitioned by status code, method and request URI

Inside .examples directory you can find basic examples for each of the middlewares. Just run **go run main.go** and
check the output in your browser, localhost:3000. To check the metrics in PromMetrics example, use localhost:3001.

## [Generate your JWT secret here](https://www.grc.com/passwords.htm)
