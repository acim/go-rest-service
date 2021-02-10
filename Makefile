.PHONY: lint test test-verbose start stop test-with-coverage

lint:
	@golint ./...
	@golangci-lint run --enable-all

test:
	@go test ./...

test-verbose:
	@go test -v ./...

test-with-coverage:
	@go test -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out

start:
	@docker-compose up --build

stop:
	@docker-compose down
