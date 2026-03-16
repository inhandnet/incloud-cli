# Alert 模块实现计划

> 基于 nezha-alert 的 API 调研。告警创建/关闭/删除为 InternalApi，CLI 主要实现查询和告警规则管理。

## 设计备注

### 时间参数统一

全局统一使用 `--after / --before`（与后端 API 参数名一致）。
`alert list` 当前使用 `--from/--to`，后续迁移为 `--after/--before`（保留 `--from/--to` 为隐藏别名兼容）。

## 子模块划分

| # | 子模块 | CLI 路径 | 用户可用端点 | 说明 |
|---|--------|---------|-------------|------|
| 1 | 告警查询 | `incloud alert` | ~8 | 告警列表、确认、导出、统计 |
| 2 | 告警规则 | `incloud alert rule` | ~6 | 规则 CRUD |
| 3 | 告警统计 | `incloud alert top` | ~2 | Top-K 设备/类型 |

---

## TODO List

### 告警查询 (`incloud alert`)

- [x] `alert list` — 列出告警（分页，支持 --after/--before/--status/--priority/--device/--group/--type 过滤）
- [x] `alert get <id>` — 查看告警详情
- [x] `alert ack <ids...>` — 确认告警（支持多个 ID）
- [x] `alert ack --all` — 确认所有告警
- [x] `alert ack-stats` — 查看未确认告警数量
- [x] `alert export` — 导出告警列表
- [ ] `alert device-status <deviceIds...>` — 批量查看设备告警状态

### 告警规则 (`incloud alert rule`)

- [x] `alert rule list` — 列出告警规则
- [x] `alert rule get <id>` — 查看规则详情
- [x] `alert rule create` — 创建告警规则（指定设备组、规则类型、通知渠道/用户/Webhook/时间窗口）
- [x] `alert rule update <id>` — 更新告警规则（全量替换 rules + notify，不可修改 groupIds）
- [x] `alert rule delete <id...>` — 删除告警规则（多 ID 自动批量删除）

### 告警统计 (`incloud alert top`)

- [x] `alert top devices` — Top-K 告警最多的设备（支持 --after/--before/--group/--n）
- [x] `alert top types` — Top-K 告警类型排名

---

## 支持的告警规则类型

供 `alert rule create` 的 `--type` 参数参考：

| 类型 | 说明 |
|------|------|
| `CONNECTED` / `DISCONNECTED` | 设备上线/离线 |
| `CONFIG_SYNC_FAILED` | 配置同步失败 |
| `SIM_SWITCH` | SIM 卡切换 |
| `LOCAL_CONFIG_UPDATE` | 本地配置更新 |
| `REBOOT` | 设备重启 |
| `FIRMWARE_UPGRADE` | 固件升级 |
| `LICENSE_EXPIRING` / `LICENSE_EXPIRED` | 许可证即将/已过期 |
| `UPLINK_SWITCH` | Uplink 切换 |
| `ETHERNET_WAN_CONNECTED` / `DISCONNECTED` | 以太网 WAN 连接/断开 |
| `MODEM_WAN_CONNECTED` / `DISCONNECTED` | Modem WAN 连接/断开 |
| `WWAN_CONNECTED` / `DISCONNECTED` | WWAN 连接/断开 |
| `CLIENT_CONNECTED` / `DISCONNECTED` | 客户端连接/断开 |
| `CELL_OPERATOR_SWITCH` | 运营商切换 |
| `BRIDGE_LOOP_DETECT` | 网桥环路检测 |
| `CELL_TRAFFIC_REACH_THRESHOLD` | 蜂窝流量达到阈值 |
| `DEVICE_POWER_OFF` | 设备断电 |

## 支持的通知渠道

供 `alert rule create` 的 `--channel` 参数参考：`SMS`、`APP`、`EMAIL`、`SYSTEM`、`WEBHOOK`、`SUBSCRIPTION`

---

## 不纳入 CLI 的功能

| 功能 | 端点数 | 理由 |
|------|--------|------|
| 告警创建/关闭/删除 | ~5 | InternalApi（系统自动触发） |
| 通知策略管理 | ~7 | InternalApi（由规则自动管理） |
| Top-K 活跃告警目标 | 1 | 与 top devices 功能重叠 |
