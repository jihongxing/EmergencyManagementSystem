# V1 详细任务清单

状态：`APPROVED DELIVERY PLAN`。日期：2026-09-28。用户已确认本计划并指定从 P0-01 开始。与[路线图](../ROADMAP.md)配套；本文件是任务状态的唯一记录处，不替代[SSOT](DECISIONS.md)、[契约](../constras/README.md)或[外部实施资料](implementation/README.md)。

## 使用方式

状态：`TODO` 未开始；`PARTIAL` 有部分成果但未达到验收；`DOING` 正在执行；`BLOCKED` 已明确阻塞；`DONE` 验收与证据齐备。当前计划已确认，按任务依赖推进，不自动跳过选型或阶段批准。尚未开始但有外部前置条件的任务仍记 TODO，进入执行后缺少输入才记 BLOCKED。

表中依赖 ID 是开始实现前的必要条件；“阶段门槛”同时适用。所有业务任务都必须交付相应模块契约、实现、自动测试与涉及的操作文档。责任以“工程/用户/实施方”区分，实际负责人在开工时登记；不假设存在额外团队或授权并行代理。

## P0 工程基线收口

依据：平台契约 C05、[技术方案](technical-design-v1.md)、[工程治理](engineering.md)。目标是可复现和可验证，不增加业务能力。

| ID / 状态 | 任务与依赖 | 交付及验收标准 |
| --- | --- | --- |
| P0-01 / DONE | 方案与工程基线对齐；用户已确认最小选型 | [基线核对](engineering-baseline.md)完成；技术 SSOT、决策索引、OPEN、技术方案及平台契约已同步，保留待决与发布边界 |
| P0-02 / DONE | 版本控制与实际 CI；用户已指定 GitHub 仓库 | 基线已推送，实际 CI 运行 36445834084 成功；main 要求 PR、最新分支与 preflight，通过 API 回读确认保护生效；不包含原生构建 |
| P0-03 / DONE | 本地与测试环境可复现；依赖 P0-01 | 已提供[运行说明](database-development.md)、连接/就绪契约与独立迁移，实际 PostgreSQL 集成及重启保留验证通过；无业务表，远程 CI 数据库集成尚未接入 |
| P0-04 / DONE | 机器契约门禁；依赖 P0-01 | 判定 JSON、身份案例按 JSON Schema 及语义校验；OpenAPI 3.1 规范/引用与 Go 实际探针响应按 Schema 校验；损坏、未知字段、错误案例、断引用等负测会使门禁失败 |
| P0-05 / DONE | 双端工程验证；依赖 P0-01 | 保留已有 Flutter analyze/widget 测试；实际 Android debug 构建、macOS iOS 无签名构建与至少基础启动验证；缺少 macOS 时记录阻塞，不以原生目录存在代替构建 |
| P0-06 / DONE | 基线验收；依赖 P0-02 至 P0-05 | 从干净检出按 README 安装并预检，记录工具版本、命令、CI 与双端证据；更新路线图，请用户验收 |

已有证据：本地和干净检出 preflight 均已通过，Web 构建/SSR 测试、Go test/vet、Flutter analyze/widget 测试、原生目录及锁文件存在；P0-02 至 P0-05 的远程 CI、数据库独立验证和双端原生证据已齐备。P0-06 已完成工程验收，P0 阶段仍等待用户确认后才进入 P1。

## P1 身份与最小组织

依据：C02-01 至 C02-04、C02-07、C05；进入条件 P0-06。交付一个真实的账号与授权链路，不扩成多级组织架构。

| ID / 状态 | 任务与依赖 | 交付及验收标准 |
| --- | --- | --- |
| P1-01 / DONE | 身份/组织模块设计与契约；依赖 P0-06 | 明确登录凭据、激活/重置/失效、会话、首位管理员操作入口、四角色与兼任边界；OpenAPI、数据约束、错误/状态案例齐备；凭据选择待确认，不暗中引入公众注册 |
| P1-02 / PARTIAL | 组织及首位管理员开通；依赖 P1-01、P0-03 | 持久化组织/成员及线下材料来源、开通人；受控实施入口不成为日常第五角色；重复开通防护、更换核验留痕；测试材料只在隔离环境 |
| P1-03 / PARTIAL | 会话与逐次身份校验；依赖 P1-02 | 已完成服务层、PostgreSQL 存储、HTTP 最小入口、迁移和软件侧测试；仍待专用 PostgreSQL 集成与端到端客户端验收 |
| P1-04 / PARTIAL | 本方成员管理与审计；依赖 P1-03 | 已完成成员 API、数据库事务、审计表、角色/范围校验和软件侧回归测试；仍待专用 PostgreSQL 集成、HTTP 端到端和客户端联调验收 |
| P1-05 / PARTIAL | Web/App 身份入口；依赖 P1-03、P1-04 | 已完成入口契约、Web 身份驱动壳、Flutter 单 App 最小工作区状态和软件侧测试；仍待真实服务端联调、Flutter 网络/安全存储接入及撤权端到端验收 |
| P1-06 / PARTIAL | 最小组织阶段验收；依赖 P1-05 | 已完成两个单位与一个部门的隔离验收测试、管理员兼任、跨组织拒绝、撤权即时生效和审计证据；仍待真实 PostgreSQL、Web/App 联调和完整阶段演示 |

