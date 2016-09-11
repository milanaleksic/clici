PACKAGE := $(shell go list -e)
APP_NAME = $(lastword $(subst /, ,$(PACKAGE)))
MAIN_APP_DIR = cmd/main

include gomakefiles/common.mk
include gomakefiles/metalinter.mk
include gomakefiles/upx.mk
include gomakefiles/proto.mk
include gomakefiles/wago.mk

DATA_DIR := $(SOURCEDIR)/$(MAIN_APP_DIR)/data
BINDATA_DEBUG_FILE := $(SOURCEDIR)/$(MAIN_APP_DIR)/bindata_debug.go
BINDATA_RELEASE_FILE := $(SOURCEDIR)/$(MAIN_APP_DIR)/bindata_release.go
include gomakefiles/bindata.mk

SOURCES := $(shell find $(SOURCEDIR) -name '*.go' \
	-not -path '${BINDATA_DEBUG_FILE}' \
	-not -path '${BINDATA_RELEASE_FILE}' \
	-not -path './vendor/*')

EXCLUDES_METALINTER := .*.pb.go|bindata_.*.go

$(APP_NAME): $(MAIN_APP_DIR)/$(APP_NAME)

$(MAIN_APP_DIR)/$(APP_NAME): $(SOURCES) $(BINDATA_DEBUG_FILE)
	cd $(MAIN_APP_DIR)/ && go build -ldflags '-X main.Version=${VERSION}' -o ${APP_NAME}

cmd/server/$(APP_NAME)_server: $(SOURCES)
	cd cmd/server/ && go build -ldflags '-X main.Version=${VERSION}' -o ${APP_NAME}_server

${RELEASE_SOURCES}: ${BINDATA_RELEASE_FILE} $(SOURCES)

include gomakefiles/semaphore.mk

.PHONY: prepare
prepare: prepare_metalinter prepare_upx prepare_bindata prepare_github_release prepare_wago

.PHONY: clean
clean: clean_common clean_bindata
	rm -rf $(MAIN_APP_DIR)/${APP_NAME}
	rm -rf $(MAIN_APP_DIR)/${APP_NAME}.exe
	rm -rf cmd/server/${APP_NAME}_server
	rm -rf cmd/server/${APP_NAME}_server.exe
