FROM alpine AS builder
RUN mkdir -p /var/run/secrets && \
  chmod 0700 /var/run/secrets && \
  chown 65534:65534 /var/run/secrets

FROM scratch
ARG SERVICE
COPY ./builders/passwd /etc/passwd
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER nobody
COPY ./bin/linux/arm64/${SERVICE} /server
COPY --from=builder /var/run /var/run
COPY --from=builder /var/run/secrets /var/run/secrets

ENTRYPOINT ["/server"]