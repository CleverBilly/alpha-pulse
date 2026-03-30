# alpha-pulse

AI Crypto Futures Direction Copilot（BTC / ETH / SOL）

## 当前状态

当前仓库对应的版本已经完成 `V2.0 Futures Direction Copilot` 主线开发，可作为单用户公网部署的合约方向判断与告警终端使用。

当前上线范围：

- `BTCUSDT / ETHUSDT / SOLUSDT`
- `1m / 5m / 15m / 1h / 4h`
- Dashboard / Chart / Review（`/review`，`/signals` 兼容） / Market
- 订单流、结构、流动性、信号、AI 解释统一快照分析
- Futures 基础因子快照：mark / funding / open interest / long-short ratio / liquidation pressure proxy
- 完整 Futures Direction Engine：`4h / 1h / 15m / 5m`
- No-Trade 过滤与 A 级可跟踪判断
- Alert Center：浏览器通知 + 飞书机器人推送
- 告警配置中心：事件开关、最小置信度、静默时段、标的过滤
- 告警历史回放：recent feed + `/review` 复盘
- Auto Trading：真实 Binance Futures 限价开仓、超时撤单、保护单补挂、持仓同步与 `/auto-trading` 控制台

当前不包含：

- 回测平台
- 多交易所接入

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
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env.local
./scripts/bootstrap.sh
./scripts/dev.sh
```

- `backend/.env` 会在后端启动时自动加载，适合本地直接调试 MySQL 8
- `frontend/.env.local` 用于配置 API 地址和登录拦截开关
- `./scripts/dev.sh` 默认直接使用本地 MySQL / Redis，不依赖 Docker
- `./scripts/dev.sh` 会把本地登录所需的鉴权变量从 `backend/.env` 同步给前端进程，避免登录成功后又被 middleware 踢回 `/login`
- 本地和 Docker 默认都会把自动交易保持关闭，只有显式配置 `TRADE_ENABLED=true`、`TRADE_AUTO_EXECUTE=true` 并在 `/auto-trading` 页面保存运行时开关后，才会碰真实账户
- 如果你仍想用 Docker 起本地依赖，可使用 `USE_DOCKER_DEPS=1 ./scripts/dev.sh`

## 服务器部署

下面这套流程面向 `单台 Linux 服务器 + 公网域名 + Nginx`。
当前仓库支持两种生产部署方式：

- `宿主机 + PM2`：更适合宝塔面板，绕开容器 DNS、镜像构建和 Node/Go 多层环境问题
- `Docker Compose`：更适合纯容器化环境

推荐使用 `同一个域名` 对外提供服务，例如 `https://app.example.com`：

- `https://app.example.com/` -> 前端 Next.js
- `https://app.example.com/api/` -> 前端同域 API 代理，再转发到后端 Gin
- `https://app.example.com/healthz` -> 后端健康检查

这样前端、登录 Cookie、浏览器通知和飞书深链都最简单，不需要再拆前后端两个子域名。

推荐的公网链路是：

- 浏览器 -> `Nginx`
- `Nginx` -> `frontend:3000`
- 前端 Next.js 通过同域 `/api/*` -> `API_PROXY_TARGET=http://backend:8080`
- `Nginx` 只额外保留 `/healthz` 直连后端做健康检查

这样部署最稳，浏览器永远只认识一个域名和一套 `/api` 路径，也最不容易出现登录态和 404 混乱。

### 1. 服务器前置条件

- 一台 Linux 服务器，建议 `2C4G` 起步
- 已解析好的域名，例如 `app.example.com`
- 已安装以下两类组件中的至少一套：
  - `宿主机方案`：`go`、`node`、`npm`、`pm2`、`nginx`
  - `Docker 方案`：`docker`、`docker compose`、`nginx`
- 防火墙只开放：
  - `80`
  - `443`

说明：

- 不建议把 `3000 / 8080 / 3306 / 6379` 直接暴露公网
- 生产环境必须使用 HTTPS，否则登录 Cookie 安全性会下降

### 2. 拉代码

```bash
git clone <your-repo-url> /opt/alpha-pulse
cd /opt/alpha-pulse
```

### 3. 生成登录密码哈希

如果你要继续使用单用户登录，先生成 `bcrypt` 哈希：

```bash
docker run --rm httpd:2.4-alpine htpasswd -nbBC 10 "" 'admin123' | tr -d ':\n'
```

说明：

