---
name: test-scenarios
description: |
  Use after implementing CLI commands to run scenario-based acceptance testing.
  Simulates a real network engineer using the CLI to solve operational problems,
  then reports usability issues and design suggestions. Trigger phrases include:
  "测试场景", "跑场景测试", "test scenarios", "验收测试", "场景验收",
  "测一下新命令", "test the new commands in real scenarios".
  Also triggered automatically at the end of implement-module Phase 3.
---

# 场景化验收测试

模拟真实网络工程师使用 CLI 完成运维任务，验证命令在实际场景中是否好用，发现问题后输出具体的改进建议。

## 模式

### 快速模式（implement-module 自动触发）

仅测试本次变更的命令，验证基本可用性。通常 1-2 个最相关的场景。

触发方式：implement-module Phase 3 后自动调用，或手动指定 `quick`：
- `/test-scenarios quick`
- "快速验收一下新加的命令"

### 完整模式（默认）

推断所有相关场景，完整验收。每个场景启动独立的 network-engineer agent 并行执行。

触发方式：
- `/test-scenarios`
- `/test-scenarios diagnostics` （指定场景）
- "跑一下完整的场景测试"

---

## 工作流

### Step 1: 识别待测命令

**有明确指定时**：直接使用用户指定的命令或场景。

**未指定时**：从 git diff 推断：

```bash
# 找出最近变更涉及的命令文件
git diff --name-only HEAD~5 -- internal/cmd/
# 或对比某个 base
git diff --name-only main -- internal/cmd/
```

从变更的文件路径提取命令名（如 `internal/cmd/device/signal.go` → `device signal`）。

用 `./bin/incloud <cmd> --help` 确认命令存在且可用。

### Step 2: 推断相关场景

读取 incloud-skills 的场景文档：

```bash
# 主 skill 文档
cat /Users/j3r0lin/Workspace/ai/incloud-skills/skills/incloud/SKILL.md

# 所有 reference 场景文档
ls /Users/j3r0lin/Workspace/ai/incloud-skills/skills/incloud/references/
```

根据待测命令，判断它们涉及哪些运维场景。匹配逻辑：

1. **直接匹配**：reference 文档中显式使用了该命令 → 必须测
2. **间接关联**：命令属于某个功能域（如 `device signal` 属于诊断域）→ 优先测
3. **快速模式下**：只取最相关的 1-2 个场景

将匹配结果展示给用户确认：

```
待测命令：device config schema get, device config schema validate
匹配场景：
  1. ai-config-workflow（直接匹配）
  2. diagnostics（间接关联 — 配置相关排查）

是否调整？
```

### Step 3: 启动 network-engineer agent

为每个场景启动一个 network-engineer agent。prompt 模板：

```
你需要完成一个场景化验收测试。

## 场景
<场景名称和描述>

## 重点验收的命令
<本次需要重点关注的新增/修改命令列表>

## 场景文档
请先读取以下文件获取操作指南：
- /Users/j3r0lin/Workspace/ai/incloud-skills/skills/incloud/SKILL.md
- /Users/j3r0lin/Workspace/ai/incloud-skills/skills/incloud/references/<scenario>.md

## 要求
1. 按场景文档的思路完成完整的运维流程
2. 重点关注上述待验收命令的使用体验
3. 输出结构化验收报告（按 agent 定义中的报告格式）
4. 如果某些命令不存在或报错，如实记录，不要跳过
```

**并行策略**：
- 场景之间互相独立，可以并行启动多个 agent
- 快速模式下串行即可（只有 1-2 个场景）

### Step 4: 汇总报告

收集所有 agent 的验收报告，汇总为最终结论：

```markdown
# 场景验收汇总

## 测试范围
- 待测命令：<列表>
- 覆盖场景：<列表>

## 各场景结果
| 场景 | 结果 | 问题数 | 建议数 |
|------|------|--------|--------|
| diagnostics | ✅ PASS | 0 | 1 |
| fleet-inspection | ⚠️ WARN | 2 | 3 |

## 问题汇总
（合并各场景的问题项，去重）

## 设计建议汇总
（合并各场景的建议，去重，按优先级排序）
- [ ] P0: <阻塞性问题>
- [ ] P1: <体验问题>
- [ ] P2: <优化建议>

## 结论
<PASS: 所有场景通过 / WARN: 有体验问题但可用 / FAIL: 有阻塞性问题>
```

### Step 5: 反馈与跟进

- **PASS**：验收通过，可以继续后续流程（sync-skills、提交等）
- **WARN**：展示问题和建议，由用户决定是否修复后再继续
- **FAIL**：展示阻塞性问题，建议修复后重新验收

如果用户决定修复，修复完成后可以用 `/test-scenarios quick` 快速复验。

---

## 注意事项

- **必须打真实 API**：场景测试的价值在于端到端验证，不做 mock
- **不自动修复**：只报告问题和建议，修不修由开发者决定
- **先确认 `make build` 通过**：确保测试的是最新构建
- **测试数据清理**：如果场景中创建了资源（设备、告警规则等），测试后清理
- **没有场景文档的命令**：如果待测命令不属于任何已有的 reference 场景，agent 应自行构造一个合理的使用场景进行测试，并在报告中标注"自拟场景"
