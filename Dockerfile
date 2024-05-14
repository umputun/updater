FROM umputun/baseimage:buildgo-v1.13.0  as build

ARG GIT_BRANCH
ARG GITHUB_SHA
ARG CI

ENV CGO_ENABLED=0

ADD . /build
WORKDIR /build

RUN apk add --no-cache --update git tzdata ca-certificates

RUN \
    if [ -z "$CI" ] ; then \
    echo "runs outside of CI" && version=$(git rev-parse --abbrev-ref HEAD)-$(git log -1 --format=%h)-$(date +%Y%m%dT%H:%M:%S); \
    else version=${GIT_BRANCH}-${GITHUB_SHA:0:7}-$(date +%Y%m%dT%H:%M:%S); fi && \
    echo "version=$version" && \
    cd app && go build -o /build/updater -ldflags "-X main.revision=${version} -s -w"


FROM ghcr.io/umputun/baseimage/app:v1.13.0
LABEL org.opencontainers.image.source="https://github.com/umputun/updater"

RUN apk add docker openssh-client

RUN \
  mkdir -p /home/app/.ssh && \
  echo "StrictHostKeyChecking=no" > /home/app/.ssh/config && \
  echo "LogLevel=quiet" >> /home/app/.ssh/config && \
  chown -R app:app /home/app/.ssh/ && \
  chmod 600 /home/app/.ssh/* && \
  chmod 700 /home/app/.ssh

COPY --from=build /build/updater /srv/updater
WORKDIR /srv
CMD ["/srv/updater"]
