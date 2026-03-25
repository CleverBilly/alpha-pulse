# Trading Decision Loop — 设计规格

**日期：** 2026-03-25
**状态：** 待实现
**作者：** Claude Code
**目标用户：** 单用户个人实盘辅助

---

## 1. 背景与目标

Alpha Pulse V2.0 已完成 Futures Direction Copilot 主线开发，能够生成方向告警。但当前存在四个使用侧痛点：

| 痛点 | 现状 |
|------|------|
| 信不信？ | 信号发出后无历史结果追踪，无胜率数据 |
| 怎么执行？ | `entry_price / stop_loss / target_price` 已存在，但无仓位计算工具，图表无三线叠加 |
| 反应慢？ | 调度器默认 60s，无声音告警 |
| 复盘难？ | Review 页仅为列表，无 K 线上下文还原 |

本次设计目标：构建"交易决策闭环"，分快慢两条线并行推进。

---

## 2. 范围

### 快线（Week 1，零/极小后端改动）
- **F1** 声音告警（Web Audio API）
- **F2** 调度间隔降至 15s（.env 配置调整 + 文档）
- **F3** 图表三线叠加（entry / stop / target SVG 层）
- **F4** 浮动仓位计算器（纯前端组件）

### 慢线（Week 2-3，需后端新增）
- **S1** 信号结果追踪（OutcomeTracker，`alert_records` 加结果字段）
- **S2** Review 页升级（K 线上下文跳转 + 三线历史还原）
- **S3** 胜率统计面板（滚动胜率 + 平均 R:R）

### 不在范围
- 自动下单
- 多交易所
- 多用户系统

---

## 3. 架构

```
┌─────────────────── 快线 ───────────────────────┐
│  F1: AlertCenter.tsx                           │
│      ↳ 监听新 alert → Web Audio API 播放提示音   │
│      ↳ AlertConfigPanel 新增声音开关             │
│      ↳ 偏好持久化至后端 alert_preferences 表     │
│                                                │
│  F2: backend/.env SCHEDULER_INTERVAL_SECONDS=15│
│      （无代码改动，Jobs.Start 已支持可配间隔）    │
│                                                │
│  F3: KlineChart.tsx 拆分（必要前置）             │
│      ↳ KlineCandleLayer.tsx    蜡烛+成交量       │
│      ↳ StructureLiquidityLayer.tsx 结构+流动性  │
│      ↳ SignalOverlayLayer.tsx  信号+三线  ←F3   │
│      数据：alertApi.getAlertHistory(limit=20)   │
│            前端过滤取最新 setup_ready alert      │
│                                                │
│  F4: PositionCalculator.tsx（新组件）            │
│      ↳ 挂载于 Dashboard 右侧 + Chart 浮动按钮    │
│      ↳ 输入：账户余额、风险比例%                  │
│      ↳ 自动填入：entry_price、stop_loss（最新告警）│
│      ↳ 输出：仓位大小、止损距离%、预计最大亏损$    │
└────────────────────────────────────────────────┘

┌─────────────────── 慢线 ───────────────────────┐
│  S1: OutcomeTracker                            │
│      ↳ alert_records 新增字段：                 │
│          outcome ENUM(pending/target_hit/      │
│                       stop_hit/expired)        │
│          outcome_price DECIMAL(18,8)           │
│          outcome_at BIGINT (Unix ms)           │
│          actual_rr DECIMAL(10,4)               │
│          interval VARCHAR(8) （记录触发周期）    │
│      ↳ backend/internal/service/               │
│          outcome_tracker.go (新文件)            │
│      ↳ scheduler/jobs.go 新增 runOutcomeTrack()│
│      ↳ 判断逻辑（逐根 K 线，止损优先）：         │
│          多头信号：low≤stop→stop_hit             │
│                   high≥target→target_hit       │
│          空头信号：high≥stop→stop_hit            │
│                   low≤target→target_hit        │
│          过期：event_time + 60分钟 < now         │
│                                                │
│  S2: Review 页升级                              │
│      ↳ AlertEventCard 新增"复盘"按钮            │
│      ↳ 点击 → ReviewChartModal.tsx             │
│          ∘ 两次调用：FindBefore(前20根)          │
│                     FindAfter(后40根) 合并      │
│          ∘ SignalOverlayLayer 叠加三线           │
│          ∘ 若已有 outcome，标注结果点             │
│                                                │
│  S3: 胜率统计面板 WinRatePanel.tsx              │
│      ↳ 挂载于 /review 页顶部                    │
│      ↳ 后端：GET /api/alerts/stats             │
│          胜率 = target_hit / (target_hit+stop_hit)│
│          pending 和 expired 不计入分母           │
│          支持 ?symbol=X&limit=N 过滤             │
└────────────────────────────────────────────────┘
```

