# incloud CLI Dogfood 场景验收汇总

## 测试概览

- **测试时间**：2026-03-24
- **场景总数**：20
- **结果分布**：PASS 1 / WARN 19 / FAIL 0
- **执行者**：exec-1（9 个场景）、exec-2（10 个场景）、原 exec-1（#2, #3 后替换为 sonnet 模型）

## 各场景结果

| # | 场景 | 执行者 | 结果 | 关键问题数 |
|---|------|--------|------|-----------|
| 1 | VIP 客户季度巡检 | exec-2 | ⚠️ WARN | 2 |
| 2 | 批量设备突然离线的紧急排查 | exec-1 | ⚠️ WARN | 4+5建议 |
| 3 | 跨站点 VPN 隧道故障诊断 | exec-1 | ⚠️ WARN | 3+4建议 |
| 4 | 连锁门店批量配置 WiFi 和端口映射 | exec-2 | ⚠️ WARN | 3 |
| 5 | 固件升级计划制定与执行 | exec-1 | ⚠️ WARN | 4 |
| 6 | 告警规则调优与噪音治理 | exec-2 | ⚠️ WARN | 4 |
| 7 | SD-WAN 组网与链路质量监控 | exec-2 | ⚠️ WARN | 5 |
| 8 | 新客户批量设备上架与初始化 | exec-1 | ⚠️ WARN | 6 |
| 9 | 用户权限体系梳理与安全加固 | exec-2 | ⚠️ WARN | 4 |
| 10 | 蜂窝设备信号劣化趋势分析与优化 | exec-2 | ⚠️ WARN | 4 |
| 11 | 配置变更回滚与事故复盘 | exec-1 | ⚠️ WARN | 5 |
| 12 | InCloud Connector 远程访问搭建 | exec-2 | ⚠️ WARN | 7 |
| 13 | 设备流量异常与安全排查 | exec-1 | ⚠️ WARN | 4 |
| 14 | OOBM 带外管理紧急救援 | exec-2 | ⚠️ WARN | 6 |
| 15 | 多组织设备迁移与交接 | exec-1 | ⚠️ WARN | 4 |
| 16 | 设备影子文档实现自定义标签管理 | exec-2 | ⚠️ WARN | 6 |
| 17 | 设备定位与资产地图核实 | exec-1 | ⚠️ WARN | 3 |
| 18 | WiFi 终端连接问题深度排查 | exec-2 | ⚠️ WARN | 6 |
| 19 | 平台运营月报数据采集 | exec-1 | ✅ PASS | 3建议 |
| 20 | DNS 配置错误导致的间歇性断网 | exec-1 | ⚠️ WARN | 3 |

---

## 问题汇总（按优先级排序）

### P0 — 安全问题

| # | 场景 | 问题 |
|---|------|------|
| 1 | #12 Connector | PKI 私钥在 `device add`/`account create`/`list-all` 中明文返回 |
| 2 | #14 OOBM | `serial list` 返回密码明文 |

### P1 — 高优问题（阻塞核心工作流）

| # | 场景 | 问题 |
|---|------|------|
| 1 | #6 告警 | `alert rule delete` 缺少 `--yes` flag，AI 无法跳过确认 |
| 2 | #7 SD-WAN | `sdwan network create` 后端 500（Uplink Status enum "unknown"） |
| 3 | #8 上架 | `device import` 失败时错误信息只显示行号不含 SN |
| 4 | #8 上架 | `device import` 不支持 `--group`，30 台需 30 次 `device assign` |
| 5 | #9 用户 | `user list` 不返回 roles 字段，批量权限审计完全不可行 |
| 6 | #9 用户 | `user unlock` 缺少 `--yes` flag |
| 7 | #10 信号 | 缺少批量信号查询命令，40 台设备需逐个调用 |
| 8 | #11 回滚 | 无批量快照回滚（按分组），12 台只能逐台操作 |
| 9 | #11 回滚 | 活动日志不记录配置变更内容 |
| 10 | #12 Connector | `--subnet` 参数声明支持但实际报 400 |
| 11 | #13 流量 | `flowscan` 在线设备上启动失败（HTTP 400 invalid_state） |
| 12 | #15 迁移 | `device transfer` 只支持单台，35 台需 35 次命令 |
| 13 | #15 迁移 | 迁移后设备分组丢失，需逐台 assign |
| 14 | #16 影子 | 无跨设备影子文档查询（50 台需 50 次调用） |

### P2 — 中优问题（影响效率）

