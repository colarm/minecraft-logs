# 项目目标

构建一个：

```text id="8o3gg8"
Minecraft 世界服务器日志实时分析平台
```

系统需要：

* 实时读取游戏服务器日志
* 解析日志事件
* 将日志事件发送到远程 Web VPS
* 存储与分析数据
* 提供 Web Dashboard 展示
* 尽量不影响游戏服务器性能

---

# 核心设计原则

## 1. 游戏服极轻量

游戏服务器 VPS 只负责：

```text id="a14k1m"
日志采集
```

避免：

* 数据库
* Web服务
* 大量计算
* 图表
* 前端
* 重型框架

目标：

```text id="r0cfj9"
不影响 TPS
```

---

## 2. Web VPS 负责所有重任务

包括：

* 日志分析
* 聚合统计
* 数据库存储
* API
* 前端 Dashboard
* 历史数据
* 图表
* 用户系统

---

# 最终整体架构

```text id="g7gx3o"
┌──────────────────────┐
│ VPS1 游戏服务器       │
│                      │
│ Minecraft Server     │
│        ↓             │
│ world.log            │
│        ↓             │
│ Go Log Agent         │
│  - tail日志          │
│  - 解析事件          │
│  - publish MQ        │
└────────┬─────────────┘
         │
         │ NATS
         │
┌────────▼─────────────┐
│ VPS2 Web服务器       │
│                      │
│ NATS Consumer        │
│        ↓             │
│ NestJS Backend       │
│  - 消费消息          │
│  - 数据分析          │
│  - 聚合统计          │
│  - REST API          │
│                      │
│ PostgreSQL           │
│ + TimescaleDB        │
│                      │
│ Next.js Frontend     │
│ + ECharts            │
└──────────────────────┘
```

---

# VPS1（游戏服务器）设计

# 目标

```text id="6wn7zc"
超低资源占用
```

---

# 技术栈

| 功能      | 技术                                                                     |
| ------- | ---------------------------------------------------------------------- |
| Agent语言 | Go                                                                     |
| 日志监听    | [hpcloud/tail](https://github.com/hpcloud/tail?utm_source=chatgpt.com) |
| 消息发送    | NATS                                                                   |

---

# Agent职责

## 1. 实时监听日志

类似：

```bash id="7mmbkc"
tail -f latest.log
```

---

## 2. 解析日志事件

例如：

日志：

```text id="4hbdcf"
Player Steve joined the game
```

解析为：

```json id="lq4z6o"
{
  "event": "player_join",
  "player": "Steve",
  "timestamp": 1716123123
}
```

---

## 3. 发布消息到 NATS

```text id="s3b1fk"
publish(event)
```

---

# Agent 不负责

## 不做：

* 数据库存储
* Web API
* Dashboard
* 历史分析
* 图表
* 聚合统计

---

# Agent资源目标

| 项目  | 目标     |
| --- | ------ |
| 内存  | 5~20MB |
| CPU | 接近0    |
| IO  | 极低     |

---

# 消息队列设计

# 选择

## NATS

---

# 选择原因

| 特性   | 原因      |
| ---- | ------- |
| 轻量   | 适合小机器   |
| 单二进制 | 部署简单    |
| Go生态 | 完美适配    |
| 低延迟  | 实时日志适合  |
| 运维简单 | 不需要复杂集群 |

---

# 不选择 Kafka/RabbitMQ 原因

| 技术       | 问题               |
| -------- | ---------------- |
| Kafka    | JVM太重            |
| RabbitMQ | Erlang runtime偏重 |

---

# 消息格式设计

## 使用 JSON Event

例如：

```json id="hy2kqa"
{
  "event": "chat",
  "player": "Steve",
  "message": "hello",
  "time": 1716123123
}
```

---

# 不直接发送原始日志

原因：

* 降低网络流量
* 降低 Web VPS 解析压力
* 结构更稳定
* 更易扩展

---

# VPS2（Web服务器）设计

# 目标

```text id="we5tcr"
现代化 Dashboard 平台
```

---

# 后端技术栈

| 功能    | 技术          |
| ----- | ----------- |
| Web后端 | NestJS      |
| MQ消费  | NATS Client |
| 数据库   | PostgreSQL  |
| 时序扩展  | TimescaleDB |

---

# NestJS职责

## 1. 消费 MQ 消息

```text id="4z3rwl"
subscribe events
```

---

## 2. 数据分析

例如：

* 在线人数
* TPS
* 玩家行为
* 聊天统计
* 死亡统计
* 异常日志

---

## 3. 聚合统计

例如：

```text id="1l2rj7"
每分钟平均TPS
```

---

## 4. REST API

提供：

```text id="yq1j9s"
GET /api/stats
GET /api/players
GET /api/logs
GET /api/history
```

---

# 数据库设计

# 数据库

## PostgreSQL

---

# 时序优化

## TimescaleDB

适合：

* TPS
* 在线人数
* 日志事件
* 时间序列数据

---

# 数据库保存内容

## 实时状态

例如：

* 当前在线人数
* 当前 TPS

---

## 历史数据

例如：

* TPS趋势
* 在线峰值
* 玩家历史
* 聊天记录

---

# 前端设计

# 技术栈

| 功能   | 技术             |
| ---- | -------------- |
| 前端框架 | Next.js        |
| 图表   | Apache ECharts |

---

# 前端功能

## Dashboard

显示：

* 在线人数
* TPS
* CPU
* 内存
* 最近日志
* 聊天
* 玩家活动

---

## 图表

例如：

* TPS曲线
* 在线人数趋势
* 玩家统计

---

## 日志页面

实时显示：

```text id="99b0jv"
latest.log
```

解析后的事件。

---

# 通信方式

## 前端 → 后端

使用：

```text id="t2du7u"
HTTP REST API
```

---

# 不使用 WebSocket

原因：

* 系统复杂度降低
* 轮询已足够
* 更稳定
* 更省资源

---

# 推荐轮询频率

```text id="6j7n4j"
3~5秒
```

---

# 最终数据流

```text id="hlr6g8"
Minecraft Log
    ↓
Go Agent
    ↓
NATS
    ↓
NestJS Consumer
    ↓
PostgreSQL/TimescaleDB
    ↓
REST API
    ↓
Next.js Dashboard
```

---

# 系统优势

## 游戏服压力极低

因为：

```text id="ofrcf1"
只负责采集
```

---

# Web平台可扩展

未来可以增加：

| 功能           | 是否支持 |
| ------------ | ---- |
| 多服务器         | 支持   |
| 多世界          | 支持   |
| Discord Bot  | 支持   |
| AI日志分析       | 支持   |
| Prometheus监控 | 支持   |
| 告警系统         | 支持   |
| 玩家统计         | 支持   |
| 权限系统         | 支持   |

---

# 最终技术栈总结

| 模块      | 技术             |
| ------- | -------------- |
| 采集Agent | Go             |
| 日志监听    | hpcloud/tail   |
| 消息队列    | NATS           |
| Web后端   | NestJS         |
| 数据库     | PostgreSQL     |
| 时序扩展    | TimescaleDB    |
| 前端      | Next.js        |
| 图表      | Apache ECharts |