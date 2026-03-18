# TODO

## [中] API Client 统一业务错误处理

**现状**：后端部分 API 在业务错误时仍返回 HTTP 200，但 body 中包含 `{"error":"resource_not_found","status":404,...}` 结构。`APIClient.execute` 只检查 HTTP 状态码（`resp.IsError()`），不检查响应体中的业务错误码，导致：
- `client.Get` 返回 `(body, nil)` 但 body 实际是错误
- 每个命令需要自行解析 body 判断是否有 `error` 字段（如 `group_delete.go` 中的处理）

**改动方案**：
- [ ] 在 `APIClient.execute` 中增加响应体业务错误检测：解析 body，如果存在 `error` 字段则返回结构化错误
- [ ] 定义 `APIError` 类型，包含 `Error`、`Status`、`Message`、`Path` 字段，便于调用方用 `errors.As` 判断
- [ ] 调用方可用 `var apiErr *api.APIError; errors.As(err, &apiErr)` 获取详细错误信息
- [ ] 迁移已有的 body 内错误检查逻辑到统一机制

**预期收益**：消除命令层重复的错误检测代码；所有业务错误统一处理；新增命令不需要关心后端的"200 但实际是错误"问题。

## [低] 简化 IOStreams 抽象

**现状**：`IOStreams` 从 gh CLI 照搬，将 `Out`/`ErrOut`/`In` 全部抽象为可注入的 `io.Writer`/`io.Reader`。但实际上唯一真正需要抽象的是 `IsStdoutTTY()`（决定 table vs TSV、是否着色），其余的 `Out`/`ErrOut` 注入属于 YAGNI——测试中可以用 `os.Pipe()` 或只验证 API 调用正确性，不需要注入 buffer。

**改动方向**：
- [ ] 评估将 `Out`/`ErrOut` 改回直接用 `os.Stdout`/`os.Stderr`，仅保留 TTY 检测抽象
- [ ] 测试改为验证 API 调用参数（mock server 断言）而非捕获输出内容
- [ ] 如果保留注入，至少去掉 `In`（从未使用过）
