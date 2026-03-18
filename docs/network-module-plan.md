# Network 模块实现计划

> 基于 nezha-network 和 device-live 的 API 调研。共 103 个端点，几乎全部用户可用。
> 网络模块是端点最多的模块之一，按功能域拆分为 7 个子模块。

## 子模块划分

| # | 子模块 | CLI 路径 | 端点数 | 说明 |
|---|--------|---------|--------|------|
| 1 | AutoVPN | `incloud network vpn` | ~14 | SD-WAN 网络、隧道、连接 |
| 2 | InConnect 连接器 | `incloud network connector` | ~39 | 连接器网络 + 账号 + 设备 + 端点 + 日志 |
| 3 | OOBM | `incloud network oobm` | ~13 | 带外管理（Web/RDP/SSH/串口） |
| 4 | 网络资产 | `incloud network asset` | ~7 | MAC 追踪的网络资产 |
| 5 | 网络客户端 | `incloud network client` | ~19 | Wi-Fi/LAN 客户端监控 |
| 6 | 远程访问 | `incloud network touch` | ~10 | InTouch 远程连接 |

---

## TODO List

### AutoVPN (`incloud network vpn`)

- [ ] `network vpn list` — 列出 AutoVPN 网络
- [ ] `network vpn get <id>` — 查看网络详情（含隧道信息）
- [ ] `network vpn create` — 创建 AutoVPN 网络
- [ ] `network vpn update <id>` — 更新网络（添加/移除设备，重建隧道）
- [ ] `network vpn delete <id>` — 删除网络
- [ ] `network vpn export` — 导出网络列表
- [ ] `network vpn devices <id>` — 查看网络中的设备
- [ ] `network vpn tunnels <id>` — 查看网络隧道
- [ ] `network vpn connections <id>` — 查看网络连接
- [ ] `network vpn candidates` — 查找可加入网络的候选设备
- [ ] `network vpn verify-subnets` — 验证子网列表
- [ ] `network vpn device-subnets <deviceId>` — 查看设备子网信息

### InConnect 连接器 — 网络 (`incloud network connector`)

- [ ] `network connector list` — 列出连接器网络
- [ ] `network connector get <id>` — 查看连接器详情
- [ ] `network connector create` — 创建连接器网络
- [ ] `network connector update <id>` — 更新连接器网络
- [ ] `network connector delete <id>` — 删除连接器网络
- [ ] `network connector delete --bulk <ids>` — 批量删除
- [ ] `network connector stats` — 连接器统计概览
- [ ] `network connector export` — 导出连接器列表
- [ ] `network connector export <id>` — 导出单个连接器配置

### InConnect 连接器 — 账号 (`incloud network connector account`)

- [ ] `network connector account list <networkId>` — 列出连接器账号（VPN 用户）
- [ ] `network connector account create <networkId>` — 添加账号
- [ ] `network connector account create --batch <networkId>` — 批量添加
- [ ] `network connector account update <networkId> <accountId>` — 更新账号
- [ ] `network connector account delete <networkId> <accountId>` — 删除账号
- [ ] `network connector account delete --bulk <networkId>` — 批量删除
- [ ] `network connector account download-ovpn <accountId>` — 下载 OpenVPN 配置文件

### InConnect 连接器 — 设备 (`incloud network connector device`)

- [ ] `network connector device list <networkId>` — 列出连接器中的设备
- [ ] `network connector device list-all` — 列出所有连接器设备
- [ ] `network connector device add <networkId>` — 添加设备
- [ ] `network connector device add --batch <networkId>` — 批量添加
- [ ] `network connector device update <networkId> <deviceId>` — 更新设备设置
- [ ] `network connector device delete <networkId> <deviceId>` — 移除设备
- [ ] `network connector device delete --bulk <networkId>` — 批量移除
- [ ] `network connector device candidates` — 查找候选设备

### InConnect 连接器 — 端点 (`incloud network connector endpoint`)

