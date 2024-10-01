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
BUILD_BIN_DIR=${BUILD_DIR}/bin
EXAMPLE_DIR = ./website/example

.PHONY: all
all: build

.PHONY: deps/stringer
deps/stringer:
ifeq (, $(shell which stringer))
	@echo "Installing stringer..."
	@go install golang.org/x/tools/cmd/stringer@latest
endif

.PHONY: autogen
autogen: deps/stringer
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
	@echo "Installing to /usr/local/bin/next..."
	@cp ${BUILD_BIN_DIR}/next /usr/local/bin/

define release_unix
	$(eval dir := next.$(subst v,,${BUILD_VERSION}).$(1)-$(2))
	@echo "Building ${BUILD_DIR}/${dir}/next..."
	@mkdir -p ${BUILD_DIR}/${dir}/bin
	@cp README.md ${BUILD_DIR}/${dir}/
	@GOOS=$(if $(filter mingw,$(1)),windows,$(1)) GOARCH=$(2) ${GOBUILD} -o ${BUILD_DIR}/${dir}/bin/
	@cd ${BUILD_DIR} && tar zcf ${dir}.tar.gz ${dir} && rm -r ${dir}
endef

define release_windows
	$(eval dir := next.$(subst v,,${BUILD_VERSION}).windows-$(1))
	@echo "Building ${BUILD_DIR}/${dir}/next..."
	@mkdir -p ${BUILD_DIR}/${dir}/bin
	@GOOS=windows GOARCH=$(1) ${GOBUILD} -o ${BUILD_DIR}/${dir}/bin/
	@cp ./scripts/install.bat ${BUILD_DIR}/${dir}/
	@cp README.md ${BUILD_DIR}/${dir}/
	@cd ${BUILD_DIR} && zip ${dir}.zip -r ${dir} >/dev/null && rm -r ${dir}
endef

define release_js
	$(eval filename := next.$(subst v,,${BUILD_VERSION}).wasm)
	@echo "Building ${BUILD_DIR}/${filename}..."
	@mkdir -p ${BUILD_DIR}/
	@GOOS=js GOARCH=wasm go build -o ${BUILD_DIR}/${filename}
endef

.PHONY: release
release: autogen go/vet
	@rm -f ${BUILD_DIR}/next.*.tar.gz ${BUILD_DIR}/next.*.zip ${BUILD_DIR}/*.wasm
	$(call release_unix,darwin,amd64)
	$(call release_unix,darwin,arm64)
	$(call release_unix,linux,amd64)
	$(call release_unix,linux,arm64)
	$(call release_unix,linux,386)
	$(call release_unix,mingw,amd64)
	$(call release_unix,mingw,386)
	$(call release_windows,amd64)
	$(call release_js)

.PHONY: test/src
test/src: autogen go/vet
	@echo "Running tests..."
	@go test -v ./...

.PHONY: example
example: example/clean build example/gen

.PHONY: example/clean
example/clean:
	@if [ -d ${EXAMPLE_DIR}/gen ]; then rm -r ${EXAMPLE_DIR}/gen; fi

.PHONY: example/gen
example/gen:
	@echo "Running generate example..."
	@rm -rf ${EXAMPLE_DIR}/gen
	@NEXTNOCOPYBUILTIN=1 ${BUILD_BIN_DIR}/next build ${EXAMPLE_DIR}
	@if [ -d ${EXAMPLE_DIR}/gen/rust ]; then cd ${EXAMPLE_DIR}/gen/rust && cargo init --vcs none -q; fi

.PHONY: clean
clean:
	@if [ -d ${BUILD_DIR} ]; then rm -r ${BUILD_DIR}; fi