- 输出形如 `$2y$10$...`
- 放进 `.env` 时，建议用单引号包起来
- 否则像 `$2a$...` 这类 bcrypt 字符串可能会被 dotenv 当成变量展开

### 4. 准备后端环境变量

在服务器上创建 [backend/.env](/Users/billy/go/src/alpha-pulse/backend/.env)：

```bash
APP_MODE=prod
GIN_MODE=release
APP_PORT=8080

MYSQL_DSN=root:change-me@tcp(mysql:3306)/alpha_pulse?charset=utf8mb4&parseTime=True&loc=Local

REDIS_ADDR=redis:6379
REDIS_PASSWORD=
REDIS_DB=0
MARKET_SNAPSHOT_CACHE_TTL=5
ANALYSIS_VIEW_CACHE_TTL=15

MARKET_SYMBOLS=BTCUSDT,ETHUSDT,SOLUSDT
AUTO_MIGRATE=true
ENABLE_REDIS_CACHE=true
ENABLE_STREAM_COLLECTOR=true
ENABLE_SCHEDULER=true
ALLOW_MOCK_BINANCE_DATA=false
SCHEDULER_INTERVAL_SECONDS=60

ENABLE_SINGLE_USER_AUTH=true
AUTH_USERNAME=admin
AUTH_PASSWORD_HASH='<your-bcrypt-hash>'
AUTH_SESSION_SECRET=<long-random-secret>
AUTH_SESSION_TTL_HOURS=168
AUTH_COOKIE_NAME=alpha_pulse_session
AUTH_COOKIE_DOMAIN=
AUTH_COOKIE_SECURE=true

CORS_ALLOW_ORIGINS=https://app.example.com
ALERT_HISTORY_LIMIT=40
ALERT_PUBLIC_BASE_URL=https://app.example.com
FEISHU_BOT_WEBHOOK_URL=
FEISHU_BOT_SECRET=

BINANCE_API_KEY=
BINANCE_SECRET_KEY=
TRADE_ENABLED=false
TRADE_AUTO_EXECUTE=false
TRADE_ALLOWED_SYMBOLS=BTCUSDT,ETHUSDT,SOLUSDT
TRADE_WATCHER_INTERVAL_SECONDS=3
TRADE_SYNC_INTERVAL_SECONDS=15
```

关键说明：

- `APP_MODE=prod`
- `ALLOW_MOCK_BINANCE_DATA=false`
- 如果计划开启真实自动交易，再把 `TRADE_ENABLED=true` 和 `TRADE_AUTO_EXECUTE=true`
- `AUTO_MIGRATE=true`
  - 第一次上线建议先开启，让系统自动建表
  - 确认表结构稳定后，再改成 `false`
- `AUTH_PASSWORD_HASH` 要用单引号包住
- `AUTH_SESSION_SECRET` 建议使用下面命令生成：

```bash
openssl rand -hex 32
```

### 5. 准备前端生产环境变量

在服务器上创建 `frontend/.env.production`：

```bash
NEXT_PUBLIC_API_BASE_URL=/api
NEXT_PUBLIC_AUTH_ENABLED=true
AUTH_COOKIE_NAME=alpha_pulse_session
AUTH_SESSION_SECRET=<same-secret-as-backend>
```

说明：

- 这是前端 `build` 时要读的配置，不是本地开发用的 `frontend/.env.local`
- `AUTH_SESSION_SECRET` 必须和后端完全一致
- 推荐把 `NEXT_PUBLIC_API_BASE_URL` 保持为 `/api`，让浏览器统一走同域请求

### 6A. 宿主机 + PM2（推荐给宝塔）

如果你的服务器已经装好了宿主机版 `MySQL / Redis / Go / Node`，而且类似宝塔这种环境里 `docker build` 经常被 DNS 或源站拖慢，优先走宿主机部署会更稳。

先把后端 `.env` 改成宿主机连接：

```bash
MYSQL_DSN=<db-user>:<db-password>@tcp(127.0.0.1:3306)/<db-name>?charset=utf8mb4&parseTime=True&loc=Local
REDIS_ADDR=127.0.0.1:6379
AUTH_COOKIE_SECURE=false
CORS_ALLOW_ORIGINS=http://app.example.com
ALERT_PUBLIC_BASE_URL=http://app.example.com
```

说明：

- 上面这组适用于 `HTTP` 部署；如果已经接好 HTTPS，再把 `AUTH_COOKIE_SECURE=true`
- 如果你使用的是宝塔自带 MySQL，建议优先使用宝塔自带客户端做检查，例如 `/www/server/mysql/bin/mysql`

