# V1 契约索引

状态：`DERIVED CONTRACT`。本目录将[现行 SSOT](../docs/DECISIONS.md)转换为可验收的行为、数据和交互约束，不是新的产品权威，也不包含试点行政规则。若与 SSOT 冲突，以专题 SSOT 为准，暂停冲突实现并按[治理规则](../docs/GOVERNANCE.md)处理。研究、归档和本目录都不能独立授予新功能或行政权限。

| 契约 | 责任范围 | 权威来源 |
| --- | --- | --- |
| [01-scope-and-identity.md](01-scope-and-identity.md) | V1 边界、组织/场所/对象身份、二维码和监管绑定 | [产品](../docs/ssot/01-product.md)、[场所](../docs/ssot/03-site-identity.md) |
| [02-access-and-evidence.md](02-access-and-evidence.md) | 开通、角色、逐次授权和证据可见性 | [权限](../docs/ssot/04-access.md)、[权威](../docs/ssot/02-authority.md) |
| [03-rules-and-records.md](03-rules-and-records.md) | 规则版本、自检和行政记录、整改与历史 | [权威](../docs/ssot/02-authority.md)、[双轨流程](../docs/ssot/05-workflows.md) |
| [04-subscription.md](04-subscription.md) | 免费/付费权益、购买与到期 | [产品](../docs/ssot/01-product.md) |
| [05-platform.md](05-platform.md) | 客户端、服务端、存储及在线提交边界 | [技术](../docs/ssot/06-technology.md) |

## 机器可读判定

同名 JSON 与 Markdown 并列存放：[01](01-scope-and-identity.json)、[02](02-access-and-evidence.json)、[03](03-rules-and-records.json)、[04](04-subscription.json)、[05](05-platform.json)。Markdown 保留完整依据、禁止行为与验收场景；JSON 是其派生的可执行判定子集，不是另一份产品权威，也不表示未列出的条款已被机器证明。`manualClauses` 明确列出尚需人工或后续模块测试验收的条款；同一条款可同时有机器判定与人工要求。

每项 `policy` 列出所依据的 `clauses`、布尔输入 `facts`、准入条件 `allowWhen`，以及 `cases` 中的允许/拒绝样例。只有所有列明事实都与准入条件相符才允许；事实缺失、待核验或不确定时，业务实现必须按不满足条件处理，不得自行补成 `true`。这些事实是已经核验后的判定输入，JSON 不负责判断材料真伪，也不确定实际县域、行政职权或外部事项。输出 `allowed` 只代表该项操作满足所列准入条件，不代表全部业务权限均已获得；组合动作仍需逐项核验。

`enterprise-executor-record-access` 和 `free-site-workflow` 仅描述已分配场所的单位执行人员；单位管理员的本组织权限仍以 [权限契约](02-access-and-evidence.md) 为准。`official-inspection-create` 中的现场扫码事实须是有效二维码与现场对象匹配后的核验结果，不能由客户端自报。

运行 [治理校验](../scripts/test/check-doc-governance.ps1) 会解析每份 JSON，检查来源条款覆盖、规则结构，并执行正反案例。新增模块时须将这些条件落实为服务端行为、数据校验和实际自动化测试；这里的策略案例不能替代真实 API、数据库或客户端测试。具体接口和存储契约随模块开发补齐。

## 使用约定

P0-03 数据库连接与迁移见[基础契约](platform/database.md)和[就绪 OpenAPI](platform/ready.openapi.json)，Go HTTP 测试实际解析状态响应，专用数据库集成测试验证迁移及就绪。

身份模块首个内部授权单元见 [identity](identity/README.md)：机器案例直接由 Go 测试解析执行；不提供登录、账号开通或行政资格认定。

工程存活接口见 [live.openapi.json](platform/live.openapi.json)，派生于平台契约的 Go 单体运行形态，仅表示进程存活，不代表业务或行政依据就绪。由统一预检进行基本结构解析，由 Go HTTP 测试检查响应。

- 各条以 `Cxx` 编号，便于实现与测试逐条追溯；验收场景是行为断言，不指定 API 路径、表名或 UI 控件。`必须` 表示现行 SSOT 约束；`不得` 表示禁止行为。
- 实际县、部门、区划、职责、事项、共享范围及临时检查依据属于[实施输入](../docs/implementation/README.md)。未取得并核验时，契约要求待确认或阻断正式行政操作；不能用测试数据冒充正式依据。
- 具体接口、数据库迁移、字段权限映射和支付适配器应在对应功能编码前补充模块契约，并标明其依据的 `Cxx` 条目。纯内部重构无需预建新契约；对外或跨模块行为变更则先修改契约及测试，若触及产品决定须先获批修改 SSOT。
- 首次实现可使用隔离的测试资料验证成功路径；测试环境必须明确标识，不得让模拟的行政确认、权限或检查结果流入正式业务。