| # | 场景 | 问题 |
|---|------|------|
| 1 | #1 巡检 | 缺少分组级信号汇总命令 |
| 2 | #4 配置 | Schema 对 MR805/V2.0.20 完全不可用 |
| 3 | #4 配置 | `config copy` 全量复制，缺少按 key 选择性复制 |
| 4 | #5 固件 | `device group firmwares` 只返回版本号，无设备数量 |
| 5 | #5 固件 | 分批升级工作流断裂（firmware status → 手工构造 --target） |
| 6 | #6 告警 | `alert ack --all` 不返回确认数量 |
| 7 | #6 告警 | 缺少按设备聚合的告警统计视图 |
| 8 | #7 SD-WAN | `sdwan verify-subnets` 对重叠子网不报冲突 |
| 9 | #7 SD-WAN | 缺少 add-spoke/remove-spoke 增量操作 |
| 10 | #9 用户 | `user list` 缺少按角色/锁定状态的过滤选项 |
| 11 | #10 信号 | `device antenna` 返回大量 null 值 |
| 12 | #10 信号 | `device antenna` 无时间粒度切换（固定 30 分钟） |
| 13 | #11 回滚 | 快照列表无时间过滤 |
| 14 | #11 回滚 | 缺少 `snapshots diff` 命令 |
| 15 | #13 流量 | `exec ping <id> <host>` 不接受位置参数，必须写 `--host` |
| 16 | #13 流量 | `client datausage-daily` 返回 InfluxDB 原始格式 |
| 17 | #14 OOBM | update 是全字段替换，紧急场景下改一个参数需传所有字段 |
| 18 | #15 迁移 | `device export` 短标志 `-f` 不支持（`alert export` 支持） |
| 19 | #16 影子 | `shadow delete` 输出不规范（原始 JSON） |
| 20 | #17 定位 | `device location get` 无数据时 exit code=1（应为 0） |
| 21 | #17 定位 | `device export` CSV 不含 GPS 坐标 |
| 22 | #18 WiFi | `online-stats` 每事件嵌套完整 device 对象（30 事件 = 30 次重复） |
| 23 | #18 WiFi | RSSI/SINR 时序返回 200+ 行 null（有效数据仅 30 点） |
| 24 | #18 WiFi | `client list` 缺少 `--ssid` 过滤 |
| 25 | #19 月报 | `overview devices/offline/traffic` 缺时间范围过滤 |
| 26 | #20 DNS | `device config update` 不支持 `--set key=value` 简写 |
| 27 | #20 DNS | `config copy --module` 不支持按模块复制 |

### P3 — 低优/体验问题

| # | 场景 | 问题 |
|---|------|------|
| 1 | #5 固件 | `firmware list` 缺少 `-q/--search` |
| 2 | #5 固件 | `firmware job executions <jobId>` 位置参数被静默忽略 |
| 3 | #6 告警 | `alert rule update` 全量替换不支持增量 |
| 4 | #7 SD-WAN | 隧道缺少质量指标（延迟/丢包/抖动） |
| 5 | #10 信号 | 缺少信号变化告警（只有绝对阈值） |
| 6 | #13 流量 | `flowscan-status` 返回空 `{}` 无法区分状态 |
| 7 | #14 OOBM | connect 失败时缺少操作建议 |
| 8 | #16 影子 | shadow list 只有 name 无元信息 |
| 9 | #17 定位 | `device location set --address` 强制必填 |
| 10 | #19 月报 | 活动日志无法按 actor 去重统计活跃用户数 |

---

## 跨场景共性问题

### 1. 批量操作严重缺失（出现 8+ 场景）

几乎所有涉及多台设备的场景都遇到了"只能逐台操作"的问题：
- `device transfer`：单台迁移
- `device assign`：单台分组
- `device location`：单台定位
- `device antenna/signal`：单台查询
- `shadow`：单台操作
- `config snapshots restore`：单台回滚

**建议**：优先实现批量操作支持（`--group` 参数或 `--device` 多值）。

### 2. `--yes` flag 缺失（3 场景）

`alert rule delete`、`user unlock` 等危险操作缺少 `--yes` 跳过确认，AI 工具无法自动化。

### 3. 位置参数 vs flag 不一致（2 场景）

`firmware job executions` 和 `exec ping` 的位置参数行为与同族命令不一致，且静默忽略不报错。

### 4. 敏感信息明文返回（2 场景）

Connector PKI 私钥和 OOBM 串口密码在 list/get 输出中明文展示。

### 5. 全量替换 vs 增量更新（3 场景）

`alert rule update`、OOBM update、SD-WAN 设备操作都是全量替换，缺少增量操作能力。

---

## 亮点

- **device exec ping** 实时流式输出效果很好
- **device uplink** 数据极为丰富
- **device online --daily** 的 signalInfo 含日均值很实用
- **Connector** 端到端 CRUD 完整、批量删除 + `--yes` 对 AI 友好、写操作反馈规范统一
- **OOBM** 双通道完整、多服务选择性连接/关闭、close 幂等性好、logs 提供连接审计
- **shadow update** merge 语义正确（部分更新 + null 删除）、版本号自动递增
- **client** 子命令体系完整（8 个子命令）、过滤选项丰富
- **overview** 命令族提供了很好的聚合视图
- **--jq** 全局 flag 在多个场景中被自然使用，实用性高

---

## 结论

**整体评价：⚠️ WARN — CLI 基础功能可用，但批量操作能力严重不足**

20 个场景中 19 个 WARN、1 个 PASS，核心问题集中在：

1. **批量操作缺失**（最高频痛点）：几乎所有多设备场景都退化为循环单台调用
2. **安全问题**：2 处敏感信息明文返回需立即修复
3. **后端 bug**：SD-WAN create 500、flowscan 400 等需要后端协调
4. **一致性问题**：`--yes` flag、位置参数、`-f` 短标志在命令间不统一

**优先修复建议**：
- [ ] **P0**：Connector/OOBM 敏感信息脱敏
- [ ] **P1**：补全 `--yes` flag（alert rule delete、user unlock）
- [ ] **P1**：`device import --group` 支持
- [ ] **P1**：`user list` 返回 roles 字段
- [ ] **P2**：批量操作框架设计（transfer、assign、signal 等）
- [ ] **P2**：位置参数一致性修复（firmware job executions、exec ping）
