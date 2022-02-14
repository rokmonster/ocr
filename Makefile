# ########################################################## #
# Makefile for Golang Project
# Includes cross-compiling, installation, cleanup
# ########################################################## #

# Check for required command tools to build or stop immediately
EXECUTABLES = go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

VERSION=1.0.0
PLATFORMS=darwin linux windows
ARCHITECTURES=amd64 arm64

# Setup linker flags option for build that interoperate with variable names in src code
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

default: all

all: build_all

build_all: build_scanner build_server build_remote

build_scanner:
	go build ${LDFLAGS} -o ./bin/ ./cmd/rok-scanner

build_server:
	go build ${LDFLAGS} -o ./bin/ ./cmd/rok-server

build_remote:
	go build ${LDFLAGS} -o ./bin/ ./cmd/rok-remote
# # Remove only what we've created
# clean:
# 	find ${ROOT_DIR} -name '${BINARY}[-?][a-zA-Z0-9]*[-?][a-zA-Z0-9]*' -delete

.PHONY: check build_all all