已有成果仅为 [identity 内部契约](../constras/identity/README.md)与 Go 内部判定测试；P1-04 不因该函数已完成而标 DONE。

## P1-01 执行记录

- 完成日期：2026-10-03，Codex；依赖 P0-06 已由用户正式验收通过。本任务只交付身份/组织模块设计与契约，不实现登录、数据库业务表、账号开通或客户端工作区。
- 契约交付：[模块边界与验收说明](../constras/identity/module.md)、[身份数据 Schema](../constras/identity/data.schema.json)、[生命周期案例](../constras/identity/lifecycle.json)、[身份 OpenAPI](../constras/identity/identity.openapi.json)。现有内部授权案例继续保留并纳入同一模块索引。
- 已确定的最小实现边界：不开放公众注册；组织类型仅为单位和部门；日常角色仍为四类；首位管理员由受控实施入口依据已核验线下材料建立；本方管理员管理本方成员；单位/部门管理员可分别兼任本方执行/检查角色；撤权后后续访问拒绝且历史署名保留。
- 凭据与会话：组织内唯一登录标识加密码；Web 服务端会话 Cookie；Flutter 短时访问凭据加可撤销刷新凭据；每次受保护请求重新核对服务端成员状态、组织和角色。恢复请求不泄露账号存在性。
- 机器门禁：扩展 `scripts/contracts/validate.mjs` 实际解析身份数据 Schema、生命周期 Schema 和身份 OpenAPI；扩展 `validate.test.mjs` 覆盖生命周期未知字段负测。文档治理、机器契约检查和契约测试均通过。
- 未冻结/未伪造内容：恢复渠道、密码策略强度、密钥托管、生产部署、实施人员凭据、行政资格材料及试点外部输入仍不在本任务内；OpenAPI 中受控实施安全方案只是接口边界，不表示实施凭据已经实现。
- 结果：P1-01 验收通过，下一步为 P1-02 组织及首位管理员开通；在用户明确推进前不自动开始 P1-02。

## P1-02 执行记录

- 日期/负责人：2026-10-03，Codex；用户已正式批准实施 P1-02。
- 依赖核验：P1-01 身份/组织模块契约已完成；P0-03 数据库连接、迁移和就绪基础已完成；当前未进入 P1-03 登录/会话实现。
- 契约：新增[组织及首位管理员开通契约](../constras/identity/bootstrap.md)及机器契约[bootstrap.json](../constras/identity/bootstrap.json)，新增专用 JSON Schema 并接入 `scripts/contracts` 实际解析；扩展身份 OpenAPI、组织数据形状和模块约束。
- 实现：新增 `ems.organizations`、`ems.members`、`ems.organization_materials`、`ems.identity_audit_events` 迁移；实现受控开通与首位管理员更换服务、内存验收存储及 PostgreSQL 事务存储。开通只创建 `pending` 首位管理员，不接收初始密码；材料核验、稳定外部标识、组织唯一性、旧管理员撤销和审计均在服务/数据库约束中执行。
- 测试与验收：
  - `pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/test/check-doc-governance.ps1`：通过。
  - `pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/test/check-machine-contracts.ps1`：通过。
  - `npm run check --prefix scripts/contracts`：通过，2 个契约测试通过。
  - `go test ./...`（工作目录 `backend`）：通过。
  - `go vet ./...`（工作目录 `backend`）：通过。
  - 数据库真实集成验证：本轮未设置专用 `TEST_DATABASE_URL`，未将单元测试冒充 PostgreSQL 集成通过；迁移文件已纳入 Goose，下一次提供隔离数据库后执行。
- 结果：P1-02 的契约、代码、单元测试和治理门禁完成；因本轮未提供专用 PostgreSQL 数据库，真实迁移/事务集成验收尚未完成，任务保持 `PARTIAL`。P1-03 登录、会话、凭据恢复和激活仍未开始；P1-06 阶段验收未完成。

## P2 场所、对象与二维码

依据：C01、C02-04、C05-04 至 C05-05；进入条件 P1-06。

