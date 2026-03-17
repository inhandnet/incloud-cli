# Device 模块实现计划

> 基于后端 API 调研，device 模块拆分为子模块。
> 后端服务：nezha-link（监控/信号/uplink/接口/流量）、nezha-iot（在线状态）、nezha-device-manager（日志/诊断/远程方法）。

## 设计原则

### 命令层级：扁平优先

去掉无动作的中间命名空间（如 `device monitor`），监控类命令直接作为 `device` 的子命令。
层级越浅越好用（对比 `kubectl get pods` vs `kubectl resource get pods`）。
子命令多时用 cobra command groups 做分组展示，而非加嵌套层。

### 时间参数：全局统一 `--after / --before`

所有涉及时间范围的命令统一使用 `--after` / `--before`（与后端 API 参数一致）。
已实现的 `alert list --from/--to` 后续迁移为 `--after/--before`（保留 `--from/--to` 为隐藏别名）。

### signal vs antenna 区分

- `device signal`：Modem 级信号（RSRP/RSRQ/SINR），来自 `/api/v1/devices/{id}/signal`
- `device antenna`：天线级信号（多天线、GPS 关联），来自 `/api/v1/devices/{id}/antenna-signal`

两者后端 API 独立，查询参数不同，保留为两个命令。

---

## Phase 1a: 核心 CRUD + 设备组

### `incloud device` — 核心 CRUD (~25 端点)

- [x] `device list` — 设备列表（分页、过滤、排序）
- [x] `device get <id>` — 查看设备详情
- [x] `device create` — 创建设备（交互式输入 SN/凭证）
- [x] `device update <id>` — 更新设备属性（name/description/tags）
- [x] `device delete <id>` — 删除设备
- ~~`device summary` — 设备统计概览~~ → 迁移至 `overview devices`
- [x] `device export` — 导出设备列表为文件
- [x] `device assign <id> --group <gid>` — 分配设备到指定组
- [x] `device unassign <id>` — 将设备从组中移出
- [x] `device transfer <id> --org <oid>` — 转移设备到其他组织
- [x] `device import <file>` — 批量导入（CSV/XLSX）
- ~~`device properties <id>` — 查看 IoT 属性/状态~~
- [ ] `device features <id>` — 查看设备特性标志
- [x] `device location <id>` — 查看/更新位置信息

### `incloud device group` — 设备组 (~14 端点)

- [ ] `device group list` — 列出设备组
- [ ] `device group get <id>` — 查看设备组详情
- [ ] `device group create` — 创建设备组
- [ ] `device group update <id>` — 更新设备组
- [ ] `device group delete <id>` — 删除设备组
- [ ] `device group summary <id>` — 组内设备统计
- [ ] `device group candidates <id>` — 可加入该组的设备
- [ ] `device group export` — 导出设备组
- [ ] `device group firmware-versions <id>` — 组内固件版本分布

## Phase 1b: 诊断工具 + 远程方法

### `incloud device diag` — 诊断工具 (~13 端点)

- [ ] `device diag ping <id> --host <target>` — Ping 诊断
- [ ] `device diag traceroute <id> --host <target>` — Traceroute
- [ ] `device diag speedtest <id>` — 测速
- [ ] `device diag speedtest history <id>` — 测速历史记录
- [ ] `device diag capture <id> --interface <iface>` — 抓包 (tcpdump)
- [ ] `device diag capture status <id>` — 查看抓包状态
- [ ] `device diag flowscan <id>` — 流量扫描
- [ ] `device diag flowscan status <id>` — 流量扫描状态
- [ ] `device diag cancel <diagId>` — 取消诊断任务
- [ ] `device diag interfaces <id>` — 列出可用网络接口

### `incloud device exec` — 远程方法 (~3 端点)

- [ ] `device exec <id> reboot` — 重启设备
- [ ] `device exec <id> restore-defaults` — 恢复出厂设置
- [ ] `device exec <id> <method> [--payload '{}']` — 调用自定义方法
- [ ] `device exec --bulk <ids> <method>` — 批量执行方法

## Phase 1c: 配置管理 + 日志

### `incloud device config` — 配置管理 (~12 端点)

