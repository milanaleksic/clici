APP_NAME := jenkins_ping
SHELL := /bin/bash
GOPATH := ${GOPATH}
SOURCEDIR = .
DATA_DIR := ./data
BINDATA_DEBUG_FILE := $(SOURCEDIR)/bindata_debug.go
BINDATA_RELEASE_FILE := $(SOURCEDIR)/bindata_release.go
SOURCES := $(shell find $(SOURCEDIR) -name '*.go' -not -path '${BINDATA_DEBUG_FILE}' -not -path '${BINDATA_RELEASE_FILE}')

VERSION := $(shell git name-rev --tags --name-only `git rev-parse HEAD`)
IS_DEFINED_VERSION := $(shell [ ! "${VERSION}" == "undefined" ] && echo true)

.DEFAULT_GOAL: ${APP_NAME}

${APP_NAME}: ${BINDATA_DEBUG_FILE} $(SOURCES)
	go get ./...
	go build -ldflags '-X main.Version=${TAG}' -o ${APP_NAME}

.PHONY: metalinter
metalinter: ${APP_NAME} ## Executes gometalinter tool to check the quality of the code
	gometalinter --exclude=bindata_* --deadline=2m ./...

.PHONY: deploy-if-tagged
deploy-if-tagged: ${BINDATA_RELEASE_FILE} $(SOURCES) ## In case a git tag is detected to point to HEAD, we shall call _release_to_github
ifeq ($(IS_DEFINED_VERSION),true)
	$(MAKE) _release_to_github TAG=$(VERSION)
endif

.PHONY: deploy
deploy: ${BINDATA_RELEASE_FILE} $(SOURCES) ## Explicit call of _release_to_github task. Expects you have GITHUB_TOKEN and TAG env params set
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN parameter must be set)
endif
ifndef TAG
	$(error TAG parameter must be set: make TAG=<TAG_VALUE>)
endif
	echo Creating and pushing tag
	git tag ${TAG}
	git push --tags
	echo Sleeping 5 seconds before trying to create release...
	sleep 5
	echo Creating release
	$(MAKE) _release_to_github

.PHONY: _release_to_github
_release_to_github: ${BINDATA_RELEASE_FILE} $(SOURCES) ## Executes creation of GitHub release and uploads UPXed executables. Expects you have GITHUB_TOKEN and TAG env params set
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN parameter must be set)
endif
ifndef TAG
	$(error TAG parameter must be set: make TAG=<TAG_VALUE>)
endif
	github-release release -u milanaleksic -r ${APP_NAME} --tag "${TAG}" --name "v${TAG}"

	echo Building and shipping Windows
	GOOS=windows go build -ldflags '-X main.Version=${TAG}'
	./upx ${APP_NAME}.exe
	github-release upload -u milanaleksic -r ${APP_NAME} --tag ${TAG} --name "${APP_NAME}-${TAG}-windows-amd64.exe" -f ${APP_NAME}.exe

	echo Building and shipping Linux
	GOOS=linux go build -ldflags '-X main.Version=${TAG}'
	PATH=$$PATH:. goupx ${APP_NAME}
	github-release upload -u milanaleksic -r ${APP_NAME} --tag ${TAG} --name "${APP_NAME}-${TAG}-linux-amd64" -f ${APP_NAME}

.PHONY: run
run: ${APP_NAME} ## Runs application
	${APP_NAME}

.PHONY: test
test: ## Executes all tests
	go test -v

.PHONY: ci
ci: ${BINDATA_RELEASE_FILE} $(SOURCES) ## Task meant to be run by a CI tool which does all critical testing and building
	go get ./...
	$(MAKE) metalinter
	go test ./...
	go build -ldflags '-X main.Version=${TAG}' -o ${APP_NAME}

${BINDATA_DEBUG_FILE}: ${SOURCES_DATA} ## Create bindata debug file with all data inside data dir mentioned inside a .go file (with redirects to real files)
	rm -rf ${BINDATA_RELEASE_FILE}
	go-bindata --debug -o=${BINDATA_DEBUG_FILE} ${DATA_DIR}/...

${BINDATA_RELEASE_FILE}: ${SOURCES_DATA} ## Create bindata production file with all data inside data dir packaged inside a .go file
	rm -rf ${BINDATA_DEBUG_FILE}
	go-bindata -nocompress=true -nomemcopy=true -o=${BINDATA_RELEASE_FILE} ${DATA_DIR}/...

.PHONY: prepare
prepare: ${GOPATH}/bin/github-release ## First step that needs to be run on clean environments that will make sure all deps are available \
	${GOPATH}/bin/go-bindata \
	${GOPATH}/bin/goupx \
	${GOPATH}/bin/gometalinter \
	upx

${GOPATH}/bin/gometalinter: ## Installs gometalinter tool
	go get github.com/alecthomas/gometalinter
	gometalinter --install --update

${GOPATH}/bin/goupx: ## Installs goupx tool
	go get github.com/pwaller/goupx

${GOPATH}/bin/github-release: ## Installs github-release tool
	go get github.com/aktau/github-release

${GOPATH}/bin/go-bindata: ## Installs go-bindata tool & library
	go get github.com/jteeuwen/go-bindata/go-bindata

upx: ## Fetches and makes locally available a UPX executable
	curl http://upx.sourceforge.net/download/upx-3.91-amd64_linux.tar.bz2 | tar xjvf - && mv upx-3.91-amd64_linux/upx upx && rm -rf upx-3.91-amd64_linux

.PHONY: clean
clean: ## Removes all known intermediary files
	rm -rf ${BINDATA_DEBUG_FILE}
	rm -rf ${BINDATA_RELEASE_FILE}
	rm -rf ${APP_NAME}
	rm -rf ${APP_NAME}.exe

.PHONY: help
help:
	@grep -P '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'