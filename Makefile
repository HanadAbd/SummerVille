SRC_DIR := web/src
DIST_DIR := web/dist

TS_EXT := .ts
JS_EXT := .js

# TypeScript Compiler
TSC := tsc

all: clean copy_and_compile start_app

clean:
	@echo "Removing dist directory..."
	@rm -rf $(DIST_DIR)
	@echo "Dist directory removed."

copy_and_compile:
	@echo "Copying and compiling files..."
	@mkdir -p $(DIST_DIR)
	@$(MAKE) copy_files SRC_DIR=$(SRC_DIR) DIST_DIR=$(DIST_DIR)
	@echo "Files copied and compiled."

copy_files:
	@for src_file in $(shell find $(SRC_DIR) -type f ! -name "*.ts"); do \
		dist_file=$(DIST_DIR)/$$src_file; \
		mkdir -p $$(dirname $$dist_file); \
		cp $$src_file $$dist_file; \
	done

convert_to_js:
	@echo "Starting TypeScript compiler..."
	@$(TSC) --watch
	@echo "TypeScript compiler is watching for changes..."

.PHONY: all clean copy_and_compile copy_files convert_to_js

start_app:
	@echo "Starting myProject..."
	@trap 'echo "Stopping server..."; exit 0' INT; go run .
	@echo "Server started."