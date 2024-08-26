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
	@echo "Building ${BUILD_DIR}/next..."
	@mkdir -p ${BUILD_DIR}
	@${GOBUILD} -o ${BUILD_DIR}/ ./cmd/next

.PHONY: install
install: build
	@echo "Installing to /usr/local/bin/next..."
	@cp ${BUILD_DIR}/next /usr/local/bin/
	@cp -r next.d /usr/local/etc/

define release_cmd
	$(eval dir := next.$(subst v,,${BUILD_VERSION}).$(1)-$(2))
	@echo "Building ${BUILD_DIR}/${dir}/next..."
	@mkdir -p ${BUILD_DIR}/${dir}
	@GOOS=$(1) GOARCH=$(2) ${GOBUILD} -o ${BUILD_DIR}/${dir}/ ./cmd/next
	@cp -r next.d ${BUILD_DIR}/${dir} 
	@cd ${BUILD_DIR} && tar zcf ${dir}.tar.gz ${dir} && rm -r ${dir}
endef

.PHONY: release
release: go/generate go/vet
	$(call release_cmd,windows,amd64)
	$(call release_cmd,darwin,amd64)
	$(call release_cmd,darwin,arm64)
	$(call release_cmd,linux,amd64)
	$(call release_cmd,linux,arm64)
	$(call release_cmd,linux,386)

.PHONY: test/src
test/src: go/generate go/vet
	@echo "Running tests..."
	@go test -v ./...

.PHONY: test/template
test/template:
	@echo "Running template tests..."
	@next -O go=testdata/gen/go -T go=testdata/templates/go testdata/a.next testdata/b.next
	@next -O go=testdata/gen/go/c -T go=testdata/templates/go testdata/c.next

.PHONY: clean
clean:
	@if [ -d ${BUILD_DIR} ]; then rm -r ${BUILD_DIR}; fi