# mobycron

[![Build Status](https://drone.pfillion.com/api/badges/pfillion/mobycron/status.svg?branch=master)](https://drone.pfillion.com/pfillion/mobycron)
[![Go Report Card](https://goreportcard.com/badge/github.com/pfillion/mobycron)](https://goreportcard.com/report/github.com/pfillion/mobycron)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/pfillion/mobycron)](https://golang.org/ "The Go Programming Language")
![GitHub](https://img.shields.io/github/license/pfillion/mobycron)
[![GitHub last commit](https://img.shields.io/github/last-commit/pfillion/mobycron?logo=github)](https://github.com/pfillion/mobycron "GitHub projet")

[![Docker Image Version (tag latest semver)](https://img.shields.io/docker/v/pfillion/mobycron/latest?logo=docker)](https://hub.docker.com/r/pfillion/mobycron "Docker Hub Repository")
[![Docker Image Size (latest by date)](https://img.shields.io/docker/image-size/pfillion/mobycron/latest?logo=docker)](https://hub.docker.com/r/pfillion/mobycron "Docker Hub Repository")

A simple cron deamon for docker written in go. It use the [robfig cron library v3](https://github.com/robfig/cron/tree/v3) engine and all cron jobs can be confgurated by two ways. The first mode will perform actions on others Docker containers based on cron schedule. The second mode is by a JSON file acting like a contab file.

## Versions

* [latest](https://github.com/pfillion/mobycron/tree/master) available as ```pfillion/mobycron:latest``` at [Docker Hub](https://hub.docker.com/r/pfillion/mobycron/)

## Go packages

You can use the [mobycron library](https://github.com/pfillion/mobycron) directly by importing the package ```github.com/pfillion/mobycron/pkg/cron``` directly in your project.

## Tools included in the docker image

* bash
* ca-certificates
* curl
* tzdata

## Environnement variables

Cron job support any environnement variables specified by docker and replace it by the real value before executing the command.

```MOBYCRON_DOCKER_MODE``` definde how ```mobycron``` will interact with Docker. The value possible is ```none``` (default), ```container``` or ```swarm```. When mobycron is up and running in ```container``` or ```swarm``` mode, it watches Docker socket and try to find any containers with label ```mobycron.schedule``` and add them to the crontab based on the schedule. Go to [docker mode](#docker-mode) section for more detail with this mode.

```MOBYCRON_PARSE_SECOND``` is false by default. When activate, schedule accept an optional seconds field at the beginning of the cron spec. This is non-standard and has led to a lot of confusion. The new default parser conforms to the standard as described by the [Cron wikipedia page.](https://en.wikipedia.org/wiki/Cron)

```MOBYCRON_CONFIG_FILE``` is file path to schedule all job like a crontab file. Go to [configuration file](#configuration-file) section for more detail with this mode.

```TZ``` configure local time zone of the container.

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

```CRON_TZ``` is now the recommended way to specify the timezone of a single schedule, which is sanctioned by the specification. The legacy ```TZ``` prefix will continue to be supported since it is unambiguous and easy to do so.

The ```container``` mode is the classic Docker mode. Labels can be applied are:

* ```mobycron.action``` is requied and indicate wich action must be performed on the container. Possible choices are ```start```, ```restart```, ```stop``` or ```exec```.
* ```mobycron.command``` specifie the commande line to execute and is requied when the action is ```exec```.
* ```mobycron.timeout``` override the default 10 second timeout to do the action.

The second mode is ```swarm``` mode. Docker need to be in a swarm node. Label can be applied is:

* ```mobycron.action``` is required and indicate which action must be performed on the container. Possible choices are only ```update``` due to the mechanic of services in Docker Swarm.

### Examples

```sh
# Start mobycron in container mode
> docker run -d -e MOBYCRON_DOCKER_MODE=container -v /var/run/docker.sock:/var/run/docker.sock pfillion/mobycron:latest

# Start the job container and print date every minute
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
            "https://example.com"
        ]
    }
]
```

This file will schedule 2 cron job.

* The first one will replace ```$NAME``` by the environnement variable configured in the container and print ```Hello``` + ```$NAME``` every minutes.
* The second will execute a ```curl``` command every 2 minutes. It may be usefull when you need to call any simple **webcron** or **webhook** URL like with [EasyCron](https://www.easycron.com)

## Docker Secrets

As an alternative to passing sensitive information via environment variables, `__FILE` may be appended to any environment variables, causing the job to load the values for those variables from files present in the container. In particular, this can be used to load passwords from Docker secrets stored in `/run/secrets/<secret_name>` files.

## Docker compose

See the docker swarm [examples](https://github.com/pfillion/mobycron/tree/master/examples) in the source code or below.

```yml
version: '3.7'
services:
  cron:
    image: pfillion/mobycron:latest
    environment:
      MOBYCRON_DOCKER_MODE: 'swarm'
      MOBYCRON_PARSE_SECOND: 'true'
      TZ: America/New_York
    volumes:
      - "/var/run/docker.sock:/var/run/docker.sock"
    deploy:
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
        window: 30s
      replicas: 1

  busybox:
    image: busybox:latest
    command: echo 'Hello World!!'
    deploy:
      restart_policy:
        condition: none
      replicas: 6
      labels:
        mobycron.schedule: "*/30 * * * * *"
        mobycron.action: "update"
```

## Authors

* [pfillion](https://github.com/pfillion)

## License

MIT
