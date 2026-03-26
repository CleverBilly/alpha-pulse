# Auto Trading (半自动合约交易) Design

## Goal

在现有 Alpha Pulse 信号体系之上，新增半自动合约交易能力：系统产生 `setup_ready` 信号后，用户在 UI 确认执行，后端自动开仓、挂止损/止盈单；同时定时同步币安持仓，兼容用户手动开单的场景。

## Architecture

```
AlertService (setup_ready 信号)
        ↓
AlertEventCard UI "执行" 按钮
        ↓
POST /api/trades/execute {alert_id}
        ↓
TradeExecutorService
  ├─ 检查重复持仓（同标的已有 open 订单则拒绝）
  ├─ GetFuturesBalance() → 可用 USDT 余额
  ├─ 计算手数 = (余额 × TRADE_RISK_PCT / 100) / (entry_price / 杠杆)
  ├─ PlaceFuturesOrder(MARKET) → 开仓
  ├─ PlaceFuturesOrder(STOP_MARKET, stop_loss) → 止损单
  ├─ PlaceFuturesOrder(TAKE_PROFIT_MARKET, target_price) → 止盈单
  └─ 写入 trade_orders 表
        ↓
币安服务器执行止损/止盈，无需本地监控

PositionSyncService（每 15s 调度）
  ├─ /fapi/v2/positionRisk → 当前全部持仓
  ├─ 本地无记录的持仓 → 创建 source=manual
  └─ 币安已平（size=0）→ 更新 status=closed
```

## Tech Stack

- 后端：Go，`github.com/adshao/go-binance/v2/futures`（已引入）
- 数据库：MySQL，新增 `trade_orders` 表
- 前端：React + Ant Design，新增执行按钮和持仓面板

## Database

### `trade_orders` 表

| 字段 | 类型 | 说明 |
|------|------|------|
| id | uint64 PK | 自增主键 |
| alert_id | varchar(64) | 关联告警 ID（manual 来源时为空） |
| symbol | varchar(20) | 交易对，如 BTCUSDT |
| side | varchar(8) | LONG / SHORT |
| qty | decimal(18,6) | 合约数量（张数） |
| entry_order_id | varchar(64) | 币安市价开仓订单 ID |
| sl_order_id | varchar(64) | 币安止损单 ID |
| tp_order_id | varchar(64) | 币安止盈单 ID |
| filled_price | decimal(18,4) | 实际成交均价 |
| entry_price | decimal(18,4) | 信号入场价 |
| stop_loss | decimal(18,4) | 止损价 |
| target_price | decimal(18,4) | 止盈价 |
| leverage | int | 下单时账户对该标的设置的杠杆倍数（审计用） |
| risk_pct | decimal(5,2) | 下单时使用的余额百分比 |
| source | varchar(16) | system（本系统下单）/ manual（手动） |
| status | varchar(16) | pending / open / closed / failed |
| fail_reason | varchar(255) | 失败原因（status=failed 时填写） |
| created_at | bigint | Unix ms |
| closed_at | bigint | Unix ms，0 表示未平仓 |

索引：`symbol + status`（查询当前持仓），`alert_id`（关联查询）

## Backend Components

### `pkg/binance/client.go` — 新增 futures 交易方法

```go
// GetFuturesBalance 获取 USDT 可用余额
func (c *Client) GetFuturesBalance() (available float64, err error)

// GetFuturesLeverage 获取指定标的当前账户设置的杠杆倍数
func (c *Client) GetFuturesLeverage(symbol string) (leverage int, err error)

// PlaceFuturesMarketOrder 市价下单，返回订单 ID 和成交均价
func (c *Client) PlaceFuturesMarketOrder(symbol, side string, qty float64) (orderID string, filledPrice float64, err error)

// PlaceFuturesStopOrder 挂条件单（STOP_MARKET 或 TAKE_PROFIT_MARKET）
func (c *Client) PlaceFuturesStopOrder(symbol, side, orderType string, stopPrice, qty float64) (orderID string, err error)

// CancelFuturesOrder 撤销指定订单
func (c *Client) CancelFuturesOrder(symbol string, orderID string) error

// GetFuturesPositions 获取所有合约持仓（size != 0 的）
func (c *Client) GetFuturesPositions() ([]FuturesPosition, error)
```

`FuturesPosition` 结构：
```go
type FuturesPosition struct {
    Symbol        string
    Side          string  // LONG / SHORT
    Qty           float64
    EntryPrice    float64
    UnrealizedPnL float64
}
```

### `backend/repository/trade_order_repo.go`

```go
func (r *TradeOrderRepository) Create(order *models.TradeOrder) error
func (r *TradeOrderRepository) FindOpen(symbol string) ([]models.TradeOrder, error)  // status=open
func (r *TradeOrderRepository) FindAll(limit int) ([]models.TradeOrder, error)
func (r *TradeOrderRepository) UpdateStatus(id uint64, status, failReason string, closedAt int64) error
func (r *TradeOrderRepository) FindByAlertID(alertID string) (models.TradeOrder, error)
```

