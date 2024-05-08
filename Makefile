B=$(shell git rev-parse --abbrev-ref HEAD)
BRANCH=$(subst /,-,$(B))
GITREV=$(shell git describe --abbrev=7 --always --tags)
REV=$(GITREV)-$(BRANCH)-$(shell date +%Y%m%d-%H:%M:%S)

docker:
	docker build -t umputun/updater:master --progress=plain .

dist:
	- @mkdir -p dist
	docker build -f Dockerfile.artifacts --progress=plain -t updater.bin .
	- @docker rm -f updater.bin 2>/dev/null || exit 0
	docker run -d --name=updater.bin updater.bin
	docker cp updater.bin:/artifacts dist/
	docker rm -f updater.bin

race_test:
	cd app && go test -race -timeout=60s -count 1 ./...

build: info
	- cd app && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-X main.revision=$(REV) -s -w" -o ../dist/updater

site:
	@rm -f  site/public/*
	docker build -f Dockerfile.site --progress=plain -t updater.site .
	docker run -d --name=updater.site updater.site
	docker cp updater.site:/build/public site/
	docker rm -f updater.site
	rsync -avz -e "ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" --progress ./site/public/ updater.io:/srv/www/reproxy.io

info:
	- @echo "revision $(REV)"

.PHONY: dist docker race_test bin info site build_site
