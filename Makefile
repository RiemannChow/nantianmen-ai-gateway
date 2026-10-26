# 南天门 AI 网关 — 本地与 CI 统一构建/测试入口
#
# 常用目标：
#   make test           运行后端 Go 单元测试（go test ./...）
#   make test-frontend  前端 lint + jest 单测
#   make build          构建后端二进制（嵌入 frontend/dist）
#   make build-frontend 仅构建前端 core 包
#   make run            本地以 config.yml 启动网关
#   make gen-key        生成一个 32 字节 AES 主密钥（写入 config.yml 的 secret.encryption_key）

GO ?= go
FRONTEND_DIR := frontend
BINARY := nantianmen-ai-gateway
CONFIG := config.yml

.PHONY: test test-frontend build build-frontend run gen-key clean

test:
	$(GO) test ./...

test-frontend:
	cd $(FRONTEND_DIR) && pnpm install && pnpm run lint && pnpm test -- --ci

build-frontend:
	cd $(FRONTEND_DIR) && pnpm install && ./node_modules/.bin/lerna run build --scope=core

build: build-frontend
	$(GO) build -o $(BINARY) .

run: build
	./$(BINARY) -c $(CONFIG)

gen-key:
	$(GO) run ./scripts/genkey

clean:
	rm -f $(BINARY)
