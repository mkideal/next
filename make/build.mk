# make/build.mk

.PHONY: build
build: build/relife-server

.PHONY: build/relife-server
build/relife-server:
	$(call build_cmd,relife-server)

.PHONY: copy
copy:
	@cp scripts/*.sh ${TARGET_DIR}/