# P1-06 最小组织阶段验收

状态：`SOFTWARE ACCEPTANCE PASSED`。日期：2026-10-03。对应任务：[P1-06](../development-tasks.md)。统一验收记录见[ P1-02 至 P1-06 身份与最小组织统一验收](p1-02-to-p1-06-acceptance.md)。

## 验收边界

本验收只验证最小组织模型的软件隔离与身份生命周期：

- 两个独立企业组织：`org_enterprise_a`、`org_enterprise_b`；
- 一个部门组织：`org_department`；
- 企业管理员兼任企业执行人员；
- 部门管理员兼任检查人员；
- 跨组织读取被拒绝；
- 成员变更留下组织、操作者、目标、动作和前后状态审计证据；
- 已建立会话的成员撤权后，下一次服务端身份校验立即拒绝。

测试组织、成员和密码均为隔离测试资料，不是实际企业、部门、辖区、行政资格或检查授权。

## 内存验收证据

执行：

```text
pwsh -NoProfile -ExecutionPolicy Bypass -Command "Set-Location backend; go test ./internal/identity -run TestP1MinimumOrganizationAcceptance -count=1"
```

测试实现：[p1_acceptance_test.go](../../backend/internal/identity/p1_acceptance_test.go)。

## 统一验收后的结论边界

本文件的内存测试证明当前服务实现满足 P1-06 的最小隔离场景；统一验收另已补充真实 PostgreSQL 身份链路证据。仍不证明：

- PostgreSQL 事务、锁和真实数据迁移已完成集成验收；
- Web 与 Flutter 已完成真实服务端联调；
- 平台角色等同于行政资格；
- 已确认任何真实县域、部门、管辖范围或检查事项。

因此 P1 软件侧统一验收通过，但发布级客户端验证和外部行政资料仍是独立缺口，不得据此自动进入 P2 或把行政业务视为已授权。
