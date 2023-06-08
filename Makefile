# Copyright (c) 2021-2023 Jahway603 & The Hush Developers
# Released under the GPLv3
#
# Hush Lightwalletd Makefile
# author: jahway603
# 
PROJECT_NAME := "lightwalletd"
GOCMD=go
GOTEST=$(GOCMD) test
GOVET=$(GOCMD) vet

#.PHONY: build

build:
	# Build binary
	./util/build.sh

build-arm:
	# Build binary for ARM architecture (aarch64)
	GOOS=linux GOARCH=arm64 ./util/build.sh

protobuf:
	# Generate protobuf shizzle
	cd walletrpc && protoc --go_out=paths=source_relative:. service.proto compact_formats.proto && protoc --go-grpc_out=paths=source_relative:. service.proto

# Stop the hushd process in the hushdlwd container
#docker_img_stop_hushd:
#	docker exec -i hushdlwd hush-cli stop

# Remove and delete ALL images and containers in Docker; assumes containers are stopped
#docker_remove_all:
#	docker system prune -f

dep:
	@go get -v -d ./...

vendor:
	go mod tidy && go mod vendor

clean:
	@echo "Cleaning project $(PROJECT_NAME) files..."
	rm -f $(PROJECT_NAME)
	rm -rf /tmp/$(PROJECT_NAME)-*
