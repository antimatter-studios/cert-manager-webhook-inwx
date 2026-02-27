GO ?= $(shell which go)
OS ?= $(shell $(GO) env GOOS)
ARCH ?= $(shell $(GO) env GOARCH)

IMAGE_NAME := "cert-manager-webhook-inwx"
IMAGE_TAG := "latest"

OUT := $(shell pwd)/_out

HELM_FILES := $(shell find deploy/cert-manager-webhook-inwx)

.PHONY: build
build:
	$(GO) build -o $(OUT)/webhook .

.PHONY: test
test:
	$(GO) test -v ./...

.PHONY: docker-build
docker-build:
	docker build -t "$(IMAGE_NAME):$(IMAGE_TAG)" .

.PHONY: clean
clean:
	rm -rf $(OUT)

$(OUT):
	mkdir -p $(OUT)

.PHONY: rendered-manifest.yaml
rendered-manifest.yaml: $(OUT)/rendered-manifest.yaml

$(OUT)/rendered-manifest.yaml: $(HELM_FILES) | $(OUT)
	helm template \
		--name-template cert-manager-webhook-inwx \
		--set image.repository=$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG) \
		deploy/cert-manager-webhook-inwx > $@