| ID / 状态 | 任务与依赖 | 交付及验收标准 |
| --- | --- | --- |
| P2-01 / TODO | 场所/责任/对象数据和接口契约；依赖 P1-06 | 申报与行政确认分离、稳定 ID、实际地址、责任历史及对象生命周期；约束不把组织注册地址当场所地址、不写死消防对象 |
| P2-02 / TODO | 场所申报与变更；依赖 P2-01 | App 单位管理员申报、查看本组织场所、维护证明与历史；先落实证明材料的私有存储/受控引用契约及授权测试，存储方案未定则记录阻塞；未行政确认不阻断企业流程，不伪装已确认 |
| P2-03 / TODO | 对象登记与二维码生命周期；依赖 P2-02 | 稳定对象、码生成/补打/停用/更换、打印输出和 App 扫码；旧映射保留、停用码不能成为有效入口、码不固定检查表或授予权限 |
| P2-04 / TODO | 场所/对象人员分配；依赖 P2-02、P1-04 | 管理员在本组织分配执行范围；执行人员读写受当前分配限制；撤销分配后重放请求及猜测对象 ID 均拒绝 |
| P2-05 / TODO | 对象基础验收；依赖 P2-03、P2-04 | App 建场所 -> 建对象 -> 出码 -> 扫码 -> 范围内访问演示；地址/责任/二维码更改历史不被覆盖；不要求已有监管绑定 |

## P3 企业免费闭环

依据：C03-01 至 C03-05、C02-06、C04-02、C05-03 至 C05-05；进入条件 P2-05。

| ID / 状态 | 任务与依赖 | 交付及验收标准 |
| --- | --- | --- |
| P3-01 / TODO | 规则版本与双轨记录基础契约；依赖 P2-05 | 区分外部权威/内部安排/待确认；记录引用执行时版本、来源及对象；规定更正审计、结果/问题/复查状态及失败行为；不由软件规定法定周期 |
| P3-02 / TODO | 检查证据存储；依赖 P3-01、P2-02 | 复用 P2-02 的私有文件基础能力，补检查记录引用与上传/完成/读取协议、大小类型校验、孤立上传清理；下载逐次授权；已知文件键不能越权取件；材料用途不同也不能混用授权 |
| P3-03 / TODO | 自检记录与问题创建；依赖 P3-01、P3-02 | App 扫码、显示适用规则来源、提交自检并关联问题；服务端确认前不显示完成；幂等重试、提交超时查结果和规则更新历史测试 |
| P3-04 / TODO | 工作分配、整改与自主复查；依赖 P3-03、P2-04 | 本组织逐场所分配工作、整改证据、复查及需继续整改的路径；操作权限与原始记录关联保持一致；不是行政合格结论 |
| P3-05 / TODO | 历史与纠错；依赖 P3-04 | 免费逐场所查完整本组织历史，审计更正而非静默覆盖；对象关系与规则更新不改旧记录，撤权后记录保留署名但无新增访问权 |
| P3-06 / TODO | 免费闭环验收；依赖 P3-05 | 未订阅组织完成全链路，覆盖拒绝路径、网络中断、重复提交、证据越权及分配变化；不得设置必要人员、扫码、次数或二维码付费墙 |

## P4 自动监管绑定与行政闭环

依据：C01-05 至 C01-07、C02-05/06、C03-06 至 C03-08；进入条件 P3-06。正式资料由实施方获取并核验，工程仅定义载入、状态和执行方式。

| ID / 状态 | 任务与依赖 | 交付及验收标准 |
| --- | --- | --- |
| P4-01 / TODO | 外部依据与共享配置契约；依赖 P3-06、X-01/X-02 的结构需求 | 区划/职责/事项/人员及启动程序依据记录来源、版本、效力和核验状态；共享字段/证据最小可见；隔离测试夹具可开发，未核验正式资料不得启用 |
| P4-02 / TODO | 确定性匹配与重算；依赖 P4-01、P2-02 | 实际地址明确归属、已核验唯一范围命中才绑定；重复/争议/多候选阻断；地址/责任/范围更新重算当前绑定并留存依据，历史不回写 |
| P4-03 / TODO | 部门待处理与范围维护；依赖 P4-02、P1-05 | Web 待处理列表、理由/依据/操作人/时间；配置权限不扩大法定职责；人工处理不能跳过核验，并发处理不重复建立有效关系 |
| P4-04 / TODO | 扫码临时抽检；依赖 P4-02、P3-02、P3-01 | 同 App 行政工作区扫码 -> 逐次核验 -> 独立行政记录；无需预建任务/计划/派单；身份资格、监管关系、事项或启动依据任一缺失都不能提交正式记录 |
| P4-05 / TODO | 行政整改、复查和外部文书；依赖 P4-04 | 企业必要资料提交与行政记录分别留痕；行政要求、复查引用来源；外部处罚文书只关联，不生成裁量/决定，不当作每次抽检必经步骤 |
| P4-06 / TODO | 部门监管视图与共享隔离；依赖 P4-03、P4-05 | Web 只呈现监管/职责/字段共享范围内的数据和证据；双方不能互改记录；更换绑定和撤权后的访问测试；订阅状态不改变行政权限 |
| P4-07 / TODO | 行政阶段分层验收；依赖 P4-06 | 软件验收使用隔离夹具验证成功/拒绝路径；正式验收另需 X-01/X-02/X-03 的真实已核验资料；分别记录结果，缺资料时不写“正式可上线” |

## P5 订阅与便利功能

依据：C04、C02-07；默认 P4 后推进，技术入口需 P3-06。价格引用[产品 SSOT](ssot/01-product.md)，不在任务清单复制售价。

