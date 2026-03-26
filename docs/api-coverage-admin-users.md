# API 覆盖分析 - 管理与用户域

> 数据来源：apps/network 和 apps/universal-login（仅限 src/services/ 目录）

## 汇总统计

- Portal 中的 API 总数：12
- CLI 已覆盖：8
- CLI 未覆盖（Gap）：4
- 覆盖率：67%

## 详细对比表

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | `/api/v1/orgs/self` | 获取当前用户所属组织信息 | `incloud org self` | ✅ 已覆盖 |
| GET | `/api/v1/frontend/settings` | 获取前端全局配置 | — | ❌ 未覆盖 |
| POST | `/oauth2/v1/login` | 用户名密码登录（OAuth2） | `incloud auth login` (间接) | — 认证流程 |
| POST | `/api/v1/verifications/send` | 发送短信验证码 | — | ❌ 未覆盖 |
| POST | `/api/v1/forget-password` | 发送忘记密码邮件 | — | ❌ 未覆盖 |
| POST | `/api/v1/reset-password` | 重置密码（凭 token） | — | ❌ 未覆盖 |
| GET | `/api/v1/reset-password/verify-token` | 校验重置密码 token 有效性 | — | ❌ 未覆盖 |
| POST | `/api/v1/register` | 注册新账号 | — | ❌ 未覆盖 |
| POST | `/api/v1/invitation/{id}/register` | 通过邀请链接注册 | — | ❌ 未覆盖 |
| GET | `/api/v1/verify-email/send` | 发送邮箱验证邮件 | — | ❌ 未覆盖 |
| GET | `/api/v1/verify-email/confirm` | 确认邮箱验证 | — | ❌ 未覆盖 |
| GET | `/api/v1/phone/validate` | 校验手机号格式/可用性 | — | ❌ 未覆盖 |

> 注：以下 CLI 命令覆盖的 API 未出现在上述两个 app 的 service 文件中，属于 CLI 额外能力：

| 方法 | API 路径 | CLI 命令 |
|------|---------|---------|
| GET | `/api/v1/users` | `incloud user list` |
| POST | `/api/v1/users` | `incloud user create` |
| GET | `/api/v1/users/{id}` | `incloud user get` |
| PUT | `/api/v1/users/{id}` | `incloud user update` |
| DELETE | `/api/v1/users/{id}` | `incloud user delete` |
| PUT | `/api/v1/users/{id}/lock` | `incloud user lock` |
| PUT | `/api/v1/users/{id}/unlock` | `incloud user unlock` |
| GET | `/api/v1/users/me` | `incloud user me` |
| GET | `/api/v1/user/identities` | `incloud user identity list` |
| GET | `/api/v1/orgs` | `incloud org list` |
| POST | `/api/v1/orgs` | `incloud org create` |
| GET | `/api/v1/orgs/{id}` | `incloud org get` |
| PUT | `/api/v1/orgs/{id}` | `incloud org update` |
| DELETE | `/api/v1/orgs/{id}` | `incloud org delete` |
| PUT | `/api/v1/orgs/self` | `incloud org update-self` |
| GET | `/api/v1/roles` | `incloud role list` |
| GET | `/api/v1/products` | `incloud product list` |
| POST | `/api/v1/products` | `incloud product create` |
| GET | `/api/v1/products/{id}` | `incloud product get` |
| PUT | `/api/v1/products/{id}` | `incloud product update` |
| DELETE | `/api/v1/products/{id}` | `incloud product delete` |
| GET | `/api/v1/audit/logs` | `incloud activity list` |
| GET | `/api/v1/datausage/overview` | `incloud overview traffic` |
| GET | `/api/v1/datausage/topk` | `incloud overview traffic` |

## Gap 分析

### 重要缺口

这两个 app 的 service 文件主要覆盖登录、注册、密码重置等**认证流程 API**，这类操作属于面向最终用户的交互式流程，不适合 CLI 直接支持。但以下条目值得关注：

- **`/api/v1/frontend/settings`**（GET）：获取前端全局配置（如功能开关、UI 主题等），CLI 中无等效命令。若需排查配置问题，需手动调用。
- **`/api/v1/orgs/self`**：Portal 已有调用，CLI 也已有 `incloud org self` 覆盖，对齐良好。

### 次要缺口

- **注册/邀请注册类 API**（`/api/v1/register`、`/api/v1/invitation/{id}/register`）：面向最终用户的自助流程，CLI 不适合覆盖。
- **密码重置类 API**（`/api/v1/forget-password`、`/api/v1/reset-password`、`/api/v1/reset-password/verify-token`）：交互式用户流程，CLI 不适合覆盖。
- **邮箱/手机验证类 API**（`/api/v1/verify-email/*`、`/api/v1/verifications/send`、`/api/v1/phone/validate`）：同上，属于用户自助注册流程的一部分。

### 总结

这两个 app（network、universal-login）的 service 文件整体较精简，主要聚焦认证/注册流程。CLI 在**组织管理、用户管理、角色管理、产品管理、审计日志、流量统计**等管理域有远超 Portal 这两个 app 的覆盖深度。Portal 覆盖而 CLI 缺失的唯一有意义条目是 `/api/v1/frontend/settings`，其余缺口均属交互式用户流程，不在 CLI 的目标场景内。
