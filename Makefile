SHELL = /bin/sh
.PHONY: help
.DEFAULT_GOAL := help

ifeq ($(MODE_LOCAL),true)
	GIT_CONFIG_GLOBAL := $(shell git config --global --add safe.directory /go/src/github.com/pfillion/mobycron > /dev/null)
endif

# Version
DESCRIBE           := $(shell git describe --match "v*" --always --tags)
DESCRIBE_PARTS     := $(subst -, ,$(DESCRIBE))

VERSION_TAG        := $(word 1,$(DESCRIBE_PARTS))
COMMITS_SINCE_TAG  := $(word 2,$(DESCRIBE_PARTS))

VERSION            := $(subst v,,$(VERSION_TAG))
VERSION_PARTS      := $(subst ., ,$(VERSION))
VERSION_ALPINE     := 3.23

MAJOR              := $(word 1,$(VERSION_PARTS))
MINOR              := $(word 2,$(VERSION_PARTS))
MICRO              := $(word 3,$(VERSION_PARTS))

NEXT_MICRO          = $(shell echo $$(($(MICRO)+$(COMMITS_SINCE_TAG))))

ifeq ($(strip $(COMMITS_SINCE_TAG)),)
CURRENT_VERSION_MICRO := $(MAJOR).$(MINOR).$(MICRO)
else
CURRENT_VERSION_MICRO := $(MAJOR).$(MINOR).$(NEXT_MICRO)
endif
CURRENT_VERSION_MINOR := $(MAJOR).$(MINOR)
CURRENT_VERSION_MAJOR := $(MAJOR)

DATE                = $(shell date -u +"%Y-%m-%dT%H:%M:%S")
COMMIT             := $(shell git rev-parse HEAD)
AUTHOR             := $(firstword $(subst @, ,$(shell git show --format="%aE" $(COMMIT))))

# Bats parameters
TEST_FOLDER ?= $(shell pwd)/tests

# Go parameters
ROOT_FOLDER=$(shell pwd)
BIN_FOLDER=$(ROOT_FOLDER)/bin
APP_FOLDER=$(ROOT_FOLDER)/cmd/mobycron
APP_NAME=mobycron
GOOS=linux 
GOARCH=amd64

# Docker parameters
NS ?= pfillion
IMAGE_NAME ?= mobycron
CONTAINER_NAME ?= mobycron
CONTAINER_INSTANCE ?= default

help: ## Show the Makefile help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

version: ## Show all versionning infos
	@echo CURRENT_VERSION_MICRO="$(CURRENT_VERSION_MICRO)"
	@echo CURRENT_VERSION_MINOR="$(CURRENT_VERSION_MINOR)"
	@echo CURRENT_VERSION_MAJOR="$(CURRENT_VERSION_MAJOR)"
	@echo VERSION_ALPINE="$(VERSION_ALPINE)"
	@echo DATE="$(DATE)"
	@echo COMMIT="$(COMMIT)"
	@echo AUTHOR="$(AUTHOR)"
	@echo DESCRIBE="$(DESCRIBE)"
	@echo COMMITS_SINCE_TAG="$(COMMITS_SINCE_TAG)"

bats-test: ## Test bash scripts
	bats $(TEST_FOLDER)

go-install:
	go install golang.org/x/lint/golint@latest
	go install github.com/golang/mock/mockgen@latest

go-mock: ## Generate mock file
	mockgen -source=$(ROOT_FOLDER)/cmd/mobycron/main.go -destination=$(ROOT_FOLDER)/cmd/mobycron/main_mock.go -package=main
	mockgen -source=$(ROOT_FOLDER)/pkg/cron/interface.go -destination=$(ROOT_FOLDER)/pkg/cron/interface_mock.go -package=cron

go-build: ## Build go app
	golint -set_exit_status ./...
	go vet -v ./...
	GOOS=${GOOS} GOARCH=${GOARCH} go build -o $(BIN_FOLDER)/$(APP_NAME) -v $(APP_FOLDER)
	chmod 755 $(BIN_FOLDER)/$(APP_NAME)

go-rebuild: go-clean go-build ## Rebuild go app

go-test: ## Test go app
	go test -cover -v ./...

