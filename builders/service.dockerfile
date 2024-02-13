FROM golang:1.22 AS builder

RUN mkdir -p /var/run/secrets && \
  chmod 0700 /var/run/secrets && \
  chown 65534:65534 /var/run/secrets

WORKDIR /build

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/root/.cache/go-build go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-build CGO_ENABLED=0 go build -o ./bin/animeman ./cmd/service/main.go

FROM scratch
ARG SERVICE
COPY ./builders/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER nobody
COPY --from=builder ./build/bin/animeman /server
COPY --from=builder /var/run /var/run
COPY --from=builder /var/run/secrets /var/run/secrets

ENTRYPOINT ["/server"]