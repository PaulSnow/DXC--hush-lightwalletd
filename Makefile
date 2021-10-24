# Copyright (c) 2021 Jahway603 & The Hush Developers
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

# Stop the hushd process in the hushdlwd container
#docker_img_stop_hushd:
#	docker exec -i hushdlwd hush-cli stop

# Remove and delete ALL images and containers in Docker; assumes containers are stopped
#docker_remove_all:
#	docker system prune -f

clean:
	@echo "Cleaning project $(PROJECT_NAME) files..."
	rm -f $(PROJECT_NAME)
	rm -rf /tmp/$(PROJECT_NAME)-*
