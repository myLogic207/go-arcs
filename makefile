TAG := v0.1.0
GIT_HOST := git.mylogic.dev
GIT_PATH := homelab/go-arcs

build: build-client build-server

publish: publish-client publish-server

run: build-server
	docker run go-arcs/server:$(TAG)

.PHONY: buf-generate
buf-generate: build-buf
	docker run -v $(CURDIR):/api $(GIT_HOST)/$(GIT_PATH)/buf-generate-container:$(TAG)

.PHONY: build-buf
build-buf:
	docker build -t $(GIT_HOST)/$(GIT_PATH)/buf-generate-container:$(TAG) -f ./dockerfile.buf .

.PHONY: build-server
build-server: buf-generate
	docker build -t $(GIT_HOST)/$(GIT_PATH)/server:$(TAG) -f ./dockerfile.server . --provenance=false

.PHONY: publish-server
publish-server: docker-login build-server
	docker push $(GIT_HOST)/$(GIT_PATH)/server:$(TAG)

.PHONY: build-client
build-client: buf-generate
	docker build -t $(GIT_HOST)/$(GIT_PATH)/client:$(TAG) -f ./dockerfile.client . --provenance=false

.PHONY: publish-client
publish-client: docker-login build-client
	docker push $(GIT_HOST)/$(GIT_PATH)/client:$(TAG)

.PHONY: docker-login
docker-login:
	docker login $(GIT_HOST)
