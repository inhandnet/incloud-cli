---
name: sync-skills
description: |
  Use when CLI commands have been added, modified, or removed and the incloud-skills
  documentation needs to be updated to match. Trigger phrases include:
  "同步 skill", "更新 skill 文档", "sync skills", "同步 incloud-skills",
  "skill 文档需要更新", or after completing a feature that adds new CLI commands.
  Also trigger proactively when you notice CLI changes that aren't reflected
  in the incloud-skills repo.
---

# Sync CLI Changes to incloud-skills

当 incloud-cli 新增、修改或删除了命令后，同步更新 incloud-skills 仓库中的运维技能文档，确保 AI 工具通过 skill 获取的命令知识与 CLI 实际能力一致。

## 路径约定

- **CLI 项目**：当前目录（`/Users/j3r0lin/Workspace/nezha/incloud-cli`）
- **Skills 项目**：`/Users/j3r0lin/Workspace/ai/incloud-skills`
- **主 Skill 文件**：`skills/incloud/SKILL.md`
- **Reference 目录**：`skills/incloud/references/`

## 工作流

### Step 1: 识别 CLI 变更

确定本次需要同步的变更范围。信息来源（按优先级）：

1. **当前对话上下文** — 如果刚完成了功能开发，直接从对话中提取变更内容
2. **Git 历史** — 对比 incloud-skills 上次更新后的 CLI commit：
   ```bash
   # 查看 CLI 最近的命令相关 commit
   git log --oneline --since="<last-sync-date>" -- internal/cmd/
   ```
3. **CLI help 输出** — 用 `./bin/incloud <cmd> --help` 确认当前命令结构

归纳出变更清单：
- 新增了哪些命令/子命令
- 修改了哪些命令的 flag、参数、行为
- 删除/废弃了哪些命令

### Step 2: 评估影响范围

读取 incloud-skills 的当前内容，确定哪些文件需要更新：

```bash
# 快速查看当前 skill 结构
cat skills/incloud/SKILL.md
ls skills/incloud/references/
```

对照变更清单，判断每项变更影响哪些文件：

| 变更类型 | 可能需要更新的文件 |
|---------|------------------|
| 新增命令组（如 `config schema`） | SKILL.md 命令速查 + 新建/更新 reference |
| 新增子命令（如 `alert rule create`）| SKILL.md 命令速查 + 对应 reference |
| 修改 flag/参数 | 对应 reference 中的命令示例 |
| 修改输出格式 | 对应 reference 中的分析说明 |
| 废弃命令 | SKILL.md 命令速查 + 对应 reference |

### Step 3: 更新 SKILL.md

**命令速查区**是最常需要更新的部分。原则：

- 按已有的分组结构（设备、告警、固件、网络服务、平台）添加新命令
- 保持一行一条，注释简洁（中文，5-15 字）
- 新功能域如果不属于现有分组，新建分组

**description 区**：如果新增了全新的功能域（不只是现有功能的子命令），考虑在 description 中追加对应的触发关键词。但小范围增删无需改 description。

**核心能力区**：如果新增了之前未覆盖的能力类别，在"核心能力"列表中补充。

### Step 4: 更新 Reference 文档

判断变更是否需要新建 reference 文档或更新已有文档。

**新建 reference 的条件**：
- 新增了一个完整的功能域（有自己的工作流、多个命令协同）
- 该功能的使用指南超过了 SKILL.md 命令速查能承载的信息量

**更新已有 reference 的情况**：
- 命令名、flag 名变更 → 更新对应的命令示例
- 新增子命令 → 补充到对应流程中
- 输出格式变更 → 更新分析说明

**Reference 文档风格规范**（与已有文档保持一致）：
- 场景驱动：从用户要解决的问题出发，不是从命令出发
- 命令块 + 思路说明交替编排
- 中文为主，命令和参数保持英文原样
- 注意事项放在流程末尾

### Step 5: 验证一致性

更新完成后，做一次交叉检查：

1. SKILL.md 命令速查中列出的命令，在 CLI 中确实存在：
   ```bash
   # 抽查几个新增命令
   ./bin/incloud <new-cmd> --help
   ```
2. Reference 文档中的命令示例语法正确（flag 名、参数顺序）
3. 没有遗漏：对比变更清单，确认每项都已覆盖

### Step 6: 向用户汇报

展示变更摘要：
- 更新了哪些文件
- 每个文件改了什么（diff 摘要）
- 是否有需要用户决策的地方（如是否新建 reference）

等用户确认后，由用户决定是否提交。

## 注意事项

- **不要过度文档化**：skill 文档的目标是让 AI 工具能正确使用命令，不是替代 `--help`。重点写工作流和注意事项，不要逐个 flag 罗列。
- **保持 SKILL.md 精简**：命令速查区一行一命令，复杂内容放 reference。SKILL.md 是全量加载的，每多一行都增加 token 消耗。
- **reference 按场景组织，不按命令组织**：`diagnostics.md` 而非 `ping-command.md`。一个场景可能用到多个命令，一个命令可能出现在多个场景中。
- **更新 incloud-skills 后不要忘记同步 design doc 的状态**：如果 CLI 的 plan 文档中有"同步 skill"的 TODO 项，一并勾选。
