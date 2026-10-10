REGISTRY := registry.internal.taucetifive.com
IMAGE    := book-scrape
TAG      ?= latest

.PHONY: build push tag all

build:
	docker buildx build --platform linux/amd64 -t $(REGISTRY)/$(IMAGE):$(TAG) .

push:
	docker buildx build --platform linux/amd64 --push -t $(REGISTRY)/$(IMAGE):$(TAG) .

tag:
	git tag -f $(TAG)
	git push -f origin $(TAG)

all: tag push
