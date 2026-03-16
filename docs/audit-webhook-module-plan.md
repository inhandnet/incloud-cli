# Activity + Webhook 模块实现计划

> 基于 nezha-support（审计日志）和 nezha-message（Webhook + 系统消息）的 API 调研。

## 模块划分

activity 和 webhook 为两个独立的小模块，另外发现 system message（应用内通知）也适合 CLI。

| # | 模块 | CLI 路径 | 用户可用端点 | 说明 |
|---|------|---------|-------------|------|
| 1 | 操作日志 | `incloud activity` | 1 | 查询操作日志 |
| 2 | Webhook | `incloud webhook` | ~5 | WeCom Webhook 管理 |
| 3 | 系统消息 | `incloud notification` | ~3 | 应用内通知（可选） |

---

## API 验证结果

### `GET /api/v1/audit/logs` 真实响应

```json
{
  "result": [
    {
      "_id": "69b78492d264244aed6d942a",
      "app": "nezha",
      "oid": "000000000000000000000000",
      "org": "admin",
      "action": "device_deleted",
      "ipAddress": "127.0.0.1",
      "actor": { "type": "user", "name": "管家", "_id": "5f1e5605cf562757b857a7b9" },
      "entity": { "type": "device", "name": "CLI-Delete-Test", "_id": "69b78491d264244aed6d9428" },
      "details": {},
      "timestamp": "2026-03-16T04:18:26.056Z"
    }
  ],
  "total": 44,
  "page": 0,
  "limit": 3
}
```

**已验证的过滤参数**：`action`、`actor`（传 actor._id）、`app`、`from`/`to`（ISO 8601）、`sort`（如 `timestamp,asc`）

**权限**：admin 用户可正常访问（scope `system:auditlog:admin`，admin 账号默认拥有）。

---

## TODO List

### 操作日志 (`incloud activity`)

- [x] `activity list` — 查询操作日志（分页，支持 `--after/--before/--app/--action/--actor` 过滤，`--fields`，`--sort`，`--count`）

> 注：`--after/--before` 映射到 API 的 `from/to` 参数。delete 端点不暴露给 CLI（审计日志不应被随意删除）。

**默认字段**：`_id, app, action, actor.name, entity.type, entity.name, ipAddress, timestamp`

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
| 审计日志删除 (DELETE) | 1 | 审计日志不应被随意删除，高风险操作 |
| Webhook 批量查询 | 1 | InternalApi（nezha-alert 调用） |
| Stripe 支付回调 | 2 | 外部入站（Stripe→平台），非用户 API |
| 广播消息管理 | 9 | 全部 InternalApi（notify:admin） |

## 备注

- Webhook 目前仅支持企业微信（WeCom），provider 枚举值为 `WECHAT`
- 审计日志通过 RabbitMQ 或 Feign 投递，CLI 只需实现查询端
- 系统消息（notification）优先级较低，可延后到 P4 阶段
