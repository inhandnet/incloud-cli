# InCloud CLI — 未覆盖 API 调研

> 调研日期：2026-03-26
> 每个 API 均已调研前端调用位置和后端实现细节

---

## Device 统计与导入

### GET /api/v1/devices/summary — Dashboard 设备统计摘要 ✅ `incloud overview devices`

- **作用**: 获取设备管理仪表板的统计摘要数据（设备数量、配置、升级、网络、产品等）
- **使用背景**: Dashboard 首页显示设备相关的关键指标，每 10 秒自动刷新
- **前端调用**: `apps/network/src/pages/dashboard/overview/index.tsx` → `fetchStatsData()`，数据用于 `DevicesStats`、`PieChartStats` 等组件
- **后端实现**: `nezha-iot` → `DeviceController.listDeviceSummary()`，通过多个 `SummaryHandler` 策略模式处理不同统计类型（count/config/upgrade/networking/product）
- **权限**: `devices:read`
- **参数**: `fields` 可选，指定返回的统计类型

### GET /api/v1/devices/offline/statistics — 离线设备统计明细 ✅ `incloud overview offline`

- **作用**: 获取离线设备的详细统计数据（分页列表），包括离线次数、离线时长等
- **使用背景**: Dashboard → 离线分析 → 参数设备分析页面，支持多维度过滤
- **前端调用**: `apps/network/src/pages/dashboard/overview/OfflineAnalysis/ParamDeviceAnalysis.tsx` → `fetchOfflineStatistics()`，支持导出 Excel
- **后端实现**: `nezha-iot` → `PresenceController.getOfflineStatistics()`，查询 `DeviceOfflineDailyStats` 集合，MongoDB 聚合管道
- **权限**: `devices:read`
- **参数**: `after/before`（时间范围）、`q`（关键词）、`offlineTimesGreaterThan`、信号参数（rsrp/rsrq/sinr 等）、分页

### GET /api/v1/devices/offline/topn — 离线排名 Top-N ✅ `incloud overview offline`

- **作用**: 获取离线次数最多的 Top-N 设备排名
- **使用背景**: Dashboard → 离线排名卡片，快速识别问题设备
- **前端调用**: `apps/network/src/pages/dashboard/overview/OfflineRanking/index.tsx` → `fetchTopOfflineDevices()`，用 `RankingTable` 展示
- **后端实现**: `nezha-iot` → `PresenceController.getOnlineEventsTopN()`，MongoDB 聚合排序后合并设备名称
- **权限**: `devices:read`
- **参数**: `after/before`、`topN`、`groupId`、`deviceIds`

### PUT /api/v1/devices/imports/cancel/{id} — 取消设备导入任务 ⏭️ 跳过

- **作用**: 取消正在进行中的设备导入任务
- **使用背景**: 设备导入记录页面，当导入任务处于 checking/waiting/importing 状态时点击"取消导入"
- **前端调用**: `apps/network/src/pages/devices/list/Import/Records.tsx` → `cancelImport(id)`
- **后端实现**: `nezha-iot` → `DeviceImportController.cancel()`，原子更新状态为 CANCEL，将未完成的导入明细标记为失败
- **权限**: `devices:write`
- **约束**: 只能取消 CHECKING/WAITING/IMPORTING 状态的任务
- **跳过原因**: 导入过程很快，取消窗口期极短；低频操作可用 `incloud api` 兜底

---

## Device 详情

### GET /api/v1/devices/{id}/datausage-{type}/overview — 设备流量概览 ⏭️ 跳过

- **作用**: 获取设备指定时间段和粒度（hourly/daily/monthly）的流量数据概览
- **使用背景**: 设备流量管理页面（Traffic Overview），按接口类型（蜂窝/有线/无线）展示流量统计
- **前端调用**: `apps/network/src/pages/devices/list/profile/traffic/service.ts` → `fetchOverviewTraffic()`
- **后端实现**: `nezha-link` → `DataUsageController`，从 InfluxDB 获取流量数据，返回 overview + trend 两部分
- **参数**: 按 hourly/daily/monthly 三种粒度查询，支持按接口类型过滤
- **跳过原因**: 已有 `incloud device datausage hourly/daily/monthly` 提供详细时间序列数据；overview 汇总仅是按 cellular/wired/wireless 聚合的几个数字，CLI 场景下价值有限

