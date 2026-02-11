# --- Stage 1: Build ---
FROM golang:1.26 AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy dependency files first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

ENV CGO_ENABLED=0
ARG VERSION=dev

RUN go build -ldflags "-extldflags '-static' -X main.version=$VERSION" -o /app/bin/server /app/cmd/server

# --- Stage 2: Runtime ---
FROM scratch

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /usr/share/ca-certificates /usr/share/ca-certificates

WORKDIR /bin

# Copy the binary from the builder stage
COPY --from=builder /app/bin/server .

# Expose the port your Go app listens on
EXPOSE 8080

# Run the binary
CMD ["server"]