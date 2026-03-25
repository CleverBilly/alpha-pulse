# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

Alpha Pulse V2.0 是 AI 驱动的加密货币合约方向判断与告警终端。支持标的：`BTCUSDT / ETHUSDT / SOLUSDT`，时间周期：`1m / 5m / 15m / 1h / 4h`。

核心能力：现货 + 合约多周期快照分析、Futures Direction Engine（A级信号/方向切换/No-Trade 过滤）、Alert Center（浏览器通知 + 飞书机器人）、单用户登录保护。

## 开发命令

### 初始化
```bash
./scripts/bootstrap.sh        # go mod tidy + npm install
```

### 启动开发环境
```bash
./scripts/dev.sh                        # 需要本地 MySQL + Redis
USE_DOCKER_DEPS=1 ./scripts/dev.sh      # 用 Docker 跑 MySQL/Redis，本地跑应用
cd docker && docker compose up --build  # 全 Docker 部署
```

### 后端（Go）
```bash
cd backend
go run ./cmd/server              # 启动服务（自动加载 .env）
go test ./...                    # 全部测试
go test ./internal/signal/...    # 单包测试
go build -o server ./cmd/server  # 编译二进制
```

### 前端（Next.js）
```bash
cd frontend
npm run dev          # 开发服务器 localhost:3000
npm run build        # 生产构建
npm run lint         # ESLint
npm test             # Vitest 单元测试
npm run test:watch   # 监听模式
npm run test:e2e     # Playwright E2E 测试
```

## 架构

```
Binance API（REST + WebSocket）
  ↓
采集层 backend/internal/collector/
  ↓
仓储层 backend/repository/ → MySQL
  ↓
分析引擎（并行）
  ├─ 指标引擎  backend/internal/indicator/       → EMA20/50、VWAP、布林带
  ├─ 订单流引擎 backend/internal/orderflow/      → 真实成交驱动分析
  ├─ 结构引擎  backend/internal/structure/       → HH/HL/LH/LL/BOS/CHOCH
  └─ 流动性引擎 backend/internal/liquidity/      → 深度与换手分析
  ↓
信号引擎 backend/internal/signal/               → 综合评分
  ↓
Futures Direction Copilot                        → 4h/1h/15m/5m 多周期方向判断
  backend/internal/service/direction_copilot.go  → A级 setup / 方向切换 / No-Trade
  backend/internal/service/futures_snapshot.go   → mark/funding/OI/LSR/清算压力
  ↓
AI 解释引擎 backend/internal/ai/                → 规则驱动文本
  ↓
Alert 服务 backend/internal/service/alert_service.go
  ├─ 飞书机器人推送 feishu_notifier.go
  └─ 站内 Alert Center（history + preferences）
  ↓
服务层 backend/internal/service/market_service.go → buildMarketSnapshot()
  ↓
Redis 缓存（可选，TTL：快照 5s / 序列 15s）
  ↓
HTTP/WebSocket API（Gin，端口 8080）
  ├─ Auth 拦截 backend/internal/auth/            → bcrypt + session cookie
  └─ Observability backend/internal/observability/
  ↓
前端 Zustand Store frontend/store/marketStore.ts
  ↓
React 组件（所有面板共享同一快照，无数据撕裂）
```

### 关键架构决策

- **单快照模型**：前端所有面板统一消费 `/api/market-snapshot` 的一份 `MarketSnapshot`，避免多接口并发导致的数据撕裂。
- **三种运行模式**，通过 `APP_MODE` 环境变量切换：

  | 模式 | 自动迁移 | Redis | 流式采集 | 调度任务 | Mock 数据 |
  |------|--------|-------|--------|--------|---------|
  | dev  | ✅ | ✅ | ✅ | ✅ | ✅ |
  | test | ❌ | ❌ | ❌ | ❌ | ✅ |
  | prod | ❌ | ✅ | ✅ | ✅ | ❌ |

- **自定义 SVG 图层**：图表页面使用自定义 SVG 渲染（而非第三方图表库），精确叠加 K 线、指标线、结构点、流动性区域和信号点。
- **单用户登录**：`ENABLE_SINGLE_USER_AUTH=true` 时后端以 bcrypt 校验密码并签发 session cookie，前端 middleware.ts 做路由守卫；前后端必须共享同一个 `AUTH_SESSION_SECRET`。

## 后端目录结构

```
backend/
├── cmd/server/main.go          # 入口：DB → Redis → Engines → Services → Router → Scheduler → Collector → HTTP
├── config/config.go            # 所有环境变量及各模式默认值
├── internal/
│   ├── handler/                # HTTP 层（MarketHandler、SignalHandler、AuthHandler、AlertHandler）
│   ├── service/                # 编排层（MarketService、SignalService、AlertService、DirectionCopilot）
│   ├── collector/              # Binance REST + WebSocket 数据采集
│   ├── indicator/              # 技术指标计算
│   ├── orderflow/              # 订单流分析
│   ├── structure/              # 市场结构识别
│   ├── liquidity/              # 流动性分析
│   ├── signal/                 # 信号评分引擎
│   ├── ai/                     # 规则驱动解释引擎
│   ├── auth/                   # 单用户认证（bcrypt + session）
│   ├── observability/          # 日志 / 指标可观测
│   └── scheduler/              # 定时调度任务
├── models/                     # GORM 模型（11 张表）
├── repository/                 # 数据访问层
├── middleware/                 # CORS、Logger、Recovery
├── router/router.go            # 全部路由定义（含 auth 和 alert 路由）
└── pkg/
    ├── binance/client.go       # Binance SDK 封装
    ├── database/               # MySQL + Redis 初始化
    └── utils/                  # 通用工具
```

