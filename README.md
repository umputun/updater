<div align="center">
  <img class="logo" src="https://raw.githubusercontent.com/umputun/updater/master/site/src/logo-bg.svg" width="355px" height="142px" alt="Updater | Simple Remote Updater"/>
</div>

Updater is a simple web-hook-based receiver executing things via HTTP requests and invoking remote updates without exposing any sensitive info, like ssh keys, passwords, etc. The updater is usually called from CI/CD system (i.e., Github action), and the actual http call looks like `curl https://<server>/update/<task-name>/<access-key>`

List of tasks defined in the configuration file, and each task has its custom section for the command.

---

<div align="center">

[![Build Status](https://github.com/umputun/updater/workflows/build/badge.svg)](https://github.com/umputun/updater/actions) &nbsp;[![Coverage Status](https://coveralls.io/repos/github/umputun/updater/badge.svg?branch=master)](https://coveralls.io/github/umputun/updater?branch=master) 

</div>

Example of `updater.yml`:

```yaml
tasks:

  - name: remark42-site
    command: |
      echo "update remark42-site"
      docker pull ghcr.io/umputun/remark24-site:master
      docker rm -f remark42-site
      docker run -d --name=remark42-site

  - name: feed-master
    command: |
      echo "update feed-master"
      docker pull umputun/feed-master
      docker restart feed-master
```

By default the update call synchronous but can be switched to non-blocking mode with `async` query parameter, i.e. `curl https://example.com/update/remark42-site/super-seecret-key?async=1`

## Install

Updater distributed as multi-arch docker container as well as binary files for multiple platforms. Container has the docker client preinstalled to allow the typical "docker pull & docker restart" update sequence.

Containers available on both [github container registry (ghcr)](https://github.com/umputun/updater/pkgs/container/updater) and [docker hub](https://hub.docker.com/repository/docker/umputun/updater)


This is an example of updater usage inside of the docker compose. It uses [reproxy](https://reproxy.io) as the reversed proxy, but any other (nginx, apache, haproxy, etc) can be used as well.

```yaml
services:
  
  reproxy:
    image: ghcr.io/umputun/reproxy:master
    restart: always
    hostname: reproxy
    container_name: reproxy
    logging: &default_logging
      driver: json-file
      options:
        max-size: "10m"
        max-file: "5"
    ports:
      - "80:8080"
      - "443:8443"
    environment:
      - TZ=America/Chicago
      - DOCKER_ENABLED=true
      - SSL_TYPE=auto
      - SSL_ACME_EMAIL=umputun@gmail.com
      - SSL_ACME_FQDN=jess.umputun.com,echo.umputun.com
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock:ro
      - ./var/ssl:/srv/var/ssl

  echo:
    image: ghcr.io/umputun/echo-http
    hostname: echo
    container_name: echo
    command: --message="echo echo 123"
    logging: *default_logging
    labels:
      reproxy.server: 'echo.umputun.com'
      reproxy.route: '^/(.*)'

  updater:
    image: ghcr.io/umputun/updater:master
    container_name: "updater"
    hostname: "updater"
    restart: always
    logging: *default_logging
    environment:
      - LISTEN=0.0.0.0:8080
      - KEY=super-secret-password
      - CONF=/srv/etc/updater.yml
    ports:
      - "8080"
    volumes:
      - ./etc:/srv/etc
      - /var/run/docker.sock:/var/run/docker.sock
    labels:
      reproxy.server: 'jess.umputun.com'
      reproxy.route: '^/(.*)'
```

## Working with docker-compose

For a simple container, started with all the parameters manually, the typical update sequence can be as simple as "kill container and recreate it", however docker compose-based container can be a little trickier. If user runs updater directly on the host (not from a container) the update command can be as trivial as "docker-compose pull <service> && docker-compose up -d <service>". In case if updater runs from a container the simplest way to do the same is "ssh user@bridge-ip docker-compose ...". To simplify the process the openssh-client already preinstalled. 

This is an example of ssh-based `updater.yml`

```yaml
tasks:

  - name: remark42-site
    command: |
      echo "update remark42-site with compose"
      ssh app@172.17.42.1 "cd /srv && docker-compose pull remark42-site && docker-compose up -d remark42-site"

  - name: reproxy-site
    command: |
      echo "update reproxy-site"
      ssh app@172.17.42.1 "cd /srv && docker-compose pull reproxy-site && docker-compose up -d reproxy-site"

```

## Other use cases

The main goal of this utility is to update containers; however, all it does is the remote activation of predefined commands. Such command can do anything user like, not just "docker pull && docker restart." For instance, it can be used to schedule remote jobs from some central orchestrator, run remote cleanup jobs, etc.

## All parameters

```
  -f, --file=   config file (default: updater.yml) [$CONF]
  -l, --listen= listen on host:port (default: localhost:8080) [$LISTEN]
  -k, --key=    secret key [$KEY]
  -b, --batch   batch mode for multi-line scripts
      --dbg     show debug info

Help Options:
  -h, --help    Show this help message

```
