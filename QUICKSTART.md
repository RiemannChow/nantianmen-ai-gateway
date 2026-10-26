# nantianmen-ai-gateway 快速启动 / 接入指南

本项目是一个 **AI / API 网关的控制面（Control Plane）**：本身不做流量代理，而是把 AI 服务、订阅、策略、上游等配置写入共享存储，由外部的 Apinto 数据面（Data Plane）真正执行代理。

本指南帮助你**从源码快速跑起来**，并说明各依赖组件的作用。

---

## 1. 环境要求

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | **≥ 1.23.4**（toolchain 1.23.6） | 低于此版本编译会失败 |
| Node.js | ≥ 18 | 仅前端构建需要 |
| pnpm | latest | 前端包管理（仓库无 lockfile，需全新 `pnpm install`） |
| MySQL | 8.0 | 元数据主存储 |
| Redis | 7 | 缓存 |
| Docker（可选） | — | 一键拉起全套依赖/官方镜像 |

> 前端通过 `//go:embed dist` 嵌入后端二进制，**必须先构建前端生成 `frontend/dist`，否则 `go build` 会报 embed 错误**。

---

## 2. 方式一：本地源码构建运行（推荐，用于开发 / 验证你的改动）

### 2.1 准备依赖（MySQL + Redis）
最简方式用 Docker 拉起两个依赖（也可连你已有的实例）：

```bash
docker run -d --name ntm-mysql -p 33306:3306 \
  -e MYSQL_ROOT_PASSWORD=123456 \
  -e MYSQL_DATABASE=nantianmen-ai-gateway \
  mysql:8.0.37 --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci

docker run -d --name ntm-redis -p 6379:6379 \
  redis:7.2.4 redis-server --requirepass 123456
```

### 2.2 构建前端（生成 `frontend/dist`）
```bash
cd frontend
pnpm install --registry https://registry.npmmirror.com   # 国内可加镜像加速
pnpm build                                               # 等价于 lerna run build --scope=core
cd ..
```

### 2.3 准备后端配置
```bash
cp scripts/resource/config.yml.tpl config.yml
```
按需修改 `config.yml`（字段说明见第 4 节），示例（连上面两个容器）：

```yaml
port: 8088
mysql:
  user_name: "root"
  password: "123456"
  ip: "127.0.0.1"
  port: 33306
  db: "nantianmen-ai-gateway"
redis:
  user_name: ""
  password: "123456"
  addr:
    - 127.0.0.1:6379
error_log:
  dir: work/logs
  file_name: error.log
  log_level: info
  log_expire: 7d
  log_period: day
```

### 2.4 构建并启动后端
```bash
go mod tidy            # 需要网络拉取依赖
go build -o nantianmen-ai-gateway .
./nantianmen-ai-gateway -c config.yml     # 默认读取 config.yml；-c 可指定路径
```

看到监听日志后，打开浏览器访问：**http://localhost:8088**（端口取 `config.yml` 中的 `port`）。

> 也可用仓库自带的进程脚本：`bash scripts/resource/run.sh start|stop|restart`（需在编译产物同目录放置名为 `nantianmen-ai-gateway` 的二进制）。

---

## 3. 方式二：Docker Compose 一键启动（运行官方已发布版本）

```bash
cd scripts
docker compose -f docker-compose.yml up -d
```

启动后访问：**http://localhost:18288**（端口映射 `18288:8288`）。

⚠️ **注意**：compose 中的主服务镜像是 `apipark/apipark:v1.2.0-beta`（上游官方发布版），**不会包含你本地的 rebrand 改动**。它适合快速体验产品形态；要运行你改过的源码，请走第 2 节本地构建，或把本地构建的镜像替换进 compose。

该 compose 已一并拉起：MySQL、Redis、InfluxDB（监控）、以及外部数据面 Apinto（`eolinker/apinto-gateway`）。

---

## 4. 配置说明（`config.yml`）

| 字段 | 必填 | 说明 |
|------|------|------|
| `port` | 是 | 后端监听端口（默认示例 8088） |
| `mysql` | 是 | 元数据库地址 / 账号 / 库名 |
| `redis` | 是 | 缓存地址与密码 |
| `error_log` | 否 | 日志目录、文件名、级别、切割周期 |
| `nsq` | 否 | AI 事件队列（`addr` / `topic_prefix`），用于用量统计；不配则相关统计不可用 |
| `influxdb` / `loki` | 否 | 监控指标与日志外送，需搭配数据面开启 |

---

## 5. 组件依赖关系

| 组件 | 是否在本仓库 | 作用 |
|------|--------------|------|
| 本服务（控制面） | ✅ | 配置编排、管理后台、API |
| MySQL / Redis | ❌ 外部 | 元数据存储 / 缓存 |
| **Apinto 数据面** (`eolinker/apinto-gateway`) | ❌ 外部 | **真正执行 AI/REST 代理**；本项目只生成并推送其插件配置 |
| NSQ | ❌ 外部 | 接收数据面发来的 AI 用量事件 |
| `app/ai-event-handler` | ✅ 独立二进制 | NSQ 消费者，落库用量统计 |
| InfluxDB / Loki | ❌ 外部 | 监控指标 / 日志 |

> 要点：即使只跑本服务 + MySQL + Redis，也能启动管理后台并做配置；但要**实际代理 AI 请求 / 看到监控**，必须有外部 Apinto 数据面（见 compose 或单独部署它）。

---

## 6. 快速验证启动成功

服务起来后，可用管理 API 探活（返回系统通用配置即代表运行正常）：

```bash
curl -s http://localhost:8088/api/v1/system/general | head
```

或查看错误日志：

```bash
tail -f work/logs/error.log
```

---

## 7. 常见问题（FAQ）

**Q1：`go build` 报错类似 `pattern dist: no matching files` 或 embed 失败**
→ 没构建前端。先执行第 2.2 节 `pnpm install && pnpm build`，确保 `frontend/dist` 存在。

**Q2：编译报 `go.mod requires go >= 1.23.4`**
→ 本机 Go 版本过低。升级到 1.23.4 及以上（项目使用 1.23 toolchain）。

**Q3：`go mod tidy` 拉不到依赖 / 超时**
→ 需要外网。确保本机可访问 Go 模块代理（如 `proxy.golang.org` 或 `goproxy.cn`），依赖 `github.com/RiemannChow/nantianmen-ai-gateway` 通过模块代理拉取（或配置本地 `replace`）。

**Q4：启动后访问 8088 没反应**
→ 确认 `config.yml` 的 `port` 与访问端口一致；确认 MySQL/Redis 连通；看 `error.log`。

**Q5：想要 AI 调用真正生效**
→ 单独部署 Apinto 数据面，并在本服务「集群」配置里填入其管理地址，把配置发布过去（参见架构文档的「发布到数据面」流程）。

---

## 8. 下一步

- 阅读架构分析了解 `controller → module → service → stores` 分层与发布流程。
- 从 `init.go`（所有模块的 blank import 注册入口）和 `plugins/core/core.go`（API 聚合）开始通读代码。
- 本地改动后端后，重新 `go build` 即可；前端改动需重新 `pnpm build`。