| ID / 状态 | 任务与依赖 | 交付及验收标准 |
| --- | --- | --- |
| P5-01 / TODO | 订阅与交易契约；依赖 P3-06、X-04 | 组织计费、订单/权益/退款/到期状态、回调校验与去重；渠道、续费、开票和退款方案先明确，不能擅定规则；Flutter 不默认导流站外支付 |
| P5-02 / TODO | 跨场所企业工作台；依赖 P5-01、P3-05 | App 管理员汇总本组织自检/整改/复查待办与进度；服务端查权益并授权；他组织数据拒绝，免费逐场所历史照常 |
| P5-03 / TODO | 批量协作；依赖 P5-02、P3-04 | 批量分配/提醒本组织工作；明确每项成功/失败、重复提交和撤权/到期并发；提醒基于已确认规则或内部安排，不创设法定周期 |
| P5-04 / TODO | Web 购买与支付权益；依赖 P5-01、P1-05、X-04 | 组织购买入口、服务端支付核实、签名/金额/币种/订单校验、重复/乱序通知及对账；客户端成功不单独开权益；真实交易验收与沙箱分开 |
| P5-05 / TODO | 到期/退款/免费回归验收；依赖 P5-03、P5-04、P4-07 软件部分 | 关闭便利而非基础流程；免费与到期组织自检、证据、历史、监管必要交互继续；付费不扩大授权；按确认方案验证退款/续费 |

## P6 试点发布

依据：C01-C05、[实施清单](implementation/README.md)、[工程治理](engineering.md)。不以通过本地预检代替上线审批。

| ID / 状态 | 任务与依赖 | 交付及验收标准 |
| --- | --- | --- |
| P6-01 / TODO | 发布环境与迁移；依赖 P0-P5 软件退出门槛、X-05 | 测试/生产隔离、密钥与 TLS、最小权限、健康检查和版本化部署；从空库安装及已有库迁移演练，生产不带模拟权限/行政资料 |
| P6-02 / TODO | 数据恢复与观测；依赖 P6-01 | 数据库与证据一致恢复演练、审计与日志脱敏、告警与故障排查说明；按确认的恢复目标验收，不把“有备份”当“能恢复” |
| P6-03 / TODO | 安全、容量与全链路；依赖 P6-01、P1-P5 | 跨组织/分配/字段/证据/双轨负测；会话、重试、支付、上传故障注入；以确认的试点规模制定负载与指标并保留报告，不虚构容量承诺 |
| P6-04 / TODO | 双端分发与真机；依赖 P6-01、P0-05、X-04/X-05 | Android/iOS 真机扫码、拍照、权限拒绝、弱网、退出/撤权与工作区隔离；正式标识/签名/分发策略和购买边界核验；不宣称商店已审核 |
| P6-05 / TODO | 真实输入与试点验收；依赖 P6-02 至 P6-04、P4-07 正式部分 | 真实已核验组织/部门/规则，完成免费、监管与订阅链路；演示记录、未解决风险、回退步骤和授权范围清单齐备 |
| P6-06 / TODO | 放行与观察；依赖 P6-05、用户发布批准 | 发布清单签认，记录版本与回退触发条件；受控试点观察故障、使用/支付/存储成本等真实数据；收入与盈利仍是待验证事实 |

## 外部输入与设计阻塞

这些项目与开发同步准备，不是让软件制定行政规则。清单记录“谁提供”和“阻塞什么”，具体资料仍只放在相应权威位置。

| ID | 输入 / 协作责任 | 需要解决的时间点 | 缺失时处理 |
| --- | --- | --- | --- |
| X-01 | 实施方取得并由有权主体确认区划、地址归属与监管职责/范围 | P4-02 正式绑定前 | 只做隔离测试；正式场所待处理 |
| X-02 | 有权主体提供适用事项/规则、人员资格、启动及程序依据、共享字段/证据 | P4-04 正式记录及 P4-06 正式共享前 | 不形成正式行政记录，不推断共享权 |
| X-03 | 单位/部门提供首位管理员材料；实施方线下核验 | P1-02 真实开通前 | 测试环境可演练，禁止真实账号无依据开通 |
| X-04 | 用户/运营明确支付、退款/开票/续费方案及分发地区/渠道，取得商户与平台条件 | P5-01、P5-04 和 P6-04 前 | 不集成真实交易、不提供未经核验 App 支付引导 |
| X-05 | 用户/实施方明确部署地区、规模、资料留存、恢复目标、签名/分发资源及 macOS 验证环境 | P0-05 所需构建资源及 P6 前 | 记录具体阻塞，不虚构性能/合规/恢复达标 |

## 每任务的完成定义

