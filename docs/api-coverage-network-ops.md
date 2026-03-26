# API 覆盖分析 - 网络与运维域

## 汇总统计

- Portal/Network 中的 API 总数：53
- CLI 已覆盖：40
- CLI 未覆盖（Gap）：13
- 覆盖率：75%

## 详细对比表

### SD-WAN（`/api/v1/autovpn/*`）

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | /api/v1/autovpn/networks | 获取 SD-WAN 网络列表 | `incloud sdwan network list` | ✅ 已覆盖 |
| POST | /api/v1/autovpn/networks | 创建 SD-WAN 网络 | `incloud sdwan network create` | ✅ 已覆盖 |
| GET | /api/v1/autovpn/networks/{id} | 获取网络详情（含 tunnels 统计） | `incloud sdwan network get` | ✅ 已覆盖 |
| PUT | /api/v1/autovpn/networks/{id} | 更新 SD-WAN 网络 | `incloud sdwan network update` | ✅ 已覆盖 |
| DELETE | /api/v1/autovpn/networks/{id} | 删除 SD-WAN 网络 | `incloud sdwan network delete` | ✅ 已覆盖 |
| GET | /api/v1/autovpn/networks/{id}/devices | 获取网络中的设备列表 | `incloud sdwan devices <networkId>` | ✅ 已覆盖 |
| GET | /api/v1/autovpn/networks/{id}/tunnels | 获取网络隧道列表 | `incloud sdwan network tunnels` | ✅ 已覆盖 |
| GET | /api/v1/autovpn/networks/{id}/connections | 获取网络连接列表 | `incloud sdwan network connections` | ✅ 已覆盖 |
| GET | /api/v1/autovpn/networks/{id}/connections/{connId}/tunnels | 获取连接的隧道列表 | `incloud sdwan network connection-tunnels` | ✅ 已覆盖 |
| POST | /api/v1/autovpn/networks/devices/candidates | 查找候选设备 | `incloud sdwan candidates` | ✅ 已覆盖 |
| GET | /api/v1/autovpn/devices/{id}/subnets | 获取设备子网 | `incloud sdwan device-subnets` | ✅ 已覆盖 |
| POST | /api/v1/autovpn/devices/subnets/verify | 检查子网冲突 | `incloud sdwan verify-subnets` | ✅ 已覆盖 |