### GET /api/v1/devices/{id}/online-events-chart/statistics — 设备连接历史图表 ⏭️ 跳过

- **作用**: 获取设备的在线/离线事件统计数据，用于绘制连接历史图表
- **使用背景**: 设备概览页面的连接历史部分
- **前端调用**: `apps/network/src/pages/devices/list/profile/overview/ConnectHistory/service.ts` → `fetchDeviceConnectHistory()`
- **后端实现**: `nezha-iot` → `PresenceController.findEventsStatistics4Chart()`
- **参数**: `from`、`to` 时间范围
- **跳过原因**: 汇总值（onlineTime/onlineRate）agent 可从 `device online --daily` 数据推算；list 采样点是前端画图用，对 agent 无价值

### POST /api/v1/devices/{id}/connections — 创建设备客户端连接 ⏭️ 跳过

- **作用**: 创建远程访问连接（SSH、RDP、串口等）
- **使用背景**: 设备客户端管理界面点击"创建连接"
- **前端调用**: `apps/network/src/pages/devices/list/profile/clients/service.ts` → `CreateConnection()`
- **后端实现**: `device-live`（边缘计算服务）→ `ConnectionController.startClients()`，获取 TURN 服务器信息后创建连接
- **参数**: `method`（ssh/rdp/serial）、`idleTime`、`connTime`、`port`、`clientIp`、`serial`（波特率等）
- **备注**: 前端路径 `/api/v1/devices/{id}/connections`，后端实际为 `/api/v1/touch/connections`，通过网关路由映射
- **跳过原因**: device-live 服务的 DeviceTouch 功能，依赖 TURN/WebRTC 交互式会话，不属于 incloud-manager 管理 API

### GET /api/v1/devices/{id}/jobs — 设备任务记录 ⏭️ 跳过

- **作用**: 获取设备的所有任务执行记录（OTA 升级、配置下发等）
- **使用背景**: 设备详情页面展示升级/任务历史
- **前端调用**: `apps/network/src/pages/devices/list/profile/info/service.ts` → `fetchUpgradeHistory()`
- **后端实现**: `nezha-device-manager` → `JobExecutionController.findAllDeviceJobs()`
- **参数**: 分页，可扩展 creator、org
- **跳过原因**: OTA 任务已有 `firmware job executions --device` 覆盖，配置下发已有 `device config history` 覆盖

### GET /api/v1/devices/{id}/clients — 设备连接客户端列表 ✅ `incloud device client list --device <id>`

- **作用**: 获取连接到指定设备的所有客户端列表
- **使用背景**: 设备客户端管理界面显示所有连接的客户端
- **前端调用**: `apps/network/src/pages/devices/list/service.ts` → `fetchClients()`，用于 ClientsTable 组件
- **后端实现**: `nezha-network` → `ClientController.getClients()`，异步获取（DeferredResult），支持超时参数
- **返回字段**: name、ip、mac、vlan、connection、tx、rx、uptime 等

---

## Network Clients — 网络客户端

### GET /api/v1/network/clients — 列出客户端 ✅ `incloud device client list`

- **作用**: 获取已连接的客户端列表，支持分页和过滤
- **使用背景**: 客户端列表页面展示所有网络客户端
- **前端调用**: `apps/network/src/pages/clients/list/index.tsx` → `fetchClients()`
- **后端实现**: `nezha-network` → `ClientController`(L123-126)，查询 `lastMessageAt` 存在的客户端
- **权限**: `clients:read`
- **参数**: 分页、name、deviceName、online、type（wired/wireless）

### GET /api/v1/network/clients/statistics — 客户端统计 ⏭️ 跳过

- **作用**: 获取客户端的在线/离线统计摘要
- **使用背景**: 客户端列表页顶部统计卡片（总数、在线数、离线数）
- **前端调用**: `service.ts` → `fetchClientsStatusStats()`
- **后端实现**: `nezha-network` → `ClientController`(L134-138)，MongoDB 聚合查询
- **权限**: `clients:read`
- **跳过原因**: 仅返回 online/offline/total 三个数字，CLI 场景下 `client list` 已可查看状态，无独立命令价值

### PUT /api/v1/network/clients/mark-assets — 标记为资产 ✅ `incloud device client mark-asset <id...>`

