TAG := v0.1.3
GIT_HOST := git.mylogic.dev
GIT_PATH := homelab/go-arcs

_VERSION   := $(subst ., ,$(TAG))
TAG_MAJOR := $(word 1,$(_VERSION))
TAG_MINOR := $(TAG_MAJOR).$(word 2,$(_VERSION))

.PHONY: build
build: build-client build-server

.PHONY: publish
publish: publish-client publish-server

.PHONY: run
run: build-server
	docker run go-arcs/server:$(TAG)

.PHONY: buf-generate
buf-generate: build-buf
	docker run -v $(CURDIR):/api $(GIT_HOST)/$(GIT_PATH)/buf-generate-container:$(TAG)

.PHONY: build-buf
build-buf:
	docker build -t go-arcs-buf-generate \
	-t $(GIT_HOST)/$(GIT_PATH)/buf-generate-container:$(TAG) \
	-t $(GIT_HOST)/$(GIT_PATH)/buf-generate-container:$(TAG_MAJOR) \
	-t $(GIT_HOST)/$(GIT_PATH)/buf-generate-container:$(TAG_MINOR) \
	-t $(GIT_HOST)/$(GIT_PATH)/buf-generate-container:latest \
	-f ./dockerfile.buf .

.PHONY: build-server
build-server: buf-generate
	docker build -t go-arcs-server \
	-t $(GIT_HOST)/$(GIT_PATH)/server:$(TAG) \
	-t $(GIT_HOST)/$(GIT_PATH)/server:$(TAG_MAJOR) \
	-t $(GIT_HOST)/$(GIT_PATH)/server:$(TAG_MINOR) \
	-t $(GIT_HOST)/$(GIT_PATH)/server:latest \
	-f ./dockerfile.server . --provenance=false

.PHONY: publish-server
publish-server: docker-login build-server
	docker push $(GIT_HOST)/$(GIT_PATH)/server:$(TAG)
	docker push $(GIT_HOST)/$(GIT_PATH)/server:$(TAG_MAJOR)
	docker push $(GIT_HOST)/$(GIT_PATH)/server:$(TAG_MINOR)
	docker push $(GIT_HOST)/$(GIT_PATH)/server:latest

.PHONY: build-client
build-client: buf-generate
	docker build -t go-arcs-client \
	-t $(GIT_HOST)/$(GIT_PATH)/client:$(TAG) \
	-t $(GIT_HOST)/$(GIT_PATH)/client:$(TAG_MAJOR) \
	-t $(GIT_HOST)/$(GIT_PATH)/client:$(TAG_MINOR) \
	-t $(GIT_HOST)/$(GIT_PATH)/client:latest \
	-f ./dockerfile.client . --provenance=false

.PHONY: publish-client
publish-client: docker-login build-client
	docker push $(GIT_HOST)/$(GIT_PATH)/client:$(TAG)
	docker push $(GIT_HOST)/$(GIT_PATH)/client:$(TAG_MAJOR)
	docker push $(GIT_HOST)/$(GIT_PATH)/client:$(TAG_MINOR)
	docker push $(GIT_HOST)/$(GIT_PATH)/client:latest

.PHONY: docker-login
docker-login:
	docker login $(GIT_HOST)