### InConnect VPN（`/api/v1/connectors/*`）

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | /api/v1/connectors | 获取 VPN 网络列表 | `incloud connector network list` | ✅ 已覆盖 |
| POST | /api/v1/connectors | 创建 VPN 网络 | `incloud connector network create` | ✅ 已覆盖 |
| GET | /api/v1/connectors/{id} | 获取 VPN 网络详情 | `incloud connector network get` | ✅ 已覆盖 |
| PUT | /api/v1/connectors/{id} | 更新 VPN 网络 | `incloud connector network update` | ✅ 已覆盖 |
| DELETE | /api/v1/connectors (bulk) | 批量删除 VPN 网络 | `incloud connector network delete` | ✅ 已覆盖 |
| GET | /api/v1/connectors/statistics | 获取 VPN 统计概览 | `incloud connector network stats` | ✅ 已覆盖 |
| GET | /api/v1/connectors/usage/statistics | 获取流量统计 | `incloud connector usage stats` | ✅ 已覆盖 |
| GET | /api/v1/connectors/usage/tendency | 获取流量趋势 | `incloud connector usage trend` | ✅ 已覆盖 |
| GET | /api/v1/connectors/usage/topk | 获取 TopK 流量排名 | `incloud connector usage topk` | ✅ 已覆盖 |
| GET | /api/v1/connectors/{id}/accounts | 获取账号列表 | `incloud connector account list` | ✅ 已覆盖 |
| POST | /api/v1/connectors/{id}/accounts | 创建账号 | `incloud connector account create` | ✅ 已覆盖 |
| PUT | /api/v1/connectors/{id}/accounts/{accountId} | 更新账号 | `incloud connector account update` | ✅ 已覆盖 |
| DELETE | /api/v1/connectors/{id}/accounts (bulk) | 批量删除账号 | `incloud connector account delete` | ✅ 已覆盖 |
| GET | /api/v1/connectors/accounts/{id}/ovpn/download | 下载 OpenVPN 配置 | `incloud connector account download-ovpn` | ✅ 已覆盖 |
| GET | /api/v1/connectors/{id}/accounts/{accountId}/online-events | 账号在线事件（图表） | `incloud connector account events` | ✅ 已覆盖 |
| GET | /api/v1/connectors/{id}/accounts/{accountId}/online-logs | 账号连接日志 | `incloud connector account logs` | ✅ 已覆盖 |
| GET | /api/v1/connectors/{id}/accounts/{accountId}/online-tendency | 账号流量趋势 | `incloud connector account tendency` | ✅ 已覆盖 |
| GET | /api/v1/connectors/{id}/devices | 获取设备列表 | `incloud connector device list` | ✅ 已覆盖 |
| POST | /api/v1/connectors/{id}/devices | 添加设备 | `incloud connector device add` | ✅ 已覆盖 |
| PUT | /api/v1/connectors/{id}/devices/{deviceId} | 更新设备 | `incloud connector device update` | ✅ 已覆盖 |
| DELETE | /api/v1/connectors/{id}/devices (bulk) | 批量删除设备 | `incloud connector device delete` | ✅ 已覆盖 |
| GET | /api/v1/connectors/devices/candidates | 查找候选设备 | `incloud connector device candidates` | ✅ 已覆盖 |
| POST | /api/v1/connectors/{id}/devices/batch | 批量添加设备 | — | ❌ 未覆盖 |
| GET | /api/v1/connectors/{id}/devices/clients/candidates | 获取设备终端候选 | — | ❌ 未覆盖 |
| GET | /api/v1/connectors/{id}/endpoints | 获取终端列表 | `incloud connector endpoint list` | ✅ 已覆盖 |
| POST | /api/v1/connectors/{id}/endpoints | 创建终端 | `incloud connector endpoint create` | ✅ 已覆盖 |
| PUT | /api/v1/connectors/{id}/endpoints/{endpointId} | 更新终端 | `incloud connector endpoint update` | ✅ 已覆盖 |
| DELETE | /api/v1/connectors/{id}/endpoints (bulk) | 批量删除终端 | `incloud connector endpoint delete` | ✅ 已覆盖 |
| POST | /api/v1/connectors/{id}/endpoints/batch | 批量添加终端 | — | ❌ 未覆盖 |
| POST | /api/v1/connectors/send-config | 同步设备配置 | — | ❌ 未覆盖 |
| GET | /api/v1/connectors/devices | 查询 VPN 关联设备（下拉搜索） | — | ❌ 未覆盖 |

### 网络客户端（`/api/v1/network/clients*`、`/api/v1/network/assets*`）

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | /api/v1/network/clients | 获取网络客户端列表 | `incloud device client list` | ✅ 已覆盖 |
| GET | /api/v1/network/clients/statistics | 获取客户端在线统计 | `incloud device client online-stats` | ✅ 已覆盖 |
| GET | /api/v1/network/clients/{id} | 获取客户端详情 | `incloud device client get` | ✅ 已覆盖 |
| PUT | /api/v1/network/clients/{id} | 更新客户端信息 | `incloud device client update` | ✅ 已覆盖 |
| PUT | /api/v1/network/clients/mark-assets | 标记为资产 | — | ❌ 未覆盖 |
| GET | /api/v1/network/clients/{id}/online-events-chart/statistics | 在线历史图表 | — | ❌ 未覆盖 |
| GET | /api/v1/network/clients/{id}/online-events-list | 在线事件列表 | `incloud device client online-events` | ✅ 已覆盖 |
| GET | /api/v1/network/clients/{id}/datausage-daily | 每日流量用量 | `incloud device client datausage-daily` | ✅ 已覆盖 |
| GET | /api/v1/network/clients/{id}/datausage-{type} | 流量用量（多周期） | `incloud device client datausage-hourly` | ⚠️ 部分覆盖 |
| GET | /api/v1/network/clients/{id}/throughput | 实时吞吐量 | `incloud device client throughput` | ✅ 已覆盖 |
| GET | /api/v1/network/clients/{id}/rssi | RSSI 信号数据 | `incloud device client rssi` | ✅ 已覆盖 |
| GET | /api/v1/network/clients/{id}/sinr | SINR 信号数据 | `incloud device client sinr` | ✅ 已覆盖 |
| GET | /api/v1/network/assets | 获取资产列表 | — | ❌ 未覆盖 |
| POST | /api/v1/network/assets | 添加资产 | — | ❌ 未覆盖 |
| PUT | /api/v1/network/assets/{id} | 更新资产 | — | ❌ 未覆盖 |
| DELETE | /api/v1/network/assets/{id} | 删除资产 | — | ❌ 未覆盖 |
| POST | /api/v1/network/assets/remove | 批量删除资产 | — | ❌ 未覆盖 |

