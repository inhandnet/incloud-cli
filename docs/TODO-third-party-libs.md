# TODO: 引入三方库替代手写逻辑

## 1. [高] ~~引入 go-resty 封装 API 客户端层~~ ✅ 已完成

**实现**：在 `internal/api/rest.go` 中基于 resty 封装了 `APIClient`，提供 `Get`、`Post`、`Put`、`Delete`、`Upload`、`Do` 方法，统一错误处理和 JSON 序列化。`Factory.APIClient()` 自动注入 BaseURL + auth transport。

- [x] 封装 `APIClient`（`internal/api/rest.go`）
- [x] 所有 50+ 命令文件从 `HttpClient()` 迁移到 `APIClient()`
- [x] 删除 `Factory.HttpClient()` 方法
- [x] 更新所有测试适配新 API
- ~~迁移 TokenTransport 到 resty middleware~~ — 保留 `http.RoundTripper` 方式注入（resty 支持 `SetTransport`），无需改动

**收益**：每个命令减少 ~10 行样板代码；错误处理集中在 `APIClient.execute`；新增命令开发更快。

---

## 2. [高] 替换 TablePrinter 为 go-pretty

**现状**：`internal/iostreams/table.go` 手写了一个 ~70 行的 TablePrinter，仅支持：
- 按 `len(col)` 计算列宽（不支持 Unicode 宽字符，中文会错位）
- 固定两空格列间距
- TTY 模式下列宽对齐，非 TTY 输出 TSV

不支持：终端宽度感知、列截断/换行、数字右对齐、header 分隔线、自动省略过长内容。

**库**：`github.com/jedib0t/go-pretty/v6/table`（star 1.5k+）

**改动方案**：
- [x] 引入 `go-pretty/v6` 依赖
- [x] 重写 `internal/iostreams/table.go`：
  - 用 `table.NewWriter()` 替代手写的 `TablePrinter`
  - 配置自定义无边框样式（保持当前视觉风格）
  - 非 TTY 时保留 TSV 输出
- [x] `json_table.go` 中的 `renderArray` / `renderObject` 无需改动（TablePrinter API 保持兼容）
- [x] 验证中文字符显示效果（go-pretty 使用 runewidth 正确计算 Unicode 宽字符）

**预期收益**：修复中文字符错位；长字段自动截断不撑破终端；后续可轻松切换输出格式（CSV、HTML、Markdown）。

---

## 3. [中] 引入 gjson 替代手写 JSON 路径查询

**现状**：`internal/iostreams/json_table.go` 中手写了：
- `resolveField(obj, "actor.name")` — 手动 split dot path 并逐层 walk map（第 218-229 行）
- `unwrapResult(raw)` — 硬编码提取 `result` 字段（第 117-126 行）
- `flattenKeys(m, prefix)` — 递归收集叶子节点路径（第 233-248 行）

这些只支持简单的 `a.b.c` 路径，不支持数组索引、通配符、条件过滤。

**库**：`github.com/tidwall/gjson`（star 14k+）

**改动方案**：
- [x] 引入 `gjson` 依赖
- [x] 重写 `resolveField`：用 `gjson.Result.Get(path)` 替代手动 walk，删除旧函数
- [x] 重写 `unwrapResult`：用 `gjson.ParseBytes(data).Get("result")` 替代，删除旧函数
- [x] 重写 `flattenKeys`：用 `gjson.Result.ForEach` 遍历对象 key，删除旧函数
- [ ] 评估是否可以利用 gjson 的 multipath 特性（`{name,status,product}`）简化 `--fields` 实现
- [x] 更新 `json_table_test.go` 验证兼容性

**预期收益**：路径查询更健壮；为后续支持 `--query` 过滤（类似 `jq`）打下基础。

---

## 4. [低] 替换 JSON 着色为 colorjson

**现状**：`internal/iostreams/json.go:42-88` 用 4 个正则逐行匹配着色：
```go
jsonKeyRe    = regexp.MustCompile(`^(\s*)"([^"]+)":`)
jsonStringRe = regexp.MustCompile(`: "(.*)"(,?)$`)
jsonNumberRe = regexp.MustCompile(`: (-?\d+\.?\d*)(,?)$`)
jsonBoolRe   = regexp.MustCompile(`: (true|false|null)(,?)$`)
```

已知问题：
- `jsonStringRe` 用 `.*` 贪婪匹配，含引号的字符串值会出错（如 `"desc": "say \"hello\""`)
- 数组内的值不会被着色（只匹配 `: value` 模式）
- key 含冒号或特殊字符时匹配失败

**库**：`github.com/TylerBrock/colorjson`（star 700+，API 极简）

**改动方案**：
- [x] 引入 `colorjson` 依赖
- [x] 重写 `colorizeJSON` 函数：
  ```go
  f := colorjson.NewFormatter()
  f.Indent = 2
  // 配置颜色映射对应当前 termenv 色彩方案
  ```
- [x] 移除 4 个正则变量
- [x] 确认非 TTY 路径不受影响（`FormatJSON` 中非 TTY 走 `json.Compact`，不调用着色）

**预期收益**：修复转义字符和嵌套数组的着色问题；减少 ~40 行正则代码。

