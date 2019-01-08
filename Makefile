IMAGE ?= docker.medium.build/brigade
COMMIT ?= $(shell git rev-parse --short HEAD)
TAG ?= $(shell date -u +%Y%m%d-%H%M%S)-$(COMMIT)

.PHONY: all brigade image push

all: image

brigade:
	go build

image:
	docker build -t $(IMAGE):$(TAG) .

push: image
	docker push $(IMAGE):$(TAG)
