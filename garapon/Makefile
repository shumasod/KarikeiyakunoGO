# ==============================================================================
# garapon — ビルドシステム
# ==============================================================================
# 使い方:  make <target>
# 例:      make build   make test   make run   make docker
# ==============================================================================

BINARY    := garapon
GO        := go
BUILD_DIR := bin
PKG       := ./...
COVERAGE  := $(BUILD_DIR)/coverage.out

# --- バージョン情報（git から自動取得、CI 環境では上書き可能）---
VERSION   ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT    ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -ldflags "\
  -X main.version=$(VERSION) \
  -X main.commit=$(COMMIT) \
  -X main.buildDate=$(BUILD_DATE) \
  -w -s"

# --- デフォルトターゲット ---
.DEFAULT_GOAL := help

.PHONY: all build run test test-race test-cover bench \
        vet fmt tidy clean docker docker-run help

# ==============================================================================
# 開発フロー
# ==============================================================================

## all          : fmt → vet → test-race → build を順次実行
all: fmt vet test-race build

## build        : バイナリをビルド（bin/garapon）
build: $(BUILD_DIR)
	@echo ">> Building $(BINARY) $(VERSION) ..."
	$(GO) build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY) .
	@echo ">> Output: $(BUILD_DIR)/$(BINARY)"

## run          : ビルドしてサーバー起動
run: build
	@echo ">> Starting $(BINARY) ..."
	./$(BUILD_DIR)/$(BINARY)

# ==============================================================================
# テスト
# ==============================================================================

## test         : ユニットテスト（verboseモード）
test:
	@echo ">> Running tests ..."
	$(GO) test -v $(PKG)

## test-race    : データ競合検出つきテスト（推奨）
test-race:
	@echo ">> Running tests with race detector ..."
	$(GO) test -race -v $(PKG)

## test-cover   : カバレッジ計測 → bin/coverage.html を生成
test-cover: $(BUILD_DIR)
	@echo ">> Running tests with coverage ..."
	$(GO) test -race -coverprofile=$(COVERAGE) -covermode=atomic $(PKG)
	$(GO) tool cover -func=$(COVERAGE)
	$(GO) tool cover -html=$(COVERAGE) -o $(BUILD_DIR)/coverage.html
	@echo ">> Report: $(BUILD_DIR)/coverage.html"

## bench        : ベンチマークテスト実行
bench:
	@echo ">> Running benchmarks ..."
	$(GO) test -bench=. -benchmem $(PKG)

# ==============================================================================
# コード品質
# ==============================================================================

## vet          : go vet による静的解析
vet:
	@echo ">> Running go vet ..."
	$(GO) vet $(PKG)

## fmt          : go fmt によるコードフォーマット
fmt:
	@echo ">> Formatting code ..."
	$(GO) fmt $(PKG)

## tidy         : go.mod / go.sum の整理
tidy:
	@echo ">> Tidying modules ..."
	$(GO) mod tidy

# ==============================================================================
# Docker
# ==============================================================================

## docker       : Docker イメージをビルド（garapon:VERSION, garapon:latest）
docker:
	@echo ">> Building Docker image garapon:$(VERSION) ..."
	docker build \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg COMMIT=$(COMMIT) \
	  --build-arg BUILD_DATE=$(BUILD_DATE) \
	  -t garapon:$(VERSION) \
	  -t garapon:latest \
	  .
	@echo ">> Image: garapon:$(VERSION)"

## docker-run   : Docker コンテナを起動（ポート 8081）
docker-run:
	@echo ">> Running garapon:latest on :8081 ..."
	docker run --rm -p 8081:8081 garapon:latest

# ==============================================================================
# ユーティリティ
# ==============================================================================

## clean        : ビルド成果物を削除
clean:
	@echo ">> Cleaning ..."
	rm -rf $(BUILD_DIR)
	@echo ">> Done"

$(BUILD_DIR):
	@mkdir -p $(BUILD_DIR)

## help         : このヘルプを表示
help:
	@echo ""
	@echo "  ╔══════════════════════════════════════════╗"
	@echo "  ║  garapon ガラガラポン抽選システム        ║"
	@echo "  ║  Build System                            ║"
	@echo "  ╚══════════════════════════════════════════╝"
	@echo ""
	@echo "  Usage: make <target>"
	@echo ""
	@grep -E '^## ' $(MAKEFILE_LIST) | \
	  sed -e 's/## //' | \
	  awk -F ':' '{ printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2 }'
	@echo ""
	@echo "  VERSION=$(VERSION)  COMMIT=$(COMMIT)"
	@echo ""
