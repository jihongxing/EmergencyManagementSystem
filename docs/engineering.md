# 工程治理

状态：`ENGINEERING BASELINE`。在现有文档上增量采用 full-product-system 的变更、验证与交付纪律，不复制另一套 SSOT。产品事实以 [SSOT](DECISIONS.md) 为准，工程建议见[技术方案](technical-design-v1.md)。

## 变更单元

按[路线图](../ROADMAP.md)与[任务清单](development-tasks.md)选择依赖就绪的任务，先确认任务 ID 再开发。阶段结束提交可演示成果和验收证据，经用户确认后推进下一阶段；不得以单元测试通过替代模块或阶段验收。任务状态统一在任务清单维护，路线图只汇总阶段。

每次变更明确目标、来源条款、修改范围、验收标准。对外或跨模块行为先更新 `constras/` 再实现并测试；未确认的产品决定留在 OPEN。交付时列明执行命令、结果、未执行项、配置/迁移变化、回退方式，并同步路线图及入口。

## 代码边界

- `backend/internal/` 按实际模块增长，不创建空业务包；跨模块禁止直接写入他方私有表。
- 数据库和证据存储是外部基础设施；新增业务表、文件适配、支付适配须先有模块契约。
- 单 App 工作区不是授权边界；未接入身份系统时不提供假登录或角色切换绕过授权。
- 配置从环境注入，秘密不入库，日志不输出密码、令牌或证据原文。
- Go 使用 gofmt、test、vet；Web 使用严格类型、构建及测试；Flutter 使用 analyze 和 widget tests。依赖安装后提交锁文件，新增依赖说明必要性。

## 验证与交付

[统一预检](../scripts/test/preflight.ps1)执行文档检查、机器契约与 OpenAPI 规范/引用校验、探针实际响应验证、Go 检查、Web 构建测试、Flutter 检查及原生目录存在性。契约校验依赖 `scripts/contracts` 的锁文件和 `npm ci`；`-DocsOnly` 明确是部分验证；预检不替代 Android/iOS 实际构建、签名或生产验收。原生 iOS 在 macOS 上验收。

现有 Markdown/JSON 由文档脚本和[机器契约门禁](../scripts/test/check-machine-contracts.ps1)检查；首个工程接口见 [OpenAPI](../constras/platform/live.openapi.json)。现有存活/就绪 OpenAPI 已做规范、引用和实际响应校验；后续新增业务接口仍须逐一增加真实请求/响应及权限测试，不以探针覆盖代替业务契约验收。

本机容器使用 Podman；PostgreSQL 开发配置见 [compose](../infra/compose.yaml)。启动前设置本地密码，禁止无意删除命名数据卷。正式部署、文件服务和备份方案仍待实施选型。

GitHub origin 已建立，采用任务短分支和 PR。main 要求 preflight 成功、分支最新及讨论解决，保护适用于管理员且禁止强推/删除；单人维护暂设额外审批数 0，不宣称独立人工审查。保护配置见[记录](../.github/main-protection.json)，实际运行证据见[任务清单](development-tasks.md)。Node 使用 `.nvmrc` 与 workflow 对齐的 22.23.2，Web 使用 `npm ci` 安装锁定依赖。

已运行 [GitHub Actions 基础检查](../.github/workflows/verify.yml)，使用锁文件安装与统一预检，首次远程运行已成功。macOS 原生构建任务仍需在 P0-05 补齐，不将 analyze/widget 测试当原生构建。

本机使用 v2rayN。CLI 不一定继承系统代理；网络超时时可仅在当前进程设置 `HTTP_PROXY`、`HTTPS_PROXY` 及本地地址的 `NO_PROXY`，端口以本机代理实际配置为准，不提交个人代理设置，不擅改全局配置。Flutter 即使使用 `--offline` 创建模板也可能先补齐缺失的 SDK 组件，须查看实际下载日志。

发布前记录版本、实际测试、迁移、配置、备份恢复结果和已知风险。迁移采用兼容扩展再收缩，不能简单回滚应用后假定数据可逆。证据留存、恢复目标和运行指标按试点要求确认；工程骨架不代表生产就绪。