---

## 4. 详细设计

### 4.1 F1 — 声音告警

**触发条件：** `AlertCenter.tsx` 检测到 `notifiedIdsRef` 中没有记录的新 AlertEvent 时播放音频。

**实现：**
- 使用 Web Audio API（`AudioContext.createOscillator` 合成音或预加载 `/public/alert.mp3`）
- `AudioContext` 在用户首次点击 Alert 铃铛按钮后初始化，避免浏览器自动播放限制
- 在 `AlertCenter.tsx` 的告警检测逻辑中，新告警到达且 `sound_enabled=true` 时调用 `playSoundAlert()`
- 无法初始化 AudioContext 时静默降级（不报错，不影响其他功能）

**后端数据模型：**
- `backend/models/alert_preference.go`：新增 `SoundEnabled bool gorm:"column:sound_enabled;not null;default:false"`
- `backend/internal/service/alert_preferences.go`：读写 `SoundEnabled` 字段
- `backend/internal/handler/alert_handler.go`：`alertPreferencesRequest` 结构体新增 `SoundEnabled bool json:"sound_enabled"`，`UpdateAlertPreferences` 映射时传递该字段

**文件改动：**
- `backend/models/alert_preference.go` — 新增 `SoundEnabled`
- `backend/internal/service/alert_preferences.go` — 读写 `SoundEnabled`
- `backend/internal/handler/alert_handler.go` — request 结构体 + 映射逻辑新增 `sound_enabled`
- `frontend/types/alert.ts` — `AlertPreferences` 新增 `sound_enabled: boolean`
- `frontend/components/alerts/AlertCenter.tsx` — AudioContext 初始化 + 触发声音告警
- `frontend/components/alerts/AlertConfigPanel.tsx` — 声音开关 UI

---

### 4.2 F2 — 调度间隔

`backend/internal/scheduler/jobs.go` 已支持可配 `interval`，由 `config.SchedulerIntervalSeconds` 驱动，无需代码改动。

**改动：**
- `backend/.env.example`：`SCHEDULER_INTERVAL_SECONDS=15`
- `CLAUDE.md` 更新默认值说明

> **速率安全分析：** 公开行情接口限制 1200 权重/分钟。当前调度每轮约 3 标的 × 4 周期 = 12 次请求，15s 间隔下约 48 次/分，安全。若触发限制，降回 30s。

---

### 4.3 F3 — 图表三线叠加（前置：KlineChart.tsx 拆分）

#### KlineChart.tsx 拆分

当前 `KlineChart.tsx` 2145 行，须拆分为：

| 文件 | 职责 | 目标行数 |
|------|------|---------|
| `KlineChart.tsx` | 主容器：坐标系计算、状态、事件、图层组合 | ≤420 |
| `KlineCandleLayer.tsx` | 蜡烛实体 + 影线 + 成交量柱 SVG | ≤280 |
| `StructureLiquidityLayer.tsx` | 结构点（BOS/CHOCH/HH/HL）+ 流动性墙区域 SVG | ≤320 |
| `SignalOverlayLayer.tsx` | 信号标记 + 微结构事件点 + entry/stop/target 三线 | ≤300 |

所有图层接收统一的 `ChartCoords`（坐标映射函数）和各自所需数据 props，无内部状态，纯渲染。拆分完成后立即运行 `npm test` + Playwright 主路径确认回归，再追加三线代码。

#### 三线数据来源

F3 使用现有 `alertApi.getAlertHistory(limit=20)`（无需改后端），在前端过滤出与当前 `symbol` 匹配且 `kind === "setup_ready"` 的最新一条 alert 作为 `activeSignal`。

