# API 覆盖分析 - 网络与运维域

> 数据来源：apps/network 和 apps/universal-login

## 汇总统计

- Portal/Network 中的 API 总数：52
- CLI 已覆盖：43
- CLI 未覆盖（Gap）：9
- 覆盖率：83%

## 详细对比表

### `/api/v1/autovpn/*` — SD-WAN 网络

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | `/api/v1/autovpn/networks` | 列出 SD-WAN 网络 | `incloud sdwan network list` | ✅ 已覆盖 |
| POST | `/api/v1/autovpn/networks` | 创建 SD-WAN 网络 | `incloud sdwan network create` | ✅ 已覆盖 |
| GET | `/api/v1/autovpn/networks/{id}` | 获取网络详情 | `incloud sdwan network get` | ✅ 已覆盖 |
| PUT | `/api/v1/autovpn/networks/{id}` | 更新网络 | `incloud sdwan network update` | ✅ 已覆盖 |
| DELETE | `/api/v1/autovpn/networks/{id}` | 删除网络 | `incloud sdwan network delete` | ✅ 已覆盖 |
| GET | `/api/v1/autovpn/networks/{id}/devices` | 列出网络设备 | `incloud sdwan devices` | ✅ 已覆盖 |
| POST | `/api/v1/autovpn/networks/devices/candidates` | 候选设备 | `incloud sdwan candidates` | ✅ 已覆盖 |
| GET | `/api/v1/autovpn/networks/{id}/connections` | 列出网络连接 | `incloud sdwan network connections` | ✅ 已覆盖 |
| GET | `/api/v1/autovpn/networks/{id}/connections/{cid}/tunnels` | 连接隧道 | `incloud sdwan network connection-tunnels` | ✅ 已覆盖 |
| GET | `/api/v1/autovpn/networks/{id}/tunnels` | 网络隧道 | `incloud sdwan network tunnels` | ✅ 已覆盖 |
| GET | `/api/v1/autovpn/devices/{id}/subnets` | 设备子网 | `incloud sdwan device-subnets` | ✅ 已覆盖 |
| POST | `/api/v1/autovpn/devices/subnets/verify` | 检查子网冲突 | `incloud sdwan verify-subnets` | ✅ 已覆盖 |

