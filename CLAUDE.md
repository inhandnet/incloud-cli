# incloud-cli 项目说明

## 构建与测试

```bash
make build    # 输出到 bin/incloud
make install  # 安装到 $GOPATH/bin
make test     # 运行测试（CGO_ENABLED=0）
make lint     # golangci-lint
```

> 注意：macOS darwin/amd64 必须 `CGO_ENABLED=0`，否则会遇到 dyld LC_UUID 错误。

### Lint

修改 Go 代码后手动运行确保通过：

```bash
goimports -w -local github.com/inhandnet/incloud-cli <changed files>
golangci-lint run ./...
```

## 命令设计约定

### 双用户群体：人类 + AI 工具

incloud-cli 的用户有两类：自然人（终端操作）和 AI 工具（如 Claude Code，通过 shell 调用 CLI 完成任务）。设计时需同时兼顾两者：

**输出格式**
- table 和 JSON 两种格式都可能被人类和 AI 使用，都要保证可读、结构稳定
- 确认信息写 stderr，数据写 stdout——方便管道处理

**错误信息**
- 错误消息要包含足够上下文让 AI 能自行诊断和重试，而非只说"操作失败"
- 包含具体的错误原因、相关资源 ID、建议的修复动作
- exit code 要有区分度：参数错误、认证失败、资源不存在、服务端错误用不同 code

**命令可发现性**
- `--help` 输出要完整：包含用法示例、参数说明、关联命令提示——AI 工具依赖 help 文本理解命令能力
- 子命令的 help 里引用相关命令时用完整路径（如 `incloud device list`），方便 AI 直接复制执行

**非交互性**
- 危险操作（delete、批量修改）需要 `--yes`/`-y` 跳过确认——AI 工具无法响应交互式 prompt，必须能通过 flag 跳过

**可组合性**
- 命令输出应适合管道组合：`incloud device list -o json | jq '.[] | .id'` 这类用法要能正常工作
- 单条资源查询（get/show）直接输出对象，列表查询输出数组，不要额外包装

### 时间参数

所有涉及时间范围过滤的命令统一使用 `--after / --before`（与后端 API 参数名一致）。
- 格式：ISO 8601（`2024-01-01T00:00:00`）
- 已有的 `--from/--to`（如 `alert list`）后续迁移为 `--after/--before`，保留 `--from/--to` 为隐藏别名

### 搜索参数

全文搜索 flag 统一使用 `--search`（短标志 `-q`），对应 API 参数 `q`。

### 命令层级

扁平优先，避免无动作的中间命名空间。子命令多时用 cobra command groups 分组展示，而非加嵌套层。
- 正确：`device signal <id>`、`device perf <id>`
- 避免：`device monitor signal <id>`（`monitor` 层无自身动作）
- 例外：子资源有自身 CRUD 时可嵌套（如 `device group list`、`alert rule create`）

### 跨资源 ID 引用提示

当 flag 引用了其他资源的 ID 且对应的 list 命令已存在时，应在 help 描述里用完整命令提示用户如何获取该 ID，例如：
- `"Role ID to assign (required; use 'incloud role list' to find IDs)"`
- `"Filter by device group ID (use 'incloud device group list' to find IDs)"`

### 分页参数

分页命令统一使用 `--page`（默认 1，1-based，发给 API 时减 1）、`--limit`（默认 20）、`--sort`（如 `createdAt,desc`）。

### 字段选择

- 领域命令（`device list`、`alert list` 等）使用 `--fields`/`-f` 控制返回字段，同时传给 API `fields` 参数减少传输量
- 通用 `api` 命令使用 `--column`/`-c` 做纯客户端列过滤（不传给 API）
- table 模式下若未指定 `--fields`，默认显示全部字段
- 仅当 API 返回字段过多、全显示不可读时，才定义 `defaultXxxFields` 控制 table 默认列（如 `device list` 返回 20+ 字段）

### 写操作反馈

所有写操作（create/update/delete）成功后必须在 stderr 输出确认信息，格式统一为：
- create: `<Resource> "<name>" created. (id: <id>)`
- update: `<Resource> "<name>" (<id>) updated.`
- delete: `<Resource> "<name>" (<id>) deleted.`

delete 前先 GET 拿到名称用于确认提示；create/update 从响应体解析名称和 ID。确认信息写 stderr，响应数据写 stdout，互不干扰管道。

### Top-K 参数

排名类命令统一使用 `--n`（默认 10）表示返回条数。

## 开发流程

- 实现了功能模块（或其中的子命令）后，必须同步更新 `docs/` 下对应的计划文档，将已完成的 TODO 项勾选（`- [ ]` → `- [x]`）。
- Worktree 分支合并回 main 后，必须在 main 上跑一次 `make build && make test` 确认构建和测试通过。
