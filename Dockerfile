FROM alpine:latest AS alpine
RUN apk add -U --no-cache ca-certificates

FROM scratch
ARG TARGETPLATFORM

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/animeman"]
COPY ${TARGETPLATFORM}/animeman /