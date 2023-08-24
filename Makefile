all: build

export BIN_TIMESTAMP ?=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
SOURCE_GIT_TAG :=$(shell git describe --tags --always --abbrev=7 --match 'v*')
GOARCH := $(shell go env GOARCH)
SRC_ROOT :=$(shell pwd)
DOCKERFILE := $(SRC_ROOT)/build/Dockerfile

ifdef CTR_CMD
	CTR_CMD := $(CTR_CMD)
else
	CTR_CMD :=$(or $(shell which podman 2>/dev/null), $(shell which docker 2>/dev/null))
endif

ifdef IMAGE_REPO
	IMAGE_REPO := $(IMAGE_REPO)
else
	IMAGE_REPO := quay.io/sustainability
endif

ifdef IMAGE_TAG
	IMAGE_TAG := $(IMAGE_TAG)
else
	IMAGE_TAG := v0.1
endif

.PHONY: build
build: tidy-vendor
	@mkdir -p bin/
	CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o bin/label-exporter ./cmd/exporter.go

### toolkit ###
tidy-vendor:
	go mod tidy -v
	go mod vendor

build_image:
	@if [ -z '$(CTR_CMD)' ] ; then echo '!! ERROR: containerized builds require podman||docker CLI, none found $$PATH' >&2 && exit 1; fi

	$(CTR_CMD) build -t $(IMAGE_REPO)/label-exporter:$(IMAGE_TAG) \
		-f $(DOCKERFILE) \
		--network host \
		--build-arg SOURCE_GIT_TAG=$(SOURCE_GIT_TAG) \
		--build-arg BIN_TIMESTAMP=$(BIN_TIMESTAMP) \
		--platform="linux/$(GOARCH)" \
		.
