# v0.2.0 (2026-03-24)

## ✨ 新功能

### 查询与输出
- **`--jq` 全局过滤器** — 对任意命令的 JSON/YAML 输出做 jq 表达式过滤（基于 gojq）。单 key `{"result": [...]}` 响应自动解包，无需手写 `.result[]`。字符串结果输出纯文本（无 JSON 引号），其他类型输出紧凑 JSON。用法：`incloud device list --jq '.[].name'`
- **智能输出格式** — 自动根据 TTY 检测选择输出格式：终端默认 table，管道/重定向自动切换 JSON。`--output` flag 可强制覆盖。时序数据（series）命令同样默认 table 输出

### 设备管理
- **配置 Schema 命令集** — 新增 `device config schema list/get/overview/validate` 四个子命令，支持按设备 ID（`--device`）或产品型号+固件版本（`--product` + `--version`）查询。`list` 支持 `--name` 正则过滤；`validate` 通过 `--payload` 或 `--file` 传入载荷，用 JSON Schema 预校验配置；Schema 未找到时自动提示可用固件版本
- **设备创建增强** — SN 预校验（创建前自动校验序列号有效性和所需验证字段）、条件式 MAC/IMEI 提示（终端下交互输入，管道模式报错并提示所需 flag）、富错误信息（设备名重复、SN 重复、MAC/IMEI 不匹配等场景均有明确诊断和修复建议）
- **抓包同步下载** — `device exec capture` 从异步改为同步模式，阻塞等待完成。`--download <file>` 完成后自动下载 pcap 文件，Ctrl+C 可取消设备端任务，下载失败自动清理不完整文件
- **测速交互式改造** — `device exec speedtest` 改为交互式引导：自动获取可用接口列表供选择，根据接口获取匹配测速节点再选择，流式展示进度（TTY 实时刷新，非 TTY 仅输出最终结果）。新增 `speedtest-config` 子命令查看可用选项
- **Syslog 实时上传** — `device log syslog --fetch` 触发设备主动上传当前日志缓冲区（等待最多 40s）。无 `--fetch` 时查询平台已有日志（`--after`/`--before` 必填）；有 `--fetch` 时时间范围可选，默认今天

### 告警
- **规则类型参数** — `alert rule create/update` 支持按类型传入特定参数（离线保持时长、CPU 阈值、信号强度门限等）。`--type` 支持纯类型名、逗号分隔参数、JSON 三种格式，可重复指定多种类型
- **类型发现命令** — 新增 `alert rule types` 列出全部 26 种告警类型及参数说明，`alert rule types <type>` 查看单个类型详情

### 开发调试
- **调试模式** — `--debug` flag 或 `INCLOUD_DEBUG=1` 环境变量启用。输出 HTTP 请求/响应头和状态码、请求体（截断 4KB）、响应耗时、Token 刷新事件、配置上下文来源，Authorization 始终脱敏为 `****`
- **用户模拟（Sudo）** — 超级管理员可模拟任意用户身份执行操作。`--sudo <user>` 隐藏 flag 或 `INCLOUD_SUDO=<user>` 环境变量启用。仅超管可用，非超管调用被后端静默忽略。Sudo header 仅对同域请求注入，防止凭证泄露

## 🔧 改进
- 时间戳本地化：table 输出中的 ISO 8601 时间自动转为本地时间显示
- Flag 重命名优化 AI 可发现性（`--to` → `--target`、`--out` → `--output-file`），旧名保留为隐藏别名
- 必填参数统一使用 `MarkFlagRequired` 替代手动校验，help 自动标注 required
- Connector 删除优化：批量查找确认名称，新增 HTTPError 类型化错误

## 🐛 修复
- 修复 `device perf` 缺少磁盘和 microSD 格式化器，导致相关指标无法展示
- 修复 `overview` 平均/最大离线时长显示为原始秒数，现已格式化为人类可读时间
- 修复 `device config history list` 返回过大的 mergedConfig 字段
- 修复设备解析误用 `partNumber` 字段（改为 `product`）
- 修复 config schema 查询中 UTF-8 字符截断问题
- 修复流式命令在未显式指定 `--output` 时误报 warning
- 修复 `alert rule --type` help 文本和示例引用了不可发现的类型名
- 修复跨域重定向时 Authorization header 未被剥离，存在凭证泄露风险

---

# v0.1.0

InCloud CLI 首个发布版本，提供对 InHand Cloud 平台的完整命令行管理能力。

## ✨ 新功能

### 核心能力
- OAuth 浏览器登录（PKCE），Token 自动刷新，登录状态查看
- 多格式输出：JSON、YAML、表格，内置人性化格式化
- 通用 `api` 命令，支持查询参数、请求体和自定义 Header
- 配置上下文管理（use/list/set/delete/current）
- 实时 SSE 流式 ping 和 traceroute 诊断

### 设备管理
- 设备全生命周期：list、get、create、update、delete
- 设备分组、配置、影子、已连接客户端管理
- 信号、接口、上下线事件、syslog 日志查看
- 位置管理、流量统计、性能监控、在线状态
- 天线信息、上行链路、远程执行
- 批量导入（CSV/XLSX）和导出（CSV）
- 分配、取消分配、转移设备

### 固件管理
- 固件列表和详情查看
- 升级任务：创建、列表、取消、执行详情、重试

### 告警
- 告警列表、详情、确认、确认统计
- 告警规则 CRUD、告警导出

### 概览仪表盘
- 仪表盘、设备、告警、流量、离线汇总
- Top 设备和 Top 告警类型分析

### 网络与连接
- SD-WAN 模块：网络、设备、隧道、连接管理
- OOBM 带外管理命令
- Connector 用量：统计、趋势、TopK

### 组织管理
- 组织、用户、角色管理
- 操作审计日志查询

### 产品
- 产品 CRUD：list、get、create、update、delete

## 🔧 改进
- 字节、比特率、百分比、延迟、抖动、时长等人性化格式化
- 表格排除列模式（! 前缀）、点分路径解析嵌套字段
- TTY 表格分页摘要头、自动展平嵌套对象
- 使用 charmbracelet/huh 交互式确认提示

## 🐛 修复
- 修复未指定 --output 时默认输出格式
- 修复分页头 ANSI 样式嵌套导致的颜色错乱
- 修复导出文件权限（0o600）
- 修复 OAuth 登录后浏览器关闭失败的回退提示
- 修复设备导入轮询和验证状态处理