```typescript
// SignalOverlayLayer 接收可选的 activeSignal prop
interface ActiveSignal {
  entryPrice: number;
  stopLoss: number;
  targetPrice: number;
  direction: "long" | "short"; // 由 direction_state 推导
}
```

**视觉规范：**
- Entry 线：绿色 `#52c41a` 虚线（`stroke-dasharray: 4 3`）
- Stop 线：红色 `#ff4d4f` 虚线
- Target 线：金色 `#faad14` 虚线
- 三线右侧附文字标签（`entry` / `SL` / `TP`）+ 价格值

**文件改动：**
- `frontend/components/chart/KlineChart.tsx` — **拆分**（主容器 ≤420 行）
- `frontend/components/chart/KlineCandleLayer.tsx` — **新建**
- `frontend/components/chart/StructureLiquidityLayer.tsx` — **新建**
- `frontend/components/chart/SignalOverlayLayer.tsx` — **新建**（含三线）

---

### 4.4 F4 — 仓位计算器

**组件：** `frontend/components/trading/PositionCalculator.tsx`（新建，纯前端，无后端依赖）

**挂载位置：**
- `/dashboard` 页：右侧面板区域，默认展开
- `/chart` 页：右下角浮动按钮，点击弹出 Drawer

**输入字段：**

| 字段 | 默认值 | 说明 |
|------|--------|------|
| 账户余额 (USDT) | localStorage 上次值 | 不上传后端 |
| 风险比例 % | 1%（localStorage） | 建议范围 0.5%-3% |
| 进场价 | 最新 setup_ready alert 的 `entry_price` | 可手动覆盖 |
| 止损价 | 最新 setup_ready alert 的 `stop_loss` | 可手动覆盖 |
| 目标价 | 最新 setup_ready alert 的 `target_price` | 仅展示，不可编辑 |

**输出（实时计算）：**

```
止损距离% = |entry - stop| / entry × 100
仓位大小(USDT) = balance × riskPct% / 止损距离%
预计最大亏损$ = balance × riskPct%
预计最大盈利$ = 仓位大小 × |target - entry| / entry
实际 R:R = 预计最大盈利 / 预计最大亏损
```

**安全约束：**
- 仓位大小 > 账户余额时显示警告（橙色）
- 止损距离为 0 时禁止计算并提示"止损价不能等于进场价"
- 账户余额、风险比例 存 `localStorage`，不通过网络传输

**文件改动：**
- `frontend/components/trading/PositionCalculator.tsx` — **新建**
- `frontend/app/dashboard/page.tsx` — 挂载 PositionCalculator
- `frontend/app/chart/page.tsx` — 挂载浮动 PositionCalculator

---

### 4.5 S1 — 信号结果追踪

#### 数据库变更

`alert_records` 表新增字段（AutoMigrate 处理）：

```go
Interval     string  `gorm:"size:8;not null;default:'15m';comment:触发告警时的参考周期"`
Outcome      string  `gorm:"size:24;index;not null;default:'pending';comment:pending/target_hit/stop_hit/expired"`
OutcomePrice float64 `gorm:"type:decimal(18,8);not null;default:0"`
OutcomeAt    int64   `gorm:"not null;default:0"`  // Unix ms，0=未结算
ActualRR     float64 `gorm:"type:decimal(10,4);not null;default:0"`
```

> **注意：** `alert_record_repo.go` 的 `Create` 方法使用 `OnConflict DoUpdates` upsert。该 upsert 的更新列列表**不得包含** `outcome / outcome_price / outcome_at / actual_rr / interval`，确保调度重跑时不会覆盖已结算的结果。`OutcomeTrackerService` 通过独立的 `UpdateOutcome` 方法更新这四个字段。

#### alert_persistence.go 变更

`projectAlertRecord` 函数新增 `Interval` 字段填充，值取 `alertBiasInterval`（`"1h"`，即触发周期），以供过期计算使用。

#### OutcomeTrackerService

新文件：`backend/internal/service/outcome_tracker.go`