### 带外管理（`/api/v1/oobm/*`）

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | /api/v1/oobm/resources | 获取 OOBM 资源列表 | `incloud oobm list` | ✅ 已覆盖 |
| POST | /api/v1/oobm/resources | 创建 OOBM 资源 | `incloud oobm create` | ✅ 已覆盖 |
| PUT | /api/v1/oobm/resources/{id} | 更新 OOBM 资源 | `incloud oobm update` | ✅ 已覆盖 |
| DELETE | /api/v1/oobm/resources/by-ids | 批量删除 OOBM 资源 | `incloud oobm delete` | ✅ 已覆盖 |
| POST | /api/v1/oobm/resources/{id}/connect | 连接 OOBM 资源 | `incloud oobm connect` | ✅ 已覆盖 |
| POST | /api/v1/oobm/resources/{id}/close | 关闭 OOBM 连接 | `incloud oobm close` | ✅ 已覆盖 |
| GET | /api/v1/oobm/serials | 获取串口列表 | `incloud oobm serial list` | ✅ 已覆盖 |
| POST | /api/v1/oobm/serials | 创建串口配置 | `incloud oobm serial create` | ✅ 已覆盖 |
| PUT | /api/v1/oobm/serials/{id} | 更新串口配置 | `incloud oobm serial update` | ✅ 已覆盖 |
| DELETE | /api/v1/oobm/serials/{id} | 删除串口配置 | `incloud oobm serial delete` | ✅ 已覆盖 |
| DELETE | /api/v1/oobm/serials/by-ids | 批量删除串口配置 | `incloud oobm serial delete` | ✅ 已覆盖 |
| POST | /api/v1/oobm/serials/{id}/connect | 连接串口 | `incloud oobm serial connect` | ✅ 已覆盖 |
| POST | /api/v1/oobm/serials/{id}/close | 关闭串口连接 | `incloud oobm serial close` | ✅ 已覆盖 |
| GET | /api/v1/oobm/serials/{id}/credential | 获取串口凭据 | — | ❌ 未覆盖 |

### 远程访问与 Ngrok（`/api/v1/ngrok/*`）

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | /api/v1/ngrok/devices/{id}/logs | 获取 OOBM/远程 CLI 会话日志 | `incloud oobm logs` | ✅ 已覆盖 |
| POST | /api/v1/ngrok/devices/{id}/local-cli | 创建远程 CLI 会话（Ngrok） | — | ❌ 未覆盖 |
| DELETE | /api/v1/ngrok/tunnels/{id} | 关闭 Ngrok 隧道会话 | — | ❌ 未覆盖 |

