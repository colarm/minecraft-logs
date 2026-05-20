# 部署流程

## 整体架构

```
VPS1 (Minecraft 服务器)                  VPS2 (核心服务)
┌───────────────────────────┐            ┌──────────────────────────────────┐
│  mc-log-agent (Go)         │   ──────►  │  NATS (消息队列, :4222)           │
│  读取MC日志 → 解析 → 发到NATS │  NATS    │      │                             │
└───────────────────────────┘            ├──────────────────────────────────┤
                                         │  Server (NestJS, :3001)          │
                                         │  消费NATS消息 → 写入PostgreSQL     │
                                         │  提供 REST API                     │
                                         ├──────────────────────────────────┤
                                         │  PostgreSQL (TimescaleDB, :5432) │
                                         │  时序数据库, 自动压缩/保留策略     │
                                         ├──────────────────────────────────┤
                                         │  Web (Next.js, :3000)            │
                                         │  前端面板, 查询API展示数据         │
                                         └──────────────────────────────────┘
```

- **Agent**（Go）：运行在 Minecraft 服务器所在机器上，tail 日志、解析事件、通过 NATS 发送
- **NATS**：消息中间件，接收 Agent 上报的日志事件
- **Server**（NestJS）：订阅 NATS，消费消息写入 TimescaleDB，提供 REST API
- **PostgreSQL**（TimescaleDB）：时序数据库，自动压缩、保留策略、物化视图
- **Web**（Next.js）：前端面板，调用 Server API 展示数据

---

## 第一步：VPS2 部署核心服务

```bash
cd deploy
```

### 1. 准备环境变量

```bash
cp ../.env.example .env
```

修改 `.env` 中的关键值：

| 变量 | 说明 |
|------|------|
| `NATS_TOKEN` | 改成强随机 token，用于 NATS 鉴权 |
| `POSTGRES_USER` | 数据库用户名，默认 `mclogs` |
| `POSTGRES_PASSWORD` | 数据库密码，默认 `mclogs_secret` |
| `POSTGRES_DB` | 数据库名，默认 `mclogs` |
| `NEXT_PUBLIC_API_URL` | 改成 `http://<VPS2公网IP>:3001` |

### 2. 启动服务

```bash
docker compose up -d --build
```

这会启动 4 个容器：

| 服务 | 端口 | 说明 |
|------|------|------|
| `nats` | 4222 | 消息队列，Agent 发送日志的目标 |
| `postgres` | 5432 | TimescaleDB，内部网络不暴露 |
| `server` | 3001 | NestJS 后端，消费 NATS 写 DB，提供 API |
| `web` | 3000 | Next.js 前端面板 |

### 3. 数据库初始化

首次启动时 PostgreSQL 自动执行 `deploy/postgres/init/` 下的脚本：

- **`001_extensions.sql`** — 启用 TimescaleDB 和 `uuid-ossp` 扩展
- **`002_schema.sql`** — 建表与策略：
  - `events` — 事件表，超表（按天分区），30 天后压缩，365 天后自动清理
  - `players` — 玩家信息表
  - `server_status` — 服务器状态表
  - `player_sessions` — 玩家会话表
  - `tps_per_minute` — 每分钟 TPS 物化视图（自动刷新）
  - `player_activity_per_hour` — 每小时玩家活动物化视图（自动刷新）

### 4. 开放防火墙

VPS2 需要对外开放以下端口：

- `4222` — NATS（Agent 连接）
- `3000` — Web 前端
- `3001` — Server API（如需直接调用）

```bash
# UFW 示例
sudo ufw allow 4222/tcp
sudo ufw allow 3000/tcp
sudo ufw allow 3001/tcp
```

---

## 第二步：VPS1 部署 Agent

Agent 是 Go 程序，运行在 Minecraft 服务器所在机器上。

### 1. 编译

```bash
cd agent
go mod download
go build -o mc-log-agent ./cmd/
```

### 2. 配置

参考 `config.yaml.example` 创建 `config.yaml`：

```yaml
server_id: survival-01
container_name: mc                    # MC 的 Docker 容器名
nats_url: nats://<VPS2公网IP>:4222
nats_token: <与 VPS2 .env 中 NATS_TOKEN 一致>
log_path: ""                          # 留空表示读 Docker 日志
```

如果 MC 没跑在 Docker 里，设置 `log_path` 为日志文件路径（如 `/opt/minecraft/logs/latest.log`），并留空 `container_name`。

### 3. 安装为 systemd 服务

```bash
sudo cp mc-log-agent /usr/local/bin/
sudo cp systemd/mc-log-agent.service /etc/systemd/system/
sudo cp config.yaml /etc/mc-log-agent/
sudo systemctl daemon-reload
sudo systemctl enable --now mc-log-agent
```

查看运行状态：

```bash
sudo systemctl status mc-log-agent
sudo journalctl -u mc-log-agent -f
```

---

## 数据流

```
MC 日志文件
    │
    ▼
Agent tail + 正则解析
    │
    ▼ (NATS over TCP :4222, token 鉴权)
NATS 消息队列
    │
    ▼ (Docker 内部网络)
Server 订阅消息 → 写入 PostgreSQL
    │
    ▼ (Docker 内部网络)
Web 前端调用 :3001 API → 查询 DB → 展示图表
```

1. MC 玩家加入/聊天/死亡/TPS 变化 → 写入日志文件
2. Agent 实时 tail 日志 → 正则匹配解析为结构化 JSON → 通过 NATS 发送
3. Server 订阅 NATS 主题 → 解析消息 → 写入 TimescaleDB 对应表
4. 用户访问 `http://VPS2_IP:3000` → Web 面板调 API → DB 查询 → 返回数据

---

## 网络与安全

- NATS `:4222` 对外开放但有 token 鉴权
- Server、PostgreSQL 跑在 `mc-logs-internal` 内部网络，不暴露到宿主机
- Web 前端 `:3000` 和 Server API `:3001` 对外可访问
- PostgreSQL 默认凭证需在 `.env` 中修改
