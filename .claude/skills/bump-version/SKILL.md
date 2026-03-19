---
name: bump-version
description: |
  项目版本发布：生成中文 changelog、打 tag、创建 Release。
  当用户说"发版"、"bump version"、"打 tag"、"创建 release"、"发布新版本"、
  "升级版本号"、"准备发布 vX.Y.Z"时触发。
  即使用户只是提到版本号（如 "v0.2.0"）并暗示要发布，也应触发。
---

# bump-version

发布项目新版本。

## 流程

1. 确认工作区干净
2. 查看 `git tag` 和自上个 tag 以来的提交，建议版本号，让用户确认
3. 生成中文 changelog：如果有 `/changelog-generator` skill 则用它，否则自行从提交历史生成。展示给用户确认
4. 将 changelog 写入 CHANGELOG.md（追加在文件顶部，保留历史版本记录），提交
5. 创建 annotated tag 并 push（push 会触发 CI，CI 读取 CHANGELOG.md 自动创建带二进制附件的 Release）
6. 等待 CI pipeline 完成，确认构建通过

## CHANGELOG.md 格式

每个版本用一级标题 `# vX.Y.Z` 分隔，CI 通过 sed 提取对应版本的内容作为 release description。

## 要点

- Changelog 用中文写，技术术语保持英文
- 把 commit message 转化为用户能理解的变更描述，不是直接搬 commit
