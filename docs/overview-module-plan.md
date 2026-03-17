# Overview 模块实现计划

> 全局运维概览，聚合跨模块统计数据，对标前端 Dashboard 页面。
> 后端服务：nezha-iot（设备统计）、nezha-alert（告警统计）、nezha-link（流量/离线统计）。

## 设计原则

### 定位：全局聚合，非单设备

overview 只放**跨设备/跨模块的全局统计**，单设备维度的查询留在各自模块（如 `device signal <id>`）。

### 单命令聚合输出

每个子命令聚合多个 API 的数据，一次输出完整视图，不再拆子命令。用 flag 控制细节（如 `--n` 控制排名条数）。

### 从 alert 模块迁移

`alert top devices` / `alert top types` 迁移至 `overview alerts`，alert 模块只保留告警 CRUD + 规则管理。

### 从 device 模块迁移

`device summary`、`device presence offline topn`、`device presence offline stats`、`device datausage topk` 迁移至 overview，device 模块只保留单设备维度的命令。

---

## 命令清单

### `incloud overview` — 综合摘要

> 无子命令时输出一屏关键指标，并发调用各 API。

输出内容：
- **设备**：总数 / 在线 / 离线 / 不活跃
- **告警**：活跃告警数 + Top 3 告警类型
- **流量**：近 24h 总流量（tx/rx）
- **离线**：Top 3 离线设备

参数：
- `--after / --before` — 自定义时间范围（默认近 24 小时）

API：
- `GET /api/v1/devices/summary`
- `GET /api/v1/alerts/stats`
- `GET /api/v1/alert/top-alert-types` (n=3)
- `GET /api/v1/datausage/overview`
- `GET /api/v1/devices/offline/topn` (topN=3)

TODO：
- [x] `overview` — 综合摘要（并发拉取，一屏输出）

---

### `incloud overview devices` — 设备状态分布

> 全局设备统计：在线/离线/不活跃数量及占比。

API：`GET /api/v1/devices/summary`

参数：
- `--fields` — 选择返回字段

TODO：
- [x] `overview devices` — 设备状态分布

---

### `incloud overview alerts` — 告警统计

> 聚合告警摘要 + Top-K 告警设备 + Top-K 告警类型，一个命令输出。

API：
- `GET /api/v1/alerts/stats` — 告警统计摘要
- `GET /api/v1/alert/top-alert-devices` — Top-K 告警最多的设备
- `GET /api/v1/alert/top-alert-types` — Top-K 告警类型排名

参数：
- `--after / --before` — 时间范围
- `--group` — 按设备组过滤
- `--n` — 排名条数（默认 10）

TODO：
- [x] `overview alerts` — 告警统计 + Top 设备 + Top 类型（从 `alert top` 迁移）

---

### `incloud overview traffic` — 流量概览

> 全局流量统计 + Top-K 流量设备。

API：
- `GET /api/v1/datausage/overview` — 流量时序（tx/rx/total）
- `GET /api/v1/datausage/topk` — Top-K 流量设备

参数：
- `--after / --before` — 时间范围
- `--type cellular|wifi|wired|all` — 流量类型（默认 all）
- `--n` — Top-K 条数（默认 10）

TODO：
- [x] `overview traffic` — 流量概览 + Top-K 设备（从 `device datausage topk` 迁移）

---

### `incloud overview offline` — 离线分析

> Top-N 离线设备 + 离线统计列表。

API：
- `GET /api/v1/devices/offline/topn` — 离线最多的设备排名
- `GET /api/v1/devices/offline/statistics` — 离线统计列表（分页）

参数：
- `--after / --before` — 时间范围
- `--group` — 按设备组过滤
- `--n` — Top-N 条数（默认 10）
- `--page / --limit` — 统计列表分页

TODO：
- [x] `overview offline` — 离线 Top-N + 统计列表（从 `device presence offline topn/stats` 迁移）

---

## 不纳入 overview 的功能

| 功能 | 理由 | 归属 |
|------|------|------|
| 单设备离线事件/每日汇总 | 单设备维度 | `device presence` |
| 单设备流量（hourly/daily/monthly） | 单设备维度 | `device datausage` |
| License/服务状态统计 | 独立业务域 | 未来 `billing` 模块 |
| SIM 卡统计 | 独立业务域 | 未来 `link` 模块 |
| 组织维度汇总 | 管理功能 | 未来 `org` 模块 |
| 时序指标通用查询 (`stats/{name}/data`) | 太底层 | 各命令内部调用 |