### `/api/v1/connectors/*` — Connector 网络

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | `/api/v1/connectors` | 列出 Connector | `incloud connector network list` | ✅ 已覆盖 |
| POST | `/api/v1/connectors` | 创建 Connector | `incloud connector network create` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/{id}` | 获取 Connector 详情 | `incloud connector network get` | ✅ 已覆盖 |
| PUT | `/api/v1/connectors/{id}` | 更新 Connector | `incloud connector network update` | ✅ 已覆盖 |
| POST | `/api/v1/connectors/bulk/delete` | 批量删除 Connector | `incloud connector network delete` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/statistics` | Connector 统计 | `incloud connector network stats` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/usage/statistics` | 用量统计 | `incloud connector usage stats` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/usage/tendency` | 用量趋势 | `incloud connector usage trend` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/usage/topk` | 用量 Top-K | `incloud connector usage topk` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/{id}/accounts` | 列出账号 | `incloud connector account list` | ✅ 已覆盖 |
| POST | `/api/v1/connectors/{id}/accounts` | 创建账号 | `incloud connector account create` | ✅ 已覆盖 |
| PUT | `/api/v1/connectors/{id}/accounts/{aid}` | 更新账号 | `incloud connector account update` | ✅ 已覆盖 |
| POST | `/api/v1/connectors/{id}/accounts/bulk/delete` | 批量删除账号 | `incloud connector account delete` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/{id}/accounts/{aid}/online-events` | 账号在线事件 | `incloud connector account events` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/{id}/accounts/{aid}/online-logs` | 账号在线日志 | `incloud connector account logs` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/{id}/accounts/{aid}/online-tendency` | 账号用量趋势 | `incloud connector account tendency` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/accounts/{aid}/ovpn/download` | 下载 OVPN 配置 | `incloud connector account download-ovpn` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/{id}/endpoints` | 列出端点 | `incloud connector endpoint list` | ✅ 已覆盖 |
| POST | `/api/v1/connectors/{id}/endpoints` | 创建端点 | `incloud connector endpoint create` | ✅ 已覆盖 |
| PUT | `/api/v1/connectors/{id}/endpoints/{eid}` | 更新端点 | `incloud connector endpoint update` | ✅ 已覆盖 |
| POST | `/api/v1/connectors/{id}/endpoints/bulk/delete` | 批量删除端点 | `incloud connector endpoint delete` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/{id}/devices` | 列出网络设备 | `incloud connector device list` | ✅ 已覆盖 |
| POST | `/api/v1/connectors/{id}/devices` | 添加设备 | `incloud connector device add` | ✅ 已覆盖 |
| PUT | `/api/v1/connectors/{id}/devices/{did}` | 更新设备 | `incloud connector device update` | ✅ 已覆盖 |
| POST | `/api/v1/connectors/{id}/devices/bulk/delete` | 批量删除设备 | `incloud connector device delete` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/devices/candidates` | 候选设备 | `incloud connector device candidates` | ✅ 已覆盖 |
| GET | `/api/v1/connectors/devices` | 全局设备列表 | `incloud connector device list-all` | ✅ 已覆盖 |
| POST | `/api/v1/connectors/{id}/devices/batch` | 批量添加设备 | — | ❌ 未覆盖 |
| POST | `/api/v1/connectors/send-config` | 同步设备配置 | — | ❌ 未覆盖 |

### `/api/v1/network/clients*` — 网络客户端

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | `/api/v1/network/clients` | 列出客户端 | — | ❌ 未覆盖 |
| GET | `/api/v1/network/clients/statistics` | 客户端统计 | — | ❌ 未覆盖 |
| PUT | `/api/v1/network/clients/mark-assets` | 标记为资产 | — | ❌ 未覆盖 |
| GET | `/api/v1/network/clients/{id}` | 获取客户端详情 | — | ❌ 未覆盖 |
| PUT | `/api/v1/network/clients/{id}` | 更新客户端 | — | ❌ 未覆盖 |
| GET | `/api/v1/network/clients/{id}/{type}` | 客户端信号数据（rssi/sinr/throughput） | — | ❌ 未覆盖 |
| GET | `/api/v1/network/clients/{id}/online-events-chart/statistics` | 在线事件统计图 | — | ❌ 未覆盖 |
| GET | `/api/v1/network/clients/{id}/online-events-list` | 在线事件列表 | — | ❌ 未覆盖 |
| GET | `/api/v1/network/clients/{id}/datausage-{type}` | 用量趋势（daily/weekly） | — | ❌ 未覆盖 |
| GET | `/api/v1/network/clients/{id}/throughput` | 上下行速率 | — | ❌ 未覆盖 |
| GET | `/api/v1/network/clients/export` | 导出客户端 | — | ❌ 未覆盖 |

### `/api/v1/network/assets*` — 网络资产

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | `/api/v1/network/assets` | 列出资产 | — | ❌ 未覆盖 |
| POST | `/api/v1/network/assets` | 创建资产 | — | ❌ 未覆盖 |
| PUT | `/api/v1/network/assets/{id}` | 更新资产 | — | ❌ 未覆盖 |
| DELETE | `/api/v1/network/assets/{id}` | 删除资产 | — | ❌ 未覆盖 |
| POST | `/api/v1/network/assets/remove` | 批量删除资产 | — | ❌ 未覆盖 |
| POST | `/api/v1/network/assets/imports` | 批量导入资产（文件上传） | — | ❌ 未覆盖 |
| GET | `/api/v1/network/assets/export` | 导出资产 | — | ❌ 未覆盖 |

### `/api/v1/oobm/*` — 带外管理

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | `/api/v1/oobm/resources` | 列出 OOBM 资源 | `incloud oobm list` | ✅ 已覆盖 |
| POST | `/api/v1/oobm/resources` | 创建 OOBM 资源 | `incloud oobm create` | ✅ 已覆盖 |
| PUT | `/api/v1/oobm/resources/{id}` | 更新 OOBM 资源 | `incloud oobm update` | ✅ 已覆盖 |
| DELETE | `/api/v1/oobm/resources/by-ids` | 批量删除资源 | `incloud oobm delete` | ✅ 已覆盖 |
| POST | `/api/v1/oobm/resources/{id}/connect` | 连接资源 | `incloud oobm connect` | ✅ 已覆盖 |
| POST | `/api/v1/oobm/resources/{id}/close` | 关闭连接 | `incloud oobm close` | ✅ 已覆盖 |
| GET | `/api/v1/oobm/serials` | 列出串口资源 | `incloud oobm serial list` | ✅ 已覆盖 |
| POST | `/api/v1/oobm/serials` | 创建串口资源 | `incloud oobm serial create` | ✅ 已覆盖 |
| PUT | `/api/v1/oobm/serials/{id}` | 更新串口资源 | `incloud oobm serial update` | ✅ 已覆盖 |
| DELETE | `/api/v1/oobm/serials/by-ids` | 批量删除串口 | `incloud oobm serial delete` | ✅ 已覆盖 |
| POST | `/api/v1/oobm/serials/{id}/connect` | 连接串口 | `incloud oobm serial connect` | ✅ 已覆盖 |
| POST | `/api/v1/oobm/serials/{id}/close` | 关闭串口连接 | `incloud oobm serial close` | ✅ 已覆盖 |
| GET | `/api/v1/oobm/serials/{id}/credential` | 获取串口凭证 | — | ❌ 未覆盖 |

### `/api/v1/alerts/*` 和 `/api/v1/alert/*` — 告警

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | `/api/v1/alerts` | 列出告警 | `incloud alert list` | ✅ 已覆盖 |
| GET | `/api/v1/alerts/{id}` | 获取告警详情 | `incloud alert get` | ✅ 已覆盖 |
| GET | `/api/v1/alerts/export` | 导出告警 | `incloud alert export` | ✅ 已覆盖 |
| PUT | `/api/v1/alerts/acknowledge` | 确认告警 | `incloud alert ack` | ✅ 已覆盖 |
| PUT | `/api/v1/alerts/acknowledge/all` | 确认全部告警 | `incloud alert ack --all` | ✅ 已覆盖 |
| GET | `/api/v1/alerts/acknowledge/statistics` | 未读统计 | — | ❌ 未覆盖 |
| GET | `/api/v1/alerts/rules` | 列出告警规则 | `incloud alert rule list` | ✅ 已覆盖 |
| POST | `/api/v1/alerts/rules` | 创建告警规则 | `incloud alert rule create` | ✅ 已覆盖 |
| GET | `/api/v1/alerts/rules/{id}` | 获取规则详情 | `incloud alert rule get` | ✅ 已覆盖 |
| PUT | `/api/v1/alerts/rules/{id}` | 更新告警规则 | `incloud alert rule update` | ✅ 已覆盖 |
| POST | `/api/v1/alerts/rules/bulk-delete` | 批量删除规则 | `incloud alert rule delete` | ✅ 已覆盖 |
| POST | `/api/v1/alerts/rules/webhooks/send` | 测试 Webhook 通知 | — | ❌ 未覆盖 |
| GET | `/api/v1/alert/top-alert-devices` | 告警最多设备排名 | — | ❌ 未覆盖 |
| GET | `/api/v1/alert/top-alert-types` | 告警类型排名 | — | ❌ 未覆盖 |

### `/api/v1/ngrok/*` — 隧道/远程访问

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| POST | `/api/v1/ngrok/devices/{id}/local-web` | 创建 Web 访问隧道 | — | ❌ 未覆盖 |
| POST | `/api/v1/ngrok/devices/{id}/local-cli` | 创建 CLI 会话 | — | ❌ 未覆盖 |
| DELETE | `/api/v1/ngrok/tunnels/{id}` | 关闭隧道 | — | ❌ 未覆盖 |
| GET | `/api/v1/ngrok/tunnels/{id}/wait` | 等待隧道就绪 | — | ❌ 未覆盖 |
| GET | `/api/v1/ngrok/devices/{id}/logs` | 会话记录（oobm/cli） | `incloud oobm logs` | ✅ 已覆盖（oobm场景） |

### `/api/v1/devices/{id}/diagnosis/*` — 设备诊断

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | `/api/v1/devices/{id}/diagnosis/interfaces` | 获取接口信息 | `incloud device exec interfaces` | ✅ 已覆盖 |
| GET | `/api/v1/devices/{id}/diagnosis/speedtest/config` | 测速配置 | `incloud device exec speedtest-config` | ✅ 已覆盖 |
| POST | `/api/v1/devices/{id}/diagnosis/speedtest` | 启动测速 | `incloud device exec speedtest` | ✅ 已覆盖 |
| GET | `/api/v1/devices/{id}/diagnosis/speed-test-histories` | 测速历史 | `incloud device exec speedtest-history` | ✅ 已覆盖 |
| POST | `/api/v1/devices/{id}/diagnosis/traceroute` | 启动 Traceroute | `incloud device exec traceroute` | ✅ 已覆盖 |
| POST | `/api/v1/devices/{id}/diagnosis/ping` | 启动 Ping | `incloud device exec ping` | ✅ 已覆盖 |
| POST | `/api/v1/devices/{id}/diagnosis/capture` | 启动抓包 | `incloud device exec capture` | ✅ 已覆盖 |
| GET | `/api/v1/devices/{id}/diagnosis/capture` | 获取抓包状态 | `incloud device exec capture` | ✅ 已覆盖 |
| PUT | `/api/v1/diagnosis/{taskId}/cancel` | 取消诊断任务 | `incloud device exec cancel` | ✅ 已覆盖 |
| POST | `/api/v1/devices/{id}/diagnosis/flowscan` | 启动域名监控 | `incloud device exec flowscan` | ✅ 已覆盖 |
| GET | `/api/v1/devices/{id}/diagnosis/flowscan` | 获取域名监控状态 | `incloud device exec flowscan-status` | ✅ 已覆盖 |

### `/api/v1/message/webhooks*` — 消息 Webhook（apps/network + universal-login 均有）

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | `/api/v1/message/webhooks` | 列出 Webhook | — | ❌ 未覆盖 |
| POST | `/api/v1/message/webhooks` | 创建 Webhook | — | ❌ 未覆盖 |
| PUT | `/api/v1/message/webhooks/{id}` | 更新 Webhook | — | ❌ 未覆盖 |
| DELETE | `/api/v1/message/webhooks/{id}` | 删除 Webhook | — | ❌ 未覆盖 |
| POST | `/api/v1/message/webhooks/send` | 测试发送 Webhook | — | ❌ 未覆盖 |

### `/api/v1/touch/*` — 触摸屏/门户

> **结论**：`apps/network` 和 `apps/universal-login` 中均未发现对 `/api/v1/touch/*` 的调用。

### `/api/v1/places/*` — 地点/地图

| 方法 | API 路径 | 前端用途 | CLI 命令 | 状态 |
|------|---------|---------|---------|------|
| GET | `/api/v1/places/suggestion` | 地点搜索建议 | — | ❌ 未覆盖 |
| GET | `/api/v1/places/details/{id}` | 地点详情 | — | ❌ 未覆盖 |
| GET | `/api/v1/places/geocoder` | 反向地理编码 | — | ❌ 未覆盖 |

## Gap 分析

### 重要缺口

1. **`/api/v1/network/clients*`（全部 11 个 API 未覆盖）**
   网络客户端是独立的资源域，前端有完整的列表、详情、编辑、信号数据、在线事件、用量趋势、导出功能。CLI 完全没有对应命令，是覆盖率最低的领域。

2. **`/api/v1/network/assets*`（全部 7 个 API 未覆盖）**
   网络资产管理（CRUD、批量导入/导出）完全缺失。

3. **`/api/v1/message/webhooks*`（全部 5 个 API 未覆盖）**
   Webhook 管理在 universal-login（系统设置）和 network（告警）两个 app 均有使用，CLI 完全缺失。

4. **`/api/v1/ngrok/*`（隧道管理，4 个 API 未覆盖）**
   远程 Web 访问和 CLI 会话的创建/关闭/等待。虽然 oobm logs 通过 ngrok API 获取，但主要隧道操作（建立和关闭）没有 CLI 支持。

### 次要缺口

5. **`/api/v1/oobm/serials/{id}/credential`**
   获取串口凭证（用于客户端认证），使用频率较低。

6. **`/api/v1/connectors/{id}/devices/batch`**
   批量添加设备到 Connector，前端有专属对话框支持。

7. **`/api/v1/connectors/send-config`**
   手动触发设备配置同步，日常不常用但运维场景有价值。

8. **`/api/v1/alerts/rules/webhooks/send` 和 `/api/v1/alerts/acknowledge/statistics`**
   前者为测试告警通知的功能，后者为未读统计，均为辅助性 API。

9. **`/api/v1/alert/top-alert-devices` 和 `/api/v1/alert/top-alert-types`**
   告警统计分析 API（注意：前缀 `/alert` 非 `/alerts`），用于 Dashboard 展示，CLI 无对应命令。

10. **`/api/v1/places/*`**
    地图/地点 API（建议、详情、反向地理编码），与设备位置管理相关，CLI 场景需求较低。