### `backend/internal/service/trade_executor.go`

**Execute(alertEvent AlertEvent) error**
1. 检查 `TRADE_ENABLED`，否则返回错误
2. 验证 `entry_price > 0 && stop_loss > 0 && target_price > 0`
3. 验证 `risk_reward >= 1.0`（R:R 下限），否则拒绝执行
4. 验证 `riskPct > 0 && riskPct <= 10`，超出范围拒绝执行（防止配置错误爆仓）
5. 查询同标的是否已有 `status=open` 的订单，有则返回 `"already have open position for BTCUSDT"`
6. 调用 `GetFuturesBalance()` 获取可用余额；余额为 0 或出错则拒绝
7. 调用 `GetFuturesLeverage(symbol)` 获取当前杠杆倍数
8. 计算手数：`qty = (balance × riskPct / 100) × leverage / entryPrice`，按标的精度取整
9. qty 低于标的最小下单量时返回错误（不静默跳过）
10. 调用 `PlaceFuturesMarketOrder` 开仓
11. 开仓成功后调用 `PlaceFuturesStopOrder` 挂止损单
12. 调用 `PlaceFuturesStopOrder` 挂止盈单
13. 止损/止盈任一挂单失败 → 调用 `CancelFuturesOrder` 撤市价单并平仓，写 `status=failed`
14. 全部成功 → 写 `status=open`，记录 `leverage` 字段

**Preview(alertEvent AlertEvent) (TradePreview, error)**
- 执行步骤 1-9 的只读计算（不下单），返回预计 qty、notional_usdt、available_balance、leverage

**CloseOrder(orderID uint64) error**
1. 查找订单，验证 `status=open`
2. 撤销 sl_order_id 和 tp_order_id；若订单已成交（`ORDER_ALREADY_CLOSED` 类错误）则忽略，继续执行
3. 发市价平仓单（方向相反）
4. 更新 `status=closed`

### `backend/internal/service/position_sync.go`

**SyncAll(ctx context.Context)**
1. 调用 `GetFuturesPositions()` 获取全部持仓
2. 查询本地 `status=open` 的订单列表
3. **币安有，本地无** → 先检查 `symbol+status=open` 是否已存在（幂等），不存在则创建记录，`source=manual`，`status=open`
4. **本地有（system/manual），币安 size=0** → 更新 `status=closed`，`closed_at=now`；同时调用 `CancelFuturesOrder` 撤销该订单对应的另一张孤单（SL 触发后撤 TP，TP 触发后撤 SL）；撤单时忽略 `ORDER_ALREADY_CLOSED` 类错误

集成到 `backend/internal/scheduler/jobs.go` 的 `runOnce()` 中，与现有 OutcomeTracker 并列调用。

### `backend/internal/handler/trade_handler.go`

```
POST /api/trades/execute        body: {alert_id: string}         → ExecuteTrade
GET  /api/trades/preview        query: alert_id                  → PreviewTrade
GET  /api/trades                query: limit, symbol, status     → ListOrders
POST /api/trades/:id/close                                       → CloseOrder
```

所有接口受 `TRADE_ENABLED` 开关保护，未开启时返回 `403 {"code":403,"message":"auto trading is disabled"}`。

### `backend/router/router.go`

在现有 alert 路由组之后注册 trade 路由组（同样受 auth 中间件保护）。

## Frontend Components

### `AlertEventCard.tsx` — 新增"执行"按钮

仅当 `event.kind === "setup_ready"` 时显示"执行"按钮（`PlayCircleOutlined` 图标）。点击后弹出 `TradeConfirmModal`。

### `TradeConfirmModal.tsx`（新建）

展示交易参数确认框：
- 标的、方向（多/空）
- 入场价、止损价、止盈价、预计 R:R
- 预计仓位大小（需先调 `/api/trades/preview` 获取）
- 确认 → 调用 `POST /api/trades/execute`
- 成功后显示"已提交"，失败显示错误原因

### `TradeOrderPanel.tsx`（新建）

展示当前和历史订单列表，列：
- 标的、方向、来源（系统/手动）
- 开仓均价、止损价、止盈价
- 状态（持仓中/已平/失败）
- "平仓"按钮（仅 status=open 时显示）

挂载位置：`frontend/app/review/page.tsx` 中，WinRatePanel 上方。

### `frontend/services/apiClient.ts` — 新增 tradeApi

```typescript
export const tradeApi = {
  preview(alertId: string): Promise<TradePreview>,
  execute(alertId: string): Promise<TradeOrder>,
  list(opts?: { limit?: number; symbol?: string; status?: TradeStatus }): Promise<TradeOrder[]>,
  close(orderId: number): Promise<void>,
};
```

### `frontend/types/trade.ts`（新建）

