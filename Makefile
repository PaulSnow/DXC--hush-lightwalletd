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

clean:
	@echo "clean project..."
	rm -f $(PROJECT_NAME)
