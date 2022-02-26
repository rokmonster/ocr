# Check for required command tools to build or stop immediately
EXECUTABLES = go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

default: build

deps: 
	go install github.com/goreleaser/goreleaser@latest

tessdata:
	curl -Lo tessdata/eng.traineddata https://github.com/tesseract-ocr/tessdata/raw/main/eng.traineddata
	curl -Lo tessdata/fra.traineddata https://github.com/tesseract-ocr/tessdata/raw/main/fra.traineddata

build: deps
	goreleaser build --snapshot --rm-dist --single-target

dev:
	docker run --rm -it -v $(shell pwd):$(shell pwd) -w $(shell pwd) golang:1.17

.PHONY: check deps build tessdata