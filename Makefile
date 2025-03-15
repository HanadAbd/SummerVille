SRC_DIR := web/src
DIST_DIR := web/dist

TS_EXT := .ts
JS_EXT := .js

# TypeScript Compiler
TSC := tsc

all: clean copy_and_compile start_docker start_app 

clean:
	@echo "Removing dist directory..."
	@rm -rf $(DIST_DIR)
	@echo "Dist directory removed."

copy_and_compile:
	@echo "Copying and compiling files..."
	@mkdir -p $(DIST_DIR)
	@$(MAKE) copy_files SRC_DIR=$(SRC_DIR) DIST_DIR=$(DIST_DIR)
	@$(MAKE) convert_to_js
	@echo "Files copied and compiled."

copy_files:
	@echo "Copying files from $(SRC_DIR) to $(DIST_DIR)..."
	@cp -r $(SRC_DIR)/* $(DIST_DIR)/

convert_to_js:
	@echo "Starting TypeScript compiler..."
	@$(TSC)
	@echo "TypeScript compiler is watching for changes..."

start_docker:
	@echo "Starting Docker..."
	@docker-compose up -d
	@echo "Docker started."

.PHONY: all clean copy_and_compile copy_files convert_to_js

start_app:
	@echo "Starting myProject..."
	@trap 'echo "Stopping server..."; exit 0' INT; go run .
	@echo "Server started."