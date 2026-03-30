# v0.4.1 (2026-03-30)

## 新功能

### 反馈管理
- **`feedback resolve`** — 更新反馈工单的解决状态

### 设备数据
- **`device datausage --interval`** — 设备流量统计支持 `--interval` 参数指定统计间隔
- **`device list` 过滤增强** — 补全设备列表缺失的过滤参数，与其他列表命令保持一致

### 查询增强
- **`--expand` 参数** — 设备列表、设备组列表等多个命令新增 `--expand`，支持展开关联字段
- **`--sort` 参数** — 多个列表命令新增 `--sort` 排序支持
- **`device group list --org`** — 设备组列表支持按组织过滤
- **离线统计过滤** — `overview offline` 命令支持按条件过滤统计结果

### 超级管理员
- **`--sudo` 参数可见性** — 超级管理员在所有子命令的帮助信息中均可看到 `--sudo` 参数

## 修复
- 修复 `feedback list` 默认输出不显示回复内容的问题
- 修复流量统计表格输出与 JSON 输出字段不一致的问题
- 修复 `--expand` 参数值与后端 API 不匹配的问题

## 调整
- 移除 `activity`、`alert`、`feedback` 列表命令的 `--count` 参数

---

# v0.4.0 (2026-03-27)

## 新功能

### 远程隧道
- **`tunnel` 命令组** — 管理设备远程访问隧道（list/get/create/delete），查看隧道连接状态和详情
- **`tunnel forward`** — 通过隧道将设备端口转发到本地，基于 smux 多路复用实现高效 TCP 转发
- **`tunnel exec`** — 通过隧道在设备上执行 CLI 命令，无需 SSH 直连设备

### Webhook 管理
- **`webhook` 命令组** — 消息 Webhook 全生命周期管理（list/get/create/update/delete）
- **`webhook test`** — 测试 Webhook 连通性，支持按 ID 测试和通用 provider 测试

### 设备资产
- **`device asset` 命令组** — 设备资产 CRUD 管理（list/get/create/update/delete）
- **`client mark-asset`** — 将客户端标记为资产

### 客户端管理
- **`client set-pos-ready`** — 设置客户端 POS 就绪状态

### 数据概览
- **`overview trend`** — 查询每日设备在线数/总数趋势

### 查询增强
- **`--order` 排序参数** — `device log mqtt` 和 `signal list` 支持 `--order` 指定排序方式
- **`--timeout` 超时参数** — `device log syslog --fetch` 支持自定义超时时间

### 其他
- **User-Agent 请求头** — 所有 CLI 请求自动携带 User-Agent，便于服务端识别 CLI 流量
- **Windows 安装文档** — INSTALL.md 新增 Windows 安装步骤

## 修复
- 修复 `api` 命令错误信息重复输出的问题
- 修复 JSON/YAML/JQ 输出模式下 0-based 页码未转换为 1-based 的问题
- 修复分页查询参数误用 `size` 而非 `limit` 的问题
- 修复 syslog 时间戳参数重复追加 `Z` 后缀的问题
- 修复时间戳参数未正确归一化的问题

## 调整
- `open-web` 移除 `--forward` 参数，端口转发功能统一到 `tunnel forward` 命令

---

# v0.3.0 (2026-03-26)

## 新功能

### 自更新
- **`incloud update` 命令** — 从 GitHub Releases 检查并安装新版本，支持 `--check` 仅检查、`--version` 指定版本、`--yes` 跳过确认。GitHub 不可达时自动回退到 S3 国内镜像源

### 反馈管理
- **`feedback create`** — 提交反馈工单，支持 `--file` 上传附件
- **`feedback list`** — 查看反馈列表，显示附件信息
- **`feedback download`** — 下载反馈附件

### 认证优化
- **零配置登录** — `incloud login` 无需任何参数即可登录（默认 global 区域、default context）
- **区域简写** — `--host` 支持 `global`、`cn`、`dev`、`beta` 等区域名称，无需输入完整域名
- **顶层 login 别名** — `incloud login` 作为 `incloud auth login` 的快捷方式
- **401 自动提示** — 收到 401 错误时提示重新登录
- **移除本地存储凭证** — OAuth client credentials 不再保存到配置文件，改为动态获取，减少磁盘上的敏感数据

### 设备管理
- **批量导入增强** — `device import` 新增 `--group` 和 `--org` 参数，导入时直接指定分组和子组织
- **`device import-status` 命令** — 查询导入任务状态，失败时显示逐行错误详情（序列号、失败原因）
- **诊断日志自动解密** — `device log diagnostic` 下载时自动检测 AES 加密并解密，直接输出 .tar.gz 文件

### 用户与组织
- **`user identity list`** — 查看当前用户在所有可访问组织中的身份角色，支持按组织名筛选
- **`--tenant` 全局参数** — 按请求切换组织上下文，多组织用户无需管理员权限即可操作外部组织

### 架构改进
- **API/Auth URL 分离** — 主机配置拆分为 API 地址（star.*）和认证地址（portal.*），支持 IP 地址直连

## 修复
- `alert rule delete` 和 `user unlock` 增加 `--yes` 确认提示，防止误操作
- `device log syslog` 始终输出纯文本行，忽略 `-o json` 参数，保持 grep 友好

---

# v0.2.0 (2026-03-24)

## Features

