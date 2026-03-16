# incloud CLI

InCloud IoT 设备管理平台的命令行工具，支持认证、多环境 context 切换、API 调用及多种输出格式。

## 安装

### 从源码构建

```bash
# 需要 Go 1.25+
make build    # 输出到 bin/incloud
make install  # 安装到 $GOPATH/bin
```

> macOS 下必须 `CGO_ENABLED=0` 构建（Makefile 已默认设置），否则可能遇到 dyld LC_UUID 错误。

### 跨平台构建

CI 会自动构建以下平台的二进制文件：

- `linux/amd64`、`linux/arm64`
- `darwin/amd64`、`darwin/arm64`
- `windows/amd64`

## 快速开始

### 1. 配置 Context

```bash
incloud config set-context dev --host https://portal.nezha.inhand.dev --org myorg
incloud config use-context dev
```

### 2. 登录

```bash
incloud auth login --context dev --host https://portal.nezha.inhand.dev
```

登录使用 OAuth 2.0 Authorization Code + PKCE 流程，会自动打开浏览器完成授权。

CLI 复用平台前端的 SPA OAuth client，登录时自动从平台 API 获取 `client_id` 和 `client_secret`，通过 `client_secret_post` 方式发送凭证（与前端行为一致）。也可通过 `--client-id` 手动指定。

### 3. 验证

```bash
incloud auth status
incloud api /api/v1/users/me
```

## 命令速查

### 认证

```bash
incloud auth login                    # 浏览器 OAuth 登录
incloud auth status                   # 查看当前认证状态
incloud auth logout                   # 登出
```

### Context 管理

```bash
incloud config set-context <name> --host <url> --org <org>
incloud config use-context <name>
incloud config current-context
incloud config list-contexts
incloud config delete-context <name>
```

### API 调用

```bash
incloud api /api/v1/users/me                            # GET 请求，彩色 JSON
incloud api /api/v1/devices -q page=0 -q limit=10       # 带 query params
incloud api /api/v1/devices -X POST -f name=test         # POST body fields
echo '{}' | incloud api /api/v1/devices -X POST --input - # 从 stdin 读取 JSON body
incloud api /api/v1/users/me -H "Sudo: user@example.com" # 自定义 header
```

### 全局 Flag

```bash
incloud --context prod api /api/v1/users/me              # 临时切换 context
incloud version                                          # 查看版本
```

## 输出格式

通过 `-o` 指定输出格式：

| 格式 | TTY 行为 | 管道行为 |
|------|---------|---------|
| `json`（默认） | 彩色 pretty JSON | 紧凑 JSON |
| `table` | 对齐表格 | TSV |
| `yaml` | YAML | YAML |

```bash
incloud api /api/v1/devices -o table                     # 表格输出
incloud api /api/v1/devices -o table -c name -c status   # 选定列
incloud api /api/v1/devices -o yaml                      # YAML 输出
```

## 环境变量

| 变量 | 作用 |
|------|------|
| `INCLOUD_CONTEXT` | 覆盖当前 context |
| `INCLOUD_HOST` | 覆盖 context 中的 host |
| `INCLOUD_TOKEN` | 覆盖 context 中的 token |

## 配置文件

路径：`~/.config/incloud/config.yaml`（权限 `0600`）

配置文件存储所有 context 信息（host、org、token 等），通过 `incloud config` 子命令管理。

## 开发指南

### 前置依赖

- Go 1.25+
- [golangci-lint](https://golangci-lint.run/)
- [lefthook](https://github.com/evilmartians/lefthook) — Git hooks 管理
- [goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports)

### 安装 Git Hooks

首次 clone 项目后，需要安装 lefthook：

```bash
lefthook install
```

这会注册 `pre-commit` hook，每次提交时自动执行：

1. **goimports** — 对暂存的 `.go` 文件执行 import 排序和格式化（本地包前缀 `github.com/inhandnet/incloud-cli`），并自动 re-stage 修复后的文件
2. **golangci-lint** — 对整个项目运行 lint 检查，不通过则阻止提交

### 构建 & 测试

```bash
make build    # 构建到 bin/incloud
make test     # 运行测试
make lint     # 运行 golangci-lint
make clean    # 清理构建产物
```

### Lint 规则

项目使用 golangci-lint v2，启用的 linter 包括：bodyclose、errcheck、gocritic、gosec、govet、ineffassign、misspell、noctx、staticcheck、unconvert、unused 等。格式化使用 gofmt + goimports。

详见 `.golangci.yml`。

### 项目结构

```
cmd/incloud/        # CLI 入口
internal/
  api/              # OAuth 认证、Token 传输
  cmd/              # 各子命令实现（api、auth、config、version）
  config/           # 配置文件读写、Context 模型
  factory/          # 依赖注入工厂
  iostreams/        # 终端输出、格式化（JSON/Table/YAML）
```