- **作用**: 将选定的客户端批量标记为资产，创建 Asset 记录
- **使用背景**: 客户端列表页"标记为资产"按钮，最多 2000 个
- **前端调用**: `apps/network/src/pages/clients/list/MarkAsAsset/index.tsx` → `handleMarkedToAssets()`
- **后端实现**: `nezha-network` → `ClientController`(L160-166)，设置 `assetId` 并创建 Asset 记录
- **权限**: `clients:write`
- **请求体**: `{ ids: [ObjectId[]] }`

### GET /api/v1/network/clients/{id} — 客户端详情 ✅ `incloud device client get <id>`

- **作用**: 获取单个客户端的详细信息
- **使用背景**: 客户端详情页面
- **前端调用**: `apps/network/src/pages/clients/list/profile/service.ts` → `fetchClient()`
- **后端实现**: `nezha-network` → `ClientController`(L140-144)
- **权限**: `clients:read`

### PUT /api/v1/network/clients/{id} — 更新客户端名称 ✅ `incloud device client update <id> --name "..."`

- **作用**: 更新客户端的自定义名称
- **使用背景**: 客户端详情页点击"编辑"按钮
- **前端调用**: `apps/network/src/pages/clients/list/EditModal/service.ts` → `updateClient()`
- **后端实现**: `nezha-network` → `ClientController`(L146-150)，更新 `name` 字段
- **权限**: `clients:read`
- **请求体**: `{ name: string }`

### GET /api/v1/network/clients/{id}/{type} — 客户端信号数据 ✅ `incloud device client rssi/sinr <id>`

- **作用**: 获取客户端的信号时间序列数据（RSSI/SINR）
- **使用背景**: 客户端详情页"连接"标签中绘制信号曲线图
- **前端调用**: `profile/connection/single-line-chart/service.ts` → `fetchSingleLineData(clientId, params, type)`
- **后端实现**: `nezha-network` → `ClientController`(L168-182)，从 InfluxDB `network.data_5m` 查询
- **权限**: `clients:read`
- **参数**: `after`、`before`（ISO 8601）

### GET /api/v1/network/clients/{id}/online-events-chart/statistics — 在线事件统计图 ✅ `incloud device client online-stats <id>`

- **作用**: 获取客户端的在线/离线事件统计数据（图表用）
- **使用背景**: 客户端详情页"连接历史"标签的统计图表
- **前端调用**: `profile/connection/history/service.ts` → `fetchClientConnectHistory()`
- **后端实现**: `nezha-network` → `ClientController`(L259-270)，MongoDB `clientLogs` 聚合
- **权限**: `clients:read`
- **参数**: `after`、`before`（必填）

### GET /api/v1/network/clients/{id}/online-events-list — 在线事件列表 ✅ `incloud device client online-events <id>`

- **作用**: 获取客户端的在线/离线事件列表（连接历史记录）
- **使用背景**: 客户端详情页"连接历史"标签的事件列表
- **前端调用**: `profile/connection/history/service.ts` → `fetchClientConnectList()`
- **后端实现**: `nezha-network` → `ClientController`(L240-257)，查询 MongoDB `clientLogs`，默认当天
- **权限**: `clients:read`
- **参数**: `after`、`before`、分页

### GET /api/v1/network/clients/{id}/datausage-{type} — 用量趋势 ✅ `incloud device client datausage-hourly/datausage-daily <id>`

- **作用**: 获取客户端的数据用量趋势（daily 按天/hourly 按 30 分钟）
- **使用背景**: 客户端详情页"连接"标签的流量趋势图
- **前端调用**: `profile/connection/trend/service.ts` → `fetchClientTrendByClientId()`
- **后端实现**: `nezha-network` → `ClientController`(L184-226)，InfluxDB 聚合查询 rx/tx/total
- **权限**: `clients:read`
- **参数**: daily 模式用 `month`（YearMonth），hourly 模式用 `after`/`before`

### GET /api/v1/network/clients/{id}/throughput — 上下行速率 ✅ `incloud device client throughput <id>`

- **作用**: 获取客户端的上下行吞吐量时间序列数据
- **使用背景**: 客户端详情页"连接"标签的速率曲线图
- **前端调用**: `profile/connection/up-down-speed/service.ts` → `fetchUpDownSpeed()`
- **后端实现**: `nezha-network` → `ClientController`(L152-158)，InfluxDB `throughputUp`/`throughputDown`
- **权限**: `clients:read`
- **参数**: `after`、`before`（必填）

