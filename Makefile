APP_NAME := jenkins_ping
SHELL := /bin/bash
GOPATH := ${GOPATH}
SOURCEDIR = .
DATA_DIR := ./data
BINDATA_DEBUG_FILE := $(SOURCEDIR)/bindata_debug.go
BINDATA_RELEASE_FILE := $(SOURCEDIR)/bindata_release.go
SOURCES := $(shell find $(SOURCEDIR) -name '*.go' -not -path '${BINDATA_DEBUG_FILE}' -not -path '${BINDATA_RELEASE_FILE}' -not -path './vendor/*')

VERSION := $(shell git name-rev --tags --name-only `git rev-parse HEAD`)
IS_DEFINED_VERSION := $(shell [ ! "${VERSION}" == "undefined" ] && echo true)

.DEFAULT_GOAL: ${APP_NAME}

${APP_NAME}: ${BINDATA_DEBUG_FILE} $(SOURCES)
	go build -ldflags '-X main.Version=${TAG}' -o ${APP_NAME}

.PHONY: metalinter
metalinter: ${APP_NAME}
	gometalinter --exclude="bindata_*" --vendor --deadline=2m ./...

.PHONY: deploy-if-tagged
deploy-if-tagged: ${BINDATA_RELEASE_FILE} $(SOURCES)
ifeq ($(IS_DEFINED_VERSION),true)
	$(MAKE) _release_to_github TAG=$(VERSION)
endif

.PHONY: deploy
deploy: ${BINDATA_RELEASE_FILE} $(SOURCES)
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
_release_to_github: ${BINDATA_RELEASE_FILE} $(SOURCES)
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
run: ${APP_NAME}
	./${APP_NAME}

.PHONY: test
test:
	go test -v $$(go list ./... | grep -v /vendor/)

.PHONY: ci
ci: ${BINDATA_RELEASE_FILE} $(SOURCES)
	$(MAKE) metalinter
	$(MAKE) test
	go build -ldflags '-X main.Version=${TAG}' -o ${APP_NAME}

${BINDATA_DEBUG_FILE}: ${SOURCES_DATA}
	rm -rf ${BINDATA_RELEASE_FILE}
	go-bindata --debug -o=${BINDATA_DEBUG_FILE} ${DATA_DIR}/...

${BINDATA_RELEASE_FILE}: ${SOURCES_DATA}
	rm -rf ${BINDATA_DEBUG_FILE}
	go-bindata -nocompress=true -nomemcopy=true -o=${BINDATA_RELEASE_FILE} ${DATA_DIR}/...

.PHONY: prepare
prepare: ${GOPATH}/bin/github-release \
	${GOPATH}/bin/go-bindata \
	${GOPATH}/bin/goupx \
	${GOPATH}/bin/gometalinter \
	upx

${GOPATH}/bin/gometalinter:
	go get github.com/alecthomas/gometalinter
	gometalinter --install --update

${GOPATH}/bin/goupx:
	go get github.com/pwaller/goupx

${GOPATH}/bin/github-release:
	go get github.com/aktau/github-release

${GOPATH}/bin/go-bindata:
	go get github.com/jteeuwen/go-bindata/go-bindata

upx:
	curl http://upx.sourceforge.net/download/upx-3.91-amd64_linux.tar.bz2 | tar xjvf - && mv upx-3.91-amd64_linux/upx upx && rm -rf upx-3.91-amd64_linux

.PHONY: clean
clean:
	rm -rf ${BINDATA_DEBUG_FILE}
	rm -rf ${BINDATA_RELEASE_FILE}
	rm -rf ${APP_NAME}
	rm -rf ${APP_NAME}.exe
