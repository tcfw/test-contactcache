GOBUILD := CGO_ENABLED=0 go build
BUILD_DIR := ./build
BUILD_FLAGS := -ldflags="-s -w" -trimpath

.PHONY:all
all: build

.PHONY: build
build:
	$(GOBUILD) $(BUILD_FLAGS) -o $(BUILD_DIR)/contactcache ./cmd/contactcache

.PHONY: run
run: build
	$(BUILD_DIR)/contactcache start

.PHONY: test
test:
	go test -v github.com/tcfw/test-contactcache/pkg/contactcache