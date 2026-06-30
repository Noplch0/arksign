# =============================================================================
# arkSign — 森空岛自动签到工具
# =============================================================================

APP_NAME    := arkSign
DIST_DIR    := dist
GO          := go
GOFLAGS     := CGO_ENABLED=0

# 版本信息（从 git tag 提取）
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME  := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
LDFLAGS     := -s -w \
	-X 'main.Version=$(VERSION)' \
	-X 'main.BuildTime=$(BUILD_TIME)'

# =============================================================================
# 平台定义
# =============================================================================

PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64

# =============================================================================
# 主目标
# =============================================================================

.PHONY: all
all: clean build  ## 清理并构建所有平台

.PHONY: build
build: $(addprefix build-,$(PLATFORMS))  ## 构建所有平台

.PHONY: dev
dev:  ## 构建当前平台（用于本地开发）
	$(GOFLAGS) $(GO) build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME) .

# =============================================================================
# 各平台构建
# =============================================================================

.PHONY: build-linux-amd64
build-linux-amd64: $(DIST_DIR)  ## 构建 Linux x86_64
	GOOS=linux GOARCH=amd64 $(GOFLAGS) $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)_linux_amd64 .

.PHONY: build-linux-arm64
build-linux-arm64: $(DIST_DIR)  ## 构建 Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOFLAGS) $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)_linux_arm64 .

.PHONY: build-darwin-amd64
build-darwin-amd64: $(DIST_DIR)  ## 构建 macOS x86_64
	GOOS=darwin GOARCH=amd64 $(GOFLAGS) $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)_darwin_amd64 .

.PHONY: build-darwin-arm64
build-darwin-arm64: $(DIST_DIR)  ## 构建 macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 $(GOFLAGS) $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)_darwin_arm64 .

.PHONY: build-windows-amd64
build-windows-amd64: $(DIST_DIR)  ## 构建 Windows x86_64
	GOOS=windows GOARCH=amd64 $(GOFLAGS) $(GO) build \
		-ldflags "$(LDFLAGS)" \
		-o $(DIST_DIR)/$(APP_NAME)_windows_amd64.exe .

# 构建目录
$(DIST_DIR):
	@mkdir -p $(DIST_DIR)

# =============================================================================
# 开发辅助
# =============================================================================

.PHONY: test
test:  ## 运行测试（含竞态检测 + 覆盖率）
	$(GO) test -v -race -coverprofile=coverage.txt ./...

.PHONY: cover
cover: test  ## 运行测试并在浏览器中查看覆盖率
	$(GO) tool cover -html=coverage.txt

.PHONY: lint
lint:  ## 运行代码检查（需安装 golangci-lint）
	@which golangci-lint > /dev/null || (echo "请先安装 golangci-lint: https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

.PHONY: fmt
fmt:  ## 格式化代码
	$(GO) fmt ./...

.PHONY: vet
vet:  ## 运行 go vet
	$(GO) vet ./...

.PHONY: tidy
tidy:  ## 整理依赖
	$(GO) mod tidy

.PHONY: clean
clean:  ## 删除构建产物
	rm -rf $(DIST_DIR)

# =============================================================================
# 帮助
# =============================================================================

.PHONY: help
help:  ## 显示此帮助信息
	@echo "arkSign Makefile"
	@echo ""
	@echo "构建目标:"
	@echo "  make all                 清理并构建所有平台"
	@echo "  make build               构建所有平台"
	@echo "  make dev                 构建当前平台（开发用）"
	@echo "  make build-linux-amd64   构建 Linux x86_64"
	@echo "  make build-linux-arm64   构建 Linux ARM64"
	@echo "  make build-darwin-amd64  构建 macOS x86_64"
	@echo "  make build-darwin-arm64  构建 macOS ARM64"
	@echo "  make build-windows-amd64 构建 Windows x86_64"
	@echo ""
	@echo "开发目标:"
	@echo "  make test                运行测试"
	@echo "  make cover               测试覆盖率（浏览器）"
	@echo "  make lint                代码检查"
	@echo "  make fmt                 格式化代码"
	@echo "  make vet                 go vet 检查"
	@echo "  make tidy                整理 go.mod"
	@echo "  make clean               删除构建产物"
	@echo "  make help                显示此帮助"
	@echo ""
	@echo "版本: $(VERSION)"