### POST /api/v1/network/devices/{id}/clients/pos-ready — POS Ready 状态 🔲 待实现

- **作用**: 设置客户端的 POS Ready 状态（开启/关闭）
- **使用背景**: 设备详情页客户端表格中通过菜单操作
- **前端调用**: `devices/list/profile/clients/service.ts` → `updateStarClientPosReady()`
- **后端实现**: `nezha-network` → `ClientController`(L87-112)，通过 `DeviceClient.invokeDirectMethod()` 发送 `nezha_client_pos_ready` 直接方法（30 秒超时），成功后更新 MongoDB
- **权限**: `clients:write`
- **请求体**: `{ mac: string, enabled: boolean }`
- **依赖**: 设备需支持 `star_pos_ready` 功能
- **实现备注**: 虽然可用 `exec method` 替代，但需独立命令以同步更新 MongoDB 状态

---

## Network Assets — 网络资产

### GET /api/v1/network/assets — 列出资产 ✅ `incloud device asset list`

- **作用**: 分页获取资产列表
- **使用背景**: 客户端资产管理页面 (`/clients-assets`)
- **前端调用**: `apps/network/src/pages/clients/clients-assets/service.ts` → `fetchClientsAssets()`，ProMaxTable 数据源
- **后端实现**: `nezha-network` → `AssetController.listAssets()`(L59-62)
- **权限**: `assets:read`
- **参数**: name/mac/number（模糊搜索）、category/status（多选过滤）、expand（client/org）

### POST /api/v1/network/assets — 创建资产 ✅ `incloud device asset create`

- **作用**: 创建新资产
- **使用背景**: 资产管理页面"新增"按钮
- **前端调用**: `service.ts` → `addClientAsset()`，通过 `AssetsEditModal` 提交
- **后端实现**: `nezha-network` → `AssetController.createAsset()`(L33-48)，检查 MAC 唯一性，创建 Asset 并关联 Client
- **权限**: `assets:write`
- **必填字段**: name(1-64)、mac(XX:XX:XX:XX:XX:XX)、category、status
- **可选字段**: number(0-64)、warrantyExpiration、notes

### PUT /api/v1/network/assets/{id} — 更新资产 ✅ `incloud device asset update <id>`

- **作用**: 更新资产信息（MAC 不可改）
- **使用背景**: 资产编辑模态框
- **前端调用**: `service.ts` → `updateClientAsset()`
- **后端实现**: `nezha-network` → `AssetController.updateAsset()`(L51-56)，findAndModify 更新
- **权限**: `assets:write`

### DELETE /api/v1/network/assets/{id} — 删除资产 ✅ `incloud device asset delete <id>`

- **作用**: 删除单个资产
- **使用背景**: 资产列表操作列删除按钮
- **前端调用**: `service.ts` → `deleteClientAsset()`
- **后端实现**: `nezha-network` → `AssetController.deleteAsset()`(L65-70)，删除后调用 `ClientService.unmarkClientAsAsset()` 清除关联
- **权限**: `assets:write`

### POST /api/v1/network/assets/remove — 批量删除资产 ✅ `incloud device asset delete <id1> <id2> ...`

- **作用**: 批量删除资产（最多 2000 条）
- **使用背景**: 资产列表多选后"批量删除"按钮
- **前端调用**: `service.ts` → `batchRemove(ids)`
- **后端实现**: `nezha-network` → `AssetController.deleteByIds()`(L73-78)，findAllAndRemove 后批量清除客户端关联
- **权限**: `assets:write`
- **请求体**: `{ ids: [ObjectId, ...] }`

### POST /api/v1/network/assets/imports — 批量导入资产 ⏭️ 跳过

- **作用**: 上传 Excel 文件批量导入资产
- **使用背景**: 资产管理页面"导入"按钮
- **前端调用**: `clients-assets/import/index.tsx`，`ProFormUploadButton` 上传 .xlsx 文件
- **后端实现**: `nezha-network` → `BulkOperationController.imports()`(L24-27)，解析 Excel，验证字段和 MAC 格式，创建 Asset + Client
- **权限**: `assets:write`
- **限制**: 最多 10000 条，Excel 列：name/mac/number/status/category/warrantyExpiration
- **返回**: totalCount、successCount、failedCount、errors
- **跳过原因**: Excel 文件上传对 agent 无价值，批量创建可用 `asset create` 循环替代

### GET /api/v1/network/assets/export — 导出资产 ⏭️ 跳过