然后编译后端：

```bash
cd /opt/alpha-pulse/backend
go mod download
go build -o ./bin/alpha-pulse ./cmd/server
```

再编译前端：

```bash
cd /opt/alpha-pulse/frontend
npm ci
npm run build
```

如果你在宝塔环境里，Node 可能不在默认 PATH。常见位置是：

```bash
/www/server/nodejs/v24.14.1/bin/node
/www/server/nodejs/v24.14.1/bin/npm
```

接着复制一份 PM2 配置：

```bash
cd /opt/alpha-pulse
mkdir -p logs
cp deploy/ecosystem.host.example.cjs ecosystem.config.cjs
```

默认示例会读取：

- `ALPHA_PULSE_ROOT`，默认 `/www/wwwroot/alpha-pulse`
- `ALPHA_PULSE_NPM_BIN`，默认 `/www/server/nodejs/v24.14.1/bin/npm`

启动并保存进程：

```bash
cd /opt/alpha-pulse
pm2 start ecosystem.config.cjs
pm2 save
pm2 startup systemd -u root --hp /root
```

常用运维命令：

```bash
pm2 status
pm2 logs alpha-pulse-backend --lines 100
pm2 logs alpha-pulse-frontend --lines 100
pm2 restart alpha-pulse-backend
pm2 restart alpha-pulse-frontend
```

后续更新代码时，推荐使用仓库自带的一键部署脚本：

```bash
cd /opt/alpha-pulse
git pull origin main
bash scripts/deploy.sh
```

`scripts/deploy.sh` 会固定完成这些动作：

- 自动补齐宝塔宿主机的 `go / npm / pm2` 路径
- 检查 `backend/.env`、`frontend/.env.production`、`ecosystem.config.cjs`
- 编译后端：`go mod download` + `go build`
- 构建前端：`npm ci` + `npm run build`
- 重启 `alpha-pulse-backend` 和 `alpha-pulse-frontend`
- 做三条本机健康检查：
  - `http://127.0.0.1:8080/healthz` -> `200`
  - `http://127.0.0.1:3000/login` -> `200`
  - `http://127.0.0.1:3000/api/trade-settings` -> `401`

说明：

- 脚本不会自动 `git pull`，这样你可以先确认分支和提交再发布
- 脚本不会修改 `backend/.env`、`frontend/.env.production`、Nginx 或 SSL
- 失败时会直接退出，并提示查看 `deploy/.tmp/` 里的阶段日志
- 如果你在宝塔面板里给站点开了“反向代理缓存”或 `proxy_cache`，请对 Next.js 主站和 `/api/*` 关闭它；这类接口必须实时回源，否则像 `/api/trade-settings` 这种保存后立即回读的页面会看到旧值

### 6B. 准备生产版 Compose 文件

建议在服务器上新建 `docker/docker-compose.prod.yml`：

```yaml
services:
  mysql:
    image: mysql:8.4
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: change-me
      MYSQL_DATABASE: alpha_pulse
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7.4-alpine
    restart: unless-stopped

  backend:
    build:
      context: ..
      dockerfile: docker/backend.Dockerfile
    restart: unless-stopped
    env_file:
      - ../backend/.env
    depends_on:
      - mysql
      - redis
    ports:
      - "127.0.0.1:8080:8080"

  frontend:
    build:
      context: ..
      dockerfile: docker/frontend.Dockerfile
    restart: unless-stopped
    environment:
      API_PROXY_TARGET: http://backend:8080
    depends_on:
      - backend
    ports:
      - "127.0.0.1:3000:3000"

volumes:
  mysql_data:
```

说明：

- 这里把 `3000 / 8080` 只绑定到 `127.0.0.1`，宿主机 Nginx 可以访问，但公网无法直接访问
- 外部流量统一走 Nginx
- `frontend` 镜像构建前，`frontend/.env.production` 必须已经存在
- 真实 Binance 密钥和自动交易开关统一放在 `backend/.env`，不要把敏感值直接硬编码进 `docker-compose.prod.yml`
- `frontend` 容器会把同域 `/api/*` 请求转发到 `API_PROXY_TARGET`，这样 `/api/trade-settings`、`/api/trades/runtime` 这类接口即使不直接暴露后端给浏览器，也能正常工作

### 7. 构建并启动

```bash
cd /opt/alpha-pulse/docker
docker compose -f docker-compose.prod.yml build
docker compose -f docker-compose.prod.yml up -d
```

查看状态：