1. 来源明确：列出 SSOT、Cxx 与模块契约；新增决定有批准记录，外部输入注明核验状态。
2. 契约先行：声明输入输出、状态、权限、失败、幂等、持久化及审计边界；适用的机器文件被实际解析测试。
3. 功能可演示：跨端任务有实际 UI/API/数据库链路，不用 mock 成功冒充业务完成。
4. 自动测试：正常路径及相应越权、重复、失效、并发/故障负测通过；统一预检通过，任务特有检查另附。
5. 交付可追溯：记录变更路径、命令/结果、CI/构建证据、未覆盖风险、配置及回退影响。
6. 状态真实：尚缺验收的任务保持 PARTIAL/BLOCKED；不因写完代码直接 DONE。

开工后在本文件追加对应任务执行记录，至少包含：任务 ID、负责人、状态、日期、依赖核验、契约/代码路径、验收命令与结果、阻塞/下一步。文档计划阶段不预建几十份空任务文件，不自动创建远程 issue。

## 建议第一批执行

计划已确认 -> P0-01 基线选型确认 -> P0-02/P0-03/P0-04/P0-05 按资源逐项推进 -> P0-06 验收 -> P1-01。P1-04 已有代码先保留，不继续扩展它来绕过基础阶段。各任务是可独立验收的工作包，不承诺一次对话完成全部；开工时可拆成子任务但保留父 ID 与相同退出标准。

## P0-01 执行记录

- 日期/负责人：2026-09-28，Codex；用户负责必要选型确认。
- 依赖：路线图及最小基线均获用户明确确认，并批准按依赖继续推进。
- 成果：[工程基线](engineering-baseline.md)列出文件事实、选型建议、版本盘点及开发/发布界限；更新计划批准状态和入口，不修改业务代码、依赖、签名或部署配置。
- 结果：用户已明确批准最小技术基线，相关权威正文及派生契约已同步，任务 DONE；不将其扩张为技术方案全篇或生产选型批准。
- 验收：文件核对完成；运行 `powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/test/preflight.ps1`，退出码 0；文档检查（166 个本地链接、24 张表、7 条策略/24 个案例）、Go test/vet、Web 类型/构建/SSR 测试、Flutter analyze/widget test 均通过。原生构建和远程 CI 不属于本轮已验证结果。

## P0-02 执行记录

- 日期/负责人：2026-09-28，Codex；用户提供远程仓库与访问方式。
- 依赖：P0-01 已完成。仅执行可独立完成的本地准备，不开展身份或其他业务开发。
- 成果：本地 `git init -b main`；`.gitignore` 排除生成目录、环境文件和本机 `.gstack`；`.nvmrc` 与 CI Node 精确对齐到已验证的 22.23.2；保留现有 Go/Flutter CI 版本及锁文件安装步骤。
- 阻塞：尚无用户指定远程，未创建远程、暂存、提交、推送或启用实际 PR 保护；CI 文件存在不能作为运行成功证据。任务保持 BLOCKED。
- 后续：用户提供仓库 URL、平台和访问方式后，检查待提交内容及本机秘密，再按授权建立远程协作，获取实际 CI 与保护规则证据。不根据机器登录信息猜测仓库归属或可见性。
- 验证：本轮完整 `scripts/test/preflight.ps1` 退出码 0，文档/策略案例、Go test/vet、Web 构建/SSR 测试、Flutter analyze/widget test 均通过；`git remote -v` 为空，`git status --short` 显示文件未暂存；`git check-ignore` 确认环境文件、本机 .gstack、Android local.properties、node_modules 和 .dart_tool 被忽略。此结果不是干净检出验证、秘密扫描或实际远程 CI 证据。

### 远程验收补记

2026-09-28：用户指定 `jihongxing/EmergencyManagementSystem` 后，确认它为空的公开仓库，使用本机已有认证建立 origin。初次基线提交 `febf121` 推送成功；对入库文件做凭据模式与忽略规则检查，未发现命中的凭据，但不宣称完成全面秘密审计。最初的无远程阻塞已解除，上文保留历史过程。

- GitHub Actions 运行 ID `36445834084`，提交 `febf12163387e4a881f4f2ef2fcfbd7412cd97b4`，`preflight` 成功，包含干净检出、npm ci、Flutter 锁文件解析和统一预检。
- main 保护已配置并回读确认：要求 PR、preflight 成功、分支最新、讨论已解决；管理员同样受约束，禁止强推和删除。单人维护暂不强制其他人审批，审批数为 0；这不等于独立人工审查。配置记录见 [.github/main-protection.json](../.github/main-protection.json)。
- 基线创建后切到 `chore/p0-02-remote-governance`，治理文档通过 PR 推进，不绕过主分支保护。
- CI 成功但提示 checkout@v4、setup-go@v5、setup-node@v4 的 Node 20 action 运行时已被 runner 转为 Node 24；列为后续兼容性维护，不与应用 Node 22.23.2 混淆。
- P0-02 已达到任务验收；Android/iOS 实际构建仍属 P0-05，整个 P0 尚未验收。

## P0-03 执行记录

- 完成验证日期：2026-09-29（本机 Asia/Shanghai，任务跨日执行）。