- [ ] `device config get <id>` — 查看当前设备配置
- [ ] `device config merge <id>` — 查看完整合并配置（设备+组+默认层）
- [ ] `device config edit <id>` — 直接更新配置（一步提交）
- [ ] `device config history <id>` — 配置变更历史
- [ ] `device config snapshot <id> <snapId>` — 查看历史快照
- [ ] `device config restore <id> <snapId>` — 从快照恢复配置
- [ ] `device config discard <id>` — 丢弃待提交的变更
- [ ] `device config error <id>` — 查看最近配置下发错误
- [ ] `device config copy --from <src> --to <dst>` — 配置复制到其他设备/组

### `incloud device log` — 日志 (~4 端点)

> 后端服务：nezha-device-manager
> 基础 API：`/api/v1/devices/{id}/logs/download`

- [x] `device log syslog <id>` — 查询历史 syslog（`GET /api/v1/devices/{id}/logs/download/syslog/history`，`--after/--before/--keywords/--limit`，后续可加 `--realtime` 支持实时采集）
- [ ] `device log download <id>` — 流式下载设备日志（`GET /api/v1/devices/{id}/logs/download`，`--type DIAGNOSTIC|SYSLOG`）
- [ ] `device log mqtt <id>` — 查看 MQTT 通信日志

## Phase 1d: 监控 + 在线状态（扁平结构）

> 去掉 `device monitor` 中间层，监控命令直接挂在 `device` 下。
> cobra command group: "Monitoring" 分组展示。

### 信号监控（后端：nezha-link）

- [x] `device signal <id>` — 信号强度时序（`GET /api/v1/devices/{id}/signal`，RSRP/RSRQ/SINR，`--after/--before`）
- [ ] `device signal current <id>` — 当前实时信号（`GET /api/v1/devices/{id}/current-signal`）
- [ ] `device signal export <id>` — 导出信号数据（`GET /api/v1/devices/{id}/signal/export`）
- [ ] `device signal summary --device <ids>` — 批量信号汇总（`POST /api/v1/signal-summary/batch`，`--after/--before`）

### 天线信号（后端：nezha-link）

- [ ] `device antenna <id>` — 天线信号数据（`GET /api/v1/devices/{id}/antenna-signal`，多天线+GPS 关联，`--after/--before` 必填）
- [ ] `device antenna export <id>` — 导出天线信号（`GET /api/v1/devices/{id}/antenna-signal/export`）

### 性能监控（后端：nezha-link）

- [ ] `device perf <id>` — 性能时序（`GET /api/v1/devices/{id}/performance`，CPU/内存，`--after/--before` 必填）
- [ ] `device perf current <id>` — 当前性能（`GET /api/v1/devices/{id}/performances`）
- [ ] `device perf refresh <id>` — 实时采集（`POST /api/v1/devices/{id}/performances/refresh`）
- [ ] `device perf export <id>` — 导出性能数据（`GET /api/v1/devices/{id}/performance/export`）

### 网络接口（后端：nezha-link）

- [x] `device interface <id>` — 网络接口状态（`GET /api/v1/devices/{id}/interfaces`，含蜂窝接口信号）
- [ ] `device interface refresh <id>` — 实时刷新（`POST /api/v1/devices/{id}/interfaces/refresh`）

### Uplink 链路（后端：nezha-link）

- [ ] `device uplink <id>` — 设备 Uplink 列表（`GET /api/v1/devices/{id}/uplinks`）
- [ ] `device uplink list` — 全局 Uplink（`GET /api/v1/uplinks`，分页，`--device/--status/--type`）
- [ ] `device uplink get <uplinkId>` — Uplink 详情（`GET /api/v1/uplinks/{id}`）
- [ ] `device uplink perf <id> --name <name>` — Uplink 性能趋势（`GET /api/v1/devices/{id}/uplinks/perf-trend`，`--after/--before`）
- [ ] `device uplink status` — Uplink 状态统计（`GET /api/v1/uplinks/status`）

### 在线状态（后端：nezha-iot）

