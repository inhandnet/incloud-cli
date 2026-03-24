---
name: dogfood
description: |
  Scenario-based usability testing (dogfooding) for CLI commands.
  Simulates a real network engineer using the CLI to solve operational problems,
  then reports usability issues and design suggestions. Trigger phrases include:
  "dogfood", "dogfooding", "测试场景", "跑场景测试", "验收测试", "场景验收",
  "测一下新命令", "test the new commands in real scenarios".
  Also triggered automatically at the end of implement-module Phase 3.
---

# Dogfooding — 场景化可用性测试

模拟真实网络工程师使用 CLI 完成运维任务，验证命令在实际场景中是否好用，发现问题后输出具体的改进建议。

## 模式

### 快速模式（implement-module 自动触发）

仅测试本次变更的命令，验证基本可用性。通常 1-2 个最相关的场景。

触发方式：implement-module Phase 3 后自动调用，或手动指定 `quick`：
- `/dogfood quick`
- "快速验收一下新加的命令"

### 完整模式（默认）

推断所有相关场景，完整验收。每个场景启动独立的 network-engineer agent 并行执行。

触发方式：
- `/dogfood`
- `/dogfood diagnostics` （指定场景）
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

### Step 3: 构造场景 prompt 并启动 network-engineer agent

**关键原则：给业务目标，不给命令清单。**

agent 应该像真实用户一样，带着业务需求去摸索 CLI，自己决定用什么命令、怎么组合。
如果你把命令步骤都列好了，agent 只会验证"这些命令能不能跑通"，永远发现不了
"用户根本想不到要用这个命令"或"这个流程需要 3 个命令才能完成但应该 1 个就够"。

prompt 模板：

```
你是一个 MSP 网络运维工程师，管理着分布在各地的几百台路由器。

## 你的任务
<用业务语言描述目标，不提任何 CLI 命令名>

例如：
- "你刚接手一个新客户，需要为他们的 200 台设备建立告警体系。核心站点要求
   离线 5 分钟内短信通知，普通站点 30 分钟邮件通知即可。蜂窝设备需要信号
   质量监控。"
- "昨晚客户投诉收到太多告警，你需要排查当前的告警规则配置，看看哪些规则
   阈值不合理，调整后验证。"

## 背景约束
<补充业务约束，让场景更真实>

例如：
- "客户有 3 个设备分组：headquarters（10 台）、branch（50 台）、remote-cellular（140 台）"
- "你只有 EMAIL 和 APP 两种通知渠道可用"
- "客户要求工作日 9:00-18:00 的告警才发短信，其余时间只发邮件"

## 操作指南
请先读取以下文件了解 CLI 用法：
- /Users/j3r0lin/Workspace/ai/incloud-skills/skills/incloud/SKILL.md
- /Users/j3r0lin/Workspace/ai/incloud-skills/skills/incloud/references/<scenario>.md（如有）

## 新增/变更的命令（供参考，不是操作清单）
<列出本次变更的命令，让 agent 知道重点关注什么，但不要列步骤>

## 要求
1. 像真实用户一样工作：先搞清楚有什么能力，再规划怎么配置，最后执行
2. 遇到不顺的地方如实记录——你第一直觉想做什么、实际能做什么、差距在哪
3. 关注流程断裂点：哪个环节需要的信息拿不到、哪个操作需要多步才能完成
4. 输出结构化验收报告
5. 清理测试数据
```

**自拟场景的构造原则**：
- 场景必须包含**决策点**（如"不同分组用不同阈值"），而非只有单一路径
- 场景必须包含**信息查询需求**（如"查看当前规则覆盖了哪些设备"），测试 list/get 的实用性
- 场景必须包含**修改需求**（如"调整阈值"），测试 update 流程是否顺畅
- 不要给 agent 具体的 CLI 命令或参数值，让它自己从 help 和 types 等发现命令中获取

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

如果用户决定修复，修复完成后可以用 `/dogfood quick` 快速复验。

---

## 注意事项

- **必须打真实 API**：场景测试的价值在于端到端验证，不做 mock
- **不自动修复**：只报告问题和建议，修不修由开发者决定
- **先确认 `make build` 通过**：确保测试的是最新构建
- **测试数据清理**：如果场景中创建了资源（设备、告警规则等），测试后清理
- **绝不给 agent 命令步骤清单**：prompt 中只描述业务目标和约束，让 agent 自己决定用什么命令。如果你把 `incloud alert rule create --type "disconnected,retention=600"` 这样的命令写进 prompt，agent 就只会机械执行而不会像真实用户一样思考"我该怎么做"
- **自拟场景必须有决策复杂度**：不是"创建 3 条规则"，而是"不同设备组需要不同策略，你来规划"
