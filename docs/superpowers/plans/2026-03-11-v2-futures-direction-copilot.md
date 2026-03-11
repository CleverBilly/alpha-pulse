# V2.0 Futures Direction Copilot Implementation Plan

> **For agentic workers:** REQUIRED: Use `superpowers:executing-plans` or an equivalent staged execution workflow. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在保留现有 Spot MVP 的同时, 新增一个面向 `BTC / ETH / SOL` 的 `Crypto Futures Direction Copilot`, 先完成单用户登录拦截、Futures 统一快照、方向引擎、No-Trade、A 级告警与基础复盘。

**Architecture:** Futures 作为独立链路推进, 不与现有 Spot 主链路混改。前端继续优先消费统一 snapshot; 服务端新增 Futures 采集、聚合与方向服务; 登录采用单用户模式的服务端 session / cookie 方案。

**Tech Stack:** Go, Gin, GORM, MySQL, Redis, Next.js, TypeScript, Zustand, Playwright, Vitest

---

## Summary

- Phase 1 先补公网访问下的单用户登录闸门, 让系统具备最小安全边界。
- Phase 2 建立 Futures 数据链路和统一 snapshot, 但先只做 `BTC / ETH / SOL`。
- Phase 3 实现 Direction Engine、No-Trade 与 A 级 setup 规则。
- Phase 4 重做前端 Futures cockpit, 接入 watchlist、alerts 与 review 基础页。

## Key Changes

### Phase 1: 单用户登录与访问保护

目标:

- 未登录用户无法访问业务页
- 登录后才能进入 dashboard / chart / signals / market

主要改动:

- 后端新增单用户认证配置、密码哈希校验、session 管理与登录 / 登出接口
- 前端新增 `/login` 页面
- 前端新增路由级拦截, 业务页按登录态跳转
- 公网部署文档补充认证与 cookie 要求

优先文件:

- `backend/router/router.go`
- `backend/internal/handler/`
- `backend/config/config.go`
- `frontend/app/`
- `frontend/middleware` 或等价路由保护入口

### Phase 2: Futures 数据链路与统一 snapshot

目标:

- 建立 `BTC / ETH / SOL` Futures 主链路
- 提供单一 futures snapshot 给前端使用

主要改动:

- Futures 公共行情与因子采集
- Futures Kline / Ticker / Funding / Open Interest 等服务层封装
- 新增 futures snapshot 聚合服务与缓存
- 独立于 Spot 的类型、路由与缓存 key

优先文件:

- `backend/pkg/binance/`
- `backend/internal/collector/`
- `backend/internal/service/`
- `backend/internal/handler/`
- `frontend/services/apiClient.ts`
- `frontend/types/`

### Phase 3: Direction Engine / No-Trade / Alert Rules

目标:

- 输出明确方向结论
- 形成 `No-Trade` 一级能力
- 只推送 A 级 setup

主要改动:

- 新增 Futures direction service / engine
- 建立 `4h -> 1h -> 15m / 5m` 多周期判定框架
- 定义 direction state、confidence、risk、invalidation
- 定义 `No-Trade` 判定与 A 级告警门槛
- 保存 direction snapshot 与 alert event

优先文件:

- `backend/internal/signal/` 或独立 `backend/internal/futures/`
- `backend/internal/service/`
- `backend/models/`
- `backend/repository/`
- `docs/api.md`

### Phase 4: Futures Dashboard / Watchlist / Review

目标:

- 让用户先看方向, 再看 setup, 再看证据链
- 支持 watchlist、alert history 与 snapshot review

主要改动:

- 新增或重做 Futures dashboard 页面
- 新增 watchlist 状态板
- 新增 review 页
- 页面以统一 futures snapshot 驱动
- 保留现有 Spot 页面, 不强制删除

优先文件:

- `frontend/app/dashboard/`
- `frontend/app/market/`
- `frontend/app/signals/`
- `frontend/components/`
- `frontend/store/`
- `frontend/tests/e2e/`

## Delivery Order

1. 先实现 Phase 1 登录拦截
2. 再实现 Phase 2 Futures snapshot 基线
3. 接着实现 Phase 3 direction / no-trade / alerts
4. 最后重做 Phase 4 Futures cockpit 与 review

理由:

- 公网部署下, 登录拦截是硬前置
- 没有 Futures snapshot, direction 与 dashboard 都没有稳定输入
- 没有 direction / no-trade, 告警质量无法保证

## Test Plan

- 登录:
  - 未登录访问业务页会跳转
  - 正确登录可进入
  - 错误密码不可进入
  - 登出后 session 失效
- Futures snapshot:
  - `BTC / ETH / SOL` 都能返回统一 snapshot
  - Redis 缓存与刷新逻辑正常
  - Spot 与 Futures 路由互不污染
- Direction / No-Trade:
  - 至少覆盖强偏多、偏空、观望、禁止交易四类核心场景
  - A 级 setup 只在可交易状态触发
- Frontend:
  - Dashboard 能在数秒内显示方向、风险、setup 与证据链
  - Watchlist 能同时展示三大标的状态
  - Review 能回看触发记录

## Assumptions

- V2.0 不接自动交易
- V2.0 只做单用户模式
- Spot MVP 保留, 但不再作为主产品方向
- 黄金、白银与非 Crypto Futures 品种不进入本版本
- 如果某些 Futures 因子公开数据质量不足, 可以先用最稳定可得的公开数据组合完成第一版

## First Slice Recommendation

第一个实施切片建议固定为:

`单用户登录 + 业务页访问拦截 + 公网部署安全基线`

原因:

- 这是公网访问场景的硬门槛
- 技术范围清晰, 可以独立交付
- 交付后不会影响后续 Futures 数据链路扩展
