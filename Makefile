SHELL := bash
.ONESHELL:
.SHELLFLAGS := -eu -o pipefail -c
.DELETE_ON_ERROR:
MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
.SILENT:

# go builds are fast enough that we can just build on demand instead of trying to do any fancy
# change detection
build: clean check-changes
.PHONY: build

check-changes:
	go build ./cmd/check-changes

clean:
	rm -f ./check-changes
.PHONY: clean

test:
	go test ./...
.PHONY: test
