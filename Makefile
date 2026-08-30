IMG = ghcr.io/sonalys/animeman
ARCH := $(shell go env GOARCH)

run:
	go run cmd/service/main.go

build:
	CGO_ENABLED=0 go build -o ./bin/animeman ./cmd/service/main.go

image:
	docker build -t ${IMG}:latest -f builders/Dockerfile.linux.$(ARCHITECTURE) .

push:
	docker push ${IMG}