## 前端目录结构

```
frontend/
├── app/                        # Next.js App Router 页面
│   ├── dashboard/              # 5 面板总览
│   ├── chart/                  # 多层 SVG 图表
│   ├── review/                 # 告警复盘（兼容 /signals）
│   ├── alerts/                 # Alert Center
│   ├── analysis/               # 深度分析视图
│   └── market/                 # 市场总览
├── pages/                      # Next.js Pages Router（登录等基础页面）
├── components/                 # 按领域划分的 React 组件
├── services/apiClient.ts       # HTTP + WebSocket 客户端
├── store/
│   ├── marketStore.ts          # Zustand 行情全局状态（applySnapshot、setStreamState）
│   └── signalStore.ts          # 信号状态
├── types/                      # TypeScript 类型定义（alert.ts、market.ts、signal.ts、snapshot.ts）
├── middleware.ts               # Next.js 路由守卫（登录拦截）
├── test/                       # 测试 fixtures
└── tests/e2e/                  # Playwright E2E 规范
```

## 关键配置

后端环境变量（`backend/.env`）：

| 变量 | 默认值（dev） | 说明 |
|------|------------|------|
| `APP_MODE` | `dev` | dev / test / prod |
| `APP_PORT` | `8080` | HTTP 端口 |
| `MYSQL_DSN` | `root:root@tcp(localhost:3306)/alpha_pulse?...` | MySQL 连接字符串 |
| `REDIS_ADDR` | `localhost:6379` | Redis 地址 |
| `MARKET_SYMBOLS` | `BTCUSDT,ETHUSDT,SOLUSDT` | 逗号分隔的交易对列表 |
| `MARKET_SNAPSHOT_CACHE_TTL` | `5` | 快照缓存秒数（0 关闭）|
| `ANALYSIS_VIEW_CACHE_TTL` | `15` | 分析序列缓存秒数 |
| `SCHEDULER_INTERVAL_SECONDS` | `60` | 调度间隔 |
| `ENABLE_SINGLE_USER_AUTH` | `false` | 开启单用户登录保护 |
| `AUTH_USERNAME` | `` | 登录用户名 |
| `AUTH_PASSWORD_HASH` | `` | bcrypt 哈希（单引号包裹）|
| `AUTH_SESSION_SECRET` | `` | session 签发密钥，前后端必须一致 |
| `AUTH_COOKIE_SECURE` | `false(dev)/true(prod)` | HTTPS Cookie |
| `CORS_ALLOW_ORIGINS` | `http://localhost:3000,...` | 允许的前端域名 |
| `FEISHU_BOT_WEBHOOK_URL` | `` | 飞书机器人 webhook（留空不推送）|
| `FEISHU_BOT_SECRET` | `` | 飞书机器人签名校验 |
| `ALERT_PUBLIC_BASE_URL` | `` | 飞书消息深链的前端域名 |
| `BINANCE_API_KEY` | `` | 可选；公开行情无需密钥 |

前端环境变量（`frontend/.env.local`）：
```
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api
NEXT_PUBLIC_AUTH_ENABLED=false
AUTH_COOKIE_NAME=alpha_pulse_session
AUTH_SESSION_SECRET=<same-as-backend>
```

## API 约定

- 基础 URL：`http://localhost:8080/api`
- 统一响应格式：`{"code": 0, "message": "ok", "data": {...}}`
- 通用查询参数：`symbol`、`interval`、`limit`、`refresh=1`（绕过缓存）
- WebSocket 端点：`GET /api/market-snapshot/stream`
- Auth 端点（不受 session 守卫保护）：`POST /api/auth/login`、`POST /api/auth/logout`、`GET /api/auth/session`
- Alert 端点：`GET/POST /api/alerts`、`/api/alerts/history`、`/api/alerts/preferences`

## 数据库

11 张表由 GORM AutoMigrate 管理（dev 模式或 `AUTO_MIGRATE=true` 时启动自动执行）：
- 原始数据：`kline`、`agg_trades`、`order_book_snapshots`
- 分析结果：`indicators`、`orderflow`、`large_trade_events`、`microstructure_events`、`structure`、`liquidity`、`signals`、`feature_snapshots`

## 测试说明

- 后端测试使用 `APP_MODE=test`，自动关闭 Redis、流式采集和调度器，并启用 Binance Mock 数据
- 前端单元测试使用 Vitest；E2E 测试使用 Playwright（`frontend/tests/e2e/*.spec.ts`）
- 新建或大幅修改业务文件时，在相同相对路径创建对应的 `*_test.go` / Vitest 测试文件

## 提交规范

使用短命令式主题，前缀：`feat:` / `fix:` / `refactor:`。PRs 需包含：变更摘要、关联任务、测试证据；UI 变更附截图；schema / API 约定 / 环境变量变更需显式说明。
