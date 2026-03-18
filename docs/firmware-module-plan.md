# Firmware 模块实现计划

> 基于 nezha-device-manager 和 nezha-iot 的 API 调研。
> 固件写操作（CRUD/发布/废弃/上传）均为 @InternalApi，仅系统管理员可用；OTA 任务创建和固件查询为用户可用。

## 子模块划分

| #   | 子模块       | CLI 路径                  | 用户可用端点 | 说明                      |
| --- | ------------ | ------------------------- | ------------ | ------------------------- |
| 1   | 固件查询     | `incloud firmware`        | ~8           | 列表、详情、按产品查询    |
| 2   | OTA 任务     | `incloud firmware job`    | ~9           | 创建/查看/重试升级任务    |
| 3   | OTA 模块     | `incloud firmware module` | ~2           | 查看 OTA 模块定义         |
| 4   | 设备固件状态 | `incloud firmware status` | ~2           | 查看设备固件/OTA 升级状态 |

---

## TODO List

### 固件查询 (`incloud firmware`)

- [ ] `firmware list` — 列出所有固件（分页）
- [ ] `firmware get <id>` — 查看固件详情
- [ ] `firmware list --product <product>` — 按产品列出固件
- [ ] `firmware latest --product <product>` — 查看产品最新固件版本
- [ ] `firmware stats <id>` — 查看固件的设备升级统计（按状态计数）
- [ ] `firmware set-latest <id>` — 标记为最新版本
- [ ] `firmware set-order <id> --order <n>` — 设置显示排序

### OTA 任务 (`incloud firmware job`)

- [x] `firmware job list` — 列出所有 OTA 任务
- [x] `firmware job list --firmware <id>` — 按固件列出 OTA 任务
- [x] `firmware job create <firmwareId>` — 创建 OTA 升级任务（指定目标设备/组、调度时间、超时、重试）；使用 batch endpoint，支持 --version/--filter/--scheduled-at/--upgradable-start/--upgradable-end
      ~~- [ ] `firmware job create --bulk` — 批量创建 OTA 任务（跨多个固件/设备）~~ 不做这个功能，直接使用 `firmware job create` 创建多个任务
- [x] `firmware job executions` — 列出所有 OTA 任务执行记录；支持 --status/--job/--sn/--module/--expand 过滤
- [x] `firmware job executions --firmware <id>` — 按固件列出执行记录
- [x] `firmware job executions --device <id>` — 查看设备已完成的 OTA 记录
      ~~- [ ] `firmware job next --device <id>` — 查看设备下一个待执行的 OTA 任务~~ 不做这个功能
- [x] `firmware job cancel <jobId>` — 取消 OTA 任务
- [x] `firmware job executions cancel <executionId>` — 取消单个 OTA 执行
- [x] `firmware job executions retry <executionId>` — 重试失败的 OTA 执行

### OTA 模块 (`incloud firmware module`)

- [ ] `firmware module list` — 列出 OTA 模块定义（按产品过滤）
- [ ] `firmware module get <id>` — 查看 OTA 模块详情

### 设备固件状态 (`incloud firmware status`)

- [x] `firmware status` — 列出设备固件升级状态（支持 --device/--product/--module/--status/--version/--expand 过滤）
      - 默认调用 `GET /device/firmwares`（列出所有设备默认模块状态）
      - 指定 `--device <id>` 时调用 `GET /devices/{id}/ota/modules`（列出该设备所有模块状态）

---

## 不纳入 CLI 的端点（@InternalApi）

| 功能               | 端点示例                                         | 理由           |
| ------------------ | ------------------------------------------------ | -------------- |
| 固件创建/更新/删除 | `POST/PUT/DELETE /firmwares/{id}`                | 系统管理员专用 |
| 固件发布/废弃      | `PUT /firmwares/{id}/publish\|deprecate`         | 系统管理员专用 |
| 固件上传/下载      | `POST /firmwares/upload`, `GET .../download`     | 系统管理员专用 |
| Delta 包管理       | `POST/DELETE /firmwares/{id}/delta-packages/...` | 系统管理员专用 |
| Full 包管理        | `POST/DELETE /firmwares/{id}/full-package`       | 系统管理员专用 |
| OTA 模块 CRUD      | `POST/PUT/DELETE /ota/modules/{id}`              | 系统管理员专用 |
| 全局统计           | `GET /firmwares/global-summary`                  | 系统管理员专用 |
| 配置页面代理       | `GET /config/static/...`                         | UI 专用        |

## 备注

- 固件写操作全部是 InternalApi，CLI 用户主要使用查询 + OTA 任务功能
- OTA 任务创建是核心用户场景（批量升级设备固件）
- 可考虑后续增加 `--admin` 模式暴露 InternalApi 操作