```typescript
export type TradeStatus = "pending" | "open" | "closed" | "failed";
export type TradeSource = "system" | "manual";
export type TradeSide = "LONG" | "SHORT";

export interface TradeOrder {
  id: number;
  alert_id: string;
  symbol: string;
  side: TradeSide;
  qty: number;
  filled_price: number;
  entry_price: number;
  stop_loss: number;
  target_price: number;
  leverage: number;
  risk_pct: number;
  source: TradeSource;
  status: TradeStatus;
  fail_reason?: string;
  created_at: number;
  closed_at?: number;
}

export interface TradePreview {
  symbol: string;
  side: TradeSide;
  entry_price: number;
  stop_loss: number;
  target_price: number;
  risk_reward: number;
  qty: number;
  notional_usdt: number;
  available_balance: number;
  leverage: number;
  risk_pct: number;
}
```

## Configuration

新增环境变量（`backend/.env` / `backend/.env.example`）：

```
TRADE_ENABLED=false         # 总开关，必须显式设为 true 才启用下单
TRADE_RISK_PCT=2            # 每笔使用账户余额的百分比，默认 2
```

`config.go` 新增：
```go
TradeEnabled bool    // TRADE_ENABLED
TradeRiskPct float64 // TRADE_RISK_PCT, 默认 2.0
```

## Risk Controls

1. **总开关**：`TRADE_ENABLED=false` 时所有下单接口返回 403，绝不触碰币安账户
2. **重复开仓防护**：同标的已有 `status=open` 时，`Execute` 直接返回错误，不发起任何订单
3. **价格完整性校验**：`entry_price / stop_loss / target_price` 任一为 0 时拒绝执行
4. **R:R 最小值校验**：`risk_reward < 1.0` 时拒绝执行
5. **riskPct 上限校验**：`riskPct <= 0 || riskPct > 10` 时拒绝执行，防止配置错误爆仓
6. **止损/止盈挂单失败回滚**：开仓后任一条件单失败，立即发反向市价单平仓，避免裸持仓
7. **最小 qty 限制**：计算出的 qty 低于标的最小下单量时，返回错误（不静默跳过）
8. **孤单清理**：PositionSync 检测到持仓平仓时，撤销剩余的对手方条件单；撤单时忽略 `ORDER_ALREADY_CLOSED` 类错误
9. **仓位同步幂等**：`PositionSync` 创建 manual 记录前先检查 `symbol+status=open` 是否已存在，避免重复写入

## Error Handling

- 开仓成功但止损单失败：立即撤销开仓，订单状态设为 `failed`，`fail_reason` 填写具体原因
- 币安 API 超时或网络错误：返回明确错误信息，不写任何数据库记录
- `GetFuturesBalance` 返回 0 或错误：拒绝执行，提示"余额不足或账户数据获取失败"

## Testing Strategy

- **`pkg/binance/client_trade_test.go`**：对新增的 futures 下单方法做单元测试，使用 `NewFailingHTTPClient` 验证错误分支
- **`repository/trade_order_repo_test.go`**：SQLite in-memory，验证 Create / FindOpen / UpdateStatus
- **`internal/service/trade_executor_test.go`**：mock BinanceClient 接口，覆盖：正常执行路径、重复持仓拒绝、止损挂单失败回滚
- **`internal/service/position_sync_test.go`**：mock 币安持仓，验证：新增 manual 记录、关闭已平仓单、幂等性
- **前端**：`TradeConfirmModal.test.tsx`、`TradeOrderPanel.test.tsx`，mock tradeApi，覆盖确认/取消/平仓流程

## File Changelist

**新建：**
- `backend/models/trade_order.go`
- `backend/repository/trade_order_repo.go`
- `backend/repository/trade_order_repo_test.go`
- `backend/internal/service/trade_executor.go`
- `backend/internal/service/trade_executor_test.go`
- `backend/internal/service/position_sync.go`
- `backend/internal/service/position_sync_test.go`
- `backend/internal/handler/trade_handler.go`
- `frontend/types/trade.ts`
- `frontend/components/trading/TradeConfirmModal.tsx`
- `frontend/components/trading/TradeConfirmModal.test.tsx`
- `frontend/components/trading/TradeOrderPanel.tsx`
- `frontend/components/trading/TradeOrderPanel.test.tsx`

**修改：**
- `backend/pkg/binance/client.go`（新增 5 个 futures 方法）
- `backend/pkg/binance/client_test.go`（新增 trade 方法测试）
- `backend/models/migrate.go`（注册 TradeOrder）
- `backend/config/config.go`（新增 TradeEnabled / TradeRiskPct）
- `backend/config/config_test.go`（新增配置测试）
- `backend/internal/scheduler/jobs.go`（注册 PositionSync）
- `backend/cmd/server/main.go`（注入 TradeExecutor、PositionSync、TradeHandler）
- `backend/router/router.go`（注册 trade 路由）
- `frontend/services/apiClient.ts`（新增 tradeApi）
- `frontend/components/alerts/AlertEventCard.tsx`（新增执行按钮）
- `frontend/app/review/page.tsx`（挂载 TradeOrderPanel）
