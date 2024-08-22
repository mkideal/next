# make/utils.mk

define build_cmd
	@mkdir -p ${TARGET_HUB_DIR}
	@mkdir -p ${TARGET_ETC_DIR}
	@echo "Building ${TARGET_HUB_DIR}/$(1) ..."
	@${GOBUILD} -o ${TARGET_HUB_DIR}/$(1) ./cmd/$(1)/
endef
