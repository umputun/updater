B=$(shell git rev-parse --abbrev-ref HEAD)
BRANCH=$(subst /,-,$(B))
GITREV=$(shell git describe --abbrev=7 --always --tags)
REV=$(GITREV)-$(BRANCH)-$(shell date +%Y%m%d-%H:%M:%S)

docker:
	docker build -t umputun/updater:master --progress=plain .

dist:
	goreleaser --snapshot --clean

race_test:
	cd app && go test -race -timeout=60s -count 1 ./...

build: info
	cd app && CGO_ENABLED=0 go build -ldflags "-X main.revision=$(REV) -s -w" -o /tmp/updater

site:
	@rm -f  site/public/*
	docker build -f Dockerfile.site --progress=plain -t updater.site .
	docker run -d --name=updater.site updater.site
	docker cp updater.site:/build/public site/
	docker rm -f updater.site

info:
	- @echo "revision $(REV)"

.PHONY: dist docker race_test build info site
