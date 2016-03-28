PACKAGE := $(shell go list -e)
APP_NAME = $(lastword $(subst /, ,$(PACKAGE)))

include gomakefiles/common.mk
include gomakefiles/metalinter.mk
include gomakefiles/upx.mk
include gomakefiles/bindata.mk

SOURCES := $(shell find $(SOURCEDIR) -name '*.go' \
	-not -path '${BINDATA_DEBUG_FILE}' \
	-not -path '${BINDATA_RELEASE_FILE}' \
	-not -path './vendor/*')

${APP_NAME}: ${BINDATA_DEBUG_FILE} $(SOURCES)
	go build -ldflags '-X main.Version=${TAG}' -o ${APP_NAME}

${RELEASE_SOURCES}: ${BINDATA_RELEASE_FILE} $(SOURCES)

include gomakefiles/semaphore.mk

.PHONY: prepare
prepare: prepare_metalinter prepare_upx prepare_bindata prepare_github_release

.PHONY: clean
clean: clean_common clean_bindata
