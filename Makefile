.PHONY: build test lint fmt clean run check repl structure save-progress help

# 项目信息
BINARY_NAME=jpl
VERSION=$(shell grep 'Version.*=' version.go | grep '"' | head -1 | cut -d'"' -f2)
GO=go
GOFLAGS=-v

# 代码检查（简化版，不依赖外部工具）
lint:
	@echo "运行代码检查..."
	@echo "1. 检查代码格式..."
	@$(GO) fmt ./... | xargs -r echo "需要格式化的文件:" || echo "✓ 代码格式正确"
	@echo ""
	@echo "2. 运行静态分析..."
	@$(GO) vet ./... 2>&1 | head -20 || true
	@echo ""
	@echo "3. 检查常见代码问题..."
	@# 检查是否有遗留的 fmt.Print 调试语句（非测试文件）
	@DEBUG_LINES=$$(grep -r "fmt\.Print" --include="*.go" . | grep -v "_test.go" | grep -v "scripts/" | wc -l); \
    if [ "$$DEBUG_LINES" -gt 0 ]; then \
        echo "⚠ 发现 $$DEBUG_LINES 处 fmt.Print 调试语句"; \
    else \
        echo "✓ 无调试输出语句"; \
    fi
	@echo ""
	@echo "4. 代码统计..."
	@echo "   代码行数: $$(find . -name '*.go' -not -path './vendor/*' | xargs grep -P -v '^[[:space:]]*//' | wc -l | tail -1 | awk '{print $$1}')"
	@echo "   测试覆盖率:"
	@$(GO) test -cover ./... 2>&1 | grep -E "coverage:|ok |FAIL" | head -10
	@echo ""
	@echo "检查完成！"

# 代码格式化
fmt:
	$(GO) fmt ./...

# 构建
build:
	$(GO) build $(GOFLAGS) -o bin/$(BINARY_NAME) ./cmd/jpl

# 发布版本（针对当前系统）
release: clean
	@echo "开始构建发布版本 v$(VERSION)..."
	@mkdir -p release
	@echo "构建优化版本..."
	$(GO) build -ldflags="-s -w -X github.com/gnuos/jpl.ReleaseDate=$(shell date +%Y-%m-%d)" \
		-trimpath -o release/$(BINARY_NAME) ./cmd/jpl
	@echo "复制文档..."
	@cp README.md release/
	@cp CHANGELOG.md release/
	@cp LICENSE release/ 2>/dev/null || echo "LICENSE 文件不存在，跳过"
	@cp -r examples release/ 2>/dev/null || echo "examples 目录不存在，跳过"
	@echo "创建归档..."
	@OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
	ARCH=$$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/'); \
	RELEASE_NAME="$(BINARY_NAME)-$(VERSION)-$$OS-$$ARCH"; \
	if [ "$$OS" = "darwin" ] || [ "$$OS" = "linux" ]; then \
		tar -czf release/$$RELEASE_NAME.tar.gz -C release $(BINARY_NAME) README.md CHANGELOG.md LICENSE examples 2>/dev/null; \
		echo "发布包: release/$$RELEASE_NAME.tar.gz"; \
	else \
		zip -r release/$$RELEASE_NAME.zip release/$(BINARY_NAME) release/README.md release/CHANGELOG.md release/LICENSE release/examples 2>/dev/null; \
		echo "发布包: release/$$RELEASE_NAME.zip"; \
	fi
	@echo "生成校验和..."
	@cd release && sha256sum *.tar.gz *.zip 2>/dev/null > checksums.txt || true
	@echo "发布版本构建完成！"
	@echo "发布文件位于: release/"

# 测试
test:
	$(GO) test $(GOFLAGS) ./...

# 测试覆盖率
test-cover:
	$(GO) test $(GOFLAGS) -cover ./...

# 基准测试
bench:
	$(GO) test $(GOFLAGS) -bench=. ./...

# 清理
clean:
	rm -rf bin/
	$(GO) clean

# 运行脚本
run:
	$(GO) run ./cmd/jpl run $(FILE)

# 语法检查
check:
	$(GO) run ./cmd/jpl check $(FILE)

# REPL
repl:
	$(GO) run ./cmd/jpl repl

# 安装
install:
	$(GO) install ./cmd/jpl

# 依赖整理
tidy:
	$(GO) mod tidy

# 生成文档
docs:
	$(GO) doc -all ./...

# 帮助
help:
	@echo "可用命令:"
	@echo "  make build        - 构建二进制文件"
	@echo "  make test         - 运行测试"
	@echo "  make test-cover   - 运行测试并生成覆盖率报告"
	@echo "  make bench        - 运行基准测试"
	@echo "  make lint         - 运行代码检查"
	@echo "  make fmt          - 格式化代码"
	@echo "  make clean        - 清理构建文件"
	@echo "  make release      - 构建发布版本（当前系统）"
	@echo "  make run FILE=x   - 运行脚本文件"
	@echo "  make check FILE=x - 检查脚本语法"
	@echo "  make repl         - 启动 REPL"
	@echo "  make install      - 安装到 GOPATH/bin"
	@echo "  make tidy         - 整理依赖"
	@echo "  make docs         - 生成文档"
	@echo "  make help         - 显示帮助"