go-clean: ## Clean go app
	go clean -cache -testcache -fuzzcache
	rm -f $(BIN_FOLDER)/$(APP_NAME)

go-update-mod: ## Run go module cleanup
	go clean -modcache
	go get -u -v ./...
	go mod tidy -v

go-run: ## Run go app
	$(BIN_FOLDER)/$(APP_NAME)

docker-build: ## Build the image form Dockerfile
	docker build \
		--build-arg DATE=$(DATE) \
		--build-arg CURRENT_VERSION_MICRO=$(CURRENT_VERSION_MICRO) \
		--build-arg VERSION_ALPINE=$(VERSION_ALPINE) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg AUTHOR=$(AUTHOR) \
		-t $(NS)/$(IMAGE_NAME):$(CURRENT_VERSION_MICRO) \
		-t $(NS)/$(IMAGE_NAME):$(CURRENT_VERSION_MINOR) \
		-t $(NS)/$(IMAGE_NAME):$(CURRENT_VERSION_MAJOR) \
		-t $(NS)/$(IMAGE_NAME):latest \
		-f Dockerfile .

docker-rebuild: ## Rebuild the image form Dockerfile
	docker build  \
		--build-arg DATE=$(DATE) \
		--build-arg CURRENT_VERSION_MICRO=$(CURRENT_VERSION_MICRO) \
		--build-arg VERSION_ALPINE=$(VERSION_ALPINE) \
		--build-arg COMMIT=$(COMMIT) \
		--build-arg AUTHOR=$(AUTHOR) \
		-t $(NS)/$(IMAGE_NAME):$(CURRENT_VERSION_MICRO) \
		-t $(NS)/$(IMAGE_NAME):$(CURRENT_VERSION_MINOR) \
		-t $(NS)/$(IMAGE_NAME):$(CURRENT_VERSION_MAJOR) \
		-t $(NS)/$(IMAGE_NAME):latest \
		--no-cache -f Dockerfile .

docker-push: ## Push the image to a registry
ifdef DOCKER_USERNAME
	@echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin
endif
	docker push $(NS)/$(IMAGE_NAME):$(CURRENT_VERSION_MICRO)
	docker push $(NS)/$(IMAGE_NAME):$(CURRENT_VERSION_MINOR)
	docker push $(NS)/$(IMAGE_NAME):$(CURRENT_VERSION_MAJOR)
	docker push $(NS)/$(IMAGE_NAME):latest
    
docker-shell: ## Run shell command in the container
	docker run --rm --name $(CONTAINER_NAME)-$(CONTAINER_INSTANCE) -it --entrypoint "" $(PORTS) $(VOLUMES) $(ENV) $(NS)/$(IMAGE_NAME):$(CURRENT_VERSION_MICRO) /bin/sh

docker-run: ## Run the container
	docker run --rm --name $(CONTAINER_NAME)-$(CONTAINER_INSTANCE) $(PORTS) -v /var/run/docker.sock:/var/run/docker.sock -v $(ROOT_FOLDER)/tests/configs/config.json:/configs/config.json $(VOLUMES) $(ENV) $(NS)/$(IMAGE_NAME):$(CURRENT_VERSION_MICRO)

docker-start: ## Run the container in background
	docker run -d --name $(CONTAINER_NAME)-$(CONTAINER_INSTANCE) $(PORTS) $(VOLUMES) $(ENV) $(NS)/$(IMAGE_NAME):$(CURRENT_VERSION_MICRO)

docker-stop: ## Stop the container
	docker stop $(CONTAINER_NAME)-$(CONTAINER_INSTANCE)

docker-rm: ## Remove the container
	docker rm $(CONTAINER_NAME)-$(CONTAINER_INSTANCE)

docker-test: ## Run docker container tests
	container-structure-test test --image $(NS)/$(IMAGE_NAME):$(CURRENT_VERSION_MICRO) --config tests/config.yaml

build: go-build docker-build ## Build all

rebuild: go-rebuild docker-rebuild ## Rebuild all

run: docker-run ## Run all

test: go-test bats-test docker-test ## Run all tests

test-ci: ## Run CI pipeline locally
	woodpecker-cli exec --local --repo-trusted-volumes=true --env=MODE_LOCAL=true

release: build test docker-push ## Build and push the image to a registry