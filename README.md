# Go skeleton application for REST server with additional metrics endpoint

[![Build Status](https://drone.ablab.de/api/badges/acim/go-rest-server/status.svg)](https://drone.ablab.de/acim/go-rest-server)
[![Quality Gate Status](https://sonarqube.ablab.de/api/project_badges/measure?project=acim%3Ago-rest-server&metric=alert_status)](https://sonarqube.ablab.de/dashboard?id=acim%3Ago-rest-server)

You can run the server by typing **docker-compose up --build**.

Check main.go for example usage.

## This project also includes the following middlewares which can be use independently of the server

* RenderJSON - simplifies implementation of JSON REST API endpoints
* ZapLogger - [chi](https://github.com/go-chi/chi) middleware for logging using [zap](https://github.com/uber-go/zap) logger
* PromMetrics - [chi](https://github.com/go-chi/chi) middleware providing [Prometheus](https://prometheus.io/) metrics to your HTTP server
  Tracks total number of requests and requests duration partitioned by status code, method and request URI

Inside .examples directory you can find basic examples for each of the middlewares. Just run **go run main.go** and check the output in your [browser](http://localhost:3000).

To check the metrics in PromMetrics example, use this [url](http://localhost:3001) in your browser.

## [Generate your JWT secret here](https://www.grc.com/passwords.htm)
