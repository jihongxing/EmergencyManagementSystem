# 本方成员管理与审计契约

状态：`MODULE CONTRACT`。任务：`P1-04`。依据：[身份模块契约](module.md)、[权限 SSOT](../../docs/ssot/04-access.md)、[授权机器案例](authorization.json)、[身份 OpenAPI](identity.openapi.json)。

## 最小范围

- 只有当前组织的管理员可以管理本组织普通成员；执行人员、检查人员不能管理成员。
- 成员列表、创建和更新都必须从服务端当前会话读取管理员身份；客户端提交的组织、角色和状态不能扩大权限。
- 新成员只能以 `pending` 状态创建，不创建初始密码，不开放公众注册或认领。
- 角色必须与组织类型匹配：`enterprise_admin`/`enterprise_executor` 仅用于 `enterprise`；`department_admin`/`inspector` 仅用于 `department`。
- 普通成员更新允许修改角色和状态；不能通过普通接口修改首位管理员，首位管理员替代继续使用受控实施流程。
- 成员状态变更、角色变更和成员创建必须写入组织审计；审计不得保存密码、令牌原文或恢复令牌。
- 成员撤权后，P1-03 的逐次会话校验立即拒绝后续受保护请求；历史记录保留成员身份。

## 接口

接口路径、请求响应和错误码见 `identity.openapi.json`：

- `GET /v1/organizations/{organizationId}/members`
- `POST /v1/organizations/{organizationId}/members`
- `PATCH /v1/organizations/{organizationId}/members/{memberId}`

跨组织、非管理员、停用成员和未知角色均拒绝。并发更新必须由数据库行锁和条件校验保证，不产生两个互相覆盖的成功审计事实。

## 不在本契约内

- 首位管理员更换、实施人员认证、行政资格、跨组织管理。
- 邀请邮件/短信、密码初始化渠道、工作区入口、场所或业务对象。
