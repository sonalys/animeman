IMG = ghcr.io/sonalys/animeman
ARCH := $(shell go env GOARCH)

ifeq ($(ARCH),x86_64)
	ARCHITECTURE := amd64
else ifeq ($(ARCH),aarch64)
	ARCHITECTURE := arm64
endif

TOOLS_DIR = .tools
CONFIG_DIR = .config

SQLC = go tool -modfile=$(TOOLS_DIR)/sqlc/go.mod sqlc -f $(CONFIG_DIR)/sqlc.yaml

VERSION := $(shell git describe --tags --always --dirty)

run:
	CONFIG_PATH=$(CONFIG_DIR)/config.yaml go run ./cmd/server/...

build:
	CGO_ENABLED=0 go build -o ./bin/animeman ./cmd/server/main.go

image:
	docker build -t ${IMG}:latest --build-arg VERSION=$(VERSION)  -f builders/Dockerfile.linux.$(ARCHITECTURE) .

push:
	docker push ${IMG}

generate-postgres:
	$(SQLC) generate