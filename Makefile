ifneq (,$(wildcard .env))
    include .env
    export
endif

BUILD_DIR:=.bin
SRC_DIR:=./cmd

build:
	@echo "Building the core project ..."
	go build -o $(BUILD_DIR)/main $(SRC_DIR)/main.go
	@echo "Build Completed"

run-prod:
	@echo "Running the project in production mode ..."
	$(BUILD_DIR)/core -dev=false

run-dev:
	@echo "Running the project in development mode ..."
	go run $(SRC_DIR)/main.go -dev=true

generate-proto:
	@echo "Generating schema proto ..."
	protoc --go_out=. --go_opt=paths=source_relative internal/presentation/protobuf/schema.proto