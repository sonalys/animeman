IMG = ghcr.io/sonalys/animeman

build:
	CGO_ENABLED=0 GOOS=linux go build -o /animeman

image:
	docker build -t ${IMG}:latest -f Dockerfile .

push:
	docker push ${IMG}