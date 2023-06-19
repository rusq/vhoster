SHELL=/bin/sh

REPO=ffffuuu/vhoster:latest

gateway:
	go build -o $@ -ldflags="-s -w"  ./cmd/$@
.PHONY: gateway # unconditional build.

docker:
	docker build . -t $(REPO)
.PHONY: docker

push:
	docker push $(REPO)
.PHONY: push

## sundry

testserver:
	go build -o $@ -ldflags="-s -w"  ./cmd/server
.PHONY: testserver # unconditional build.