```
type OutcomeTrackerService struct {
  alertRecordRepo *repository.AlertRecordRepository
  klineRepo       *repository.KlineRepository
}

TrackOutcomes(ctx, symbols):
  for each symbol:
    records = alertRecordRepo.FindPending(symbol, limit=100)
    for each record:
      // 过期判断：固定使用 60 分钟窗口（1h×1）
      // 对应触发周期 1h，4 根 1h K 线足够判断结果
      if now - record.EventTime > 60*60*1000:
        UpdateOutcome(record.ID, "expired", 0, now, 0)
        continue
      klines = klineRepo.FindAfter(symbol, record.Interval, record.EventTime, limit=60)
      // 逐根 K 线扫描，止损优先（偏保守）
      for each kline in time order:
        if long (direction_state in ["strong-bullish","bullish"]):
          if kline.Low <= record.StopLoss:   → stop_hit, break
          if kline.High >= record.TargetPrice: → target_hit, break
        if short (direction_state in ["strong-bearish","bearish"]):
          if kline.High >= record.StopLoss:  → stop_hit, break
          if kline.Low <= record.TargetPrice:  → target_hit, break
      if hit: UpdateOutcome(record.ID, outcome, price, kline.OpenTime, actualRR)
```

**依赖注入：** `scheduler/jobs.go` 的 `Jobs` 结构体新增 `outcomeTracker *service.OutcomeTrackerService` 字段，`NewJobs` 函数新增对应参数，`cmd/server/main.go` 在构建 `Jobs` 时传入。

**新增 repo 方法：**
- `alert_record_repo.go`：`FindPending(symbol, limit)`、`UpdateOutcome(id, outcome, price, at, rr)`
- `kline_repo.go`：`FindAfter(symbol, interval, afterMs, limit)`（已有类似方法，可复用）

**文件改动：**
- `backend/models/alert_record.go` — 新增 5 个字段（含 `interval`）
- `backend/repository/alert_record_repo.go` — `FindPending`、`UpdateOutcome`；**upsert 列表排除 outcome 相关字段**
- `backend/repository/kline_repo.go` — `FindAfter`
- `backend/internal/service/outcome_tracker.go` — **新建**
- `backend/internal/service/alert_persistence.go` — `projectAlertRecord` 填充 `Interval`
- `backend/internal/scheduler/jobs.go` — `Jobs` 新增字段 + `NewJobs` 新增参数 + `runOutcomeTrack`
- `backend/cmd/server/main.go` — 依赖注入 `OutcomeTrackerService` 到 `Jobs`

---

### 4.6 S2 — Review 页 K 线上下文

#### ReviewChartModal.tsx（新组件）

触发：`AlertEventCard` 新增"复盘"图标按钮 → 打开 `ReviewChartModal`。

Modal 内容：
1. 信号 meta 头部（时间、标的、方向、置信度、outcome badge）
2. K 线图历史模式：
   - `KlineChart` 新增 `historicalMode?: { symbol: string; interval: string; centerTs: number }` prop
   - 历史模式下绕过 Zustand store，直接发起两次请求并合并：
     - `GET /api/kline?symbol=X&interval=Y&before_ts=centerTs&limit=20`（前20根）
     - `GET /api/kline?symbol=X&interval=Y&after_ts=centerTs&limit=40`（后40根）
   - 合并后按时间排序，传入 `KlineCandleLayer`
3. `SignalOverlayLayer` 叠加历史 entry/stop/target 三线（数据来自 `AlertEvent`，不依赖实时 store）
4. 若 outcome 已结算，在命中的 K 线位置标注结果标记（绿色 ✓ / 红色 ✗）

#### 后端变更

`GET /api/kline` 新增两个可选参数：
- `before_ts`（Unix 毫秒）：返回该时间点之前的 N 根 K 线（降序取，返回时升序）
- `after_ts`（Unix 毫秒）：返回该时间点之后的 N 根 K 线（升序取）

`KlineRepository` 新增：
- `FindBefore(symbol, interval, beforeMs, limit int) ([]models.Kline, error)`
- `FindAfter(symbol, interval, afterMs, limit int) ([]models.Kline, error)`（S1 也依赖）