- [x] `device presence events <id>` — 上下线事件历史（`GET /api/v1/devices/{id}/online-events-list`，分页，`--after/--before`）
- [ ] `device presence stats <id>` — 在线事件统计图表（`GET /api/v1/devices/{id}/online-events-chart/statistics`）
- [ ] `device presence offline daily <id>` — 每日离线汇总（`GET /api/v1/devices/{id}/offline/daily`，`--after/--before`）
- ~~`device presence offline topn` — 离线最多的设备排名~~ → 迁移至 `overview offline`
- ~~`device presence offline stats` — 离线统计列表~~ → 迁移至 `overview offline`
- [ ] `device presence export <id>` — 导出事件历史（`GET /api/v1/devices/{id}/online-events-list/export`）

## Phase 1e: 流量统计 + 影子

### 流量统计（后端：nezha-link）

- [ ] `device datausage hourly <id>` — 小时流量（`GET /api/v1/devices/{id}/datausage-hourly`，`--type cellular|wifi|all --after/--before`）
- [ ] `device datausage daily <id>` — 日流量（`GET /api/v1/devices/{id}/datausage-daily`，`--type --month`）
- [ ] `device datausage monthly <id>` — 月流量（`GET /api/v1/devices/{id}/datausage-monthly`，`--type --year`）
- [ ] `device datausage overview <id>` — 流量概览（聚合 hourly/daily/monthly overview 端点）
- ~~`device datausage topk` — Top-K 流量排名~~ → 迁移至 `overview traffic`
- [ ] `device datausage details` — 设备流量明细（`GET /api/v1/devices/datausage/details`）
- [ ] `device datausage export <id>` — 导出流量（`GET /api/v1/devices/{id}/datausage/export`，`--type HOURLY|DAILY|MONTHLY`）

### 影子文档

- [ ] `device shadow list <id>` — 列出影子文档名
- [ ] `device shadow get <id> --name <n>` — 获取指定影子文档
- [ ] `device shadow update <id> --name <n>` — 更新影子期望状态
- [ ] `device shadow delete <id> --name <n>` — 删除影子文档

## Phase 2: 聚合排查命令

### `incloud device inspect <id>` — 一站式设备排查

> 聚合多个数据源，一个命令输出设备当前状态全貌，类似 `docker inspect`。
> 适用场景：设备出问题时快速了解全貌，无需逐个跑子命令。

输出包含（默认最近 24 小时）：
- **基本信息**：名称、SN、产品、固件版本、所属组
- **在线状态**：当前是否在线、最近上下线事件（取 `presence events` 最近 5 条）
- **信号快照**：当前信号值（`signal current`）
- **活跃告警**：未确认告警列表（`alert list --device <id> --ack false`）
- **网络接口**：接口状态摘要（`interface`）

参数：
- `--after / --before` — 自定义时间范围
- `--section signal,presence,alerts` — 只输出指定部分
- `-o json/yaml/table` — 输出格式

---

## 基础设施（在实现子命令前完成）

- [x] 设计 device 子命令的代码结构（`internal/cmd/device/` 目录布局）
- [x] 实现通用分页参数（`--page`, `--limit`, `--sort`）
- [ ] 实现通用过滤参数（`--filter`, `-q`）
- [ ] 实现 `--device <id>` 全局参数，支持 ID / name / SN 查找
- [x] 表格输出：device 列表的默认列选择
- [ ] 考虑异步诊断任务的轮询/等待机制（diag/exec 类命令）
- [ ] cobra command groups 分组展示（Core / Diagnostics / Monitoring / Data）
- [x] 统一时间参数 `--after/--before`，提取到公共 flag helper

## 不纳入 device 模块

| 功能域 | 理由 | 归属 |
|--------|------|------|
| Firmware CRUD + OTA Jobs | 独立业务域 ~40+ 端点 | `firmware` 模块（P1） |
| Generic Jobs | 跨域（配置下发/OTA/CLI 配置） | `job` 模块或归入 `firmware` |
| Edge/Live（容器部署） | 独立边缘计算域 | `live` 顶级命令 |
| Touch（远程访问） | 独立远程连接域 | `touch` 顶级命令 |
| Client Identification Rules | 内部管理 API | 不实现 |
| Config Documents/Schema | 内部管理 API | 不实现 |
| Stream（实时数据流 SSE） | 内部 API，前端 WebSocket 用 | 不直接暴露 |
