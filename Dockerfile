FROM alpine:edge

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
    restic

# fuse TODO: Check if dependencies for restic, useless
# openssh-client TODO: Check if dependencies for restic, useless
# restic TODO: With docker crontab, useless. Change example docker-compose to adapt with restic container job
# tzdata TODO: with v3, check if it necessary with new ENV integrated in cron dirrectly

COPY bin /usr/bin

ENTRYPOINT [ "mobycron"]