- [ ] `network connector endpoint list <networkId>` — 列出端点
- [ ] `network connector endpoint create <networkId>` — 添加端点
- [ ] `network connector endpoint create --batch <networkId>` — 批量添加
- [ ] `network connector endpoint update <networkId> <endpointId>` — 更新端点
- [ ] `network connector endpoint delete <networkId> <endpointId>` — 删除端点
- [ ] `network connector endpoint delete --bulk <networkId>` — 批量删除

### InConnect 连接器 — 日志与流量 (`incloud network connector usage`)

- [ ] `network connector usage events <networkId> <accountId>` — 账号上下线事件
- [ ] `network connector usage logs <networkId> <accountId>` — 连接日志历史
- [ ] `network connector usage tendency <networkId> <accountId>` — 连接趋势
- [ ] `network connector usage stats` — 总体流量统计
- [ ] `network connector usage stats export` — 导出流量统计
- [ ] `network connector usage trend` — 流量趋势
- [ ] `network connector usage topk` — Top-K 流量消耗排名

### OOBM — 资源 (`incloud network oobm`)

- [x] `network oobm list` — 列出 OOBM 资源
- [x] `network oobm create` — 创建 OOBM 资源（Web/RDP/SSH）
- [x] `network oobm update <id>` — 更新 OOBM 资源
- [x] `network oobm delete <ids...>` — 删除 OOBM 资源
- [x] `network oobm connect <id>` — 打开 OOBM 连接（ngrok 隧道）
- [x] `network oobm close <id>` — 关闭 OOBM 连接

### OOBM — 串口 (`incloud network oobm serial`)

- [x] `network oobm serial list` — 列出串口配置
- [x] `network oobm serial create` — 创建串口配置
- [x] `network oobm serial update <id>` — 更新串口配置
- [x] `network oobm serial delete <ids...>` — 删除串口配置
- [x] `network oobm serial connect <id>` — 打开串口隧道
- [x] `network oobm serial close <id>` — 关闭串口隧道

### OOBM — 隧道日志

- [x] `network oobm logs <deviceId>` — 查看设备的 ngrok 隧道连接日志

### 网络资产 (`incloud network asset`)

- [ ] `network asset list` — 列出网络资产
- [ ] `network asset create` — 创建网络资产（按 MAC 追踪）
- [ ] `network asset update <id>` — 更新资产
- [ ] `network asset delete <id>` — 删除资产
- [ ] `network asset delete --bulk <ids>` — 批量删除
- [ ] `network asset import <file>` — 从 CSV/Excel 导入
- [ ] `network asset export` — 导出资产列表

### 网络客户端（已移至 `incloud device client`）

- [x] `device client list` — 列出所有连接客户端
- [x] `device client get <id>` — 查看客户端详情
- [x] `device client update <id>` — 更新客户端名称
- [x] `device client throughput <id>` — 客户端吞吐量时序
- [x] `device client rssi <id>` — 客户端 RSSI 信号时序
- [x] `device client sinr <id>` — 客户端 SINR 信号时序
- [x] `device client datausage-hourly <id>` — 客户端小时流量
- [x] `device client datausage-daily <id>` — 客户端日流量
- [x] `device client online-events <id>` — 客户端上下线事件
- [x] `device client online-stats <id>` — 客户端在线统计
- ~~`network client device <deviceId>`~~ — 不实现（实时查看设备连接客户端）
- ~~`network client stats`~~ — 不实现
- ~~`network client export`~~ — 不实现
- ~~`network client export --device <deviceId>`~~ — 不实现
- ~~`network client mark-assets`~~ — 不实现

---

## 不纳入 CLI 的功能

| 功能 | 端点数 | 理由 |
|------|--------|------|
| 连接器强制推送配置 | 1 | InternalApi |
| TURN 认证密钥 | 1 | 内部 TURN 服务器使用 |

## 备注

- network 是端点最多的模块（~103），建议分阶段实现
- 实现优先级建议：AutoVPN > InConnect 连接器 > OOBM > 客户端 > 资产 > Touch
- InConnect 连接器较复杂，包含网络/账号/设备/端点/日志 5 层子资源
- OOBM connect/close 操作涉及 ngrok 隧道创建，可能需要轮询状态
- Touch 远程连接涉及 WebRTC/TURN，CLI 场景下主要用于获取连接信息
