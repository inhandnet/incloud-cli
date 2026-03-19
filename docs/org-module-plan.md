# Org 模块实现计划

> 基于 nezha-core 和 nezha-billing 的 API 调研。所有组织管理端点在 nezha-core 中。

## 子模块划分

| # | 子模块 | CLI 路径 | 用户可用端点 | 说明 |
|---|--------|---------|-------------|------|
| 1 | 组织 CRUD | `incloud org` | ~12 | 列表、详情、创建、删除、设置 |
| 2 | 客户管理 | `incloud org customer` | ~5 | 代理商管理下游客户（独立 `customers` 集合，非子组织） |
| 3 | 联系人 | `incloud org contact` | ~4 | 组织联系人管理 |
| 4 | 地址 | `incloud org address` | ~4 | 组织地址管理 |
| 5 | Azure AD SSO | `incloud org sso` | ~4 | SSO 配置 |

---

## TODO List

### 组织 CRUD (`incloud org`)

- [x] `org list` — 列出组织（分页、过滤）
- [x] `org get <id>` — 查看组织详情
- [x] `org self` — 查看当前组织
- [x] `org create` — 创建组织（含管理员用户）
- [x] `org update <id>` — 更新组织信息（名称、邮箱、标签）
- [x] `org update-self` — 更新当前组织
- [x] `org delete <id>` — 删除组织（级联删除角色、邀请、客户等）
- [ ] `org system-setting <id>` — 更新组织系统设置（如许可证告警邮箱）
- [ ] `org bill-address <id>` — 更新组织账单地址
- [ ] `org descendants` — 查找下级组织 ID
- [ ] `org export` — 导出组织列表
- [ ] `org billing-policy <oid>` — 查看组织计费策略

### 客户管理 (`incloud org customer`)

> Customer 是独立实体（`customers` 集合），表示代理商（agent org）与下游组织之间的客户关系。
> Customer._id = 客户组织的 OrgID，Customer.oid = 代理商的 OrgID。
> 字段：_id, name, email, oid, note, subscriptions[], activatedAppIds(transient), createdAt, updatedAt
> 过滤：name(like), email(like), app, subscriptionStatus(ACTIVE/EXPIRED/TRIAL/NONE)

- [ ] `org customer list` — 列出客户（`GET /customers`，分页、按 name/email/subscriptionStatus 过滤）
- [ ] `org customer get <id>` — 查看客户详情（`GET /customers/{id}`）
- [ ] `org customer update <id>` — 更新客户（`PUT /customers/{id}`，可改 name/note/apps 激活）
- [ ] `org customer delete <id>` — 删除客户（`DELETE /customers/{id}`，同时解除 org.partner 关联）
- [ ] `org customer invite` — 邀请新客户（`POST /customers/invite`，参数：name/email/pendingApps）

### 联系人 (`incloud org contact`)

- [ ] `org contact list <oid>` — 列出组织联系人
- [ ] `org contact create <oid>` — 添加联系人
- [ ] `org contact update <oid> <id>` — 更新联系人
- [ ] `org contact delete <oid> <id>` — 删除联系人

### 地址 (`incloud org address`)

- [ ] `org address list <oid>` — 列出组织地址
- [ ] `org address create <oid>` — 添加地址
- [ ] `org address update <oid> <id>` — 更新地址
- [ ] `org address delete <oid> <id>` — 删除地址

### Azure AD SSO (`incloud org sso`)

- [ ] `org sso get <oid>` — 查看 Azure AD SSO 配置
- [ ] `org sso create <oid>` — 创建 SSO 配置
- [ ] `org sso update <oid>` — 更新 SSO 配置
- [ ] `org sso callback-url <oid>` — 获取 SSO 回调 URL

---

## 不纳入 CLI 的功能

| 功能 | 端点数 | 理由 |
|------|--------|------|
| 组织注册（自助） | ~4 | 公开注册流程，浏览器操作 |
| 邮箱验证 | ~2 | 邮件链接流程 |
| 密码重置 | ~4 | 已归入 user 模块 / 浏览器流程 |
| 组织品牌定制 | ~3 | InternalApi |
| Feature Flags | ~5 | InternalApi |
| 计费周期管理 | ~3 | InternalApi |
| 组织可访问性/高级服务 | ~2 | InternalApi |
| 外部应用管理 | ~2 | InternalApi |
| 客户邀请接受 | 1 | `POST /invitation/{id}/accept` 是邮件链接流程，非 CLI 场景 |
| 客户订阅状态汇总 | 1 | `GET /customers/subscription-summary` 聚合统计，CLI 场景价值低 |
| 客户导出 | 1 | `GET /customers/export` 下载文件，CLI 场景价值低 |
