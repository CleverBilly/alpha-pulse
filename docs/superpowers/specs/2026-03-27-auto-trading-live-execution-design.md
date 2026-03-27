# Auto Trading Live Execution Design

## Goal

在现有 Alpha Pulse 告警体系上落地真实 Binance Futures 自动下单能力：

- `setup_ready` 命中后可自动触发真实交易
- 开仓改为限价单，并支持超时撤单
- 成交后自动补挂止损 / 止盈保护单
- 后台提供可视化配置页，动态控制可交易标的、风险比例和运行开关
- 同步 Binance 当前持仓，兼容用户手动开单和交易所侧平仓

## Problem

仓库里目前只有自动交易占位页，还没有完整的交易执行链路：

- 没有 `trade_settings` / `trade_orders` 数据模型
- 没有真实 Binance Futures 下单、撤单、查订单、查持仓接口
- 没有自动执行协调器和交易运行时
- 没有后台配置页来动态启停自动执行与标的白名单
- 原始设计稿默认是“人工确认 + 市价开仓”，与当前需求不一致

本轮需求已经收敛为：

- 对接真实 Binance Futures 账户
- 默认关闭，必须显式打开
- 告警命中后全自动执行，不需要人工确认
- 开仓使用限价单，不成交则按超时策略撤单
- 可交易标的走后台白名单配置，而不是写死 BTCUSDT

## Chosen Direction

采用“告警直接触发 + 独立交易运行时”的方案：

- **告警链路**继续负责发现 `setup_ready`
- **AutoTradeCoordinator** 在告警产出后立即做风控与配置校验
- **TradeExecutorService** 负责发起限价开仓、创建订单记录
- **TradeRuntime** 以更高频率轮询挂单状态和持仓状态
- **后台配置页** 驱动运行时配置，不要求改 `.env` 才能调整交易策略

这条路线比队列架构更快落地，也能保持足够清晰的边界。

## Design

### 1. 双层开关和配置模型

自动交易配置拆成两层：

- **静态底线开关**
  - `TRADE_ENABLED`
  - `TRADE_AUTO_EXECUTE`
  - `TRADE_ALLOWED_SYMBOLS`
  - 这些值来自环境变量，代表部署层是否允许触碰真实账户
- **运行时配置**
  - 存在数据库单例表 `trade_settings`
  - 由后台页面可视化修改
  - 保存 `auto_execute_enabled`、`allowed_symbols`、`risk_pct`、`min_risk_reward`、`entry_timeout_seconds`、`max_open_positions`、`sync_enabled`

真实交易要同时满足这两层：

- 环境变量允许
- 数据库配置允许

这样可以确保默认安全，同时保留运行中可调配置能力。

### 2. 限价开仓状态机

订单生命周期改成下面这套状态：

- `pending_fill`
- `open`
- `closed`
- `expired`
- `failed`

处理规则：

1. `setup_ready` 命中后，协调器先校验开关、标的白名单、盈亏比、仓位上限、重复仓位
2. 校验通过后创建本地订单，状态设为 `pending_fill`
3. 按信号 `entry_price` 发 Binance `LIMIT` 开仓单
4. 挂单未成交时保持 `pending_fill`
5. 超过 `entry_timeout_seconds` 后撤销开仓单，状态改为 `expired`
6. 一旦确认限价单已成交，才补挂 `STOP_MARKET` 和 `TAKE_PROFIT_MARKET`
7. 止损 / 止盈任一挂单失败，立即发反向市价单平仓并标记 `failed`
8. 后续正常止损、止盈或人工平仓都统一收口为 `closed`

这套状态机比“直接市价开仓”多了挂单阶段，但更符合真实账户下的入场控制要求。

### 3. 独立 TradeRuntime

交易运行时不塞进现有分析调度 `Jobs.runOnce()`，而是单独启动 goroutine。

拆成两条循环：

- **Entry watcher**
  - 默认每 `3s` 检查一次 `pending_fill` 订单
  - 负责确认成交、处理超时撤单、触发补挂保护单
- **Position sync**
  - 默认每 `15s` 拉取 Binance 持仓
  - 负责识别手动单、识别交易所侧已平仓、清理孤儿保护单

边界如下：

- 告警调度负责发现机会并触发一次开仓尝试
- TradeRuntime 负责盯单、补保护、同步持仓和收口状态

这样交易执行不会被行情预热和信号分析拖慢。

### 4. 后端组件

#### `backend/models/trade_setting.go`

新增自动交易运行时配置模型，沿用单例配置表模式。

#### `backend/models/trade_order.go`

新增交易订单模型，补齐限价挂单所需字段：

- `entry_order_type`
- `limit_price`
- `entry_status`
- `entry_expires_at`
- `close_order_id`
- `close_reason`
- `source`
- `status`

#### `backend/repository/trade_setting_repo.go`

负责单例交易配置的读取和保存。

