# Alpha Pulse V2.0 Futures Direction Copilot PRD

版本: V2.0 Proposal  
状态: Draft for implementation  
更新时间: 2026-03-11

## 1. 文档定位

本文档定义 `Alpha Pulse` 下一阶段的产品目标, 用于指导:

- 产品范围收敛
- 前后端开发实现
- 接口与数据模型演进
- 测试与验收
- 后续 V2.5 / V3.0 的边界控制

本文档是 `Spot Analysis MVP` 之后的下一阶段产品规格, 不覆盖当前 `docs/PRD.md` 中已经完成的 Spot 主线实现细节。`docs/PRD.md` 仍然是当前线上 Spot 版本的 Source of Truth; 本文档定义的是下一阶段 `Futures Direction Copilot` 方向。

## 2. 产品一句话定义

`Alpha Pulse V2.0` 是一个面向个人使用的 `Crypto Futures Direction Copilot`, 核心目标是帮助用户先看清 `BTC / ETH / SOL` 永续合约的主方向, 明确识别 `No-Trade` 区域, 并只在高质量 setup 出现时提醒用户, 为后续半自动执行奠定基础。

## 3. 背景与问题

当前系统已经完成 `Spot Analysis MVP`, 具备统一 snapshot、结构、流动性、订单流、信号与前端 cockpit 的基础能力。但对于目标用户当前最关键的问题, Spot 版本并不能直接解决:

- 用户真实交易场景以 `Crypto Futures` 为主, 不是现货
- 用户核心痛点是 `方向看不准`, 而不是缺少更多图表
- 用户日常盯盘时间过长, 容易因为疲劳与情绪导致错误交易
- 当前系统缺少 `No-Trade` 作为一级产品能力
- 当前系统没有公网访问下的登录拦截, 不适合直接作为个人私有交易工具长期使用

因此 V2.0 的重点不是继续堆叠 Spot 分析能力, 而是转向 `Futures 主决策链路`, 让系统优先回答:

- 现在偏多、偏空, 还是禁止交易
- 这个方向是否有更高周期支持
- 当前 setup 是否成熟
- 什么时候值得看盘, 什么时候不值得出手

## 4. 目标用户

V2.0 当前只服务一个用户:

- 系统所有者本人
- 通过公网域名访问
- 需要登录后才能进入查看业务页面

用户交易画像:

- 主要交易 `BTC / ETH / SOL` 永续合约
- 以 `1h` 为主判断周期
- 使用 `5m / 15m` 做进出场触发
- 有盈利时很多单子会在 `1h` 内结束
- 一部分优质单会持有到 `4h+`

## 5. 产品目标

### 5.1 V2.0 必须达成的目标

1. 支持 `BTC / ETH / SOL` 的 `Crypto Futures` 主分析链路
2. 建立 `4h -> 1h -> 15m / 5m` 的多周期方向框架
3. 输出明确的方向结论与 `No-Trade` 状态
4. 支持单用户登录拦截与公网访问下的最小安全基线
5. 将高质量 setup 变成低频、高信噪比告警
6. 支持方向结论、告警与关键 snapshot 的回看与复盘

### 5.2 V2.0 明确不做

- 自动下单
- 半自动执行
- 多用户系统
- 注册 / 找回密码 / 权限体系
- 多交易所接入
- 黄金 / 白银 / 其他非 Crypto Futures 品种
- 大范围山寨币扫描
- 完整回测平台

## 6. 支持范围

### 6.1 市场与标的

- 市场: `Crypto Futures`
- 标的: `BTC`, `ETH`, `SOL`

### 6.2 周期

- `4h`: 大方向过滤
- `1h`: 主方向判断
- `15m`: setup 预备与局部确认
- `5m`: 精细触发辅助

### 6.3 访问方式

- 公网域名访问
- 未登录不可访问业务页
- 单用户模式

## 7. 产品原则

1. `方向优先`
   系统的第一输出永远是方向结论, 不是指标清单。

2. `No-Trade 是一级能力`
   系统必须明确告诉用户什么时候不该碰, 而不是只会喊多空。

3. `统一快照优先`
   V2.0 继续以统一 snapshot 驱动前端, 避免多接口拼装导致的数据撕裂。

4. `Futures 独立推进`
   Futures 主链路作为独立方向实现, 不直接污染当前 Spot 主线。

5. `少而准的告警`
   告警不是把更多信息推送出去, 而是只推值得处理的高质量机会。

## 8. 功能模块

### 8.1 单用户登录模块

目标: 让系统在公网部署场景下具备最小可用安全边界。

能力要求:

- 登录页
- 固定单用户账号
- 密码哈希存储
- 服务端登录校验
- `HttpOnly` session cookie
- 登录态保持
- 登出
- 未登录访问业务页时自动跳转登录页

明确不做:

- 注册
- 忘记密码
- 邮箱验证码
- 多用户管理

### 8.2 Futures 数据与统一快照模块

目标: 把 Futures 方向判断所需关键数据聚合成统一 snapshot。

Futures snapshot 至少包含:

- futures price / ticker
- futures kline
- funding rate
- open interest
- long / short 情绪或等价公开指标
- 清算热区或清算压力摘要
- 结构
- 流动性
- 订单流
- direction result
- no-trade state
- 当前 setup 摘要

产品要求:

- Dashboard 继续优先读取统一 snapshot
- Futures snapshot 与现有 Spot snapshot 分离, 但设计风格尽量一致

### 8.3 Direction Engine

目标: 给出用户可直接消费的方向结论。

输出口径:

- `强偏多`
- `偏多`
- `观望`
- `偏空`
- `强偏空`
- `禁止交易`

