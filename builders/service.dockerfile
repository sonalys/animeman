FROM golang:1.22 AS builder
WORKDIR /build

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o ./bin/animeman ./cmd/service/main.go

FROM scratch

COPY ./builders/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER nobody

COPY --from=builder ./build/bin/animeman /server
COPY --from=builder /var/run /var/run

ENTRYPOINT ["/server"]