- 前置：用户要求先合并 PR #1，再执行本任务；PR #1 已正常合并为 `8902f4e`，分支 `feat/p0-03-database` 从更新后的 main 创建。
- 契约先行：[数据库基础](../constras/platform/database.md)与[就绪接口](../constras/platform/ready.openapi.json)。pgx database/sql 与 Goose 作为本任务工程实现锁定于 go.mod/go.sum，不确定生产供应商；保留显式 SQL 与独立迁移边界。
- 交付：必填且不泄露秘密的连接配置、2 秒就绪期限、版本检查、独立向前迁移命令、仅 schema/迁移元数据的首迁移；Podman 独立项目名和固定镜像 digest、回环端口 55472、命名卷及环境样例。
- 实际验证：`preflight.ps1 -Database` 通过文档、Go test/vet、真实 PostgreSQL 集成、Web 构建/测试、Flutter analyze/test；空库 503、迁移后 200、重复迁移、失败事务回滚、不可达库和存活分离、停止/重启版本保留均覆盖。
- 环境影响：仅创建并重启本项目 `ems-dev-postgres-1`；保留测试库和命名卷，未删除其他项目数据。密码为本机临时随机值，不打印/提交，接管方法见运行说明。
- 限制：Windows 远程 CI 普通预检显式跳过数据库集成；未验证生产权限拆分、备份恢复、并发部署迁移或原生移动构建。下一任务 P0-04，不进入身份业务。

## P0-04 执行记录

- 日期/负责人：2026-09-29，Codex；依赖 P0-01 已完成。依据：平台 C05、[行为与权限契约](../constras/README.md)、[身份案例](../constras/identity/authorization.json)及[探针 OpenAPI](../constras/platform/live.openapi.json)。
- 范围：增加[严格机器契约检查](../scripts/test/check-machine-contracts.ps1)、[行为与身份 JSON Schema](../scripts/contracts/behavior.schema.json)、[OpenAPI/响应校验](../scripts/contracts/validate.test.mjs)及锁定的 npm 依赖；统一预检和 CI 安装并实际执行；Go `httptest` 输出存活 200、就绪 200/503 的响应供 OpenAPI Schema 校验。
- 验收：机器门禁校验必填字段、未知字段、来源条款与案例准入预期；隔离负测覆盖损坏 JSON、缺文件、断引用、错误 operationId/状态码、错误预期和响应体/内容类型不一致。`pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/test/preflight.ps1` 通过文档、契约、Go test/vet、Web build/test、Flutter analyze/test；实际数据库集成因本次会话未设置 `POSTGRES_PASSWORD` 未重跑，沿用 P0-03 原有独立验证记录，不宣称本次数据库集成通过。
- 数据/回退：无数据库迁移或业务数据变更；回退为撤销本任务的校验脚本、测试与 CI 依赖更改。P0-05 原生双端构建尚未完成；P0-06 阶段验收未开始。

## P0-05 执行记录

- 日期/负责人：2026-09-29，Codex。依据 C05-01 和[移动工程验证契约](../constras/platform/mobile-validation.md)；P0-01 已完成。不新增业务功能、正式标识或发布签名。
- [原生工作流](../.github/workflows/mobile.yml)分别运行 Android debug 构建与模拟器启动、macOS iOS 无签名构建与 iPhone 模拟器启动；[启动测试](../apps/mobile/integration_test/startup_test.dart)只验证未认证壳。工作流存在不等于验收通过。
- 本机 Windows 配备 Flutter 3.44.0、Android SDK 36；无本地 macOS。Android 首次构建曾补下载引擎组件，本进程使用本机现有代理，不更改全局代理。
- 远程 push CI `36546950109`（提交 `3d963bb`）和对应 PR CI `36546955477` 均成功：Android debug APK 构建并在 Linux 原生模拟器运行 `integration_test/startup_test.dart`，断言未认证应用壳启动；macOS-15 使用 Xcode 16.4 完成 iOS `--debug --no-codesign` 设备构建，并完成 iPhone 模拟器构建、安装、启动、进程存活检查及启动截图。
- iOS 启动截图已人工核验：显示“应急安全检查”和身份认证未接入提示，不显示企业或行政工作区。证据工作流：[mobile.yml](../.github/workflows/mobile.yml)；原生验证契约：[mobile-validation.md](../constras/platform/mobile-validation.md)。
- 结果：P0-05 验收通过；下一步仅为 P0-06 基线验收，未进入 P1 身份业务。

## P0-06 执行记录

