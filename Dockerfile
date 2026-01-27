FROM umputun/baseimage:buildgo-latest AS build

ARG GIT_BRANCH
ARG GITHUB_SHA
ARG CI

ENV CGO_ENABLED=0

ADD . /build/updater
WORKDIR /build/updater

RUN \
    if [ -z "$CI" ] ; then \
    echo "runs outside of CI" && version=$(/script/git-rev.sh); \
    else version=${GIT_BRANCH}-${GITHUB_SHA:0:7}-$(date +%Y%m%dT%H:%M:%S); fi && \
    echo "version=$version" && \
    cd app && go build -o /build/updater/updater -ldflags "-X main.revision=${version} -s -w"


FROM umputun/baseimage:app-latest

LABEL org.opencontainers.image.source="https://github.com/umputun/updater"

RUN apk add docker openssh-client

RUN \
  mkdir -p /home/app/.ssh && \
  echo "StrictHostKeyChecking=no" > /home/app/.ssh/config && \
  echo "LogLevel=quiet" >> /home/app/.ssh/config && \
  chown -R app:app /home/app/.ssh/ && \
  chmod 600 /home/app/.ssh/* && \
  chmod 700 /home/app/.ssh

COPY --from=build /build/updater/updater /srv/updater
WORKDIR /srv
ENTRYPOINT ["/init.sh"]
CMD ["/srv/updater"]
