# Billing 模块实现计划

> 基于 nezha-billing 和 link-manager 的 API 调研。
> 计费系统分两个域：设备许可证（nezha-billing）和 SIM 计费（link-manager）。
> CLI 聚焦设备许可证管理，SIM 计费为可选扩展。

## 子模块划分

| # | 子模块 | CLI 路径 | 用户可用端点 | 说明 |
|---|--------|---------|-------------|------|
| 1 | 许可证管理 | `incloud billing license` | ~14 | 许可证 CRUD、绑定/解绑、升级 |
| 2 | 许可证类型 | `incloud billing license-type` | ~4 | 类型查询 |
| 3 | 订单管理 | `incloud billing order` | ~8 | 订单列表、详情、收据 |
| 4 | 价格查询 | `incloud billing price` | ~3 | 价格列表 |
| 5 | 优惠券 | `incloud billing coupon` | ~4 | 优惠券管理 |
| 6 | 发票信息 | `incloud billing invoice-info` | ~2 | 发票信息查询/更新 |
| 7 | 协同到期 | `incloud billing coterm` | ~3 | 许可证对齐到期日 |

---

## TODO List

### 许可证管理 (`incloud billing license`)

- [ ] `billing license list` — 列出许可证（分页，支持 --org/--device/--type/--product/--status/--archived 过滤）
- [ ] `billing license get <id>` — 查看许可证详情
- [ ] `billing license delete <id>` — 删除许可证（自动解绑设备）
- [ ] `billing license attach <id> --device <deviceId>` — 绑定许可证到设备
- [ ] `billing license detach <id>` — 从设备解绑许可证
- [ ] `billing license attach --bulk` — 批量绑定许可证到设备
- [ ] `billing license detach --bulk` — 批量解绑许可证
- [ ] `billing license move --org <orgId>` — 移动许可证到其他组织
- [ ] `billing license preassign` — 预览许可证自动分配结果
- [ ] `billing license pre-upgrade` — 预览许可证升级（显示费用差额）
- [ ] `billing license upgrade` — 执行许可证升级
- [ ] `billing license status-summary` — 按状态汇总许可证数量
- [ ] `billing license type-summary` — 按类型汇总许可证数量

### 许可证类型 (`incloud billing license-type`)

- [ ] `billing license-type list` — 列出所有许可证类型
- [ ] `billing license-type get <slug>` — 查看许可证类型详情
- [ ] `billing license-type upgradable` — 查看可升级的许可证类型
- [ ] `billing license-type list --product <product>` — 按产品查看可用许可证类型

### 订单管理 (`incloud billing order`)

- [ ] `billing order list` — 列出订单（分页、过滤）
- [ ] `billing order get <id>` — 查看订单详情
- [ ] `billing order items <id>` — 查看订单明细
- [ ] `billing order create` — 创建订单（许可证类型 + 优惠券）
- [ ] `billing order receipt <id>` — 下载订单收据（PDF）
- [ ] `billing order apply-invoice <id>` — 提交开票申请
- [ ] `billing order issue-invoice <id>` — 设置发票已开具状态
- [ ] `billing order export` — 导出订单列表

### 价格查询 (`incloud billing price`)

- [ ] `billing price list --license-type <type>` — 查看许可证类型的价格列表（含优惠券折扣）
- [ ] `billing price list --license-types <types...>` — 批量查看多个类型的价格
- [ ] `billing price payment-methods` — 查看可用支付方式

### 优惠券 (`incloud billing coupon`)

- [ ] `billing coupon list` — 列出优惠券
- [ ] `billing coupon get <id>` — 查看优惠券详情
- [ ] `billing coupon create` — 创建优惠券（金额/百分比折扣）
- [ ] `billing coupon close <id>` — 关闭/停用优惠券

### 发票信息 (`incloud billing invoice-info`)

- [ ] `billing invoice-info get` — 查看组织发票信息
- [ ] `billing invoice-info update` — 更新组织发票/税务信息

### 协同到期 (`incloud billing coterm`)

- [ ] `billing coterm create` — 创建协同到期请求（对齐许可证过期日，生成补差价订单）
- [ ] `billing coterm get <id>` — 查看协同到期请求详情
- [ ] `billing coterm apply <id>` — 执行协同到期请求

---

## 不纳入 CLI 的功能

### 设备许可证域（nezha-billing）

| 功能 | 端点数 | 理由 |
|------|--------|------|
| 许可证创建 | 1 | InternalApi |
| 试用许可证创建 | 1 | InternalApi |
| 许可证对齐到期日（内部） | 1 | InternalApi |
| 许可证类型 CRUD | ~5 | InternalApi |
| 价格 CRUD | ~5 | InternalApi（查询可用） |
| 退款 | 1 | InternalApi |
| 支付引擎内部接口 | ~6 | InternalApi |
| 支付回调（Stripe/支付宝/微信） | ~5 | 外部入站回调 |
| 结算周期管理 | ~3 | InternalApi |
| 结算策略写入 | ~3 | InternalApi（查询可用） |
| 设备试用管理 | ~2 | InternalApi |
| 高级服务列表 | 1 | 可用但优先级低 |

### SIM 计费域（link-manager）— 整体暂不实现

| 功能 | 端点数 | 理由 |
|------|--------|------|
| SIM 发票管理 | ~12 | 独立业务域，可后续作为 `link` 模块 |
| SIM 后付费订单 | ~7 | 同上 |
| SIM 预付费订单 | ~13 | 同上 |
| SIM 套餐管理 | ~7 | 同上 |
| SIM 优惠券 | ~6 | 同上 |
| SIM 流量包 | ~8 | 同上 |
| 支付链接 | ~5 | 同上 |

## 备注

- 支付流程（Stripe/支付宝/微信支付）涉及浏览器跳转，CLI 不适合直接完成支付
- 订单创建后，可输出支付链接供用户在浏览器中完成支付
- SIM 计费（link-manager）为独立域，建议后续作为 `incloud link` 模块单独实现
