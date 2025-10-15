SHELL=bash

BUILD=build
BIN_DIR?=.

BUILD_TIME=$(shell date +%s)
GIT_COMMIT=$(shell git rev-parse HEAD)
VERSION ?= $(shell git tag --points-at HEAD | grep ^v | head -n 1)
LDFLAGS=-ldflags "-X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT) -X main.Version=$(VERSION)"

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: test-component
test-component: 
	cd examples/compose; docker compose up --abort-on-container-exit
	@echo "please ignore error codes 0, like so: ERRO[xxxx] 0, as error code 0 means that there was no error"

.PHONY: build
build:
	go build ./...

.PHONY: audit
audit:
	dis-vulncheck

.PHONY: lint
lint:
	golangci-lint run ./...