**文件改动（慢线）：**
- `backend/repository/kline_repo.go` — `FindBefore`、`FindAfter`
- `backend/internal/service/market_service.go` — `GetKline` 支持 `before_ts` / `after_ts` 参数透传
- `backend/internal/handler/market_handler.go` — `GetKline` handler 读取并传递 `before_ts` / `after_ts`
- `frontend/types/alert.ts` — 新增 `AlertOutcome` 枚举、`AlertStats` 接口、`AlertRecord`（含 outcome 字段）
- `frontend/services/apiClient.ts` — `getKline` 支持 `before_ts` / `after_ts`
- `frontend/components/alerts/AlertEventCard.tsx` — 新增复盘按钮
- `frontend/components/alerts/ReviewChartModal.tsx` — **新建**
- `frontend/components/chart/KlineChart.tsx` — 新增 `historicalMode` prop

---

### 4.7 S3 — 胜率统计面板

#### 后端

新路由：`GET /api/alerts/stats`

**胜率公式（明确定义）：**
```
win_rate = target_hit / (target_hit + stop_hit) × 100
```
`pending` 和 `expired` 不计入分母。`avg_rr` 仅统计 `outcome = target_hit` 或 `stop_hit` 的记录的 `actual_rr` 均值。

**响应格式：**
```json
{
  "symbol": "BTCUSDT",
  "total": 42,
  "target_hit": 24,
  "stop_hit": 12,
  "pending": 4,
  "expired": 2,
  "win_rate": 66.7,
  "avg_rr": 1.85,
  "sample_size_label": "近 42 条"
}
```

支持 `?symbol=BTCUSDT&limit=50`（limit 控制最近 N 条 `setup_ready` 信号的统计窗口，默认 50）。

**文件改动：**
- `backend/repository/alert_record_repo.go` — 聚合查询方法 `GetStats(symbol, limit)`
- `backend/internal/service/alert_service.go` — `GetAlertStats`
- `backend/internal/handler/alert_handler.go` — `GetAlertStats` handler
- `backend/router/router.go` — 新增 `GET /api/alerts/stats`
- `frontend/services/apiClient.ts` — `getAlertStats`
- `frontend/components/alerts/WinRatePanel.tsx` — **新建**
- `frontend/app/review/page.tsx` — 挂载 WinRatePanel

#### 前端展示

`WinRatePanel.tsx` 挂载于 `/review` 页顶部：
- 每个标的一列（BTC / ETH / SOL）
- 每列：胜率%、平均 R:R、样本量
- 切换按钮："近20条 / 近50条 / 全部"

---

## 5. 完整文件改动清单

### 快线（Week 1）

**后端**
- `backend/models/alert_preference.go` — 新增 `SoundEnabled`
- `backend/internal/service/alert_preferences.go` — 读写 `SoundEnabled`
- `backend/internal/handler/alert_handler.go` — request 结构体 + 映射逻辑新增 `sound_enabled`
- `backend/.env.example` — `SCHEDULER_INTERVAL_SECONDS=15`

**前端**
- `frontend/types/alert.ts` — `AlertPreferences` 新增 `sound_enabled`
- `frontend/components/alerts/AlertCenter.tsx` — AudioContext + 声音触发
- `frontend/components/alerts/AlertConfigPanel.tsx` — 声音开关 UI
- `frontend/components/chart/KlineChart.tsx` — **拆分**（主容器 ≤420 行）
- `frontend/components/chart/KlineCandleLayer.tsx` — **新建**
- `frontend/components/chart/StructureLiquidityLayer.tsx` — **新建**
- `frontend/components/chart/SignalOverlayLayer.tsx` — **新建**（含三线）
- `frontend/components/trading/PositionCalculator.tsx` — **新建**
- `frontend/app/dashboard/page.tsx` — 挂载 PositionCalculator
- `frontend/app/chart/page.tsx` — 挂载浮动 PositionCalculator

### 慢线（Week 2-3）

