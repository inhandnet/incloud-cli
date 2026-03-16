# incloud-cli 项目说明

## 构建与安装

```bash
make build    # 输出到 bin/incloud
make install  # 安装到 $GOPATH/bin
make test     # 运行测试（CGO_ENABLED=0）
make lint     # golangci-lint
```

> 注意：macOS darwin/amd64 必须 `CGO_ENABLED=0`，否则会遇到 dyld LC_UUID 错误。

## 命令速查

```bash
# 认证
incloud auth login --context dev --host https://portal.nezha.inhand.dev
incloud auth status
incloud auth logout

# Context 管理
incloud config set-context dev --host https://portal.nezha.inhand.dev --org myorg
incloud config use-context dev
incloud config current-context
incloud config list-contexts
incloud config delete-context old

# API 调用
incloud api /api/v1/users/me                          # 默认 GET，彩色 JSON
incloud api /api/v1/devices -q page=0 -q limit=10     # 带 query params
incloud api /api/v1/devices -X POST -f name=test      # POST body fields
echo '{}' | incloud api /api/v1/devices -X POST --input -  # stdin JSON
incloud api /api/v1/users/me -H "Sudo: user@example.com"   # 自定义 header

# 输出格式
incloud api /api/v1/devices -o json                    # JSON（TTY: pretty, 管道: 紧凑）
incloud api /api/v1/devices -o table                   # 表格（TTY: 对齐列, 管道: TSV）
incloud api /api/v1/devices -o table -c name -c status # 表格选定列
incloud api /api/v1/devices -o yaml                    # YAML 格式

# 全局 flag
incloud --context prod api /api/v1/users/me            # 临时切换 context
incloud version
```

## 环境变量

| 变量 | 作用 |
|------|------|
| `INCLOUD_CONTEXT` | 覆盖 current-context |
| `INCLOUD_HOST` | 覆盖 context 中的 host |
| `INCLOUD_TOKEN` | 覆盖 context 中的 token |

## 配置文件

路径：`~/.config/incloud/config.yaml`（权限 0600）

## OAuth 认证机制

CLI 复用前端 SPA 的 Hydra OAuth2 client（`token_endpoint_auth_method: client_secret_post`），登录时自动从 `GET /api/v1/frontend/settings` 获取 `clientId` 和 `clientSecret`。SPA client 的 `redirect_uris` 中需包含 `http://localhost:18920/callback`。

## 命令设计约定

### 时间参数

所有涉及时间范围过滤的命令统一使用 `--after / --before`（与后端 API 参数名一致）。
- 格式：ISO 8601（`2024-01-01T00:00:00`）
- 已有的 `--from/--to`（如 `alert list`）后续迁移为 `--after/--before`，保留 `--from/--to` 为隐藏别名

### 命令层级

扁平优先，避免无动作的中间命名空间。子命令多时用 cobra command groups 分组展示，而非加嵌套层。
- 正确：`device signal <id>`、`device perf <id>`
- 避免：`device monitor signal <id>`（`monitor` 层无自身动作）
- 例外：子资源有自身 CRUD 时可嵌套（如 `device group list`、`alert rule create`）

### 分页参数

分页命令统一使用 `--page`（默认 0）、`--limit`（默认 20）、`--sort`（如 `createdAt,desc`）。

### Top-K 参数

排名类命令统一使用 `--n`（默认 10）表示返回条数。

## 开发流程

### 功能模块实现后更新文档

实现了功能模块（或其中的子命令）后，必须同步更新 `docs/` 下对应的计划文档，将已完成的 TODO 项勾选（`- [ ]` → `- [x]`）。
