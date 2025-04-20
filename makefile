run: build
	docker run go-arcs-container

buf-generate-container:
	docker build --target buf-build -t buf-generate-container .

.PHONY: buf-generate
buf-generate: buf-generate-container
	docker run -v $(CURDIR):/api buf-generate-container

.PHONY: build
build:
	docker build -t go-arcs-container .