```bash
docker compose -f docker-compose.prod.yml ps
docker compose -f docker-compose.prod.yml logs -f backend
docker compose -f docker-compose.prod.yml logs -f frontend
```

先做两条最重要的本机 smoke check：

```bash
curl -i http://127.0.0.1:8080/healthz
curl -i http://127.0.0.1:3000/api/trade-settings
```

预期：

- `/healthz` 返回 `200`
- `/api/trade-settings` 在未登录时返回 `401`

如果这里已经是 `401`，说明交易路由、前端 `/api` 代理和鉴权中间件都已经接通了，不应该再出现 `404`。

### 8. 配置 Nginx 反向代理

新建 `/etc/nginx/sites-available/alpha-pulse.conf`：

```nginx
server {
    listen 80;
    server_name app.example.com;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /healthz {
        proxy_pass http://127.0.0.1:8080/healthz;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

说明：

- 推荐把所有业务路径都先交给前端容器，包括 `/api/*`
- 前端会通过 `API_PROXY_TARGET` 再把 `/api/*` 转发到后端
- 这样浏览器始终只请求 `https://app.example.com/api/...`，部署和排查都更简单
- 不要再额外给这个反代层配置 `proxy_cache`，也不要忽略上游的 `Cache-Control`; 这会把交易配置、登录态相关接口缓存住，表现出来就是页面保存成功但下一次读取还是旧值

如果你明确想让 `Nginx` 直接反代 `/api/` 到后端，也可以加一条：

```nginx
location /api/ {
    proxy_pass http://127.0.0.1:8080/api/;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

但二选一就够了，不要一会儿走前端代理、一会儿走 Nginx 直连，免得排查时路径混乱。

启用站点：

```bash
ln -s /etc/nginx/sites-available/alpha-pulse.conf /etc/nginx/sites-enabled/alpha-pulse.conf
nginx -t
systemctl reload nginx
```

### 9. 申请 HTTPS 证书

```bash
certbot --nginx -d app.example.com
```

完成后确认：

- 浏览器访问 `https://app.example.com/login`
- 后端健康检查：

```bash
curl https://app.example.com/healthz
```

### 10. 上线后验证清单

- 登录页可打开
- 使用你配置的单用户账号可以成功登录
- `/dashboard`、`/market`、`/review` 正常打开
- `/auto-trading` 可正常打开并看到运行时配置页
- 浏览器开发者工具里 `/api/auth/login` 返回 `200`
- 浏览器开发者工具里 `/api/trade-settings`、`/api/trades/runtime` 返回 `200` 或 `401`，但不应是 `404`
- `docker compose -f docker-compose.prod.yml logs -f backend` 没有持续报错
- 飞书机器人如果已配置，能收到测试提醒

如果计划开启真实自动交易，再额外确认：

- `backend/.env` 已设置：
  - `ALLOW_MOCK_BINANCE_DATA=false`
  - `BINANCE_API_KEY`
  - `BINANCE_SECRET_KEY`
  - `TRADE_ENABLED=true`
  - `TRADE_AUTO_EXECUTE=true`
- 账号已开通 Binance Futures 权限，API Key 具备期货交易权限
- `/auto-trading` 页面里已经保存运行时配置，并显式打开“自动执行”
- 先只放行最少的白名单标的，例如先只开 `BTCUSDT`

### 11. 常见问题

`登录一直 401`

- 先确认后端容器已经重启
- 确认 `AUTH_PASSWORD_HASH` 是 bcrypt 哈希，不是明文
- 确认 bcrypt 哈希在 `.env` 里使用了单引号
- 确认 `AUTH_USERNAME` 和你登录时输入的一致

`前端能打开，但请求 /api 失败`

- 确认 `NEXT_PUBLIC_API_BASE_URL=/api`
- 如果由前端容器代理后端，请确认 `API_PROXY_TARGET=http://backend:8080`
- 如果由 Nginx 直接反代 `/api/`，确认这条规则已经生效；如果你走推荐方案，就不要再额外配这条规则
- 确认后端容器健康，且 `CORS_ALLOW_ORIGINS=https://app.example.com`

`/auto-trading` 页面里的交易接口返回 `404`

- 先在服务器上执行 `curl -i http://127.0.0.1:3000/api/trade-settings`
- 如果返回 `401`，说明路由存在，问题不在后端路由表
- 如果这里返回 `404`，优先检查：
  - 前端镜像是不是旧构建，没带新的 `/api` 代理路由
  - 后端是不是旧进程/旧容器，没带交易路由
  - `./scripts/dev.sh` 本地调试时是否有旧 `8080` 进程占着端口
- 本地开发里，最新的 [scripts/dev.sh](/Users/billy/go/src/alpha-pulse/scripts/dev.sh) 会在启动前直接拦住 `8080` 端口冲突，避免被旧 `/healthz` 假阳性骗过

`Binance 返回 -2015 Invalid API-key, IP, or permissions for action`

- 先检查 API Key / Secret 是否正确
- 如果 Binance API 配置了 IP 白名单，把服务器出口 IP 加进去
- 确认该 API Key 已开启 `Futures` 交易权限
- 修改 `backend/.env` 后，记得重启后端容器

`浏览器登录成功后又被踢回 /login`

- 确认前后端 `AUTH_SESSION_SECRET` 完全一致
- 确认 `NEXT_PUBLIC_AUTH_ENABLED=true`
- HTTPS 上确认 `AUTH_COOKIE_SECURE=true`
- 如果你使用 `./scripts/dev.sh`，修改 `backend/.env` 或 `frontend/.env.local` 后请重启脚本，让前端重新读取鉴权变量

`后台没有实时数据或告警`

- 确认：
  - `ENABLE_STREAM_COLLECTOR=true`
  - `ENABLE_SCHEDULER=true`
  - `ALLOW_MOCK_BINANCE_DATA=false`
- 再检查 Binance 网络连通性和后端日志

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
MARKET_SYMBOLS=BTCUSDT,ETHUSDT,SOLUSDT
AUTO_MIGRATE=true
ENABLE_REDIS_CACHE=true
ENABLE_STREAM_COLLECTOR=true
ENABLE_SCHEDULER=true
ALLOW_MOCK_BINANCE_DATA=true
SCHEDULER_INTERVAL_SECONDS=60
TRADE_ENABLED=false
TRADE_AUTO_EXECUTE=false
TRADE_ALLOWED_SYMBOLS=BTCUSDT,ETHUSDT,SOLUSDT
TRADE_WATCHER_INTERVAL_SECONDS=3
TRADE_SYNC_INTERVAL_SECONDS=15
```

说明：

- `GIN_MODE` 默认随 `APP_MODE` 推导：`dev -> debug`，`test -> test`，`prod -> release`
- 所有 mode 默认值都可以被显式环境变量覆盖
- 本地 `docker compose` 当前默认按 `dev` 模式启动
- 本地 Docker 默认把自动交易开关保持为关闭态，避免容器一启动就接触真实账户

## 单用户登录拦截

当前支持单用户登录模式，用于公网域名部署时保护 Dashboard、Chart、Review、Market 等业务页。

后端需要配置：

```bash
ENABLE_SINGLE_USER_AUTH=true
AUTH_USERNAME=alpha-admin
AUTH_PASSWORD_HASH=<bcrypt-hash>
AUTH_SESSION_SECRET=<same-long-random-secret>
AUTH_COOKIE_NAME=alpha_pulse_session
AUTH_COOKIE_DOMAIN=
AUTH_COOKIE_SECURE=true
CORS_ALLOW_ORIGINS=https://your-frontend-domain.example.com
```

前端需要配置：

```bash
NEXT_PUBLIC_AUTH_ENABLED=true
NEXT_PUBLIC_API_BASE_URL=/api
AUTH_COOKIE_NAME=alpha_pulse_session
AUTH_SESSION_SECRET=<same-long-random-secret>
```

说明：

- `AUTH_PASSWORD_HASH` 必须使用 `bcrypt` 哈希，后端不会接受明文密码
- `AUTH_SESSION_SECRET` 前后端必须一致，供后端签发和前端 middleware 校验登录态
- 推荐把 `NEXT_PUBLIC_API_BASE_URL` 保持为 `/api`，走同域请求；只有在你明确拆分前后端域名时，才改成完整后端 URL
- 公网 HTTPS 部署时建议开启 `AUTH_COOKIE_SECURE=true`
- `CORS_ALLOW_ORIGINS` 必须精确列出允许访问后端的前端域名

## Alert Center / 飞书机器人

当前告警链路由后端定时评估 `BTC / ETH / SOL` 的 `4h / 1h / 15m / 5m` 多周期方向状态，并在出现以下事件时生成 feed：

- `A 级 setup 已就绪`
- `方向切换`
- `进入 No-Trade`

后端可选接入飞书自定义机器人：

```bash
ALERT_HISTORY_LIMIT=40
ALERT_PUBLIC_BASE_URL=https://your-frontend-domain.example.com
FEISHU_BOT_WEBHOOK_URL=https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook
FEISHU_BOT_SECRET=
```

说明：

- `FEISHU_BOT_WEBHOOK_URL` 留空时，系统只保留站内 Alert Center 和浏览器通知，不推送飞书
- 如果飞书机器人启用了签名校验，再填写 `FEISHU_BOT_SECRET`
- `ALERT_PUBLIC_BASE_URL` 会写进飞书消息深链，指向 `/dashboard`、`/market`、`/review`
- 浏览器通知不需要额外环境变量，登录后在右上角 `Alerts` 抽屉里授权即可
- 本地开发如果关闭了 `ENABLE_SCHEDULER`，仍可在 Alert Center 里点击 `立即检查`

Alert 配置中心当前支持：

- 飞书 / 浏览器推送开关
- `setup_ready / direction_shift / no_trade` 事件开关
- 最小置信度阈值
- `BTCUSDT / ETHUSDT / SOLUSDT` 标的过滤
- 飞书静默时段

相关接口：

- `GET /api/alerts`
- `POST /api/alerts/refresh`
- `GET /api/alerts/history`
- `GET /api/alerts/preferences`
- `PUT /api/alerts/preferences`

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

## 自动交易配置

当前仓库已经支持真实 Binance Futures 自动交易，但默认保持关闭。
`/api/trade-settings`、`/api/trades/runtime`、`/api/trades` 这类接口默认走同域 `/api`，可由 Nginx 反代，也可由前端容器自身代理到后端。
真实下单链路为：

- `setup_ready` 告警命中
- 校验环境底线开关、运行时开关、白名单标的、最小盈亏比、最大持仓数
- 按信号 `entry_price` 发 `LIMIT` 开仓单
- 超过超时秒数未成交则撤单并标记 `expired`
- 成交后自动补挂 `STOP_MARKET` 和 `TAKE_PROFIT_MARKET`
- 后台循环同步 Binance 持仓，识别手动单和交易所侧已平仓

要启用真实自动交易，后端至少需要这些环境变量：

```bash
ALLOW_MOCK_BINANCE_DATA=false
BINANCE_API_KEY=your_binance_api_key
BINANCE_SECRET_KEY=your_binance_secret_key

TRADE_ENABLED=true
TRADE_AUTO_EXECUTE=true
TRADE_ALLOWED_SYMBOLS=BTCUSDT,ETHUSDT,SOLUSDT
TRADE_WATCHER_INTERVAL_SECONDS=3
TRADE_SYNC_INTERVAL_SECONDS=15
```

说明：

- `TRADE_ENABLED`
  - 部署层总开关
  - 为 `false` 时，交易写接口和后台自动执行都不会触发真实账户操作
- `TRADE_AUTO_EXECUTE`
  - 部署层自动执行底线
  - 为 `false` 时，仍可查看 `/auto-trading` 页面和订单记录，但不会自动下单
- `TRADE_ALLOWED_SYMBOLS`
  - 部署层允许触碰的标的全集
  - `/auto-trading` 页面里的白名单只能从这组标的里再做子集筛选
- `TRADE_WATCHER_INTERVAL_SECONDS`
  - LIMIT 开仓单的成交/超时检查频率
- `TRADE_SYNC_INTERVAL_SECONDS`
  - Binance 持仓同步频率，用于识别手动单和交易所侧平仓

除了环境变量，运行时还需要去 `/auto-trading` 页面保存一份配置，常用项包括：

- 自动执行开关
- 允许标的白名单
- 风险比例 `risk_pct`
- 最低盈亏比 `min_risk_reward`
- 限价单超时秒数 `entry_timeout_seconds`
- 最大持仓数 `max_open_positions`
- 持仓同步开关

也就是说，真实自动下单必须同时满足三层：

- `TRADE_ENABLED=true`
- `TRADE_AUTO_EXECUTE=true`
- `/auto-trading` 页面里的“自动执行”已开启并保存

相关接口：

- `GET /api/trade-settings`
- `PUT /api/trade-settings`
- `GET /api/trades`
- `GET /api/trades/runtime`
- `POST /api/trades/:id/close`

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
- 页面自动轮询默认走缓存；点击页面内“刷新”按钮时会显式附加 `refresh=1`

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
- `GET /api/alerts`
- `GET /api/alerts/history`
- `GET /api/alerts/preferences`
- `POST /api/alerts/refresh`
- `PUT /api/alerts/preferences`
