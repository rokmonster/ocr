# Check for required command tools to build or stop immediately
EXECUTABLES = go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
GOLANG_IMAGE:=golang:1.22

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
dockerclient: ## Install docker client CLI
	curl -Lo /tmp/docker-20.10.9.tgz https://download.docker.com/linux/static/stable/x86_64/docker-20.10.9.tgz
	tar -C /usr/local/bin --strip-components=1 -xzf /tmp/docker-20.10.9.tgz
	docker info

dockerdeps: dockerclient ## Install docker dependencies (libtesseract-dev)
	apt update && apt install -y libtesseract-dev

deps: ## Install goreleaser
	go install github.com/goreleaser/goreleaser@latest

update-deps: ## Update direct golang dependencies
	go get $(shell go list -f '{{if not (or .Main .Indirect)}}{{.Path}}{{end}}' -m all) && go mod tidy

dev: ## start a dev env (for building linux binary from mac/win)
	docker run -p8080:8080 -v ~/.ssh:/root/.ssh --rm -it -v /var/run/docker.sock:/var/run/docker.sock -v $(shell pwd):$(shell pwd) -w $(shell pwd) $(GOLANG_IMAGE) bash

opencv: ## Compile opencv for linux (from source)
	go mod vendor
	(cd vendor/gocv.io/x/gocv/ && make deps download sudo_pre_install_clean build_static sudo_install clean)
	rm -rf vendor

##@ Build
remote: ## Build remote
	go build -v -o dist/rok-remote cmd/rok-remote/main.go

server: ## Build server
	go build -v -o dist/rok-server cmd/rok-server/main.go

scanner: ## Build scanner
	go build -v -o dist/rok-scanner cmd/rok-scanner/main.go

all: remote server scanner ## Build all

##@ Run
run-remote: ## Run remote
	go run ./cmd/rok-remote/main.go

run-server: ## Run server
	go run ./cmd/rok-server/main.go

run-scanner: ## Run scanner
	go run ./cmd/rok-scanner/main.go

##@ Release
.PHONY: snapshot
snapshot: deps ## Build a snapshot release
	goreleaser release --snapshot --clean --skip=publish
