FROM golang:1.21 as build

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /animeman

FROM alpine:3.9.6

COPY --from=build /animeman /animeman

CMD ["/animeman"]