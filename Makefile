APP_NAME := jenkins_ping
GOPATH := ${GOPATH}
SOURCEDIR = .
DATA_DIR := ./data
BINDATA_DEBUG_FILE := $(SOURCEDIR)/bindata_debug.go
BINDATA_RELEASE_FILE := $(SOURCEDIR)/bindata_release.go

SOURCES := $(shell find $(SOURCEDIR) -name '*.go' -not -path '${BINDATA_DEBUG_FILE}' -not -path '${BINDATA_RELEASE_FILE}')
SOURCES_DATA := $(shell find $(DATA_DIR))

.DEFAULT_GOAL: ${APP_NAME}

${APP_NAME}: ${BINDATA_DEBUG_FILE} $(SOURCES)
	go get ./...
	go build -o ${APP_NAME}

.PHONY: deploy
deploy: ${BINDATA_RELEASE_FILE} $(SOURCES)
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN parameter must be set)
endif
ifndef TAG
	$(error TAG parameter must be set)
endif
	echo Creating and pushing tag
	git tag ${TAG}
	git push --tags
	echo Sleeping 5 seconds before trying to create release...
	sleep 5
	echo Creating release
	github-release release -u milanaleksic -r ${APP_NAME} --tag "${TAG}" --name "v${TAG}"

	echo Building and shipping Windows
	GOOS=windows go build
	upx ${APP_NAME}.exe
	github-release upload -u milanaleksic -r ${APP_NAME} --tag ${TAG} --name "${APP_NAME}-${TAG}-windows-amd64.exe" -f ${APP_NAME}.exe

	echo Building and shipping Linux
	GOOS=linux go build
	goupx ${APP_NAME}
	github-release upload -u milanaleksic -r ${APP_NAME} --tag ${TAG} --name "${APP_NAME}-${TAG}-linux-amd64" -f ${APP_NAME}

.PHONY: run
run: ${APP_NAME}
	${APP_NAME} -refresh=2s \
    -server "http://jenkins/" \
    -interface=gui \
    -mock=true \
    -doLog=true

.PHONY: test
test:
	go test

${BINDATA_DEBUG_FILE}: ${SOURCES_DATA}
	rm -rf ${BINDATA_RELEASE_FILE}
	go-bindata --debug -o=${BINDATA_DEBUG_FILE} ${DATA_DIR}/...

${BINDATA_RELEASE_FILE}: ${SOURCES_DATA}
	rm -rf ${BINDATA_DEBUG_FILE}
	go-bindata -nocompress=true -nomemcopy=true -o=${BINDATA_RELEASE_FILE} ${DATA_DIR}/...

.PHONY: ci
ci: ${BINDATA_RELEASE_FILE} $(SOURCES)
	go get ./... \
	go build -o ${APP_NAME}

.PHONY: prepare
prepare: ${GOPATH}/bin/go-bindata \
	${GOPATH}/bin/github-release \
	${GOPATH}/bin/goupx \
	gtk \
	upx

${GOPATH}/bin/goupx:
	go get github.com/pwaller/goupx

${GOPATH}/bin/github-release:
	go get github.com/aktau/github-release

${GOPATH}/bin/go-bindata:
	go get github.com/jteeuwen/go-bindata/go-bindata

upx:
	curl http://upx.sourceforge.net/download/upx-3.91-amd64_linux.tar.bz2 | tar xjvf - && mv upx-3.91-amd64_linux/upx upx && rm -rf upx-3.91-amd64_linux

.PHONY: gtk
gtk:
	dpkg -s libgtk-3-dev   > /dev/null || libgtk-3-dev
	dpkg -s libcairo2-dev  > /dev/null || libcairo2-dev
	dpkg -s libglib2.0-dev > /dev/null || libglib2.0-dev

.PHONY: clean
clean:
	rm -rf ${BINDATA_DEBUG_FILE}
	rm -rf ${BINDATA_RELEASE_FILE}
	rm -rf ${APP_NAME}
	rm -rf ${APP_NAME}.exe