**后端**
- `backend/models/alert_record.go` — 新增 `Interval` + outcome 相关 4 个字段
- `backend/repository/alert_record_repo.go` — `FindPending`、`UpdateOutcome`；upsert 列表排除 outcome 字段
- `backend/repository/kline_repo.go` — `FindBefore`、`FindAfter`
- `backend/internal/service/outcome_tracker.go` — **新建**
- `backend/internal/service/alert_persistence.go` — `projectAlertRecord` 填充 `Interval`
- `backend/internal/service/alert_service.go` — 新增 `GetAlertStats`
- `backend/internal/service/market_service.go` — `GetKline` 支持 `before_ts` / `after_ts`
- `backend/internal/scheduler/jobs.go` — `Jobs` 新增 `outcomeTracker` 字段 + `NewJobs` 参数 + `runOutcomeTrack`
- `backend/internal/handler/alert_handler.go` — 新增 `GetAlertStats`
- `backend/internal/handler/market_handler.go` — `GetKline` 支持 `before_ts` / `after_ts` 参数
- `backend/router/router.go` — 新增 `GET /api/alerts/stats`
- `backend/cmd/server/main.go` — 依赖注入 `OutcomeTrackerService`

**前端**
- `frontend/types/alert.ts` — 新增 `AlertOutcome`、`AlertStats`、`AlertRecord`（含 outcome）
- `frontend/services/apiClient.ts` — `getAlertStats`；`getKline` 支持 `before_ts` / `after_ts`
- `frontend/components/alerts/AlertEventCard.tsx` — 新增复盘按钮
- `frontend/components/alerts/ReviewChartModal.tsx` — **新建**
- `frontend/components/alerts/WinRatePanel.tsx` — **新建**
- `frontend/app/review/page.tsx` — 挂载 WinRatePanel
- `frontend/components/chart/KlineChart.tsx` — 新增 `historicalMode` prop

---

## 6. 测试策略

| 层级 | 覆盖点 |
|------|--------|
| 后端单元测试 | `OutcomeTrackerService`：多头止损优先逻辑、空头止损优先逻辑、过期（60分钟窗口） |
| 后端单元测试 | `OutcomeTrackerService`：同根 K 线同时触达止损+目标时，止损优先 |
| 后端单元测试 | `AlertService.GetAlertStats`：胜率公式（expired/pending 不计分母）、avg_rr 仅统计已结算 |
| 后端单元测试 | `alert_record_repo` upsert：调度重跑不覆盖已结算 outcome |
| 后端集成测试 | `GET /api/alerts/stats` 响应格式与字段正确性 |
| 后端集成测试 | `GET /api/kline` 的 `before_ts` / `after_ts` 参数返回正确时间范围 |
| 前端单元测试 | `PositionCalculator`：标准仓位计算、止损距离为0时禁用、仓位超账户余额警告 |
| 前端单元测试 | `WinRatePanel`：三个标的正确渲染、切换统计窗口触发重新请求 |
| 前端单元测试 | `KlineChart` 拆分后各图层 props 渲染（回归测试，拆分完成后立即运行） |
| 前端单元测试 | `ReviewChartModal`：数据不足时展示空状态，不崩溃 |
| E2E | Dashboard 打开仓位计算器，输入余额+风险比例，验证输出数值 |
| E2E | Chart 页确认 entry/stop/target 三线出现（有 alert 数据时） |

---

## 7. 风险与约束

| 风险 | 应对 |
|------|------|
| `KlineChart.tsx` 拆分引发回归 | 拆分后立即运行 `npm test` + Playwright 主路径，通过后再追加三线代码 |
| Binance API 速率（15s 调度） | 已计算安全（48次/分，远低于1200限制），触发限制时降回 30s |
| AutoMigrate `alert_records` 加列 | dev 模式自动执行；prod 需手动 `ALTER TABLE` 或临时开启 `AUTO_MIGRATE=true` |
| upsert 覆盖 outcome 数据 | `alert_record_repo.go` upsert 列表**明确排除** outcome 字段，`UpdateOutcome` 单独更新 |
| 历史 K 线缺失（data gap） | `ReviewChartModal` 展示"数据不足，无法还原当时走势"提示，不崩溃 |
| 声音告警 AudioContext 自动播放限制 | 首次用户交互（点击铃铛）后初始化，不可初始化时静默降级 |
| 过期窗口固定 60 分钟的局限 | 适用于 1h 触发周期的 setup_ready 信号，若后续支持其他周期告警，需从 `interval` 字段动态推算 |
