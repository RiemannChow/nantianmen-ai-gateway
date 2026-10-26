# 贡献指南

感谢你愿意为 nantianmen-ai-gateway 出力！在提交 PR 之前，请阅读以下约定。

## 开发环境

| 组件 | 版本要求 | 说明 |
| --- | --- | --- |
| Go | 1.23.x | 后端 |
| Node.js | 18.x | 前端 |
| pnpm | 8.x | 前端包管理（仓库根 `frontend/` 目录） |
| MySQL | 8.0 | 本地开发 |
| Redis | 7.x | 本地开发 |

## 本地构建与测试

```bash
# 后端：编译 + 单元测试 + lint
make test

# 前端：安装依赖 + lint + jest 单测
make test-frontend

# 构建完整产物（前端 dist 会嵌入二进制）
make build

# 生成 AES 主密钥（生产部署必填）
make gen-key
```

> 注意：`frontend` 下的脚本面向类 Unix shell 编写（CI 与 Makefile 均如此），
> Windows 用户请使用 WSL 或 Git Bash。

## 提交 PR 前自查清单

- [ ] `go build ./...` 通过
- [ ] `go test ./...` 通过（新增逻辑请附带单测）
- [ ] `golangci-lint run` 无新增告警
- [ ] 前端改动：`pnpm run lint` 与 `pnpm test` 通过
- [ ] 不提交任何密钥、密码、Token；配置示例使用占位符
- [ ] 涉及行为变化的改动，同步更新 README / QUICKSTART 文档

## 代码约定

- 错误必须处理或显式 `//nolint` 说明理由，不要静默丢弃
- 数据库访问统一走 `stores` 层封装，禁止跨模块直接访问他人 store
- 提交信息使用 Conventional Commits（`feat:` / `fix:` / `docs:` / `refactor:` 等）

## 报告安全问题

请**不要**通过公开 issue 报告安全漏洞，参见 [SECURITY.md](./SECURITY.md)。
