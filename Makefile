# Check for required command tools to build or stop immediately
EXECUTABLES = go find pwd
K := $(foreach exec,$(EXECUTABLES),\
        $(if $(shell which $(exec)),some string,$(error "No $(exec) in PATH)))

ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

default: build

# for usage inside `make dev` - installs docker-client
dockerclient:
	curl -Lo /tmp/docker-20.10.9.tgz https://download.docker.com/linux/static/stable/x86_64/docker-20.10.9.tgz
	tar -C /usr/local/bin --strip-components=1 -xzf /tmp/docker-20.10.9.tgz
	docker info

# for usage inside `make dev` - installs tesseractlib & adb
dockerdeps: 
	apt update && apt install -y libtesseract-dev adb
	
# get's go releaser
deps: 
	go install github.com/goreleaser/goreleaser@latest

# for usage inside `make dev` - builds binary, uploads it to demo server & restarts
sshinstall: dockerdeps build
	ssh root@rokdata.fun 'systemctl stop rokocr-server'
	scp ./dist/server_linux_amd64_v1/rok-server root@rokdata.fun:/usr/bin/rok-server
	ssh root@rokdata.fun 'systemctl start rokocr-server'
	
build: deps
	goreleaser build --snapshot --rm-dist --single-target

snapshot: deps
	goreleaser release --snapshot --rm-dist --skip-publish

# start a dev env (for building linux binary from mac/win)
dev:
	docker run -p8080:8080 -v ~/.ssh:/root/.ssh --rm -it -v /var/run/docker.sock:/var/run/docker.sock -v $(shell pwd):$(shell pwd) -w $(shell pwd) golang:1.18


.PHONY: check deps build