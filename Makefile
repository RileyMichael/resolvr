PROJECT=$(shell basename $(PWD))
GIT_BRANCH=$(shell git branch | grep \* | cut -d ' ' -f2-)
TAG=$(PROJECT)_$(GIT_BRANCH)

.PHONY: help run lint docker-run docker-stop docker-exec docker-logs

help: ## display all Make targets
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% 0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

run: ## run locally
	go run ./cmd/resolvr

lint: ## lint
	@go mod tidy
	@gofmt -s -w ./

docker-run: control/docker-build docker-stop ## run containerized
	docker run -d -p 127.0.0.1:53:53/udp --name $(TAG) $(TAG)

docker-stop: ## stop container
	-docker rm -f $(TAG)

docker-exec: ## interactive shell into container
	docker exec -it $(TAG) sh

docker-logs: ## stdout from container
	docker logs $(TAG)

control/docker-build: control/control Dockerfile .dockerignore go.mod go.sum $(shell find -type f -name "*.go")
	docker build -t $(TAG) .
	touch "$@"

control/control:
	mkdir -p control
	touch "$@"