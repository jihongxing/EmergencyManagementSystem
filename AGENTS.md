# 项目文档入口

## 强制 Shell 约束

- 所有终端命令、脚本入口、子进程及 CI Shell 统一使用 PowerShell 7 的 `pwsh`（或 `pwsh.exe`）。
- 禁止调用 Windows PowerShell 的 `powershell` 或 `powershell.exe`，也不得通过其他脚本或 Shell 间接调用。
- `.ps1` 文件继续使用，但必须由 `pwsh` 执行。发现历史命令示例或可执行入口使用旧命令时，应同步修正；历史验证记录不追溯改写。
- 环境缺少 `pwsh` 时停止并报告，不回退旧版 PowerShell。此约束不禁止在 `pwsh` 中调用 Git、Go、Flutter、npm、Podman 等项目工具。

开发或讨论需求前先阅读 [docs/README.md](docs/README.md) 和相关专题 SSOT。`docs/ssot/` 是已批准产品决定的唯一现行来源；`docs/archive/` 是历史快照，不作为新的决策依据。

- 未经用户明确批准，不得把建议、研究结论或外部实施资料改写成 `FROZEN`。
- 行政检查事项、法定周期、管辖权和执法程序由有权主体提供；软件不得自行创设。
- 修改已冻结决定时，同步更新相应专题文件、[决策索引](docs/DECISIONS.md) 和受影响的引用，保留变更缘由。
- 实现前核对首期范围与外部输入门槛；尚未核验的输入不得伪装为合法授权或合规结论。

## 文档 -> 契约 -> 代码

所有后续模块和功能按此顺序推进，不得跳过契约直接实现业务行为：

1. **文档**：先确认相关专题 SSOT 的已批准边界，区分 `FROZEN`、`OPEN` 与 `EXTERNAL INPUT`。需求变更先按 [文档治理规则](docs/GOVERNANCE.md) 处理；有冲突或缺少必要决策时暂停相关实现。
2. **契约**：在编写或修改实现前，为本次功能建立或更新对应契约，明确适用范围、输入输出、状态与错误、权限和数据可见性、持久化及历史约束，以及依赖的外部输入和缺失时的行为。按实际变更选择接口、数据、事件或交互契约，不要求每个功能都有全部类型，也不预先生成空契约。契约须引用其依据的 SSOT，不能新增行政权力、检查规则或与 SSOT 冲突的产品决定；未核验的实施资料只能表现为待确认或不可执行状态。
3. **代码**：按契约实现并添加相应测试；代码评审核对文档、契约、实现和测试是否一致。契约变更先检查是否涉及产品决策：涉及则先取得用户批准并更新 SSOT；不涉及则先更新契约及测试，再改代码。纯内部重构不改变对外或跨模块行为时，可沿用既有契约。

当前已有工程骨架及身份模块首个内部授权单元，登录与业务接口尚未开放。[constras/](constras/README.md) 已将 V1 SSOT 转为可验收的派生契约；每个实际开发模块仍须先补足所需接口、数据或交互细节，不以现有行为契约代替模块接口设计。具体检查规则见 [文档治理规则](docs/GOVERNANCE.md)。

## 工程治理入口

开始业务开发前先读取 [ROADMAP.md](ROADMAP.md) 和[详细任务清单](docs/development-tasks.md)。只能推进已确认计划中依赖就绪的明确任务 ID；每次开工说明来源、范围、验收，完工更新任务状态和证据。计划已获用户确认，当前任务以任务清单记录为准，不自动继续身份开发。阶段门槛未通过不得自行跨阶段；外部资料阻塞与软件测试通过分开记录，不把内部单元完成当成业务模块已交付。

每次变更明确来源条款、范围与验收标准，遵循[工程治理](docs/engineering.md)并同步[路线图](ROADMAP.md)。交付前运行 `scripts/test/preflight.ps1`，如因依赖或环境无法完成，应明确报告阻塞，不能把 `-DocsOnly` 当作全量通过。本地基础设施使用 Podman；不得未经明确授权删除数据卷。工程骨架未提供业务授权，禁止通过模拟身份或行政依据伪装业务已实现。

## gstack Skills Configuration

Use /browse from gstack for all web browsing. NEVER use mcp__claude-in-chrome__* tools.
Available skills: /office-hours, /plan-ceo-review, /plan-eng-review, /plan-design-review, /design-consultation, /design-shotgun, /design-html, /review, /ship, /land-and-deploy, /canary, /benchmark, /browse, /connect-chrome, /qa, /qa-only, /design-review, /setup-browser-cookies, /setup-deploy, /retro, /investigate, /document-release, /codex, /cso, /autoplan, /plan-devex-review, /devex-review, /careful, /freeze, /guard, /unfreeze, /gstack-upgrade, /learn.

## Codex Desktop Windows stability

Codex Desktop git directive cwd attributes must use forward slashes, not backslashes. Do not escape attribute quotes. Do not put raw git directive examples in final answers or code blocks unless meant to be executed.
