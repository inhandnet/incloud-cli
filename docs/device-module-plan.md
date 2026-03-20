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
- [x] `device create` — 创建设备（SN 验证 → 自动检测产品/MAC/IMEI 要求 → 条件交互提示 → 富错误信息）
- [x] `device update <id>` — 更新设备属性（name/description/tags）
- [x] `device delete <id>` — 删除设备
- ~~`device summary` — 设备统计概览~~ → 迁移至 `overview devices`
- [x] `device export` — 导出设备列表为文件
- [x] `device assign <id> --group <gid>` — 分配设备到指定组
- [x] `device unassign <id>` — 将设备从组中移出
- [x] `device transfer <id> --org <oid>` — 转移设备到其他组织
- [x] `device import <file>` — 批量导入（CSV/XLSX）
- ~~`device properties <id>` — 查看 IoT 属性/状态~~
- ~~`device features <id>` — 查看设备特性标志~~
- [x] `device location <id>` — 查看/更新位置信息

### `incloud device group` — 设备组 (~14 端点)

- [x] `device group list` — 列出设备组
- [x] `device group get <id>` — 查看设备组详情
- [x] `device group create` — 创建设备组
- [x] `device group update <id>` — 更新设备组
- [x] `device group delete <id>` — 删除设备组
- [x] `device group list --summary` — 组内设备统计（合并到 list 命令）
- ~~`device group candidates <id>` — 可加入该组的设备~~ — 不实现
- ~~`device group export` — 导出设备组~~ — 不实现
- [x] `device group firmwares <id>` — 组内固件版本分布

## Phase 1b: 诊断工具 + 远程方法

### `incloud device exec` — 诊断工具（已合并到 exec）

- [x] `device exec ping <id> --host <target>` — Ping 诊断
- [x] `device exec traceroute <id> --host <target>` — Traceroute
- [x] `device exec speedtest <id>` — 测速
- [x] `device exec speedtest-history <id>` — 测速历史记录
- [x] `device exec capture <id> --interface <iface>` — 抓包 (tcpdump)
- [x] `device exec capture-status <id>` — 查看抓包状态
- [x] `device exec flowscan <id>` — 流量扫描
- [x] `device exec flowscan-status <id>` — 流量扫描状态
- [x] `device exec cancel <diagId>` — 取消诊断任务
- [x] `device exec interfaces <id>` — 列出可用网络接口

### `incloud device exec` — 远程方法 (~3 端点)

- [x] `device exec <id> reboot` — 重启设备
- [x] `device exec <id> restore-defaults` — 恢复出厂设置
- [x] `device exec <id> <method> [--payload '{}']` — 调用自定义方法
- [x] `device exec --bulk <ids> <method>` — 批量执行方法

## Phase 1c: 配置管理 + 日志

### `incloud device config` — 配置管理 (~12 端点)

- [x] `device config get <id>` — 默认查看合并配置，`--layer` 切换分层视图
- [x] `device config update <id>` — 直接更新配置（`--payload`/`--file`，JSON merge patch）
- [x] `device config error <id>` — 查看配置下发错误（table 模式只显示 error 列表）
- [x] `device config abort <id>` — 中止待同步的配置变更（`--yes` 跳过确认）
- [x] `device config copy` — 配置复制（`--source/--source-group` → `--to/--to-group`）
- [x] `device config snapshots list <id>` — 配置快照列表（`--page/--limit/--after/--before`）
- [x] `device config snapshots get <id> <snapId>` — 查看快照详情
- [x] `device config snapshots restore <id> <snapId>` — 从快照恢复配置（`--yes` 跳过确认）

### `incloud device log` — 日志 (~4 端点)

> 后端服务：nezha-device-manager
> 基础 API：`/api/v1/devices/{id}/logs/download`

- [x] `device log syslog <id>` — 查询历史 syslog（`GET /api/v1/devices/{id}/logs/download/syslog/history`，`--after/--before/--keywords/--limit`，后续可加 `--realtime` 支持实时采集）
- [x] `device log diagnostic <id>` — 下载设备诊断日志（`GET /api/v1/devices/{id}/logs/download?type=diagnostic`，`--file`）
- [x] `device log mqtt <id>` — 查看 MQTT 通信日志

## Phase 1d: 监控 + 在线状态（扁平结构）

> 去掉 `device monitor` 中间层，监控命令直接挂在 `device` 下。
> cobra command group: "Monitoring" 分组展示。

### 信号监控（后端：nezha-link）

- [x] `device signal <id>` — 信号强度时序（`GET /api/v1/devices/{id}/signal`，RSRP/RSRQ/SINR，`--after/--before`）
- [ ] ~~`device signal current <id>` — 当前实时信号（`GET /api/v1/devices/{id}/current-signal`）~~ — 返回字段过少（仅 dBm/asu/level），价值有限，暂不实现
- [x] `device signal export <id>` — 导出信号数据（`GET /api/v1/devices/{id}/signal/export`）
- [ ] ~~`device signal summary --device <ids>` — 批量信号汇总（`POST /api/v1/signal-summary/batch`，`--after/--before`）~~ — 前端未使用，暂不实现

### 天线信号（后端：nezha-link）

