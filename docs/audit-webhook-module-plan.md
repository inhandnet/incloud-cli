# Audit + Webhook 模块实现计划

> 基于 nezha-support（审计日志）和 nezha-message（Webhook + 系统消息）的 API 调研。

## 模块划分

audit 和 webhook 为两个独立的小模块，另外发现 system message（应用内通知）也适合 CLI。

| # | 模块 | CLI 路径 | 用户可用端点 | 说明 |
|---|------|---------|-------------|------|
| 1 | 审计日志 | `incloud audit` | ~2 | 查询审计日志 |
| 2 | Webhook | `incloud webhook` | ~5 | WeCom Webhook 管理 |
| 3 | 系统消息 | `incloud notification` | ~3 | 应用内通知（可选） |

---

## TODO List

### 审计日志 (`incloud audit`)

- [ ] `audit list` — 查询审计日志（分页，支持 --from/--to/--app/--action/--actor 过滤）
- [ ] `audit delete <id>` — 删除审计日志条目

> 注：审计日志端点使用 `system:auditlog:admin` scope，可能仅系统管理员可用。实现时需验证普通用户是否有权限。

### Webhook (`incloud webhook`)

- [ ] `webhook list` — 列出 Webhook（分页，按 provider 过滤）
- [ ] `webhook create` — 创建 Webhook（名称 + URL，目前仅支持 WeCom/企业微信）
- [ ] `webhook update <id>` — 更新 Webhook URL
- [ ] `webhook delete <id>` — 删除 Webhook
- [ ] `webhook test` — 测试 Webhook（发送测试消息到指定 URL）

### 系统消息 (`incloud notification`)（可选，P4）

- [ ] `notification list` — 查看应用内通知（分页）
- [ ] `notification get <id>` — 查看通知详情
- [ ] `notification read` — 标记通知为已读（支持 --all 或指定 ID 列表）

---

## 不纳入 CLI 的功能

| 功能 | 端点数 | 理由 |
|------|--------|------|
| 审计日志写入 (POST) | 1 | 内部服务间调用，非用户操作 |
| Webhook 批量查询 | 1 | InternalApi（nezha-alert 调用） |
| Stripe 支付回调 | 2 | 外部入站（Stripe→平台），非用户 API |
| 广播消息管理 | 9 | 全部 InternalApi（notify:admin） |

## 备注

- Webhook 目前仅支持企业微信（WeCom），provider 枚举值为 `WECHAT`
- 审计日志通过 RabbitMQ 或 Feign 投递，CLI 只需实现查询端
- 系统消息（notification）优先级较低，可延后到 P4 阶段
