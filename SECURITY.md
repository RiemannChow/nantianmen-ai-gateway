# 安全政策

## 支持的版本

仅最新的正式发布版本提供安全更新。

## 报告漏洞

如果你认为在 nantianmen-ai-gateway 中发现了安全漏洞，请**不要**创建公开 issue。

请通过以下方式私下报告：

- Email: contact@nantianmen-ai-gateway.com（标题注明 `[security]`）

我们会在 72 小时内确认收到，并在 7 个工作日内给出初步评估与修复计划。
在修复发布前，请对漏洞细节保密。

## 部署安全基线

生产私有化部署时，请至少完成以下配置（开发默认值仅用于本地调试）：

1. **修改所有默认密码**：MySQL、Redis、管理员初始密码（`ADMIN_PASSWORD`）。
   镜像不再内置默认管理员密码，启动时会强制要求显式提供。
2. **配置 `secret.encryption_key`**（`make gen-key` 生成 32 字节随机值并 base64），
   否则 API Key、SMTP 密码等敏感配置将以明文存储。
3. **通过网络层限制管理端口**（默认 8088/18288）的暴露面，仅对内网或可信来源开放。
4. 关注 CI 中的依赖漏洞扫描（govulncheck / pnpm audit / Trivy）输出。
5. **admin 端点**：`/metrics` 与 `/debug/pprof` 默认只监听 `127.0.0.1:9464`。
   如需远程采集，把 `config.yml` 的 `metrics.addr` 改为对外地址时，务必通过网络层
   （防火墙/反向白名单）限制访问——pprof 会暴露堆栈与内部细节。
