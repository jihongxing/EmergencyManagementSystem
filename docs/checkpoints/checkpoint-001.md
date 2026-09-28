# Checkpoint 001

日期：2026-09-29（Asia/Shanghai）。状态：工程阶段快照，不是生产发布。
恢复引用：Git annotated tag `checkpoint-001`；标签指向包含本记录及 pwsh 约束的 main 合并提交，使用 `git show checkpoint-001` 查看实际提交。
保存工作分支：`chore/checkpoint-001`；恢复基线：`main` 上的标签。

## 已保存范围

- SSOT、机器契约、工程治理及已批准的 P0-P6 路线图。
- P0-01 最小技术基线、P0-02 版本控制与实际 CI/主分支保护。
- P0-03 PostgreSQL 连接、存活/就绪分离、独立 Goose 迁移与本地集成测试。
- PR #1 已合并为 `8902f4e`；PR #2 已合并为 `4a42c9d`。
- AGENTS.md 强制使用 pwsh；项目脚本要求 PowerShell 7，CI 和当前操作说明同步，不回退 powershell.exe。
- Go/Web/单 Flutter App 工程壳和身份内部授权单元；没有真实登录、业务表、行政记录或支付实现。

## 验证证据

2026-09-29 使用 `pwsh -NoProfile -File scripts/test/preflight.ps1 -Database`，退出码 0：文档/策略、Go test/vet、真实 PostgreSQL 空库/迁移/事务回滚/重启保留与恢复就绪、Web 构建/测试、Flutter analyze/widget test 通过。PR #2 远程 preflight 已通过；本次 pwsh 更新在 checkpoint PR 中重新检查，成功后再合并和打标签。

## 恢复与剩余工作

下一任务 P0-04：完整机器契约门禁。之后 P0-05 双端原生构建，P0-06 阶段验收；不得跳过阶段直接扩展身份业务。

标签保存代码与文档，不保存数据库、证据、密钥或缓存。依赖按锁文件恢复；数据库操作见[运行说明](../database-development.md)，任务状态见[任务清单](../development-tasks.md)。本地 Podman 容器与测试库保留，恢复代码不会恢复或回滚数据库。

尚未完成：完整 OpenAPI 校验、远程 PostgreSQL 集成、Android/iOS 原生构建与真机验证、生产部署/签名/身份凭据/支付选型。Actions 旧运行时警告仍待维护。当前快照不能被称为 V1 业务完成或可正式上线。
