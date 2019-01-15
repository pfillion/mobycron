SHELL = /bin/sh
.SUFFIXES:
.SUFFIXES: .c .o
.PHONY: help
.DEFAULT_GOAL := help

# Go parameters
ROOT_FOLDER=$(shell pwd)
BIN_FOLDER=$(ROOT_FOLDER)/bin
APP_FOLDER=$(ROOT_FOLDER)/cmd/mobycron
APP_NAME=mobycron
GOOS=linux 
GOARCH=amd64

# Docker parameters
NS ?= pfillion
VERSION ?= latest
IMAGE_NAME ?= mobycron
CONTAINER_NAME ?= mobycron
CONTAINER_INSTANCE ?= default
VCS_REF=$(shell git rev-parse --short HEAD)
BUILD_DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%S")

help: ## Show the Makefile help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

go-get: ## Get external packages
	go get -u -v github.com/golang/lint/golint
	go get -u -v github.com/robfig/cron
	go get -u -v github.com/sirupsen/logrus
	go get -u -v github.com/pkg/errors
	go get -u -v github.com/spf13/afero
	go get -u -v gotest.tools/assert
	go get -u -v github.com/golang/mock/gomock

go-mock: ## Generate mock file
	mockgen -source=$(ROOT_FOLDER)/cron/cron.go -destination=$(ROOT_FOLDER)/cron/cron_mock.go -package=cron

go-build: go-get ## Build go app
	golint -set_exit_status ./...
	go vet -v ./...
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o $(BIN_FOLDER)/$(APP_NAME) -v $(APP_FOLDER)

go-rebuild: go-clean go-build ## Rebuild go app

go-test: ## Test go app
	go test -cover -v ./...

go-clean: ## Clean go app
	go clean -cache -testcache
	rm -f $(BIN_FOLDER)/$(APP_NAME)

go-run: ## Run go app
	$(BIN_FOLDER)/$(APP_NAME)

docker-build: ## Build the image form Dockerfile
	docker build \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg VCS_REF=$(VCS_REF) \
		--build-arg VERSION=$(VERSION) \
		-t $(NS)/$(IMAGE_NAME):$(VERSION) -f Dockerfile .

docker-rebuild: ## Rebuild the image form Dockerfile
	docker build  \
		--build-arg BUILD_DATE=$(BUILD_DATE) \
		--build-arg VCS_REF=$(VCS_REF) \
		--build-arg VERSION=$(VERSION) \
		--no-cache -t $(NS)/$(IMAGE_NAME):$(VERSION) -f Dockerfile .

docker-push: ## Push the image to a registry
ifdef DOCKER_USERNAME
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin
endif
	docker push $(NS)/$(IMAGE_NAME):$(VERSION)
    
docker-shell: ## Run shell command in the container
	docker run --rm --name $(CONTAINER_NAME)-$(CONTAINER_INSTANCE) -i -t $(PORTS) $(VOLUMES) $(ENV) $(NS)/$(IMAGE_NAME):$(VERSION) /bin/sh

docker-run: ## Run the container
	docker run --rm --name $(CONTAINER_NAME)-$(CONTAINER_INSTANCE) $(PORTS) $(VOLUMES) $(ENV) $(NS)/$(IMAGE_NAME):$(VERSION)

docker-start: ## Run the container in background
	docker run -d --name $(CONTAINER_NAME)-$(CONTAINER_INSTANCE) $(PORTS) $(VOLUMES) $(ENV) $(NS)/$(IMAGE_NAME):$(VERSION)

docker-stop: ## Stop the container
	docker stop $(CONTAINER_NAME)-$(CONTAINER_INSTANCE)

docker-rm: ## Remove the container
	docker rm $(CONTAINER_NAME)-$(CONTAINER_INSTANCE)

docker-test: ## Run docker container tests
	container-structure-test test --image $(NS)/$(IMAGE_NAME):$(VERSION) --config tests/config.yaml

build: go-build docker-build ## Build all

rebuild: go-rebuild docker-rebuild ## Rebuild all

run: docker-run ## Run all

test: go-test docker-test ## Run all tests

release: build test docker-push ## Build and push the image to a registry