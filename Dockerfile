FROM restic/restic:0.9.3 as restic_builder

FROM alpine:latest

# Build-time metadata as defined at http://label-schema.org
ARG BUILD_DATE
ARG VCS_REF
ARG VERSION
LABEL \
    org.label-schema.build-date=$BUILD_DATE \
    org.label-schema.name="mobycron" \
    org.label-schema.description="A simple cron deamon for docker written in go" \
    org.label-schema.url="https://hub.docker.com/r/pfillion/mobycron" \
    org.label-schema.vcs-ref=$VCS_REF \
    org.label-schema.vcs-url="https://github.com/pfillion/mobycron" \
    org.label-schema.vendor="pfillion" \
    org.label-schema.version=$VERSION \
    org.label-schema.schema-version="1.0"

RUN apk add --update --no-cache \
    ca-certificates \
    fuse \
    openssh-client \
    curl \
    bash \
    tzdata \
    mariadb-backup

COPY --from=restic_builder /usr/bin/restic /usr/bin
COPY bin /usr/bin

ENTRYPOINT [ "mobycron"]

