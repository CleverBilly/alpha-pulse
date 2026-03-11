# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此代码库中工作时提供指引。

## 项目概述

Alpha Pulse 是一个 AI 驱动的加密货币现货分析工作台。当前 MVP 专注于 `BTCUSDT` 和 `ETHUSDT` 的现货分析，支持时间周期：`1m/5m/15m/1h/4h`。

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
  ├─ 指标引擎  backend/internal/indicator/   → EMA20/50、VWAP、布林带
  ├─ 订单流引擎 backend/internal/orderflow/  → 真实成交驱动分析
  ├─ 结构引擎  backend/internal/structure/   → HH/HL/LH/LL/BOS/CHOCH
  └─ 流动性引擎 backend/internal/liquidity/  → 深度与换手分析
  ↓
信号引擎 backend/internal/signal/ → 综合评分
  ↓
AI 解释引擎 backend/internal/ai/ → 规则驱动文本
  ↓
服务层 backend/internal/service/ → buildMarketSnapshot()
  ↓
Redis 缓存（可选，TTL：快照 5s / 序列 15s）
  ↓
HTTP/WebSocket API（Gin，端口 8080）
  ↓
前端 Zustand Store frontend/store/marketStore.ts
  ↓
React 组件（所有面板共享同一快照，无数据撕裂）
```

### 关键架构决策

- **单快照模型**：前端所有面板（K线图、信号、订单流、流动性、AI 分析）统一消费 `/api/market-snapshot` 的一份 `MarketSnapshot`，避免多接口并发导致的数据撕裂。
- **三种运行模式**，通过 `APP_MODE` 环境变量切换：

  | 模式 | 自动迁移 | Redis | 流式采集 | 调度任务 | Mock 数据 |
  |------|--------|-------|--------|--------|---------|
  | dev  | ✅ | ✅ | ✅ | ✅ | ✅ |
  | test | ❌ | ❌ | ❌ | ❌ | ✅ |
  | prod | ❌ | ✅ | ✅ | ✅ | ❌ |

- **自定义 SVG 图层**：图表页面使用自定义 SVG 渲染（而非第三方图表库），精确叠加 K 线、指标线、结构点、流动性区域和信号点。

## 后端目录结构

```
backend/
├── cmd/server/main.go          # 入口：DB → Redis → Engines → Services → Router → Scheduler → Collector → HTTP
├── config/config.go            # 所有环境变量及各模式默认值
├── internal/
│   ├── handler/                # HTTP 层（MarketHandler、SignalHandler）
│   ├── service/                # 编排层（MarketService、SignalService）
│   ├── collector/              # Binance REST + WebSocket 数据采集
│   ├── indicator/              # 技术指标计算
│   ├── orderflow/              # 订单流分析
│   ├── structure/              # 市场结构识别
│   ├── liquidity/              # 流动性分析
│   ├── signal/                 # 信号评分引擎
│   ├── ai/                     # 规则驱动解释引擎
│   └── scheduler/              # 定时调度任务
├── models/                     # GORM 模型（11 张表，dev 模式自动迁移）
├── repository/                 # 数据访问层（8 个 repo）
├── middleware/                 # CORS、Logger、Recovery
├── router/router.go            # 全部 15+ 路由定义
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
│   ├── signals/                # 信号 + AI 解释
│   └── market/                 # 市场总览
├── components/                 # 按领域划分的 React 组件
├── services/apiClient.ts       # HTTP + WebSocket 客户端
├── store/marketStore.ts        # Zustand 全局状态（applySnapshot、setStreamState）
└── types/                      # TypeScript 类型定义
```

## 关键配置

后端环境变量（配置于 `backend/.env`）：

| 变量 | 默认值（dev） | 说明 |
|------|------------|------|
| `APP_MODE` | `dev` | dev / test / prod |
| `APP_PORT` | `8080` | HTTP 端口 |
| `MYSQL_DSN` | `root:root@tcp(localhost:3306)/alpha_pulse?...` | MySQL 连接字符串 |
| `REDIS_ADDR` | `localhost:6379` | Redis 地址 |
| `MARKET_SYMBOLS` | `BTCUSDT,ETHUSDT` | 逗号分隔的交易对列表 |
| `BINANCE_API_KEY` | `` | 可选；公开行情无需密钥 |

前端环境变量（配置于 `frontend/.env.local`）：
```
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api
```

## API 约定

- 基础 URL：`http://localhost:8080/api`
- 统一响应格式：`{"code": 0, "message": "ok", "data": {...}}`
- 通用查询参数：`symbol`、`interval`、`limit`、`refresh=1`（绕过缓存）
- WebSocket 端点：`GET /api/market-snapshot/stream`

## 数据库

11 张表由 GORM AutoMigrate 管理（dev 模式或 `AUTO_MIGRATE=true` 时启动自动执行）：
- 原始数据：`kline`、`agg_trades`、`order_book_snapshots`
- 分析结果：`indicators`、`orderflow`、`large_trade_events`、`microstructure_events`、`structure`、`liquidity`、`signals`、`feature_snapshots`

## 测试说明

- 后端测试使用 `APP_MODE=test`，自动关闭 Redis、流式采集和调度器，并启用 Binance Mock 数据
- 前端单元测试使用 Vitest，E2E 测试使用 Playwright
- 新建或大幅修改业务文件时，在测试目录下的相同相对路径创建对应的测试文件
