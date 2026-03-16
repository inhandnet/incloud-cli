# incloud-cli 模块规划

> 基于 Nezha API 文档（796 个端点，18 个功能域）分析，规划 CLI 模块。

## 当前已有模块

| 模块 | 功能 |
|------|------|
| `auth` | 登录/登出/状态查看 |
| `config` | Context 管理 |
| `api` | 通用 API 调用（支持 JSON/Table/YAML 输出） |
| `version` | 版本信息 |

## 建议实现的模块

### P0 - 核心功能

#### `device` — 设备管理（158 端点）

devices.md 包含多个子域，CLI 用子命令分组：

```
incloud device list/get/create/update/delete   # 核心 CRUD
incloud device diag ping/traceroute/speedtest   # 诊断
incloud device group list/get/create/delete     # 设备组
incloud device config get/history/restore       # 配置
incloud device log syslog/mqtt                  # 日志
incloud device status                           # 在线/离线/性能/信号
```

子域拆分明细：

| 子域 | 端点数 | 说明 |
|------|--------|------|
| 设备 CRUD + 批量操作 | ~30 | 核心设备管理 |
| 设备诊断 (diagnosis) | ~10 | ping/traceroute/speedtest/capture/flowscan |
| 设备组 (devicegroups) | ~17 | 独立的分组管理 |
| 设备配置 (config) | ~8 | 配置查看/历史/恢复 |
| 设备流量 (datausage) | ~15 | 按日/时/月的流量统计 |
| 设备性能/信号/接口 | ~15 | 监控数据 |
| 设备任务 (jobs/ota) | ~8 | 任务执行和 OTA |
| 设备日志 (logs/mqtt) | ~8 | syslog/mqtt/cloudwatch |
| 设备影子 (shadow) | ~4 | Device Shadow |
| 客户端识别规则 | ~10 | 内部管理 API |
| 许可证/序列号等 | ~10 | 与 billing 交叉 |

### P1 - 管理基础

#### `user` — 用户管理（73 端点）

```
incloud user list/get/create/delete
incloud user invite
incloud user role list/assign/remove
incloud user me
incloud user lock/unlock
```

#### `org` — 组织与客户管理（60 端点）

```
incloud org list/get/create/delete
incloud org self
incloud org customer list/get/invite
incloud org contact list/create
```

#### `firmware` — 固件与 OTA（46 端点）

```
incloud firmware list/get/upload/publish/deprecate
incloud firmware job list/create
incloud firmware download
```

### P2 - 运维关键

#### `network` — 网络管理（92 端点）

92 个端点横跨多个子域：AutoVPN (~14)、InConnect 连接器 (~18)、OOBM (~15)、网络资产 (~10)、客户端 (~10)、远程访问 (~8)。

```
incloud network vpn list/get/create/delete
incloud network connector list/get/create
incloud network oobm list/session
incloud network asset list
```

#### `alert` — 告警管理（29 端点）

```
incloud alert policy list/get/create/enable/disable
incloud alert rule list/get/create
incloud alert top-devices/top-types
```

#### `product` — 产品管理（64 端点）

```
incloud product list/get/create
incloud product type list
incloud product compatibility list/validate
```

### P3 - 业务支撑

#### `billing` — 计费与许可（80 端点，~50% 内部 API）

```
incloud billing license list/get/attach/detach/upgrade
incloud billing order list/get/create
incloud billing license-type list
```

#### `audit` — 审计日志（support.md 拆出）

```
incloud audit log list
incloud audit log export
```

#### `webhook` — Webhook 管理（message.md 拆出）

```
incloud webhook list/get/create/update/delete
incloud webhook test
```

## 不建议实现的模块

| API 域 | 端点数 | 不做理由 |
|--------|--------|----------|
| **iot** | 7 | 几乎全是内部 API 或移动端专用 |
| **view** | 14 | 数据可视化项目管理，GUI 操作属性强，CLI 价值低 |
| **public** | 11 | 纯内部静态资源管理 |
| **link** | 13 | 独立价值低，地理编码功能偏 GUI，uplinks 已归入 device |
| **stats** | 15 | 大部分流量统计已在 device datausage 下，剩余组织级概览可用 `api` 命令 |
| **auth 扩展** | 37 | MFA/SSO/Passkey 是交互式浏览器流程，CLI 不适合；现有 auth 模块已足够 |
| **message（邮件/短信）** | ~20 | 内部 API 占 80%，仅 webhook 部分对用户有用（已拆出） |

## 实施路径

```
Phase 1 → device（含 diag/group 子命令）
Phase 2 → user + org
Phase 3 → firmware + network + alert
Phase 4 → product + billing + audit + webhook
```
