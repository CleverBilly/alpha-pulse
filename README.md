# alpha-pulse

AI Crypto Trading Dashboard（BTC / ETH）

## Monorepo 结构

- `backend`: Golang + Gin + GORM
- `frontend`: Next.js + TypeScript + TailwindCSS
- `docker`: Dockerfile 与 docker-compose
- `scripts`: 本地开发脚本
- `docs`: 架构与 API 文档

## 快速开始

### 方式一：Docker（推荐）

```bash
cd docker
docker compose up --build
```

- Frontend: http://localhost:3000
- Backend: http://localhost:8080

### 方式二：本地开发

```bash
./scripts/bootstrap.sh
./scripts/dev.sh
```

## 运行模式

后端当前支持三种运行模式，通过 `APP_MODE` 控制：

- `dev`
  - 默认模式
  - 默认开启自动迁移、Redis 缓存、流式采集、调度任务
  - 默认允许 Binance SDK 失败时回退到 mock 行情数据
- `test`
  - 默认关闭自动迁移、Redis 缓存、流式采集、调度任务
  - 默认保留 mock 行情回退，方便隔离测试和本地演练
- `prod`
  - 默认关闭自动迁移
  - 默认开启 Redis 缓存、流式采集、调度任务
  - 默认关闭 mock 行情回退，Binance SDK 失败时直接返回错误

常用配置：

```bash
APP_MODE=dev
GIN_MODE=debug
MARKET_SYMBOLS=BTCUSDT,ETHUSDT
AUTO_MIGRATE=true
ENABLE_REDIS_CACHE=true
ENABLE_STREAM_COLLECTOR=true
ENABLE_SCHEDULER=true
ALLOW_MOCK_BINANCE_DATA=true
SCHEDULER_INTERVAL_SECONDS=60
```

说明：

- `GIN_MODE` 默认随 `APP_MODE` 推导：`dev -> debug`，`test -> test`，`prod -> release`
- 所有 mode 默认值都可以被显式环境变量覆盖
- 本地 `docker compose` 当前默认按 `dev` 模式启动

## Binance 配置

后端已接入 `github.com/adshao/go-binance/v2`。

在运行后端前建议配置：

```bash
BINANCE_API_KEY=your_binance_api_key
BINANCE_SECRET_KEY=your_binance_secret_key
```

- `BINANCE_API_KEY`
- `BINANCE_SECRET_KEY`

当前项目的运行时初始化采用 SDK 默认 endpoint：

```go
client := binance.NewClient(apiKey, secretKey)
```

当前项目的行情接口主要访问现货公开市场数据，未配置密钥时仍会尝试读取公开接口；后续如果接入下单、账户、用户数据流等签名接口，必须提供有效密钥。

## Redis 缓存

后端当前已使用 Redis 缓存以下高频接口，用于降低前端轮询带来的重复计算与重复落库：

- `GET /api/market-snapshot`
- `GET /api/signal-timeline`
- `GET /api/indicator-series`
- `GET /api/liquidity-series`

可配置项：

```bash
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
MARKET_SNAPSHOT_CACHE_TTL=5
ANALYSIS_VIEW_CACHE_TTL=15
```

- `MARKET_SNAPSHOT_CACHE_TTL` 单位为秒
- 默认值为 `5`
- 设置为 `0` 或负数时将关闭 `market-snapshot` 缓存
- `ANALYSIS_VIEW_CACHE_TTL` 单位为秒，控制 `signal-timeline / indicator-series / liquidity-series` 的缓存时长
- 默认值为 `15`
- Redis 不可用时，后端会自动退化为无缓存模式，不阻断主服务启动
- `ENABLE_REDIS_CACHE=false` 时会显式跳过 Redis 初始化

## 后端 API

- `GET /api/price`
- `GET /api/kline`
- `GET /api/indicators`
- `GET /api/indicator-series`
- `GET /api/orderflow`
- `GET /api/structure`
- `GET /api/market-structure-events`
- `GET /api/market-structure-series`
- `GET /api/liquidity`
- `GET /api/liquidity-map`
- `GET /api/liquidity-series`
- `GET /api/market-snapshot`
- `GET /api/signal`
- `GET /api/signal-timeline`
