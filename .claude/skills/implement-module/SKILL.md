---
name: implement-module
description: |
  Use when implementing multiple CLI subcommands from a plan document, such as
  "实现 device module", "实现 plan 里的命令", "批量实现子命令",
  "implement the alert rule commands", or when referencing docs/*-module-plan.md
  with unchecked TODO items to implement.
---

# Implement Module

从计划文档（`docs/*.md`）批量实现 CLI 子命令。

## 不可跳过的约束

- **必须调真实 API 验证响应结构** — 源码与运行时经常不一致（字段重命名、枚举大小写、运行时填充字段）
- **先亲自实现最复杂的命令，再分发并行** — 避免设计错误导致全员返工
- **子 agent 只创建新文件，不修改已有文件** — 你统一注册到 parent command，避免冲突
- **响应结构以 API 为准，请求参数以 Probe/DTO 源码为准** — 两者互补
- **测试数据必须清理**

---

## Phase 0: 评估规模

读计划文档，判断未完成（`- [ ]`）命令的数量和复杂度，决定：

- **是否值得并行**：命令少或彼此有依赖时，顺序实现更快，跳过 Phase 2
- **是否需要 worktree 隔离**：改动范围大、涉及新建模块骨架时，用 `/superpowers:using-git-worktrees` 创建隔离分支

## Phase 1: 调研与设计

1. 从计划文档提取待实现命令清单
2. **亲自用 Read 读后端源码**（Controller、Probe/DTO、Service），不可仅依赖 Explore agent 摘要 — 摘要会丢失参数注解、枚举值、验证规则等关键细节。后端源码在 `~/Workspace/nezha/` 下对应微服务目录
3. **调 API 验证响应**，对比源码发现差异
4. 找同模块已有命令作为代码模板

> **⏸️ 向用户展示每个命令的设计方案（Use/Args/Flags/API 映射），等待确认。**

## Phase 1.5: 验证最复杂的命令

亲自实现最复杂的写操作（通常是 create），端到端验证：实现 → `make build` → 真实 API 调用 → 确认通过。

记录发现的意外（必填字段、类型特殊处理、字段映射差异等），有设计修正在此处完成。

## Phase 2: 并行 Agent 实现

按复杂度分组，用 Agent 工具 spawn 子 agent 并行实现剩余命令，每个 agent 负责 1–3 个命令文件，总数 ≤ 4。

子 agent 无法访问你的对话历史，spawn prompt 必须自包含。模板：

```
你要在 incloud-cli 项目中实现以下 CLI 命令。

## 命令规格
- Use: `<command> <args>`
- Short: "<描述>"
- Flags: --flag1 (type, default, 说明), ...
- API: <METHOD> /api/v1/<endpoint>

## 真实 API 响应示例
<从 Phase 1 截取的完整 JSON>

## 注意事项
<Phase 1.5 发现的意外、字段映射、必填字段等>

## 代码模板
读取 `internal/cmd/<module>/<similar_cmd>.go` 作为参照，遵循同样的
Options struct + NewCmd 函数 + factory.Factory + iostreams.FormatOutput 模式。

## 约束
- 只创建新文件，不修改已有文件
- `make build` 必须通过
```

## Phase 3: 集成验证

1. 在 parent command 中 `AddCommand` 注册新命令
2. `make build` + lint（`goimports -w -local github.com/inhandnet/incloud-cli <files> && golangci-lint run ./...`）
3. 逐命令验证：`--help` → 真实 API → table/json/yaml 输出
4. 有写操作时做全链路：create → get → update → get → delete
5. 清理测试数据
6. 更新计划文档勾选已完成项（`- [ ]` → `- [x]`）

## Phase 4: 完成（Worktree 模式）

如果 Phase 0 使用了 worktree，验证全部通过后用 `/superpowers:finishing-a-development-branch` 完成合并。
