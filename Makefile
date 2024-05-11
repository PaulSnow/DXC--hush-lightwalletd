# Copyright (c) 2021-2024 Jahway603 & The Hush Developers
# Released under the GPLv3
#
# Hush Lightwalletd Makefile
# author: jahway603
# 
.PHONY: format help
# Help system from https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.DEFAULT_GOAL := help

PROJECT_NAME := "lightwalletd"
GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet

help:
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build binary for amd64 (x86_64) architecture
	./util/build.sh

build-arm: ## Build binary for ARM architecture (aarch64)
	GOOS=linux GOARCH=arm64 ./util/build.sh

protobuf: ## Generate protobuf shizzle
	cd walletrpc && protoc --go_out=paths=source_relative:. service.proto compact_formats.proto && protoc --go-grpc_out=paths=source_relative:. service.proto

# Stop the hushd process in the hushdlwd container
#docker_img_stop_hushd:
#	docker exec -i hushdlwd hush-cli stop

# Remove and delete ALL images and containers in Docker; assumes containers are stopped
#docker_remove_all:
#	docker system prune -f

dep: ## Pull dependencies (if needed)
	@go get -v -d ./...

vendor: ## Pull vendor files locally (if needed)
	go mod tidy && go mod vendor

clean: ## Clean the project
	@echo "Cleaning project $(PROJECT_NAME) files..."
	rm -f $(PROJECT_NAME)
	rm -rf /tmp/$(PROJECT_NAME)-*
