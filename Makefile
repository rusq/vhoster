SHELL=/bin/sh

REPO=ffffuuu/vhoster
TAG_LATEST=latest

TAG=$(shell git describe --tags)
# extract version from tag
VERSION=$(shell echo $(TAG) | sed -e 's/^v/v/' -e 's/-.*//')

gateway:
	go build -o $@ -ldflags="-s -w"  ./cmd/$@
.PHONY: gateway # unconditional build.

test:
	go test -v ./...
.PHONY: test

docker:
	docker build . -t $(REPO):$(TAG_LATEST)
.PHONY: docker

push: docker
	docker push $(REPO):$(TAG_LATEST)
.PHONY: push

push-stable: push
	docker tag $(REPO):$(TAG_LATEST) $(REPO):$(VERSION)
	docker push $(REPO):$(VERSION)
.PHONY: push-stable


## sundry

testserver:
	go build -o $@ -ldflags="-s -w"  ./cmd/testserver
.PHONY: testserver # unconditional build.
