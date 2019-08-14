REGISTRY ?= axot

IMAGE_NAME ?= dnstress

VERSION ?= $(shell git describe --tags --always --dirty)

ifneq (,$(findstring dirty,$(VERSION)))
	VERSION := latest
endif

.PHONY: all
all: build

.PHONY: build
build:
	docker build -t $(REGISTRY)/$(IMAGE_NAME):$(VERSION) .

.PHONY: build_bin
build_bin:
	go build -tags netgo -o dnstress

.PHONY: push
push:
	docker push $(REGISTRY)/$(IMAGE_NAME):$(VERSION)