#### `backend/repository/trade_order_repo.go`

负责订单读写、查重持仓、查询挂单、查询历史、更新状态。

#### `backend/internal/service/trade_settings.go`

封装默认配置、配置校验、环境变量静态底线投影，以及数据库记录与 API DTO 的转换。

#### `backend/internal/service/trade_executor.go`

负责：

- 风控校验
- 仓位计算
- 创建本地订单记录
- 发 Binance 限价开仓单
- 成交后补挂保护单
- 人工平仓

#### `backend/internal/service/auto_trade_coordinator.go`

负责在 `setup_ready` 事件产出后判断：

- 是否允许自动执行
- 当前标的是否在白名单内
- 是否已经存在同标的未结束订单

通过后再调用执行器。

#### `backend/internal/service/trade_runtime.go`

负责 entry watcher 和 position sync 两个轮询循环。

#### `backend/internal/handler/trade_handler.go`

暴露：

- `GET /api/trade-settings`
- `PUT /api/trade-settings`
- `GET /api/trades`
- `POST /api/trades/:id/close`
- `GET /api/trades/runtime`

### 5. Binance 客户端能力扩展

`backend/pkg/binance/client.go` 需要补齐 Futures 交易接口：

- 获取 futures USDT 可用余额
- 获取当前杠杆
- 获取交易对精度 / 最小下单量
- 发限价开仓单
- 查订单状态
- 撤单
- 发保护单
- 发反向平仓单
- 拉取当前持仓

同时要区分：

- 公开行情接口允许 mock fallback
- 真实交易接口绝不允许 fallback

任何交易接口失败都必须把错误直接暴露给上层，不能伪造成功。

### 6. 前端自动交易配置台

将 [auto-trading/page.tsx](/Users/billy/go/src/alpha-pulse/.worktrees/auto-trading-live-execution/frontend/app/auto-trading/page.tsx) 从占位页升级为真正的执行中枢：

- **顶部状态带**
  - 环境变量底线开关
  - 自动执行状态
  - 允许标的
  - 当前 `pending_fill` / `open` 数
  - 最近同步时间
- **中间配置台**
  - 自动执行开关
  - 白名单标的多选
  - 风险比例
  - 最低盈亏比
  - 限价单超时秒数
  - 最大持仓数
  - 持仓同步开关
- **底部订单面板**
  - 当前挂单
  - 当前持仓
  - 最近失败
  - 最近平仓

不做 modal，直接做常驻页面，因为这是“执行中枢”，不是附属设置。

### 7. Review 页订单区

在 [frontend/components/review/ReviewWorkspace.tsx](/Users/billy/go/src/alpha-pulse/.worktrees/auto-trading-live-execution/frontend/components/review/ReviewWorkspace.tsx) 中补入交易订单面板，让复盘页也能看见系统单和手动单的状态。

自动执行不再依赖人工确认弹窗，所以不再把“执行按钮”作为主入口；如需保留调试入口，只做受保护的开发态能力，不作为正式工作流。

## Risk Controls

必须内建以下保护：

1. `TRADE_ENABLED=false` 时，所有交易相关接口和后台自动执行都拒绝触发真实账户操作
2. `TRADE_AUTO_EXECUTE=false` 时，只允许查看配置和订单，不执行自动下单
3. 非白名单标的绝不自动下单
4. 同标的已有未结束订单时绝不重复开仓
5. `risk_pct`、`min_risk_reward`、`entry_timeout_seconds` 均做强校验
6. 限价单未成交时不允许提前挂止损止盈
7. 保护单补挂失败时，立即执行兜底强平
8. 真实交易接口绝不走 mock fallback

## Boundaries

- 本轮只支持 Binance Futures
- 自动交易触发条件只接现有 `setup_ready`
- 第一版不做复杂追价逻辑，只做“固定限价 + 超时撤单”
- 第一版不做多账户和多策略隔离
- 保留现有告警与复盘链路，不重写信号生成逻辑

## Testing

### Backend

- 配置模型与配置校验单测
- 交易配置 handler / router 测试
- 交易订单 repository 测试
- 执行器测试：
  - 成功创建限价挂单
  - 超时撤单
  - 成交后补挂保护单
  - 保护单失败触发强平
- 协调器测试：
  - 开关关闭
  - 标的不在白名单
  - 存在重复仓位
- 运行时测试：
  - watcher 推进 `pending_fill -> open`
  - watcher 推进 `pending_fill -> expired`
  - position sync 识别 manual 仓位和已平仓

### Frontend

- 自动交易页结构测试
- 配置面板加载 / 编辑 / 保存测试
- 订单面板状态渲染测试
- Review 页新增交易面板测试
- `apiClient` 新增交易接口的消费测试按组件层覆盖

### Verification

至少运行：

- `cd backend && go test ./...`
- `cd frontend && npm test`
- `cd frontend && npm run lint`
- `cd frontend && npm run build`
