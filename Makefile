BUILD_PKG = github.com/gopherd/core/builder
BUILD_BRANCH = $(shell git symbolic-ref --short HEAD)
BUILD_VERSION = $(shell git describe --tags --abbrev=0)
BUILD_COMMIT = $(shell git rev-parse HEAD)
BUILD_DATETIME = $(shell date "+%Y/%m/%dT%H:%M:%S%z")

GOBUILD = go build -ldflags " \
	-X ${BUILD_PKG}.branch=${BUILD_BRANCH} \
	-X ${BUILD_PKG}.commit=${BUILD_COMMIT} \
	-X ${BUILD_PKG}.version=${BUILD_VERSION} \
	-X ${BUILD_PKG}.datetime=${BUILD_DATETIME} \
	"

BUILD_DIR = ./build

.PHONY: all
all: build

.PHONY: deps/stringer
deps/stringer:
ifeq (, $(shell which stringer))
	@echo "Installing stringer..."
	@go install golang.org/x/tools/cmd/stringer@latest
endif

.PHONY: go/generate
go/generate: deps/stringer
	@echo "Running go generate..."
	@go generate ./...

.PHONY: go/vet
go/vet:
	@echo "Running go vet..."
	@go vet ./...

.PHONY: build
build: go/generate go/vet
	@mkdir -p ${BUILD_DIR}
	@${GOBUILD} -o ${BUILD_DIR}/ ./cmd/next

.PHONY: install
install: build
	@echo "Installing..."
	@cp ${BUILD_DIR}/next /usr/local/bin/

.PHONY: release
release: go/generate go/vet
	@echo "Building ${BUILD_DIR}/windows-amd64/next ..."
	@mkdir -p ${BUILD_DIR}/windows-amd64
	@GOOS=windows GOARCH=amd64 ${GOBUILD} -o ${BUILD_DIR}/windows-amd64/ ./cmd/next

	@echo "Building ${BUILD_DIR}/darwin-amd64/next ..."
	@mkdir -p ${BUILD_DIR}/darwin-amd64
	@GOOS=darwin GOARCH=amd64 ${GOBUILD} -o ${BUILD_DIR}/darwin-amd64/ ./cmd/next

	@echo "Building ${BUILD_DIR}/darwin-arm64/next ..."
	@mkdir -p ${BUILD_DIR}/darwin-arm64
	@GOOS=darwin GOARCH=arm64 ${GOBUILD} -o ${BUILD_DIR}/darwin-arm64/ ./cmd/next

	@echo "Building ${BUILD_DIR}/linux-amd64/next ..."
	@mkdir -p ${BUILD_DIR}/linux-amd64
	@GOOS=linux GOARCH=amd64 ${GOBUILD} -o ${BUILD_DIR}/linux-amd64/ ./cmd/next

	@echo "Building ${BUILD_DIR}/linux-arm64/next ..."
	@mkdir -p ${BUILD_DIR}/linux-amd64
	@GOOS=linux GOARCH=arm64 ${GOBUILD} -o ${BUILD_DIR}/linux-arm64/ ./cmd/next

	@echo "Building ${BUILD_DIR}/linux-386/next ..."
	@mkdir -p ${BUILD_DIR}/linux-386
	@GOOS=linux GOARCH=x86 ${GOBUILD} -o ${BUILD_DIR}/linux-386/ ./cmd/next

.PHONY: clean
clean:
	@if [ -f /usr/local/bin/next ]; then rm /usr/local/bin/next; fi
	@if [ -d ${BUILD_DIR} ]; then rm -r ${BUILD_DIR}; fi