- 完成验证日期：2026-09-29（本机 Asia/Shanghai，任务跨日执行）。
- 前置：PR #5 已合并，合并提交为 `c73212c5bb0cc4fbb47ab281840060871268871e`；本任务分支 `codex/p0-06-baseline-acceptance` 基于该提交建立。干净验收检出位于 `D:/codeSpace/EmergencyManagementSystem-p0-06-clean`，检出提交为 `c73212c`，工作区无未提交文件。
- 复现安装：按 README 和锁文件执行 `scripts/contracts/npm ci`、`apps/admin-web/npm ci`、`apps/mobile/flutter pub get --enforce-lockfile`，三项均成功；未提交 `node_modules`、`.dart_tool` 或构建产物。`apps/admin-web` 的 npm 安装报告 2 个 moderate audit advisories，未阻断本次工程基线验收，列为依赖维护项。
- 工具版本：Git `2.51.2.windows.1`；Go `1.25.5 windows/amd64`；Node `22.23.2`；npm `10.9.8`；Flutter `3.44.0`；Dart `3.12.0`；Podman `5.7.1`。版本与 `.nvmrc`、Flutter SDK 约束及现有 CI 记录一致。
- 本轮验收命令：在干净检出执行 `pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/test/preflight.ps1`，退出码 0。机器契约门禁通过；契约 npm 测试 2/2 通过；Go test/vet 通过；Web build 和 1 个 Vitest 测试通过；Flutter analyze 无问题、widget test 1/1 通过；脚本按设计提示未执行真实 PostgreSQL 集成。
- 数据库证据边界：本轮干净验收未设置 `POSTGRES_PASSWORD`，因此没有执行 `preflight.ps1 -Database`。P0-03 已完成并记录独立 PostgreSQL 连接、迁移、就绪、失败回滚及停止/重启保留验证；本记录引用该既有证据，不将其表述为本轮干净验收结果。
- 远程和双端证据：沿用 P0-02 的 GitHub Actions `36445834084`、P0-05 的 push CI `36546950109` 和 PR CI `36546955477`；P0-05 已验证 Android debug 构建与启动测试、macOS iOS 无签名构建、iPhone 模拟器安装启动和进程存活。PR #5 合并前检查全部通过。
- 验收结论：P0-01 至 P0-06 的工程基线证据齐备，P0-06 标记为 `DONE`。这只表示工程基线验收完成，不表示身份、业务、生产部署或正式发布已完成；下一步为用户确认 P0 阶段后再开始 P1-01。

## P1-03 执行记录

- 日期/负责人：2026-10-03，Codex；用户已正式批准实施 P1-03。依赖 P1-02 的身份/组织模型与 P0-03 数据库基础已核对。
- 范围：会话与逐次身份校验，包括登录、退出、刷新、凭据恢复、Web HttpOnly Cookie、Flutter access/refresh 凭据、每次请求重新读取成员和组织状态；不进入 P1-04 成员管理、P1-05 工作区或行政资格规则。
- 契约：新增[会话契约](../constras/identity/session.md)及机器文件[session.json](../constras/identity/session.json)，接入契约 Schema 门禁；同步身份 OpenAPI 的会话响应字段。
- 实现：[会话服务](../backend/internal/identity/session.go)、[PostgreSQL 存储](../backend/internal/identity/session_sql.go)、[HTTP 适配器](../backend/internal/identity/http.go)、[服务路由](../backend/cmd/server/main.go)；新增会话/密码恢复迁移 `00003_identity_sessions.sql`，数据库当前迁移版本提升为 3。
- 测试：覆盖错误密码与未知账号统一失败、pending 激活、成员撤权、组织停用、退出重放、刷新令牌一次性轮换、密码恢复一次性及撤销既有会话；Go test/vet 已通过，机器契约测试已通过。
- 验收命令与结果：`npm run check --prefix scripts/contracts` 通过；在 `backend` 目录执行 `go test ./...` 和 `go vet ./...` 通过。尚未设置 `POSTGRES_PASSWORD`，未执行本次专用 PostgreSQL 集成；Flutter 真实登录链路尚未接入，因此任务保持 `PARTIAL`。
- 数据/回退：新增不可变迁移 `00003_identity_sessions.sql`；不修改已有迁移的历史语义。回退应使用新的回退迁移或回滚发布版本，不删除已有数据卷。
- 下一步：补充专用 PostgreSQL 集成证据和 Flutter 端真实登录/凭据存储验证；在用户验收 P1-03 前不进入 P1-04。

## P1-04 执行记录