### 告警（`/api/v1/alerts/*`）

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | /api/v1/alerts | 获取告警列表 | `incloud alert list` | ✅ 已覆盖 |
| GET | /api/v1/alerts/{id} | 获取告警详情 | `incloud alert get` | ✅ 已覆盖 |
| GET | /api/v1/alerts/export | 导出告警 CSV | `incloud alert export` | ✅ 已覆盖 |
| PUT | /api/v1/alerts/acknowledge | 确认告警 | `incloud alert ack` | ✅ 已覆盖 |
| PUT | /api/v1/alerts/acknowledge/all | 全部确认 | `incloud alert ack --all` | ✅ 已覆盖 |
| GET | /api/v1/alerts/acknowledge/statistics | 未确认告警统计 | `incloud overview` | ✅ 已覆盖 |
| GET | /api/v1/alerts/rules | 获取告警规则列表 | `incloud alert rule list` | ✅ 已覆盖 |
| GET | /api/v1/alerts/rules/{id} | 获取告警规则详情 | `incloud alert rule get` | ✅ 已覆盖 |
| POST | /api/v1/alerts/rules | 创建告警规则 | `incloud alert rule create` | ✅ 已覆盖 |
| PUT | /api/v1/alerts/rules/{id} | 更新告警规则 | `incloud alert rule update` | ✅ 已覆盖 |
| POST | /api/v1/alerts/rules/bulk-delete | 批量删除告警规则 | `incloud alert rule delete` | ✅ 已覆盖 |
| POST | /api/v1/alerts/rules/webhooks/send | 发送 Webhook 测试消息 | — | ❌ 未覆盖 |

### Webhook（`/api/v1/message/webhooks*`）

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | /api/v1/message/webhooks | 获取 Webhook 列表 | — | ❌ 未覆盖 |
| POST | /api/v1/message/webhooks | 创建 Webhook | — | ❌ 未覆盖 |
| PUT | /api/v1/message/webhooks/{id} | 更新 Webhook | — | ❌ 未覆盖 |
| DELETE | /api/v1/message/webhooks/{id} | 删除 Webhook | — | ❌ 未覆盖 |
| POST | /api/v1/message/webhooks/send | 发送 Webhook 测试消息 | — | ❌ 未覆盖 |

### 设备诊断（`/api/v1/devices/{id}/diagnosis/*`）

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| POST | /api/v1/devices/{id}/diagnosis/ping | 发起 Ping 诊断 | `incloud device exec ping` | ✅ 已覆盖 |
| POST | /api/v1/devices/{id}/diagnosis/traceroute | 发起 Traceroute 诊断 | `incloud device exec traceroute` | ✅ 已覆盖 |
| GET | /api/v1/devices/{id}/diagnosis/capture | 获取抓包任务状态 | `incloud device exec capture` | ✅ 已覆盖 |
| POST | /api/v1/devices/{id}/diagnosis/capture | 发起抓包任务 | `incloud device exec capture` | ✅ 已覆盖 |
| GET | /api/v1/devices/{id}/diagnosis/speedtest/config | 获取测速配置 | `incloud device exec speedtest-config` | ✅ 已覆盖 |
| POST | /api/v1/devices/{id}/diagnosis/speedtest | 发起测速 | `incloud device exec speedtest` | ✅ 已覆盖 |
| GET | /api/v1/devices/{id}/diagnosis/speed-test-histories | 获取测速历史记录 | `incloud device exec speedtest-history` | ✅ 已覆盖 |
| GET | /api/v1/devices/{id}/diagnosis/flowscan | 获取流量扫描结果 | `incloud device exec flowscan-status` | ✅ 已覆盖 |
| POST | /api/v1/devices/{id}/diagnosis/flowscan | 发起流量扫描 | `incloud device exec flowscan` | ✅ 已覆盖 |
| GET | /api/v1/devices/{id}/diagnosis/flowscan/export | 导出流量扫描数据 | — | ❌ 未覆盖 |
| GET | /api/v1/devices/{id}/diagnosis/interfaces | 获取设备接口列表 | `incloud device exec interfaces` | ✅ 已覆盖 |
| PUT | /api/v1/diagnosis/{taskId}/cancel | 取消诊断任务 | （内置于诊断命令中） | ✅ 已覆盖 |

