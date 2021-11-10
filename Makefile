.PHONY: lint test test-verbose test-with-coverage

lint:
	@go vet ./...
	@golangci-lint run --enable-all --disable varnamelen

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
