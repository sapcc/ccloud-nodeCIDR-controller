IMAGE	 := keppel.eu-de-1.cloud.sap/ccloud/ccloud-nodecidr-controller
VERSION  ?= 2.0.0

.PHONY: all

all: build push

build:
	docker build -t $(IMAGE):$(VERSION) .

push: build
	docker push ${IMAGE}:${VERSION}

docker-push-mac:
	docker buildx build  --platform linux/amd64 . -t ${IMAGE}:${VERSION} --push