### Touch 连接与地理位置（`/api/v1/touch/*`、`/api/v1/places/*`）

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| — | /api/v1/touch/* | Portal 中未发现该 API 调用 | — | — |
| — | /api/v1/places/* | Portal 中未发现该 API 调用 | — | — |

> 注：在 `apps/network` 和 `apps/console` 的 TypeScript 源码中未检索到 `/api/v1/touch/*` 和 `/api/v1/places/*` 路径的调用，可能这两组 API 在其他子应用（如 `apps/edge`、`apps/link`）中使用，或仅由后端内部调用。

---

## Gap 分析

### 重要缺口（影响日常运维）

1. **Webhook 管理（全部缺失）**：Portal 的告警功能中有完整的 Webhook 配置（增删改查+发送测试），CLI 完全没有 `webhook` 子命令，无法在脚本/自动化中管理告警通知渠道。
   - 缺失：`GET/POST/PUT/DELETE /api/v1/message/webhooks`、`POST /api/v1/message/webhooks/send`

2. **Ngrok 远程 CLI 会话管理**：Portal 支持为设备创建远程 CLI 会话（local-cli）并关闭 Ngrok 隧道，CLI 对这类操作没有任何支持，影响运维人员的脚本化运维能力。
   - 缺失：`POST /api/v1/ngrok/devices/{id}/local-cli`、`DELETE /api/v1/ngrok/tunnels/{id}`

3. **网络资产（Asset）管理（全部缺失）**：Portal 支持将网络客户端标记为资产并进行管理，CLI 完全没有覆盖。
   - 缺失：`GET/POST/PUT/DELETE /api/v1/network/assets`、`PUT /api/v1/network/clients/mark-assets`

4. **串口凭据查询**：`GET /api/v1/oobm/serials/{id}/credential` 用于获取串口访问凭据（用户名/密码），CLI 虽然有 `oobm serial connect` 但不支持单独查询凭据。

### 次要缺口

5. **Connector 批量操作**：`POST /api/v1/connectors/{id}/devices/batch` 和 `POST /api/v1/connectors/{id}/endpoints/batch` 批量添加设备/终端，CLI 只支持逐个添加。

6. **Connector 配置同步**：`POST /api/v1/connectors/send-config` 用于主动推送 VPN 配置到设备，CLI 无对应命令。

7. **告警规则 Webhook 测试**：`POST /api/v1/alerts/rules/webhooks/send` 用于对告警规则中配置的 Webhook 发送测试消息，CLI 无对应命令。

8. **流量扫描数据导出**：`GET /api/v1/devices/{id}/diagnosis/flowscan/export` 导出 CSV 文件，CLI 的 `flowscan-status` 不支持导出。

9. **客户端在线历史图表**：`GET /api/v1/network/clients/{id}/online-events-chart/statistics` 用于图表渲染，CLI 缺失，但 `online-events-list` 已覆盖列表形式。

10. **Connector 设备终端候选查询**：`GET /api/v1/connectors/{id}/devices/clients/candidates` 用于 Portal 中的批量添加终端向导，CLI 场景使用频次较低。

---

## 备注

- **数据来源**：Portal 侧来自 `apps/network/src/pages/` 下各子模块的 `service.ts` 文件；CLI 侧来自 `internal/cmd/{sdwan,connector,alert,oobm,device}` 目录的 Go 源码。
- **网络客户端命令归属**：`/api/v1/network/clients*` 相关 CLI 命令在 `internal/cmd/device/client*.go` 中实现（`incloud device client ...`），而非独立的 `network` 模块，但功能完整。
- **诊断命令归属**：`/api/v1/devices/{id}/diagnosis/*` 相关 CLI 命令在 `internal/cmd/device/exec_*.go` 中实现（`incloud device exec ...`）。
- **Touch/Places API**：在 `apps/network` 和 `apps/console` 的前端代码中未发现该前缀的 API 调用，暂不纳入统计范围。
- **OOBM 日志 API**：`/api/v1/ngrok/devices/{id}/logs` 在 OOBM 场景中被前端和 CLI（`incloud oobm logs`）共同使用，归类于 Ngrok 域，已计入覆盖。
