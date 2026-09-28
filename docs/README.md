# 文档地图

工程入口：[项目 README](../README.md)、[工程治理](engineering.md)、[路线图](../ROADMAP.md)。已建立最小工程骨架，完整验证进度以路线图为准。

开发执行入口：[详细任务分解](development-tasks.md)。路线图定义阶段与退出门槛，任务清单维护依赖、验收及状态；计划已批准，不替代产品 SSOT。P0-01 的已批准最小选型及保留边界见[工程基线](engineering-baseline.md)。

本目录区分**产品决定**、**实施时取得的外部依据**、**研究与假设**。截至 2026-09-28，首期仅冻结单县、单监管部门的最小软件闭环；尚无选定试点县、正式检查清单或可用于上线的本地授权材料。

| 类型 | 入口 | 用途 |
| --- | --- | --- |
| 治理 | [GOVERNANCE.md](GOVERNANCE.md) | 状态、优先级、批准与修订方法 |
| 决策索引 | [DECISIONS.md](DECISIONS.md) | 已冻结专题与来源映射；索引不是另一份规则正文 |
| 派生契约 | [../constras/README.md](../constras/README.md) | 将现行 SSOT 转为可验收边界，模块接口细节随开发补充 |
| 产品定位与范围 | [ssot/01-product.md](ssot/01-product.md) | 产品目标、首期与非目标 |
| 权威边界 | [ssot/02-authority.md](ssot/02-authority.md) | 对象、检查规则、周期及外部权威的责任边界 |
| 场所与身份 | [ssot/03-site-identity.md](ssot/03-site-identity.md) | 组织、实际场所、监管匹配、检查对象和二维码 |
| 用户与权限 | [ssot/04-access.md](ssot/04-access.md) | 最小角色、管理员开通、数据可见性 |
| 业务记录 | [ssot/05-workflows.md](ssot/05-workflows.md) | 企业自检、现场临时抽检、问题处置与历史 |
| 技术边界 | [ssot/06-technology.md](ssot/06-technology.md) | 已批准技术形态，不含具体框架或数据库产品 |
| 实施输入 | [implementation/README.md](implementation/README.md) | 试点部门应提供并核验的辖区、事项和程序资料 |
| 待决与待验证 | [OPEN.md](OPEN.md) | 尚非产品 SSOT 的设计问题、外部事实和商业假设 |
| 技术方案（部分批准） | [technical-design-v1.md](technical-design-v1.md) | P0-01 最小基线已批准，其余设计、供应商及发布选型仍待对应任务确认 |
| 研究记录 | [research/market-handover.md](research/market-handover.md) | 交接时的市场判断与来源，未经重新核验 |
| 历史归档 | [archive/应急安全检查数字化平台_V1_项目交接与冻结SSOT.md](archive/应急安全检查数字化平台_V1_项目交接与冻结SSOT.md) | 2026-09-28 完整交接快照，保留原貌供追溯 |

**阅读顺序：**先看 01 与 02，再根据任务读对应专题；涉及试点数据时同时看 `implementation/` 和 `OPEN.md`。不得仅凭历史快照、研究材料或候选实体表实现功能。

旧交接稿原路径保留了[迁移说明](ssot/应急安全检查数字化平台_V1_项目交接与冻结SSOT.md)。已有工程骨架和本地预检，业务模块及远程 CI 尚未运行；开发按 [文档 -> 契约 -> 代码](../AGENTS.md) 推进。V1 Markdown 行为契约及机器文件已建立，具体模块接口和数据契约随开发补充。
