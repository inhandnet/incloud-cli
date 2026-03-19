# Network 模块实现计划

> 基于 nezha-network 和 device-live 的 API 调研。共 103 个端点，几乎全部用户可用。
> 网络模块是端点最多的模块之一，按功能域拆分为 7 个子模块。

## 子模块划分

| # | 子模块 | CLI 路径 | 端点数 | 说明 |
|---|--------|---------|--------|------|
| 1 | AutoVPN | `incloud sdwan` | ~14 | SD-WAN 网络、隧道、连接 |
| 2 | InConnect 连接器 | `incloud connector` | ~39 | 连接器网络 + 账号 + 设备 + 端点 + 日志 |
| 3 | OOBM | `incloud oobm` | ~13 | 带外管理（Web/RDP/SSH/串口） |
| 4 | 网络客户端 | `incloud device client` | ~19 | Wi-Fi/LAN 客户端监控 |

---

## TODO List

### SD-WAN (`incloud sdwan`)

- [x] `sdwan network list` — 列出 SD-WAN 网络
- [x] `sdwan network get <id>` — 查看网络详情（含隧道信息）
- [x] `sdwan network create` — 创建 SD-WAN 网络
- [x] `sdwan network update <id>` — 更新网络（添加/移除设备，重建隧道）
- [x] `sdwan network delete <id>` — 删除网络
- [x] `sdwan network tunnels <id>` — 查看网络隧道
- [x] `sdwan network connections <id>` — 查看网络连接
- [x] `sdwan network connection-tunnels <networkId> <connectionId>` — 查看连接的隧道详情
- [x] `sdwan devices <networkId>` — 查看网络中的设备
- [x] `sdwan candidates` — 查找可加入网络的候选设备
- [x] `sdwan verify-subnets` — 验证子网列表
- [x] `sdwan device-subnets <deviceId>` — 查看设备子网信息

### InConnect 连接器 — 网络 (`incloud connector network`)

- [x] `connector network list` — 列出连接器网络
- [x] `connector network get <id>` — 查看连接器详情
- [x] `connector network create` — 创建连接器网络
- [x] `connector network update <id>` — 更新连接器网络
- [x] `connector network delete <id>` — 删除连接器网络
- [ ] `connector network delete --bulk <ids>` — 批量删除
- [x] `connector network stats` — 连接器统计概览
- ~~`connector network export`~~ — 不实现
- ~~`connector network export <id>`~~ — 不实现

### InConnect 连接器 — 账号 (`incloud connector account`)

- [x] `connector account list <networkId>` — 列出连接器账号（VPN 用户）
- [x] `connector account create <networkId>` — 添加账号
- [ ] `connector account create --batch <networkId>` — 批量添加
- [x] `connector account update <networkId> <accountId>` — 更新账号
- [x] `connector account delete <networkId> <accountId>` — 删除账号
- [ ] `connector account delete --bulk <networkId>` — 批量删除
- [x] `connector account download-ovpn <accountId>` — 下载 OpenVPN 配置文件

### InConnect 连接器 — 设备 (`incloud connector device`)

- [x] `connector device list <networkId>` — 列出连接器中的设备
- [x] `connector device list-all` — 列出所有连接器设备
- [x] `connector device add <networkId>` — 添加设备
- [ ] `connector device add --batch <networkId>` — 批量添加
- [x] `connector device update <networkId> <deviceId>` — 更新设备设置
- [x] `connector device delete <networkId> <deviceId>` — 移除设备
- [ ] `connector device delete --bulk <networkId>` — 批量移除
- [x] `connector device candidates` — 查找候选设备

### InConnect 连接器 — 端点 (`incloud connector endpoint`)

- [x] `connector endpoint list <networkId>` — 列出端点
- [x] `connector endpoint create <networkId>` — 添加端点
- [ ] `connector endpoint create --batch <networkId>` — 批量添加
- [x] `connector endpoint update <networkId> <endpointId>` — 更新端点
- [x] `connector endpoint delete <networkId> <endpointId>` — 删除端点
- [ ] `connector endpoint delete --bulk <networkId>` — 批量删除

### InConnect 连接器 — 日志与流量 (`incloud connector account` / `incloud connector usage`)

- [x] `connector account events <networkId> <accountId>` — 账号上下线事件
- [x] `connector account logs <networkId> <accountId>` — 连接日志历史
- [x] `connector account tendency <networkId> <accountId>` — 连接趋势
- [x] `connector usage stats` — 总体流量统计
- ~~`connector usage stats export`~~ — 不实现
- [x] `connector usage trend` — 流量趋势
- [x] `connector usage topk` — Top-K 流量消耗排名

### OOBM — 资源 (`incloud oobm`)

- [x] `oobm list` — 列出 OOBM 资源
- [x] `oobm create` — 创建 OOBM 资源（Web/RDP/SSH）
- [x] `oobm update <id>` — 更新 OOBM 资源
- [x] `oobm delete <ids...>` — 删除 OOBM 资源
- [x] `oobm connect <id>` — 打开 OOBM 连接（ngrok 隧道）
- [x] `oobm close <id>` — 关闭 OOBM 连接

### OOBM — 串口 (`incloud oobm serial`)

- [x] `oobm serial list` — 列出串口配置
- [x] `oobm serial create` — 创建串口配置
- [x] `oobm serial update <id>` — 更新串口配置
- [x] `oobm serial delete <ids...>` — 删除串口配置
- [x] `oobm serial connect <id>` — 打开串口隧道
- [x] `oobm serial close <id>` — 关闭串口隧道

### OOBM — 隧道日志

- [x] `oobm logs <deviceId>` — 查看设备的 ngrok 隧道连接日志

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
| 网络资产 | ~7 | 不实现 |
| 远程访问 (InTouch) | ~10 | 不实现 |
| 连接器导出 | 2 | 不实现 |

## 备注

- InConnect 连接器较复杂，包含网络/账号/设备/端点/日志 5 层子资源
- OOBM connect/close 操作涉及 ngrok 隧道创建，可能需要轮询状态
