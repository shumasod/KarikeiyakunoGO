# ==============================================================================
# tse-scanner — 東証急騰スキャナー ビルドシステム
# ==============================================================================
BINARY    := tse-scanner
GO        := go
BUILD_DIR := bin
PKG       := ./...
COVERAGE  := $(BUILD_DIR)/coverage.out

VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "\
  -X main.version=$(VERSION) \
  -X main.commit=$(COMMIT) \
  -X main.buildDate=$(BUILD_DATE) \
  -w -s"

.DEFAULT_GOAL := help

.PHONY: all build run test test-race test-cover vet fmt tidy clean help

## all         : fmt → vet → test-race → build
all: fmt vet test-race build

## build       : bin/tse-scanner をビルド
build: $(BUILD_DIR)
	@echo ">> Building $(BINARY) $(VERSION) ..."
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) .
	@echo ">> Output: $(BUILD_DIR)/$(BINARY)"

## run         : ビルドして実行（60秒間隔）
run: build
	./$(BUILD_DIR)/$(BINARY) -interval 60s

## run-fast    : 30秒間隔で実行（デモ用）
run-fast: build
	./$(BUILD_DIR)/$(BINARY) -interval 30s -min-score 10

## test        : ユニットテスト
test:
	@echo ">> Running tests ..."
	$(GO) test -v $(PKG)

## test-race   : データ競合検出つきテスト
test-race:
	@echo ">> Running tests with race detector ..."
	$(GO) test -race -v $(PKG)

## test-cover  : カバレッジ計測 → bin/coverage.html
test-cover: $(BUILD_DIR)
	@echo ">> Running tests with coverage ..."
	$(GO) test -race -coverprofile=$(COVERAGE) -covermode=atomic $(PKG)
	$(GO) tool cover -func=$(COVERAGE)
	$(GO) tool cover -html=$(COVERAGE) -o $(BUILD_DIR)/coverage.html
	@echo ">> Report: $(BUILD_DIR)/coverage.html"

## vet         : go vet 静的解析
vet:
	$(GO) vet $(PKG)

## fmt         : go fmt フォーマット
fmt:
	$(GO) fmt $(PKG)

## tidy        : go.mod 整理
tidy:
	$(GO) mod tidy

## clean       : ビルド成果物を削除
clean:
	rm -rf $(BUILD_DIR)

$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

## help        : このヘルプを表示
help:
	@echo ""
	@echo "  📈 tse-scanner 東証急騰スキャナー"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | \
	  awk -F ':' '{ printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2 }'
	@echo ""
	@echo "  VERSION=$(VERSION)"
	@echo ""
