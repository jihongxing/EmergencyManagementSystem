# P1-02 至 P1-06 身份与最小组织统一验收

状态：`SOFTWARE ACCEPTANCE PASSED`。日期：2026-10-03。对应任务：[P1-02 至 P1-06](../development-tasks.md)。

## 验收范围

本次统一验收覆盖已批准的 P1 最小组织边界：

- 受控开通企业和部门组织；
- 首位管理员初始为 `pending`，经受控激活后可登录；
- 登录、退出、刷新、逐次身份校验和成员撤权即时生效；
- 本方成员创建、角色/状态变更及成员审计；
- 企业管理员兼任企业执行人员；
- 部门管理员兼任检查人员；
- 企业、部门之间以及不同企业之间的跨组织访问拒绝；
- Web Cookie 与 Flutter access/refresh 会话数据结构；
- 企业工作区与部门工作区入口隔离。

测试组织、成员、密码和外部材料均为隔离测试资料，不代表真实企业、部门、辖区、行政资格或检查授权。

## 验收证据

### 机器契约与工程门禁

```text
pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/test/preflight.ps1 -Database
```

结果：通过。包含文档治理、机器契约、OpenAPI/JSON Schema、Go test、Go vet、Web build/test、Flutter analyze/test，以及 PostgreSQL 迁移、就绪检查、损坏迁移回滚和重启后的版本持久化。

### 内存最小组织验收

```text
pwsh -NoProfile -Command "Set-Location backend; go test ./internal/identity -run TestP1MinimumOrganizationAcceptance -count=1"
```

结果：通过。验证组织隔离、管理员兼任、成员生命周期、审计和撤权后的逐次校验拒绝。

### PostgreSQL 身份链路验收

```text
pwsh -NoProfile -Command "Set-Location backend; go test ./internal/identity -run TestP1SQLMinimumOrganizationAcceptance -count=1 -v"
```

测试实现：[sql_acceptance_test.go](../../backend/internal/identity/sql_acceptance_test.go)。

结果：通过。验证真实 PostgreSQL 中的组织开通、重复 `external_key` 拒绝、首位管理员激活、成员创建与状态更新、跨组织拒绝、成员审计字段、登录、会话验证及撤权后的旧 access token 失效。

## 统一结论

P1-02 至 P1-06 的**软件侧统一验收通过**，可以合并当前实现到主干。

以下事项不属于本次软件侧统一验收的通过条件，仍作为后续发布验证缺口保留：

- 真实 Web 浏览器与后端服务的端到端联调；
- Flutter 网络请求适配和 Android/iOS 平台安全存储；
- 真机上的登录、工作区切换和撤权演示；
- 真实实施主体、县域、部门或行政授权材料。

因此，本次合并不表示 V1 发布验收完成，也不自动进入 P2；后续任务仍须按路线图和任务退出标准推进。
