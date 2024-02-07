IMG = ghcr.io/sonalys/animeman

run:
	go run cmd/service/main.go

build:
	CGO_ENABLED=0 go build -o ./bin/animeman ./cmd/service/main.go

image:
	docker build -t ${IMG}:latest -f builders/service.dockerfile .

push:
	docker push ${IMG}