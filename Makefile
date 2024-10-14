BUILD_PKG = github.com/gopherd/core/builder
BUILD_VERSION = $(shell git describe --tags --abbrev=0)
BUILD_COMMIT = $(shell git rev-parse HEAD)
BUILD_VERSION ?= $(GITHUB_REF_NAME)
BUILD_COMMIT ?= $(GITHUB_SHA)

ifeq ($(OS),Windows_NT)
BUILD_DATETIME := $(shell powershell -Command "Get-Date -Format 'yyyy/MM/ddTHH:mm:sszzz'")
else
BUILD_DATETIME := $(shell date "+%Y/%m/%dT%H:%M:%S%z")
OS_NAME := $(shell uname -s)
ifeq (${OS_NAME},Linux)
    INSTALL_DIR := $(HOME)/.local/bin
else
	INSTALL_DIR := $(HOME)/bin
endif
endif

GOBUILD = go build -ldflags "-X ${BUILD_PKG}.commit=${BUILD_COMMIT} -X ${BUILD_PKG}.version=${BUILD_VERSION} -X ${BUILD_PKG}.datetime=${BUILD_DATETIME}"

BUILD_DIR = ./build
BUILD_BIN_DIR=${BUILD_DIR}/bin
EXAMPLE_DIR = ./website/example

.PHONY: all
all: build

.PHONY: autogen
autogen:
	@echo "Running go generate..."
	@go generate ./...

.PHONY: go/vet
go/vet:
	@echo "Running go vet..."
	@go vet ./...

.PHONY: build
build: autogen go/vet
	@echo "Building ${BUILD_BIN_DIR}/next..."
	@mkdir -p ${BUILD_DIR}
	@mkdir -p ${BUILD_BIN_DIR}
	@${GOBUILD} -o ${BUILD_BIN_DIR}/

.PHONY: install
install: build
	@echo "Installing to ${INSTALL_DIR}/..."
	@mkdir -p ${INSTALL_DIR}
	@cp ${BUILD_BIN_DIR}/next ${INSTALL_DIR}/
	@cp ${BUILD_BIN_DIR}/nextls ${INSTALL_DIR}/

define create_release
	$(eval dir := next$(subst v,,${BUILD_VERSION}).$(1)-$(2))
	@if [ -d ${BUILD_DIR}/${dir} ]; then rm -r ${BUILD_DIR}/${dir}*; fi
	@mkdir -p ${BUILD_DIR}/${dir}/bin
	@cp README.md ${BUILD_DIR}/${dir}/
	@echo "Building ${BUILD_DIR}/${dir}/next..."
	@GOOS=$(1) GOARCH=$(2) ${GOBUILD} -o ${BUILD_DIR}/${dir}/bin/
	@GOOS=$(1) GOARCH=$(2) ${GOBUILD} -o ${BUILD_DIR}/${dir}/bin/ ./cmd/nextls/
	@cd ${BUILD_DIR} && \
	if [ "$(1)" = "windows" ]; then \
		zip -q -r ${dir}.zip ${dir} && rm -r ${dir}; \
	else \
		tar zcf ${dir}.tar.gz ${dir} && rm -r ${dir}; \
	fi
endef

.PHONY: release
release:
	@rm -f ${BUILD_DIR}/next*.tar.gz
	$(call create_release,darwin,amd64)
	$(call create_release,darwin,arm64)
	$(call create_release,linux,amd64)
	$(call create_release,linux,arm64)
	$(call create_release,linux,386)
	$(call create_release,windows,amd64)
	$(call create_release,windows,arm64)
	$(call create_release,windows,386)

ifeq ($(OS),Windows_NT)
.PHONY: release/windows
release/windows:
	@powershell -ExecutionPolicy Bypass -File scripts/generate-msi.ps1 -Version '$(BUILD_VERSION)' -Dir '$(BUILD_DIR)' -GoBuild '$(GOBUILD)'
endif

.PHONY: test/src
test/src: autogen go/vet
	@echo "Running tests..."
	@go test -v ./...

.PHONY: example
example: example/clean build example/gen

.PHONY: example/clean
example/clean:
	@if [ -d ${EXAMPLE_DIR}/gen ]; then rm -rf ${EXAMPLE_DIR}/gen; fi

.PHONY: example/gen
example/gen:
	@echo "Running generate example..."
	@NEXT_NOCOPYBUILTIN=1 ${BUILD_BIN_DIR}/next build ${EXAMPLE_DIR}

.PHONY: clean
clean:
	@if [ -d ${BUILD_DIR} ]; then rm -rf ${BUILD_DIR}; fi
	@rm -f *.wixobj *.wxs
