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
	@GOOS=$(1) GOARCH=$(2) ${GOBUILD} -o ${BUILD_DIR}/${dir}/bin/
	@cd ${BUILD_DIR} && tar zcf ${dir}.tar.gz ${dir} && rm -r ${dir}
endef

define release_windows
	$(eval dir := next.$(subst v,,${BUILD_VERSION}).windows-$(1))
	@echo "Building ${BUILD_DIR}/${dir}/next..."
	@mkdir -p ${BUILD_DIR}/${dir}/bin
	@GOOS=windows GOARCH=$(2) ${GOBUILD} -o ${BUILD_DIR}/${dir}/bin/
	@cp ./scripts/install.bat ${BUILD_DIR}/${dir}/
	@cp README.md ${BUILD_DIR}/${dir}/
	@cd ${BUILD_DIR} && zip ${dir}.zip -r ${dir} >/dev/null && rm -r ${dir}
endef

.PHONY: release
release: go/generate go/vet
	@rm -f ${BUILD_DIR}/next.*.tar.gz ${BUILD_DIR}/next.*.zip
	$(call release_unix,darwin,amd64)
	$(call release_unix,darwin,arm64)
	$(call release_unix,linux,amd64)
	$(call release_unix,linux,arm64)
	$(call release_unix,linux,386)
	$(call release_windows,amd64)

.PHONY: test/src
test/src: go/generate go/vet
	@echo "Running tests..."
	@go test -v ./...

.PHONY: test/template
test/template: install
	@echo "Running template tests..."
	@rm -rf testdata/gen
	@next \
		-v 1 \
		-O go=testdata/gen/go -T go=testdata/templates/go \
		-O cpp=testdata/gen/cpp -T cpp=testdata/templates/cpp \
		-O java=testdata/gen/java -T java=testdata/templates/java \
		testdata/next/

.PHONY: clean
clean:
	@if [ -d ${BUILD_DIR} ]; then rm -r ${BUILD_DIR}; fi