- **作用**: 导出资产列表为 CSV 文件
- **使用背景**: 资产管理页面"导出"按钮
- **前端调用**: `ExportButton` 组件，传递当前过滤条件
- **后端实现**: `nezha-network` → `BulkOperationController.exportAssets()`(L30-33)，按 Locale 生成中/英文 CSV
- **权限**: `assets:read`
- **文件名**: "资产信息.csv" / "asset_information.csv"
- **跳过原因**: `asset list` 已可获取全量数据，agent 无需 CSV 导出

**资产数据模型**:
- Category 枚举: router, gateway, ap, cash_register, barcode_scanner, voip_phone, printer, camera, mobile_phone, pc, pad, others
- Status 枚举: in_stock, in_use, in_repair, decommissioned
- 唯一索引: `{tid, mac}`

---

## Message Webhooks — 消息 Webhook

> 后端微服务: `nezha-message`，前端跨 `apps/network` 和 `apps/universal-login`

### GET /api/v1/message/webhooks — 列出 Webhook ✅ `incloud webhook list`

- **作用**: 列出当前组织的 Webhook 配置列表
- **使用背景**: 系统设置 → Webhook 管理页面；告警规则编辑页面查询可用 Webhook
- **前端调用**: `apps/universal-login/src/pages/account/Settings/system/Webhook/index.tsx` → `fetchWebHooks()`
- **后端实现**: `nezha-message` → `WebhookController.listWebhooks()`
- **权限**: `webhooks:read`
- **参数**: 分页、provider='wechat'、expand（creator/org）

### POST /api/v1/message/webhooks — 创建 Webhook ✅ `incloud webhook create`

- **作用**: 创建新的 Webhook 配置
- **使用背景**: Webhook 管理页面或告警规则编辑页面新增
- **前端调用**: `WebhookEditor` 组件 → `addWebhook(data)`
- **后端实现**: `nezha-message` → `WebhookController.createWebhook()`
- **权限**: `webhooks:write`
- **请求体**: `{ name, webhook, provider, oid }`
- **约束**: (oid, name) 唯一

### PUT /api/v1/message/webhooks/{id} — 更新 Webhook ✅ `incloud webhook update`

- **作用**: 更新已有 Webhook 配置
- **使用背景**: Webhook 编辑模式
- **前端调用**: `WebhookEditor` → `updateWebhook(hookId, data)`
- **后端实现**: `nezha-message` → `WebhookController.updateWebhook()`，验证 oid 权限
- **权限**: `webhooks:write`

### DELETE /api/v1/message/webhooks/{id} — 删除 Webhook ✅ `incloud webhook delete`

- **作用**: 删除 Webhook 配置
- **使用背景**: Webhook 管理页面删除按钮
- **前端调用**: `deleteWebhooks(hookId)`，有删除确认对话框
- **后端实现**: `nezha-message` → `WebhookController.deleteWebhook()`，发布 `WebhookEvent.deleted()` 事件
- **权限**: `webhooks:write`

### POST /api/v1/message/webhooks/send — 测试发送 Webhook ✅ `incloud webhook test`

- **作用**: 测试 Webhook 是否可正常推送消息
- **使用背景**: 创建/编辑 Webhook 时点击"测试推送"按钮
- **前端调用**: `WebhookEditor` → `sendTestMessage(webhook)`
- **后端实现**: `nezha-message` → `WebhookController.sendWebhook()`，发送固定测试消息（markdown 格式）
- **权限**: `webhooks:write`
- **请求体**: `{ webhook: "webhook_url" }`

---

## Alert 统计与 Webhook

### POST /api/v1/alerts/rules/webhooks/send — 测试告警 Webhook ⏭️ 跳过

- **作用**: 测试告警规则中配置的 Webhook 是否可正常推送告警消息
- **使用背景**: 告警规则编辑表单中配置 webhook_fe 后点击测试
- **前端调用**: `apps/network/src/pages/alerts/service.ts` → `sendTestMessage(data)`
- **后端实现**: `nezha-alert` → `WebhookService.sendTestNotification()`，发送模拟告警消息，支持 HMAC-SHA256 签名（X-Signature header）和指数退避重试
- **请求体**: `{ url: string, secret?: string }`
- **跳过原因**: 通用 `webhook test` 已覆盖测试需求，告警专用版重复度高

