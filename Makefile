# Check for required command tools to build or stop immediately
EXECUTABLES = go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

default: build

dockerclient:
	curl -Lo /tmp/docker-20.10.9.tgz https://download.docker.com/linux/static/stable/x86_64/docker-20.10.9.tgz
	tar -C /usr/local/bin --strip-components=1 -xzf /tmp/docker-20.10.9.tgz
	docker info

dockerdeps: 
	apt update && apt install -y libtesseract-dev
	
deps: 
	go install github.com/goreleaser/goreleaser@latest
	
tessdata:
	curl -Lo tessdata/eng.traineddata https://github.com/tesseract-ocr/tessdata/raw/main/eng.traineddata
	curl -Lo tessdata/fra.traineddata https://github.com/tesseract-ocr/tessdata/raw/main/fra.traineddata

build: deps
	goreleaser build --snapshot --rm-dist --single-target

snapshot: deps
	goreleaser release --snapshot --rm-dist --skip-publish

dev:
	docker run --rm -it -v /var/run/docker.sock:/var/run/docker.sock -v $(shell pwd):$(shell pwd) -w $(shell pwd) golang:1.18


.PHONY: check deps build tessdata