- [x] `device antenna <id>` — 天线信号数据（`GET /api/v1/devices/{id}/antenna-signal`，多天线+GPS 关联，`--after/--before` 必填）
- ~~`device antenna export <id>` — 导出天线信号（`GET /api/v1/devices/{id}/antenna-signal/export`）~~ — 不实现

### 性能监控（后端：nezha-link）

- [x] `device perf <id>` — 当前性能快照（`GET /performances`），`--refresh` 实时采集（`POST .../refresh`），`--after/--before` 历史时序（`GET /performance`）
- ~~`device perf export <id>`~~ — 不实现

### 网络接口（后端：nezha-link）

- [x] `device interface <id>` — 网络接口状态（`GET /api/v1/devices/{id}/interfaces`，含蜂窝接口信号）
- [x] `device interface <id> --refresh` — 实时刷新（`POST /api/v1/devices/{id}/interfaces/refresh`），与 `device perf --refresh` 风格一致

### Uplink 链路（后端：nezha-link）

- [x] `device uplink <id>` — 设备 Uplink 列表（`GET /api/v1/devices/{id}/uplinks`）
- ~~`device uplink list` — 全局 Uplink~~ — 不属于 device 模块，移至独立 uplink 模块
- [x] `device uplink get <uplinkId>` — Uplink 详情（`GET /api/v1/uplinks/{id}`）
- [x] `device uplink perf <id> --name <name>` — Uplink 性能趋势（`GET /api/v1/devices/{id}/uplinks/perf-trend`，`--after/--before`）
- ~~`device uplink status` — Uplink 状态统计~~ — 全局接口，不属于 device 模块，移至独立 uplink 模块

### 在线状态（后端：nezha-iot）

- [x] `device online <id>` — 上下线事件历史（`GET /api/v1/devices/{id}/online-events-list`，分页，`--after/--before`），`--daily` 每日离线统计（`GET /api/v1/devices/{id}/offline/daily-stats`）
- ~~`device presence stats <id>` — 在线事件统计图表（`GET /api/v1/devices/{id}/online-events-chart/statistics`）~~ — 前端专用接口，不实现
- ~~`device presence offline topn` — 离线最多的设备排名~~ → 迁移至 `overview offline`
- ~~`device presence offline stats` — 离线统计列表~~ → 迁移至 `overview offline`
- ~~`device online export <id>` — 导出事件历史（`GET /api/v1/devices/{id}/online-events-list/export`）~~ — 暂不实现

## Phase 1e: 流量统计 + 影子

### 流量统计（后端：nezha-link）

- [x] `device datausage hourly <id>` — 小时流量（`GET /api/v1/devices/{id}/datausage-hourly`，`--type cellular|wifi|all --after/--before`）
- [x] `device datausage daily <id>` — 日流量（`GET /api/v1/devices/{id}/datausage-daily`，`--type --month`）
- [x] `device datausage monthly <id>` — 月流量（`GET /api/v1/devices/{id}/datausage-monthly`，`--type --year`）
- ~~`device datausage overview <id>` — 流量概览（聚合 hourly/daily/monthly overview 端点）~~ — 暂不实现
- ~~`device datausage topk` — Top-K 流量排名~~ → 迁移至 `overview traffic`
- [x] `device datausage list` — 设备流量明细（`GET /api/v1/devices/datausage/details`）
- ~~`device datausage export <id>` — 导出流量（`GET /api/v1/devices/{id}/datausage/export`，`--type HOURLY|DAILY|MONTHLY`）~~ — 暂不实现

### 影子文档

- [x] `device shadow list <id>` — 列出影子文档名
- [x] `device shadow get <id> --name <n>` — 获取指定影子文档
- [x] `device shadow update <id> --name <n>` — 更新影子期望状态
- [x] `device shadow delete <id> --name <n>` — 删除影子文档

## Phase 2: 聚合排查命令

### `incloud device inspect <id>` — 一站式设备排查

> 聚合多个数据源，一个命令输出设备当前状态全貌，类似 `docker inspect`。
> 适用场景：设备出问题时快速了解全貌，无需逐个跑子命令。

输出包含（默认最近 24 小时）：
- **基本信息**：名称、SN、产品、固件版本、所属组
- **在线状态**：当前是否在线、最近上下线事件（取 `connections` 最近 5 条）
- **信号快照**：当前信号值（`signal current`）
- **活跃告警**：未确认告警列表（`alert list --device <id> --ack false`）
- **网络接口**：接口状态摘要（`interface`）

参数：
- `--after / --before` — 自定义时间范围
- `--section signal,connections,alerts` — 只输出指定部分
- `-o json/yaml/table` — 输出格式

---

## 基础设施（在实现子命令前完成）

- [x] 设计 device 子命令的代码结构（`internal/cmd/device/` 目录布局）
- [x] 实现通用分页参数（`--page`, `--limit`, `--sort`）
- [ ] 实现通用过滤参数（`--filter`, `-q`）
- [ ] 实现 `--device <id>` 全局参数，支持 ID / name / SN 查找
- [x] 表格输出：device 列表的默认列选择
- [x] 异步诊断任务的实时流机制：ping/traceroute 通过 SSE 流实时输出结果
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
| Stream（实时数据流 SSE） | 内部 API，ping/traceroute 已通过 SSE 流实现实时输出 | 不单独暴露 |
