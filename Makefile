ROOT := $(shell pwd)
goose-install:
	GOBIN=$(ROOT)/bin go install github.com/pressly/goose/v3/cmd/goose@latest
goose-gen-migration:
	$(ROOT)/bin/goose -dir $(ROOT)/migrations create init sql