### Query & Output
- **`--jq` global filter** — Apply jq expressions to filter JSON/YAML output from any command (powered by gojq). Single-key `{"result": [...]}` responses are auto-unwrapped. String results output as plain text (no JSON quotes); other types output as compact JSON. Usage: `incloud device list --jq '.[].name'`
- **Smart output format** — Auto-detect TTY to choose output format: table by default in terminal, JSON when piped or redirected. `--output` flag overrides. Time-series commands also default to table output

### Device Management
- **Config schema commands** — New `device config schema list/get/overview/validate` subcommands. Query by device ID (`--device`) or product model + firmware version (`--product` + `--version`). `list` supports `--name` regex filter; `validate` accepts payload via `--payload` or `--file` for JSON Schema pre-validation; auto-suggests available firmware versions when schema is not found
- **Enhanced device creation** — SN pre-validation (auto-checks serial number validity and required verification fields before creation), conditional MAC/IMEI prompts (interactive in TTY, error with flag hints in pipe mode), rich error messages (duplicate name, duplicate SN, MAC/IMEI mismatch all have clear diagnostics and fix suggestions)
- **Synchronous packet capture** — `device exec capture` changed from async to sync mode, blocking until completion. `--download <file>` auto-downloads pcap file on completion, Ctrl+C cancels device-side task, failed downloads auto-clean incomplete files
- **Interactive speed test** — `device exec speedtest` redesigned with interactive guidance: auto-fetches available interface list for selection, retrieves matching speed test nodes per interface, streams progress (TTY real-time refresh, non-TTY final result only). New `speedtest-config` subcommand to view available options
- **Syslog live upload** — `device log syslog --fetch` triggers device to upload current log buffer (waits up to 40s). Without `--fetch`, queries existing platform logs (`--after`/`--before` required); with `--fetch`, time range is optional, defaults to today

### Alerts
- **Rule type parameters** — `alert rule create/update` supports type-specific parameters (offline retention duration, CPU threshold, signal strength limit, etc.). `--type` accepts plain type name, comma-separated parameters, or JSON format, and can be repeated for multiple types
- **Type discovery command** — New `alert rule types` lists all 26 alert types with parameter descriptions, `alert rule types <type>` shows single type details

### Development & Debugging
- **Debug mode** — Enable with `--debug` flag or `INCLOUD_DEBUG=1` env var. Outputs HTTP request/response headers and status codes, request body (truncated at 4KB), response timing, token refresh events, config context source. Authorization is always redacted to `****`
- **User impersonation (Sudo)** — Super admins can impersonate any user. Hidden `--sudo <user>` flag or `INCLOUD_SUDO=<user>` env var. Only available to super admins; non-admin calls are silently ignored by the backend. Sudo header is only injected for same-origin requests to prevent credential leakage

## Improvements
- Localized timestamps: ISO 8601 times in table output are auto-converted to local time
- Renamed flags to improve AI discoverability (`--to` → `--target`, `--out` → `--output-file`), old names preserved as hidden aliases
- Required parameters now use `MarkFlagRequired` instead of manual validation, help auto-annotates required flags
- Connector deletion: batch name lookup for confirmation, new typed HTTPError

## Bug Fixes
- Fix `device perf` missing disk and microSD formatters, causing related metrics to not display
- Fix `overview` average/max offline duration showing raw seconds instead of human-readable time
- Fix `device config history list` returning oversized mergedConfig field
- Fix device parsing using wrong `partNumber` field (changed to `product`)
- Fix UTF-8 character truncation in config schema queries
- Fix streaming commands incorrectly warning when `--output` is not explicitly set
- Fix `alert rule --type` help text and examples referencing undiscoverable type names
- Fix Authorization header not being stripped on cross-origin redirects, preventing credential leakage

---

# v0.1.0

First release of InCloud CLI, providing full command-line management for the InHand Cloud platform.

## Features

### Core
- OAuth browser login (PKCE) with automatic token refresh and login status view
- Multi-format output: JSON, YAML, table with built-in human-friendly formatting
- Generic `api` command with query parameters, request body, and custom headers
- Config context management (use/list/set/delete/current)
- Real-time SSE streaming for ping and traceroute diagnostics

### Device Management
- Full device lifecycle: list, get, create, update, delete
- Device groups, configurations, shadows, connected clients
- Signal, interfaces, online/offline events, syslog viewing
- Location management, traffic statistics, performance monitoring, online status
- Antenna info, uplink, remote execution
- Batch import (CSV/XLSX) and export (CSV)
- Assign, unassign, transfer devices

### Firmware
- Firmware listing and details
- Upgrade tasks: create, list, cancel, execution details, retry

### Alerts
- Alert list, details, acknowledge, acknowledge statistics
- Alert rule CRUD, alert export

### Overview Dashboard
- Dashboard, device, alert, traffic, offline summaries
- Top devices and top alert type analysis

### Network & Connectivity
- SD-WAN module: networks, devices, tunnels, connections
- OOBM out-of-band management commands
- Connector usage: statistics, trends, top-K

### Organization
- Organization, user, role management
- Operation audit log queries

### Products
- Product CRUD: list, get, create, update, delete

## Improvements
- Human-friendly formatting for bytes, bitrates, percentages, latency, jitter, duration
- Table exclude-column mode (! prefix), dot-path parsing for nested fields
- TTY table pagination summary header, auto-flattening of nested objects
- Interactive confirmation prompts using charmbracelet/huh

## Bug Fixes
- Fix default output format when `--output` is not specified
- Fix ANSI style nesting in pagination header causing color corruption
- Fix exported file permissions (0o600)
- Fix fallback prompt when browser fails to close after OAuth login
- Fix device import polling and validation status handling