### GET /api/v1/alert/top-alert-devices — 告警最多设备排名 ✅ `incloud overview alerts`

- **作用**: 获取指定时间范围内告警最多的设备排名（Top N）
- **使用背景**: Dashboard 仪表板"告警最多设备"排行榜和设备组概览
- **前端调用**: `apps/network/src/pages/dashboard/overview/AlertsRanking/index.tsx` → `fetchTopAlertDevices()`，`RankingTable` 展示
- **后端实现**: `nezha-alert` → `AlertController.listDeviceAlertTopK()`(L178-184)，MongoDB `alerts.stats` 聚合，按 deviceId 分组求和
- **权限**: `ALERTS_READ`
- **参数**: `before`、`after`、`devicegroupId`（可选）
- **数据来源**: 定时任务每小时统计，数据 95 天后自动过期

### GET /api/v1/alert/top-alert-types — 告警类型排名 ✅ `incloud overview alerts`

- **作用**: 获取指定时间范围内最频繁的告警类型排名（Top N）
- **使用背景**: Dashboard 仪表板"告警类型排行"
- **前端调用**: 同上 `AlertsRanking` 组件 → `fetchTopAlertTypes()`
- **后端实现**: `nezha-alert` → `AlertController.listDeviceAlertTypeTopK()`(L187-193)，按 `type` 分组求和
- **权限**: `ALERTS_READ`
- **参数**: 同上
- **备注**: 告警类型名称通过 `RuleType` 枚举和国际化翻译显示

---

## Ngrok — 远程访问

> 隧道创建由 Go 微服务 `ngrok` 负责，日志查询由 Java 微服务 `nezha-network` 负责

### POST /api/v1/ngrok/devices/{id}/local-web — 创建 Web 访问隧道 ✅ `incloud tunnel open-web`

- **作用**: 为设备创建 Web UI 访问隧道
- **使用背景**: 设备列表/详情页"远程访问" → "Device Web UI"
- **前端调用**: `RemoteAccessButton` 组件 + `Ngrok` 组件 → `fetchDeviceWeb()`，在 Modal 的 iframe 中展示
- **后端实现**: `ngrok`(Go) → `localTunnel()`，protocol="local_web"，选择最优节点 → 调用设备 `nezha_ngrok` 方法 → 返回 HTTPS URL + token
- **备注**: 关闭时调用 DELETE tunnels/{id}

### POST /api/v1/ngrok/devices/{id}/local-cli — 创建 CLI 会话 ✅ `incloud tunnel open-cli`

- **作用**: 为设备创建远程命令行访问隧道
- **使用背景**: 设备列表/详情页"远程访问" → "Device CLI"，支持最多 3 个并发会话
- **前端调用**: `LocalCliModal` + `NezhaCli` 组件 → `createCliSession()`，支持多标签页、日志下载
- **后端实现**: `ngrok`(Go) → `localTunnel()`，protocol="local_cli"
- **事件**: 创建时通过 RabbitMQ 发布 `TunnelCreatedEvent`，`nezha-network` 记录到 `ngrok.tunnel.logs`

### DELETE /api/v1/ngrok/tunnels/{id} — 关闭隧道 ✅ `incloud tunnel close`

- **作用**: 关闭指定隧道，释放资源
- **使用背景**: 用户关闭 Web/CLI 会话或隧道超时/失败时
- **前端调用**: `Ngrok`/`NezhaCli` 组件关闭时调用
- **后端实现**: `ngrok`(Go) → `closeTunnel()`，触发关闭事件更新 MongoDB 统计

### GET /api/v1/ngrok/tunnels/{id}/wait — 等待隧道就绪

- **作用**: 长轮询等待隧道关闭信号（timeout=10s）
- **使用背景**: Web UI 隧道建立后监听状态
- **前端调用**: `Ngrok` 组件 → `fetchTokenStatus()` 递归轮询直到 closed=true
- **后端实现**: `ngrok`(Go) → `waitTunnelClose()`，返回 `{result: {closed: boolean}}`

### GET /api/v1/ngrok/devices/{id}/logs — 隧道会话记录

