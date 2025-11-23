ROOT := $(shell pwd)
goose-install:
	GOBIN=$(ROOT)/bin go install github.com/pressly/goose/v3/cmd/goose@latest
goose-gen-migration:
	$(ROOT)/bin/goose -dir $(ROOT)/migrations create init sql
golangci-lint:
	GOBIN=$(ROOT)/bin go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
lint:
	$(ROOT)/bin/golangci-lint run ./...
docker-up-app-db:
	docker-compose up -d db app
docker-up-build:
	docker-compose build --no-cache
	docker-compose up -d db app
k6-test:
	docker-compose run --rm k6