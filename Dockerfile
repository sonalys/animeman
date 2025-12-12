FROM alpine AS alpine

RUN apk add --no-cache ca-certificates

FROM scratch
ARG TARGETPLATFORM

COPY --from=alpine /etc/ssl/certs /etc/ssl/certs
COPY --from=alpine /usr/share/ca-certificates /usr/share/ca-certificates
COPY ${TARGETPLATFORM}/animeman /usr/bin

ENTRYPOINT ["/usr/bin/animeman"]