每次结论至少包含:

- direction state
- confidence
- risk level
- primary timeframe verdict (`1h`)
- higher timeframe alignment (`4h`)
- top reasons
- invalidation condition

方向框架:

- `4h` 用于过滤大方向
- `1h` 用于形成主判断
- `15m / 5m` 只用于确认与触发, 不主导宏观方向

### 8.4 No-Trade Filter

目标: 把“不交易”变成产品主功能。

系统至少需要识别以下 `No-Trade` 场景:

- `4h` 与 `1h` 方向冲突
- Futures 因子与结构方向冲突
- 结构混乱, 无法形成清晰趋势 / 反转判断
- 接近高风险关键位但没有确认
- 高波动异常阶段, 风险收益比不成立
- 追价风险过高

产品行为要求:

- `No-Trade` 作为一级页面状态
- `No-Trade` 状态下不触发 A 级 setup 告警
- 页面中必须明确展示禁止交易原因

### 8.5 Dashboard 与 Watchlist 模块

目标: 让用户在极短时间内看清主方向, 并只盯少数值得处理的机会。

Dashboard 必须优先展示:

- 当前标的方向结论
- confidence
- risk level
- 是否 `No-Trade`
- 当前 setup 是否成熟
- 方向证据链
- 最近同步时间

Watchlist 视图必须支持:

- `BTC / ETH / SOL` 并列展示
- 每个标的当前方向
- 每个标的当前是否可交易
- 每个标的当前是否存在 A 级 setup

### 8.6 告警模块

目标: 降低盯盘成本。

V2.0 告警只做高信噪比能力:

- A 级 setup 提醒
- 方向反转提醒
- `No-Trade` 进入提醒
- `No-Trade` 解除提醒

告警通道优先级:

1. 浏览器通知
2. Telegram

不做:

- 高频刷屏型事件推送
- 复杂工作流 / 多人订阅

### 8.7 Snapshot / Alert Review 模块

目标: 让系统具备最小复盘价值。

必须保存:

- 关键 futures snapshot
- 方向结论
- 告警触发记录
- 告警触发当时的主要原因

用户必须可以回看:

- 这个提醒为什么会发出
- 当时系统是否处于 `No-Trade`
- 当时方向结论与理由是什么

## 9. 页面范围

V2.0 页面建议收敛为:

- `/login`
- `/dashboard`
- `/chart`
- `/signals`
- `/market` 或 `/watchlist`
- `/review` 或等价的快照 / 告警回看页

页面权重:

- `/dashboard` 是主入口
- 其他页面是深度钻取页

## 10. 关键用户流程

### 10.1 每日开盘前查看

1. 用户登录系统
2. 打开 dashboard
3. 先查看 `BTC / ETH / SOL` 当前方向与 `No-Trade`
4. 重点查看 `1h` 主方向是否获得 `4h` 支持
5. 如果为 `No-Trade`, 则不进入执行观察

### 10.2 日内盯盘减负

1. 用户不持续盯图
2. 系统在后台持续计算 Futures snapshot 与 direction
3. 只有当 A 级 setup 出现时才发出告警
4. 用户收到提醒后进入 dashboard 或 chart 深页查看证据链

### 10.3 告警后决策

1. 用户收到提醒
2. 查看 setup 是否仍处于有效状态
3. 查看风险与 invalidation
4. 自行决定是否手动执行

### 10.4 事后复盘

1. 用户进入 review
2. 查看今天出现过的告警
3. 回看每次告警的方向、No-Trade 状态和核心证据
4. 评估哪些提醒值得保留, 哪些提醒应调低或调高门槛

## 11. 方向结论与告警规则

### 11.1 方向结论

方向结果必须具备可解释性, 不是黑盒分数。

每次方向结论必须至少给出:

- state
- confidence
- top reasons
- higher timeframe alignment
- invalidation

### 11.2 A 级 setup

V2.0 中 A 级 setup 的产品含义是:

- 大方向明确
- 主周期方向明确
- `No-Trade` 不成立
- 局部触发条件成熟
- 风险收益比达到最低门槛

### 11.3 告警降噪原则

- 不因单一因子异动就推送
- 不在 `No-Trade` 时推送 A 级 setup
- 同一 setup 在短时间内避免重复刷屏

## 12. 安全与部署要求

由于系统通过公网域名访问, V2.0 至少要求:

- 登录拦截
- `HttpOnly` cookie
- 基础 session 失效机制
- 账号与密码不硬编码在前端
- 服务端敏感配置通过环境变量管理
- 预留后续接入交易所 API key 的安全边界

## 13. 数据与存储要求

V2.0 应至少支持保存以下数据:

- futures snapshot
- direction snapshot
- alert events
- login / session 审计基础记录

说明:

- V2.0 的复盘重点是 `方向与提醒`, 不是完整交易执行流水
- 自动交易相关表结构不在本版本强制定义

## 14. 成功标准

V2.0 成功上线的标准:

1. 用户打开系统后, `3-5` 秒内能看出当前主方向
2. 系统可以明确告诉用户何时 `禁止交易`
3. 用户无需长时间盯盘才能发现高质量机会
4. A 级 setup 告警频率低于普通事件流, 但质量明显更高
5. 未登录用户无法访问业务页

## 15. V2.5 / V3.0 边界

### 15.1 V2.5 方向

- setup scanner 增强
- watchlist 强化
- 告警中心增强
- 复盘能力增强

### 15.2 V3.0 方向

- Binance Futures 账户接入
- 半自动执行
- 风控规则
- 执行审计

结论:

V2.0 必须先把 `看清方向 + 减少盯盘` 做出来, 不提前进入自动执行。
