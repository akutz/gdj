# Copyright 2022 The Go Authors. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# If you update this file, please follow
# https://www.thapaliya.com/en/writings/well-documented-makefiles/

# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL := /usr/bin/env bash

.DEFAULT_GOAL := help

all: test

## --------------------------------------
## Help
## --------------------------------------

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9._-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


## --------------------------------------
## Golang
## --------------------------------------

GO_VERSIONS_DIR  ?= $(HOME)/.go

GO_1_17_VERSIONS += 1.17.13
GO_1_18_VERSIONS += 1.18.9
GO_1_19_VERSIONS += 1.19.4
GO_1_20_VERSIONS += 1.20rc1

GO_VERSIONS ?= $(GO_1_17_VERSIONS) \
               $(GO_1_18_VERSIONS) \
               $(GO_1_19_VERSIONS) \
               $(GO_1_20_VERSIONS)

GO_1_17_BIN ?= $(HOME)/.go/$(GO_1_17_VERSION)/bin/go
GO_1_18_BIN ?= $(HOME)/.go/$(GO_1_18_VERSION)/bin/go
GO_1_19_BIN ?= $(HOME)/.go/$(GO_1_19_VERSION)/bin/go
GO_1_20_BIN ?= $(HOME)/.go/$(GO_1_20_VERSION)/bin/go


## --------------------------------------
## Tests
## --------------------------------------

define GO_TEST_DEF
ifneq (1,$(USE_DOCKER))
GO_$1_BIN := $(GO_VERSIONS_DIR)/$1/bin/go
endif
.PHONY: test-go$1
ifneq (,$$(wildcard $$(GO_$1_BIN)))
test-go$1: | $$(GO_$1_BIN)
	$$(GO_$1_BIN) test -v .
	cd canaries && $$(GO_$1_BIN) test -v -benchmem -bench . .
else
test-go$1:
	docker run --rm -v $$$$(pwd):/w -w /w golang:$1 go test -v .
	docker run --rm -v $$$$(pwd):/w -w /w/canaries golang:$1 go test -v -benchmem -bench . .
endif
endef

# Set up all of the Go, test targets.
$(foreach v,$(GO_VERSIONS),$(eval $(call GO_TEST_DEF,$(v))))

.PHONY: test
test: ## Run tests with all Go versions
test: $(addprefix test-go,$(GO_VERSIONS))
