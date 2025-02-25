GO ?= go
HAS_GO := $(shell hash $(GO) > /dev/null 2>&1 && echo yes)

ifeq ($(HAS_GO), yes)
	GOPATH ?= $(shell $(GO) env GOPATH)
	export PATH := $(GOPATH)/bin:$(PATH)
	CGO_CFLAGS ?= $(shell $(GO) env CGO_CFLAGS)
endif

TAG ?= $(shell git describe --tags --abbrev=0 HEAD)
DATE_FMT = +"%Y-%m-%dT%H:%M:%S%z"
BUILD_DATE ?= $(shell date "$(DATE_FMT)")

TEST_TAGS ?=
GOTESTFLAGS ?=
GO_TEST_PACKAGES ?= $(filter-out $(shell $(GO) list ./tests/...),$(shell $(GO) list ./...))
GO_INTEGRATION_TEST_PACKAGES ?= $(shell $(GO) list ./tests/...)

ifeq ($(IS_WINDOWS),yes)
	GOFLAGS := -v -buildmode=exe
	EXECUTABLE ?= greekmilkbot.exe
else
	GOFLAGS := -v
	EXECUTABLE ?= greekmilkbot
endif

.PHONY: help
help:
	@echo "Make Help:"
	@echo " - \"\"                               equivalent to \"build\""
	@echo " - deps                             install dependencies"
	@echo " - deps-mod                         install go mod dependencies"
	@echo " - deps-tools                       install tool dependencies"
	@echo " - tidy                             run go mod tidy"
	@echo " - test                             run go tests"

.PHONY: go-check
go-check:
	$(eval MIN_GO_VERSION_STR := $(shell grep -Eo '^go\s+[0-9]+\.[0-9]+' go.mod | cut -d' ' -f2))
	$(eval MIN_GO_VERSION := $(shell printf "%03d%03d" $(shell echo '$(MIN_GO_VERSION_STR)' | tr '.' ' ')))
	$(eval GO_VERSION := $(shell printf "%03d%03d" $(shell $(GO) version | grep -Eo '[0-9]+\.[0-9]+' | tr '.' ' ');))
	@if [ "$(GO_VERSION)" -lt "$(MIN_GO_VERSION)" ]; then \
		echo "GreekMilkBot requires Go $(MIN_GO_VERSION_STR) or greater to build. You can get it at https://go.dev/dl/"; \
		exit 1; \
	fi

.PHONY: tidy
tidy:
	$(eval MIN_GO_VERSION := $(shell grep -Eo '^go\s+[0-9]+\.[0-9.]+' go.mod | cut -d' ' -f2))
	$(GO) mod tidy -compat=$(MIN_GO_VERSION)

.PHONY: test
test: unit-test integration-test

.PHONY: unit-test
unit-test:
	@echo "Running unit-test $(GOTESTFLAGS) -tags '$(TEST_TAGS)'..."
	@$(GO) test $(GOTESTFLAGS) -tags='$(TEST_TAGS)' $(GO_TEST_PACKAGES)

.PHONY: integration-test
integration-test:
	@echo "Running integration-test with $(GOTESTFLAGS) -tags '$(TEST_TAGS)'..."
	@$(GO) test $(GOTESTFLAGS) -tags='$(TEST_TAGS)' $(GO_INTEGRATION_TEST_PACKAGES)
