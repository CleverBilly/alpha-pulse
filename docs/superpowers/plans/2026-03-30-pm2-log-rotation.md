# PM2 Log Rotation Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为宿主机 PM2 部署加入默认日志轮转和超大日志截断，避免日志文件无限增长。

**Architecture:** 通过 `scripts/deploy_lib.sh` 封装 PM2 日志轮转配置和超大日志修剪逻辑，由 `scripts/deploy.sh` 在每次宿主机部署时统一调用。PM2 模板和 README 同步更新，确保新旧服务器都按同一规则运行。

**Tech Stack:** Bash, PM2, pm2-logrotate, Node.js, README

---

### Task 1: 先写失败测试锁住部署行为

**Files:**
- Modify: `scripts/deploy_lib_test.sh`
- Test: `scripts/deploy_lib_test.sh`

- [ ] **Step 1: 为 PM2 logrotate 和超大日志截断增加失败测试**
- [ ] **Step 2: 运行 `bash scripts/deploy_lib_test.sh`，确认新增断言先失败**

### Task 2: 实现部署库里的日志治理逻辑

**Files:**
- Modify: `scripts/deploy_lib.sh`
- Modify: `scripts/deploy.sh`

- [ ] **Step 1: 在部署库中实现 `pm2-logrotate` 安装/配置函数**
- [ ] **Step 2: 在部署库中实现超大日志截断函数**
- [ ] **Step 3: 在 `scripts/deploy.sh` 中接入这两个步骤**
- [ ] **Step 4: 重新运行 `bash scripts/deploy_lib_test.sh`，确认转绿**

### Task 3: 更新模板和文档

**Files:**
- Modify: `deploy/ecosystem.host.example.cjs`
- Modify: `README.md`

- [ ] **Step 1: 调整 PM2 模板的日志说明和基础日志字段**
- [ ] **Step 2: 在 README 宿主机部署章节写清默认日志轮转策略**

### Task 4: 做完整验证并上线服务器

**Files:**
- Verify: `scripts/deploy_lib.sh`
- Verify: `scripts/deploy.sh`
- Verify: `deploy/ecosystem.host.example.cjs`

- [ ] **Step 1: 运行 `bash scripts/deploy_lib_test.sh`**
- [ ] **Step 2: 运行 `bash -n scripts/deploy.sh scripts/deploy_lib.sh scripts/deploy_lib_test.sh`**
- [ ] **Step 3: 运行 `node -e "const c=require('./deploy/ecosystem.host.example.cjs'); console.log(c.apps.length)"`**
- [ ] **Step 4: 部署到服务器并验证日志目录大小已回落、PM2 轮转配置已生效**
