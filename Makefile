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
TESTDATA_DIR = ./website/testdata

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
release: go/generate go/vet
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
test/src: go/generate go/vet
	@echo "Running tests..."
	@go test -v ./...

.PHONY: test/template
test/template: install
	@echo "Running template tests..."
	@rm -rf ${TESTDATA_DIR}/gen
	@NEXTNOCOPYBUILTIN=1 next \
		-D PROJECT_NAME=demo \
		-O go=${TESTDATA_DIR}/gen/go -T go=${TESTDATA_DIR}/templates/go \
		-O java=${TESTDATA_DIR}/gen/java -T java=${TESTDATA_DIR}/templates/java \
		-O cpp=${TESTDATA_DIR}/gen/cpp -T cpp=${TESTDATA_DIR}/templates/cpp \
		-O csharp=${TESTDATA_DIR}/gen/csharp -T csharp=${TESTDATA_DIR}/templates/csharp \
		-O c=${TESTDATA_DIR}/gen/c -T c=${TESTDATA_DIR}/templates/c \
		-O rust=${TESTDATA_DIR}/gen/rust/src -T rust=${TESTDATA_DIR}/templates/rust \
		-O protobuf=${TESTDATA_DIR}/gen/protobuf -T protobuf=${TESTDATA_DIR}/templates/protobuf \
		-O js=${TESTDATA_DIR}/gen/js -T js=${TESTDATA_DIR}/templates/js \
		-O ts=${TESTDATA_DIR}/gen/ts -T ts=${TESTDATA_DIR}/templates/ts \
		-O python=${TESTDATA_DIR}/gen/python -T python=${TESTDATA_DIR}/templates/python \
		-O php=${TESTDATA_DIR}/gen/php -T php=${TESTDATA_DIR}/templates/php \
		-O lua=${TESTDATA_DIR}/gen/lua -T lua=${TESTDATA_DIR}/templates/lua \
		-M "c.vector=void*" -M "c.map=void*" \
		${TESTDATA_DIR}/next/
	@cd ${TESTDATA_DIR}/gen/rust && cargo init --vcs none

.PHONY: clean
clean:
	@if [ -d ${BUILD_DIR} ]; then rm -r ${BUILD_DIR}; fi