- 日期/负责人：2026-10-03，Codex；用户已正式批准实施 P1-04。依赖 P1-03 的会话校验实现已核对；P1-03 本身仍保持 `PARTIAL`，本任务只复用其已交付的服务端逐次校验，不宣称 P1-03 集成验收完成。
- 范围：本组织成员列表、创建、角色/状态更新、管理员范围校验、首位管理员保护、成员变更审计；不进入 P1-05 工作区、邀请渠道、行政资格或跨组织管理。
- 契约：新增[成员管理契约](../constras/identity/members.md)、[成员机器契约](../constras/identity/members.json)和 Schema 门禁；更新身份契约索引。既有 OpenAPI 成员路径接入实际 HTTP handler。
- 实现：[成员服务](../backend/internal/identity/member.go)、[PostgreSQL 存储](../backend/internal/identity/member_sql.go)、[身份 HTTP handler](../backend/internal/identity/http.go)；新增不可变迁移 `00004_member_management.sql`，建立 `member_audit_events`，数据库当前迁移版本提升为 4。
- 规则：仅当前 active 管理员可管理同组织成员；执行人员和跨组织请求拒绝；新成员固定为 `pending`；角色按组织类型校验；首位管理员不能通过普通成员接口修改；创建、角色变化和状态变化记录审计快照。
- 测试：覆盖执行人员拒绝、跨组织拒绝、组织类型角色拒绝、pending 创建、首位管理员保护、角色/状态审计和并发重复登录标识。契约门禁、文档治理、Go test/vet 已通过。
- 验收命令与结果：`npm run check --prefix scripts/contracts`、`scripts/test/check-doc-governance.ps1`、`scripts/test/check-machine-contracts.ps1`、`go test ./...`、`go vet ./...` 均通过；`scripts/test/preflight.ps1` 待本轮最终执行。专用 PostgreSQL 集成未执行，HTTP 端到端和 Flutter 客户端联调未执行，任务保持 `PARTIAL`。
- 数据/回退：只新增 `00004_member_management.sql`，不改写已提交迁移；回退使用新的回退迁移或回滚发布版本，不删除数据卷。审计记录保留成员历史，不记录密码或令牌原文。
- 下一步：使用专用 PostgreSQL 执行迁移、并发、授权/撤权和审计集成验收，再由用户确认 P1-04；未验收前不进入 P1-05。

## P1-05 执行记录

- 日期/负责人：2026-10-03，Codex；用户已正式批准实施 P1-05。依赖 P1-03、P1-04 的软件侧接口已核对；前置任务仍为 `PARTIAL`，本任务不宣称其专用 PostgreSQL 或完整端到端验收已完成。
- 范围：Web/App 登录入口、当前身份读取、服务端身份驱动的企业/行政工作区边界、退出和失效状态清理；不进入场所、二维码、检查、订阅、支付或行政资格。
- 契约：新增[身份入口契约](../constras/identity/entry.md)、[机器契约](../constras/identity/entry.json)及 Schema 门禁；更新身份契约索引和统一机器校验。
- 实现：Web 新增 `apps/admin-web/src/api.ts`、`auth.ts`，入口由 `/v1/me` 当前身份决定，不使用 `localStorage` 令牌；Flutter 单 App 提供未认证、企业工作区、行政工作区和退出的最小状态边界。当前 Flutter 登录为 UI 状态测试替身，未冒充真实服务端链路。
- 规则：未认证不暴露工作区；`enterprise` 仅进入企业工作区并可见购买占位；`department` 仅进入行政工作区；不提供自由角色切换；前端路由/隐藏菜单不作为授权来源；撤权/停用需要由下一次受保护请求观察并清理。
- 测试：机器契约及负测、Web SSR 壳测试、Web 构建、Flutter 分析和 Flutter 测试通过。专用 PostgreSQL、真实 Web 联调、Flutter 网络请求/平台安全存储、真机撤权链路尚未执行，任务保持 `PARTIAL`。
- 数据/回退：不新增数据库迁移，不修改已提交迁移；客户端改动可随版本回退，Web 不保存令牌到 `localStorage`。Flutter 的真实 refresh token 安全存储仍需后续契约和平台验证。
- 下一步：补齐真实 Web/App API 联调与 Flutter 可替换网络/安全存储实现，再由用户确认 P1-05；不自动进入 P1-06。

## P1-06 执行记录

- 日期/负责人：2026-10-03，Codex；用户已正式批准实施 P1-06。依赖 P1-05 的软件侧入口和身份契约已核对；本轮不扩展 P2 场所/对象或任何行政业务。
- 范围：固定隔离测试资料中的两个企业组织和一个部门组织；验证组织隔离、企业管理员兼任执行角色、部门管理员兼任检查角色、跨组织拒绝、成员审计和撤权后现有会话立即失效。
- 契约/证据：[P1-06 最小组织阶段验收](acceptance/p1-06-minimum-organization.md)；专用测试为 [`TestP1MinimumOrganizationAcceptance`](../backend/internal/identity/p1_acceptance_test.go)。测试资料不代表真实企业、部门、辖区、行政资格或检查授权。
- 结果：软件侧阶段验收测试通过；两个企业和一个部门的成员列表均按组织隔离，跨组织访问被拒绝，管理员兼任场景通过，成员状态变化产生组织/操作者/目标/前后状态审计证据，已有会话在撤权后下一次服务端校验被拒绝。
- 验收命令与结果：`go test ./internal/identity -run TestP1MinimumOrganizationAcceptance -count=1` 通过；`go test ./internal/identity` 通过；`go vet ./internal/identity` 通过；统一 `scripts/test/preflight.ps1` 通过。预检明确跳过真实 PostgreSQL 集成。
- 未完成/边界：专用 PostgreSQL 集成、真实 Web/App 联调、Flutter 平台安全存储、真机撤权演示及外部行政资料均未在本轮完成；因此 P1-06 保持 `PARTIAL`，不自动进入 P2。
