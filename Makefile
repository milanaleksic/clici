APP_NAME := jenkins_ping
GOPATH := ${GOPATH}
SOURCEDIR = .

SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

.DEFAULT_GOAL: ${APP_NAME}

${APP_NAME}: $(SOURCES)
	go get ./...
	go build -ldflags '-X main.Version=${TAG}' -o ${APP_NAME}

.PHONY: metalinter
metalinter: ${APP_NAME}
	gometalinter --deadline=2m ./...

.PHONY: deploy
deploy: $(SOURCES)
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
	github-release release -u milanaleksic -r ${APP_NAME} --tag "${TAG}" --name "v${TAG}"

	echo Building and shipping Windows
	GOOS=windows go build -ldflags '-X main.Version=${TAG}'
	upx ${APP_NAME}.exe
	github-release upload -u milanaleksic -r ${APP_NAME} --tag ${TAG} --name "${APP_NAME}-${TAG}-windows-amd64.exe" -f ${APP_NAME}.exe

	echo Building and shipping Linux
	GOOS=linux go build -ldflags '-X main.Version=${TAG}'
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
	go test -v

.PHONY: ci
ci: $(SOURCES)
	go get ./...
	$(MAKE) metalinter
	go build -ldflags '-X main.Version=${TAG}' -o ${APP_NAME}

.PHONY: prepare
prepare: ${GOPATH}/bin/github-release \
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

upx:
	curl http://upx.sourceforge.net/download/upx-3.91-amd64_linux.tar.bz2 | tar xjvf - && mv upx-3.91-amd64_linux/upx upx && rm -rf upx-3.91-amd64_linux

.PHONY: clean
clean:
	rm -rf ${APP_NAME}
	rm -rf ${APP_NAME}.exe
