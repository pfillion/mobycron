# mobycron

[![Build Status](https://drone.pfillion.com/api/badges/pfillion/mobycron/status.svg?branch=master)](https://drone.pfillion.com/pfillion/mobycron)
[![Go Report Card](https://goreportcard.com/badge/github.com/pfillion/mobycron)](https://goreportcard.com/report/github.com/pfillion/mobycron)
[![microbadger image](https://images.microbadger.com/badges/image/pfillion/mobycron.svg)](https://microbadger.com/images/pfillion/mobycron "Get your own image badge on microbadger.com")
[![microbadger image](https://images.microbadger.com/badges/version/pfillion/mobycron.svg)](https://microbadger.com/images/pfillion/mobycron "Get your own version badge on microbadger.com")
[![microbadger image](https://images.microbadger.com/badges/commit/pfillion/mobycron.svg)](https://microbadger.com/images/pfillion/mobycron "Get your own commit badge on microbadger.com")

A simple cron deamon for docker written in go. It use the [robfig cron library v3](https://github.com/robfig/cron/tree/v3) engine and all cron jobs can be confgurated by two ways. The first mode will perform actions on others Docker containers based on cron schedule. The second mode is by a JSON file acting like a contab file.

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

## Environnement variables

Cron job support any environnement variables specified by docker and replace it by the real value before executing the command.

```MOBYCRON_DOCKER_MODE``` is true by defult. When ```mobycron``` is up and running in this mode, it watches Docker socket and try to find any containers with label ```mobycron.schedule``` and add them to the crontab based on the schedule. Go to [docker mode](#docker-mode) section for more detail with this mode.

```MOBYCRON_PARSE_SECOND``` is false by default. When activate, schedule accept an optional seconds field at the beginning of the cron spec. This is non-standard and has led to a lot of confusion. The new default parser conforms to the standard as described by the [Cron wikipedia page.](https://en.wikipedia.org/wiki/Cron)

```MOBYCRON_CONFIG_FILE``` is file path to schedule all job like a crontab file. Go to [configuration file](#configuration-file) section for more detail with this mode.

```CRON_TZ``` is now the recommended way to to configure local time zone of the container. The legacy ```TZ``` prefix will continue to be supported since it is unambiguous and easy to do so.

## Arguments for the executing container

You can use argument instead of environnment variables. All variables as an equivalant command line option.

* --docker-mode, -d
* --parse-second, -s
* --config-file value, -f value

```sh
> docker run -v /var/run/docker.sock:/var/run/docker.sock pfillion/mobycron:latest --docker-mode=true --parse-second=false
```

## Docker mode

Once ```mobycron``` is up and running in this mode, it watches Docker socket events for create, start and destroy events. If a container is found to have the label ```mobycron.schedule``` then it will be added to the crontab based on the schedule.

Cron scheduling rules and format is describe as follow: [CRON Expression Format](https://godoc.org/github.com/robfig/cron#hdr-CRON_Expression_Format)

Others labels can be applied.

* ```mobycron.action``` is requied and indicate wich action must be performed on the container. Possible choices are ```start```, ```restart```, ```stop``` or ```exec```.
* ```mobycron.command``` specifie the commande line to execute and is requied when the action is ```exec```.
* ```mobycron.timeout``` override the default 10 second timeout to do the action.

### Examples

```sh
# Start the container every minute
> docker run -d --label=mobycron.schedule="0/1 * * * *" --label=mobycron.action="start" busybox date
```

## Configuration file

You can mount directly a file or use docker configuration to schedule all job like a crontab file. See the [examples](https://github.com/pfillion/mobycron/tree/master/examples) in the source code or below.

* /etc/mobycron/config.json

```json
[
    {
        "schedule": "* * * * *",
        "command": "bash",
        "args": [
            "-c",
            "echo Hello $NAME"
        ]
    },
    {
        "schedule": "0/2 * * * *",
        "command": "curl",
        "args": [
            "-s",
            "-S",
            "-X",
            "GET",
            "http://example.com"
        ]
    },
    {
        "schedule": "0 3 * * *",
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

* The first one will replace ```$NAME``` by the environnement variable configured in the container and print ```Hello``` + ```$NAME``` every minutes.
* The second will execute a ```curl``` command every 2 minutes. It may be usefull when you need to call any simple **webcron** or **webhook** URL like with [EasyCron](https://www.easycron.com)

* The last one will forget and prune all restic snapshot older than 7 days every day at 3 AM. It use the secret environnement variable ```$REPO__FILE``` for telling to restic the repository to use and a password file ```/configs/passwd``` mounted in the container.

## Docker Secrets

As an alternative to passing sensitive information via environment variables, `__FILE` may be appended to any environment variables, causing the job to load the values for those variables from files present in the container. In particular, this can be used to load passwords from Docker secrets stored in `/run/secrets/<secret_name>` files.

## Docker compose

See the docker swarm [examples](https://github.com/pfillion/mobycron/tree/master/examples) in the source code or below.

```yml
version: '3.6'
services:
  cron:
    image: pfillion/mobycron:latest
    environment:
      MOBYCRON_DOCKER_MODE: 'true'
      MOBYCRON_PARSE_SECOND: 'true'
      CRON_TZ: America/New_York
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock"
    deploy:
      mode: global

  busybox:
    image: busybox:latest
    command: echo 'Hello World!!'
    labels: 
      mobycron.schedule: "*/30 * * * * *"
      mobycron.action: "start"
    deploy:
      replicas: 6
      restart_policy:
        condition: none
```

## Authors

* [pfillion](https://github.com/pfillion)

## License

MIT