- **作用**: 获取设备的隧道连接日志，支持按协议类型过滤和分页
- **使用背景**: OOBM/CLI 管理界面查看历史连接记录
- **前端调用**: `oobm/service.ts` → `fetchSessionRecords()`（按 protocol 过滤）、`local-cli-modal/service.ts`、`RecordsDrawer`、`HistoryConnectModal`
- **后端实现**: `nezha-network` → `NgrokTunnelController.getTunnelLogs()`，查询 `ngrok.tunnel.logs`（最近 30 天，60 天自动过期）
- **参数**: `protocols`（数组）、`businessId`、`type`（client/local）、`expand=creator`

---

## Config — 配置管理

> 后端微服务: `nezha-device-config`
> 配置分层: Default → Factory → Group → Device（优先级递增）

### DELETE /api/v1/config — 丢弃配置会话

- **作用**: 丢弃/删除一个配置编辑会话，清除所有未提交的配置修改
- **使用背景**: 配置编辑界面点击"放弃修改"或"关闭"
- **前端调用**: `apps/network/src/pages/devices/components/RouterEditModal/service.ts` → `discardChanges(sessionId)`
- **后端实现**: `nezha-device-config` → `SessionController`(L97-101)，删除会话记录和 `config.copy` 表中的备份
- **备注**: 会话 24 小时自动过期（TTL 索引）

### GET /api/v1/config/pending — 获取待提交配置

- **作用**: 获取会话中所有待提交的配置修改（JSON diff/增量）
- **使用背景**: 配置预览或审查变更内容
- **前端调用**: `RouterEditModal/service.ts` → `fetchPendingConfig(sessionId)`
- **后端实现**: `nezha-device-config` → `SessionController`(L117-122)，通过 JsonDiff 比较当前 config 与备份 copy，返回 merge-patch 格式
- **备注**: 通过 `x-session-id` 请求头传递会话 ID

### DELETE /api/v1/config/layer/device/{id} — 删除设备配置层

- **作用**: 删除设备的配置层，恢复到默认/出厂配置
- **使用背景**: 设备配置管理中重置设备配置
- **前端调用**: `apps/network/src/pages/devices/list/service.ts` → `deleteDeviceConfig(id)`
- **后端实现**: `nezha-device-config` → `LayerConfigController`(L88-95)，检查设备影子状态 → 重置配置 → 发送更新消息 → 记录审计日志
- **参数**: 可选 `module` 指定特定模块

### GET /api/v1/config/layer/group/{id} — 获取设备组配置层

- **作用**: 获取设备组的配置层内容
- **使用背景**: 设备组配置管理页面查看或编辑组级配置
- **前端调用**: `apps/network/src/pages/groups/service.ts` → `fetchGroupConfigInfo(id)` + `RouterConfigModal/service.ts`
- **后端实现**: `nezha-device-config` → `LayerConfigController`(L61-68)，确保组默认配置存在后查询 `configs` 表
- **参数**: 可选 `module`

### DELETE /api/v1/config/layer/group/{id} — 删除设备组配置层

- **作用**: 删除设备组的配置层，恢复为默认配置
- **使用背景**: 设备组配置管理中清除组级定制
- **前端调用**: `apps/network/src/pages/groups/service.ts` → `deleteGroupsConfig(id)`
- **后端实现**: `nezha-device-config` → `LayerConfigController`(L97-104)，重置配置 → 更新所有组成员影子状态 → 审计日志 "devicegroup_config_cleared"
- **备注**: 影响该组内所有设备的配置继承

---

## Connectors — 配置同步

### POST /api/v1/connectors/send-config — 同步配置到 Connector

- **作用**: 向 Connector 设备下发 OpenVPN 配置（子网、终端、VIP）
- **使用背景**: Connector 网络管理 → 设备列表 → "同步配置"按钮（仅 root 角色可见）
- **前端调用**: `apps/network/src/pages/networks/connector/profile/devices/index.tsx` → `syncConfig([deviceId])`
- **后端实现**: `nezha-network` → `ConnectorNetworkController`(L143-160) + `OvpnConfigService.sendDeviceConfigForce()`，通过 MQTT（主题: `nezha/{deviceId}/connector`）发送 OvpnConfig
- **权限**: `@InternalApi(CONNECTORS_WRITE)`，前端需要 `has_role_root`
- **请求体**: `{ deviceId?: [ObjectId], networkId?: [ObjectId], retain?: boolean }`
- **触发点**: 手动按钮 / 每小时定时重试失败 / AccountEventListener 监听账户变更时自动下发
- **失败处理**: MQTT 失败 + retain=true → 尝试 retain=false 重发 → 仍失败标记 failed=true → 定时任务重试
