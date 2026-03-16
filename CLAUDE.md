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
