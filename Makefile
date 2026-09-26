-include .env

.DEFAULT_GOAL := help

# Directories
MAKE_DIR := make

# Include other Makefiles
include $(MAKE_DIR)/Makefile.lint
include $(MAKE_DIR)/Makefile.fmt
include $(MAKE_DIR)/Makefile.tools
include $(MAKE_DIR)/Makefile.test
include $(MAKE_DIR)/Makefile.run
include $(MAKE_DIR)/Makefile.issue
include $(MAKE_DIR)/Makefile.gen

## Show available targets
help:
	@awk '/^## / {d = substr($$0, 4)} /^[a-zA-Z_/-]+:/ {if (d != "") {printf "\033[36m%-15s\033[0m %s\n", $$1, d; d = ""}}' $(MAKEFILE_LIST) | sort
