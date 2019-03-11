# mobycron

[![Build Status](https://drone.pfillion.com/api/badges/pfillion/mobycron/status.svg?branch=master)](https://drone.pfillion.com/pfillion/mobycron)
[![Go Report Card](https://goreportcard.com/badge/github.com/pfillion/mobycron)](https://goreportcard.com/report/github.com/pfillion/mobycron)
[![microbadger image](https://images.microbadger.com/badges/image/pfillion/mobycron.svg)](https://microbadger.com/images/pfillion/mobycron "Get your own image badge on microbadger.com")
[![microbadger image](https://images.microbadger.com/badges/version/pfillion/mobycron.svg)](https://microbadger.com/images/pfillion/mobycron "Get your own version badge on microbadger.com")
[![microbadger image](https://images.microbadger.com/badges/commit/pfillion/mobycron.svg)](https://microbadger.com/images/pfillion/mobycron "Get your own commit badge on microbadger.com")

A simple cron deamon for docker written in go. It use the [robfig cron library](https://github.com/robfig/cron) engine and all cron jobs can be confgurated by a JSON file.

The docker image include the official backup tool [restic](https://github.com/restic/restic). This may be usefull for schedule prune job and cleanup backup snaphots directly on the restic server hosting REST repositories for optimal performance.

## Versions

* [latest](https://github.com/pfillion/mobycron/tree/master) available as ```pfillion/mobycron:latest``` at [Docker Hub](https://hub.docker.com/r/pfillion/mobycron/)

## Go packages

You can use the [mobycron library](https://github.com/pfillion/mobycron) directly by importing the package ```github.com/pfillion/mobycron/pkg/cron``` directly in your project.

## Tools included in the docker image

* bash
* ca-certificates
* curl
* fuse
* openssh-client
* restic
* tzdata

## Configuration file

You can mount directly the ```config.json``` file or use docker configuration to schedule all job like a crontab file. See the [exemples](https://github.com/pfillion/mobycron/tree/master/exemples) in the source code or below.

* /configs/config.json

```json
[
    {
        "schedule": "0/2 * * * * *",
        "command": "bash",
        "args": [
            "-c",
            "echo Hello $NAME"
        ]
    },
    {
        "schedule": "0/5 * * * * *",
        "command": "curl",
        "args": [
            "-s",
            "-S",
            "-X",
            "GET",
            "http://exemple.com"
        ]
    },
    {
        "schedule": "0 0 3 ? * *",
        "command": "/usr/bin/restic",
        "args": [
            "-r",
            "$REPO__FILE",
            "-p",
            "/configs/passwd",
            "forget",
            "--keep-daily",
            "7",
            "--prune"
        ]
    },
]
```

This file will schedule three cron job.

* The first one will replace ```$NAME``` by the environnement variable configured in the container and print ```Hello``` + ```$NAME``` every 2 seconds.
* The second will execute a ```curl``` command every 5 seconds. It may be usefull when you need to call any simple **webcron** or **webhook** URL like with [EasyCron](https://www.easycron.com)

* The last one will forget and prune all restic snapshot older than 7 days every day at 3 AM. It use the secret environnement variable ```$REPO__FILE``` for telling to restic the repository to use and a password file ```/configs/passwd```mounted in the container.

## Environnement variables

Cron job support any environnement variables specified by docker and replace it by the real value before executing the command.

It also support ```TZ``` variable to confgure local time zone of the container.

## Docker Secrets

As an alternative to passing sensitive information via environment variables, `__FILE` may be appended to any environment variables, causing the job to load the values for those variables from files present in the container. In particular, this can be used to load passwords from Docker secrets stored in `/run/secrets/<secret_name>` files.

## Docker compose

See the docker swarm [exemples](https://github.com/pfillion/mobycron/tree/master/exemples) in the source code or below.

```yml
version: '3.6'
services:
  cron:
    image: pfillion/mobycron:latest
    environment:
      - TZ=America/New_York
      - NAME=World!!!
      - REPO__FILE=/run/secrets/restic-repo
    configs:
      - source: mobycron-config
        target: /configs/config.json
    secrets:
      - source: restic-passwd
        target: /configs/passwd
      - restic-repo

configs:
  mobycron-config:
    external: true

secrets:
  restic-repo:
    external: true
  restic-passwd:
    external: true
```

## Authors

* [pfillion](https://github.com/pfillion)

## License

MIT