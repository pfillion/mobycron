ARG VERSION_ALPINE=latest

FROM alpine:${VERSION_ALPINE}

# Build-time metadata as defined at https://github.com/opencontainers/image-spec
ARG DATE
ARG CURRENT_VERSION_MICRO
ARG COMMIT
ARG AUTHOR

LABEL \
    org.opencontainers.image.created=$DATE \
    org.opencontainers.image.url="https://hub.docker.com/r/pfillion/mobycron" \
    org.opencontainers.image.source="https://github.com/pfillion/mobycron" \
    org.opencontainers.image.version=$CURRENT_VERSION_MICRO \
    org.opencontainers.image.revision=$COMMIT \
    org.opencontainers.image.vendor="pfillion" \
    org.opencontainers.image.title="mobycron" \
    org.opencontainers.image.description="A simple cron deamon for docker written in go" \
    org.opencontainers.image.authors=$AUTHOR \
    org.opencontainers.image.licenses="MIT"

RUN apk add --update --no-cache \
    ca-certificates \
    curl \
    bash \
    tzdata \
    libc6-compat

COPY --chmod=0755 bin /usr/bin

ENTRYPOINT [ "mobycron"]

