MKFILE_DIR      :=  $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
.DEFAULT_GOAL   :=  help
SHELL           :=  /bin/bash
MAKEFLAGS += --no-print-directory
PLATFROM := $(go env GOOS)

.PHONY: build
build: ## run go build and go install
	@go build -v -o build/lf-cli && go install

.PHONY: build/all
build/all: ## Run `go build` to create binaries for different OS
	@export GOOS=linux; export GOARCH=amd64; go build -v -o build/linux/lf-cli_linux_amd64
	@export GOOS=darwin; export GOARCH=amd64; go build -v -o build/macos/lf-cli_macos_amd64

.PHONY: test
test: ## run tests
	@go test ./internal

.PHONY: test/verbose
test/verbose: ## run tests with verbose output
	@go test -v ./internal

.PHONY: help
help: ## Makefile Help Page
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[\/\%a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-21s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
