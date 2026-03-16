# Product 模块实现计划

> 基于 nezha-iot 的 API 调研。产品管理绝大多数端点为 InternalApi（仅系统管理员可用）。
> CLI 用户主要使用产品查询和兼容性检查功能。

## 子模块划分

| # | 子模块 | CLI 路径 | 用户可用端点 | 说明 |
|---|--------|---------|-------------|------|
| 1 | 产品查询 | `incloud product` | ~3 | 列表、详情 |
| 2 | 产品兼容性 | `incloud product compat` | ~4 | 兼容性列表、验证 |
| 3 | 产品类型 | `incloud product type` | ~1 | 类型查询 |
| 4 | 产品面板 | `incloud product panel` | ~1 | UI 面板配置查询 |
| 5 | 设备硬件规格 | — | ~1 | 已归入 device 模块 |

---

## TODO List

### 产品查询 (`incloud product`)

- [ ] `product list` — 列出所有产品（分页、过滤）
- [ ] `product get <idOrName>` — 查看产品详情（支持 ID 或名称）
- [ ] `product license-types <product>` — 查看产品可用的许可证类型

### 产品兼容性 (`incloud product compat`)

- [ ] `product compat list` — 列出所有兼容性定义（如 App、插件）
- [ ] `product compat products <id>` — 查看某兼容性定义下各产品的支持状态
- [ ] `product compat check <deviceId> <compatId>` — 检查设备是否支持某兼容性（考虑固件版本）
- [ ] `product compat validate` — 批量验证设备/设备组对兼容性的支持

### 产品类型 (`incloud product type`)

- [ ] `product type get <idOrName>` — 查看产品类型详情

### 产品面板 (`incloud product panel`)

- [ ] `product panel get <idOrName>` — 查看产品 UI 面板配置

---

## 不纳入 CLI 的功能

| 功能 | 端点数 | 理由 |
|------|--------|------|
| 产品 CRUD（创建/更新/删除/发布/废弃） | ~8 | InternalApi |
| 产品类型 CRUD | ~4 | InternalApi（仅 get 对用户可见） |
| 产品型号（PN）管理 | ~6 | InternalApi |
| 产品特性管理 | ~4 | InternalApi |
| 硬件规格管理 | ~1 | InternalApi（设备端查询已归入 device） |
| 兼容性定义 CRUD | ~3 | InternalApi（查询可用） |
| 产品属性/事件/服务/方法（IoT 数据模型） | ~15 | InternalApi |
| MQTT Topic 管理 | ~4 | InternalApi |
| 面板配置写入 | ~1 | InternalApi |
| 序列号规则管理 | ~5 | InternalApi |
| SIM/Link 产品管理 | ~5 | 电信产品域，与设备产品无关 |

## 备注

- 产品模块用户可用端点较少（~10），大部分为只读查询
- 主要使用场景：查看产品信息、检查设备兼容性、查看许可证类型
- 如需更多产品管理功能，可后续增加 `--admin` 模式
