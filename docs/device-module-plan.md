# Device 模块实现计划

> 基于后端 API 调研，device 模块拆分为 10 个子模块，约 112 个用户可用端点。

## Phase 1a: 核心 CRUD + 设备组

### `incloud device` — 核心 CRUD (~25 端点)

- [ ] `device list` — 设备列表（分页、过滤、排序）
- [ ] `device get <id>` — 查看设备详情
- [ ] `device create` — 创建设备（交互式输入 SN/凭证）
- [ ] `device update <id>` — 更新设备属性（name/description/tags）
- [ ] `device delete <id>` — 删除设备
- [ ] `device summary` — 设备统计概览（在线/离线/产品分布）
- [ ] `device export` — 导出设备列表为文件
- [ ] `device move <id> --group <gid>` — 移动设备到指定组
- [ ] `device transfer <id> --org <oid>` — 转移设备到其他组织
- [ ] `device import <file>` — 批量导入（CSV）
- [ ] `device properties <id>` — 查看 IoT 属性/状态
- [ ] `device features <id>` — 查看设备特性标志
- [ ] `device location <id>` — 查看/更新位置信息

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

### `incloud device log` — 日志 (~6 端点)

- [ ] `device log syslog <id>` — 查看实时 syslog
- [ ] `device log syslog history <id>` — 查询历史 syslog（Loki）
- [ ] `device log syslog download <id>` — 下载 syslog 文件
- [ ] `device log mqtt <id>` — 查看 MQTT 通信日志
- [ ] `device log mqtt export <id>` — 导出 MQTT 日志

## Phase 1d: 监控 + 流量 + 影子 + 在线状态

### `incloud device monitor` — 监控数据 (~16 端点)

- [ ] `device monitor perf <id>` — CPU/内存等性能时序数据
- [ ] `device monitor perf refresh <id>` — 触发实时性能采集
- [ ] `device monitor perf export <id>` — 导出性能数据
- [ ] `device monitor signal <id>` — 信号强度时序（RSRP/RSRQ/SINR）
- [ ] `device monitor signal current <id>` — 当前信号值
- [ ] `device monitor signal export <id>` — 导出信号数据
- [ ] `device monitor interface <id>` — 网络接口状态
- [ ] `device monitor interface refresh <id>` — 触发实时接口采集
- [ ] `device monitor uplink <id>` — Uplink 列表及性能趋势
- [ ] `device monitor antenna <id>` — 天线信号数据
- [ ] `device monitor antenna export <id>` — 导出天线信号数据

### `incloud device datausage` — 流量统计 (~11 端点)

- [ ] `device datausage hourly <id>` — 按小时流量趋势
- [ ] `device datausage daily <id>` — 按日流量趋势
- [ ] `device datausage monthly <id>` — 按月流量趋势
- [ ] `device datausage overview <id>` — 流量概览（含趋势）
- [ ] `device datausage topk` — Top-K 设备流量排名
- [ ] `device datausage details` — 设备流量明细列表
- [ ] `device datausage export <id>` — 导出流量数据

### `incloud device shadow` — 影子文档 (~4 端点)

- [ ] `device shadow list <id>` — 列出设备的影子文档名
- [ ] `device shadow get <id> --name <n>` — 获取指定影子文档
- [ ] `device shadow update <id> --name <n>` — 更新影子期望状态
- [ ] `device shadow delete <id> --name <n>` — 删除影子文档

### `incloud device presence` — 在线状态 (~8 端点)

- [ ] `device presence events <id>` — 上下线事件历史
- [ ] `device presence uptime <id>` — 在线率统计
- [ ] `device presence offline daily <id>` — 每日离线时长
- [ ] `device presence offline topn` — 离线最多的设备排名
- [ ] `device presence offline stats` — 离线统计（分页）
- [ ] `device presence export <id>` — 导出事件历史

---

## 基础设施（在实现子命令前完成）

- [ ] 设计 device 子命令的代码结构（`internal/cmd/device/` 目录布局）
- [ ] 实现通用分页参数（`--page`, `--limit`, `--sort`）
- [ ] 实现通用过滤参数（`--filter`, `-q`）
- [ ] 实现 `--device <id>` 全局参数，支持 ID / name / SN 查找
- [ ] 表格输出：device 列表的默认列选择
- [ ] 考虑异步诊断任务的轮询/等待机制（diag/exec 类命令）

## 不纳入 device 模块

| 功能域 | 理由 | 归属 |
|--------|------|------|
| Firmware CRUD + OTA Jobs | 独立业务域 ~40+ 端点 | `firmware` 模块（P1） |
| Generic Jobs | 跨域（配置下发/OTA/CLI 配置） | `job` 模块或归入 `firmware` |
| Edge/Live（容器部署） | 独立边缘计算域 | `live` 顶级命令 |
| Touch（远程访问） | 独立远程连接域 | `touch` 顶级命令 |
| Client Identification Rules | 内部管理 API | 不实现 |
| Config Documents/Schema | 内部管理 API | 不实现 |
