# User 模块实现计划

> 基于 nezha-core 的 API 调研。所有用户管理端点均在 nezha-core 中。
> MFA/Passkey/密码重置 等交互式浏览器流程不适合 CLI，不纳入。

## 子模块划分

| # | 子模块 | CLI 路径 | 用户可用端点 | 说明 |
|---|--------|---------|-------------|------|
| 1 | 用户 CRUD | `incloud user` | ~14 | 列表、详情、创建、删除、锁定等 |
| 2 | 当前用户 | `incloud user me` | ~4 | 个人资料、改密码、偏好 |
| 3 | 角色管理 | `incloud user role` | ~9 | 角色 CRUD + 权限分配 |
| 4 | 邀请 | `incloud user invite` | ~4 | 邮件邀请、链接邀请 |
| 5 | 登录日志 | `incloud user login-log` | ~3 | 登录历史、导出 |
| 6 | API Token | `incloud user token` | ~4 | API Token 管理 |
| 7 | 会话管理 | `incloud user session` | ~2 | 查看/撤销会话 |

---

## TODO List

### 用户 CRUD (`incloud user`)

- [x] `user list` — 用户列表（分页、过滤）
- [x] `user get <id>` — 查看用户详情
- [x] `user create` — 创建用户（指定角色、发欢迎邮件）
- [x] `user update <id>` — 更新用户信息（名称、角色、组织转移、协作者过期）
- [x] `user delete <id>` — 删除用户
- [ ] `user delete --bulk <ids>` — 批量删除用户
- [x] `user lock <id>` — 锁定用户（禁止登录）
- [x] `user unlock <id>` — 解锁用户
- [ ] `user reset-password <id>` — 管理员重置用户密码
- [ ] `user export` — 导出用户列表
- [ ] `user roles <id>` — 查看用户的角色列表
- [ ] `user permissions <id>` — 查看用户的有效权限
- [ ] `user admin-list --org <oid>` — 列出组织的管理员用户

### 当前用户 (`incloud user me`)

- [x] `user me` — 查看当前用户资料
- [ ] `user me update` — 更新个人资料
- [ ] `user me change-password` — 修改自己的密码
- [ ] `user me locale <locale>` — 更新语言偏好
- [x] `user identity list` — 查看所有组织身份（跨组织角色）
- [ ] `user me preferences` — 查看/更新 UI 偏好
- [ ] `user me settings` — 查看/更新用户设置

### 角色管理 (`incloud user role`)

- [ ] `user role list` — 列出所有角色
- [ ] `user role get <id>` — 查看角色详情
- [ ] `user role create` — 创建自定义角色
- [ ] `user role update <id>` — 更新角色
- [ ] `user role delete <id>` — 删除角色（非内置）
- [ ] `user role users <id>` — 查看角色下的用户
- [ ] `user role permissions <id>` — 查看角色权限
- [ ] `user role grant <id>` — 为角色授予权限
- [ ] `user role revoke <id>` — 移除角色权限
- [ ] `user role assign <userId> <roleId>` — 为用户分配角色
- [ ] `user role remove <userId> <roleId>` — 移除用户角色

### 邀请 (`incloud user invite`)

- [ ] `user invite email <email>` — 邮件邀请已有用户加入组织
- [ ] `user invite resend <userId>` — 重新发送邀请邮件
- [ ] `user invite link create` — 创建/获取邀请链接
- [ ] `user invite link reset` — 重置邀请链接

### 登录日志 (`incloud user login-log`)

- [ ] `user login-log list` — 查看所有用户登录日志
- [ ] `user login-log list --user <id>` — 查看特定用户登录历史
- [ ] `user login-log export` — 导出登录日志

### API Token (`incloud user token`)

- [ ] `user token list` — 列出 API Token
- [ ] `user token create` — 创建 API Token
- [ ] `user token update <id>` — 更新 Token
- [ ] `user token delete <id>` — 删除 Token

### 会话管理 (`incloud user session`)

- [ ] `user session list <userId>` — 查看用户 OAuth2 会话
- [ ] `user session revoke <userId>` — 撤销用户所有会话

---

## 不纳入 CLI 的功能

| 功能 | 端点数 | 理由 |
|------|--------|------|
| MFA/TOTP 注册 | ~11 | 交互式浏览器流程，需扫二维码 |
| Passkey/WebAuthn | ~6 | 需要浏览器 WebAuthn API |
| 手机号绑定/解绑 | ~4 | 需要 SMS 验证码交互 |
| 密码重置（自助） | ~4 | 公开端点，邮件/SMS 流程 |
| 模拟登录 | ~2 | InternalApi，超级管理员专用 |
| 快速登录 | ~2 | 内部服务间调用 |
| 权限定义 CRUD | ~2 | InternalApi |
