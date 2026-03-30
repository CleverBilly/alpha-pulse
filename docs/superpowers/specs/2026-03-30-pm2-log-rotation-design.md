# PM2 Log Rotation Design

## Goal

让宿主机 + PM2 部署默认具备日志大小上限和归档保留策略，避免 `backend.err.log`、`backend.out.log` 等文件持续膨胀到数 GB，挤爆服务器磁盘。

## Current Problem

- PM2 目前把前后端标准输出和错误输出直接写到固定日志文件。
- 现有部署链路没有任何轮转、压缩或保留数量控制。
- 一旦后端持续打印错误，日志会线性增长，当前线上已经出现 `backend.err.log` 约 `1.4G` 的情况。

## Chosen Approach

默认启用 `pm2-logrotate`，并在部署脚本里强制收敛为固定策略：

- 单个日志文件最大 `100MB`
- 最多保留 `5` 份归档
- 压缩历史归档

同时，在每次执行 `scripts/deploy.sh` 时：

- 确保 `pm2-logrotate` 已安装并写入上述配置
- 在重启进程前检查当前活动日志
- 如果当前文件已经超过上限，则直接截断，立即释放磁盘

## Scope

本次只修改宿主机部署链路，不调整 Docker 部署路径。

涉及文件：

- `deploy/ecosystem.host.example.cjs`
- `scripts/deploy.sh`
- `scripts/deploy_lib.sh`
- `scripts/deploy_lib_test.sh`
- `README.md`

## Non-Goals

- 不改业务日志内容本身
- 不引入额外的系统级 `logrotate` 配置
- 不改 Nginx / MySQL / Redis 日志策略

