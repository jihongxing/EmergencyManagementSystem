# P0-03 数据库开发环境

状态：`ENGINEERING IMPLEMENTATION`。依据：[数据库契约](../constras/platform/database.md)。PR #1 已合并；本任务使用 pgx v5.7.6 的 database/sql 适配和 Goose v3.24.3，实际版本以 go.mod/go.sum 为准。不将这些版本称为最新或已经完成安全审计。

## 本地启动

先在当前终端设置自行保管的 `POSTGRES_PASSWORD`，不写入提交或聊天。示例字段见 [.env.example](../.env.example)，脚本不会自动加载根目录 .env。已有卷的密码不能通过改变环境变量重设。

```powershell
podman compose -f infra/compose.yaml up -d
$password = [Uri]::EscapeDataString($env:POSTGRES_PASSWORD)
$env:DATABASE_URL = "postgres://emergency_dev:${password}@127.0.0.1:55472/emergency_dev?sslmode=disable"
cd backend
go run ./cmd/migrate
go run ./cmd/server
```

Podman 项目名 `ems-dev`，默认端口 55472，可通过 POSTGRES_PORT 调整；DATABASE_URL 必须同步调整。仅回环地址暴露，开发用户是容器初始化管理员，不能照搬成生产业务账户。SSL 关闭只适用此本地环境。镜像已按本机 PostgreSQL 17 镜像 digest 固定，不操作其他项目容器。

缺失/格式无效 DATABASE_URL 拒绝启动；数据库不可达时 API 可启动，但 `/health/ready` 返回 503，`/health/live` 仍返回 200。ready 检查连接及当前程序要求的迁移版本，不证明文件服务、权限或行政依据就绪。`go run ./cmd/migrate` 独立向前迁移，当前包含平台基础和已批准身份开通表；后续业务迁移需先补契约，并同步调整就绪版本。

## 验证与数据保留

```powershell
pwsh -NoProfile -ExecutionPolicy Bypass -File scripts/test/preflight.ps1 -Database
```

此命令需要 Podman 及 POSTGRES_PASSWORD，会创建随机名 `ems_test_*` 的专用空库，验证迁移、就绪 HTTP、失败事务回滚及重复执行，再停止/启动本项目 PostgreSQL 检查持久性；因此会短暂中断本项目开发连接，请勿在其他人正在使用本开发库时运行。不删除库或卷，重复运行会保留多个测试库。

普通 preflight 不要求容器，明确报告跳过数据库集成；远程现有 Windows CI 只运行普通预检，不能当作远程 PostgreSQL 集成证据。测试库由调用者显式提供 TEST_DATABASE_URL，测试会拒绝已有 Goose 元数据的数据库，不重置共享数据。

2026-09-29：本次初始化的本地开发容器密码为临时随机值，仅存于本机容器环境，未打印或提交。后续在本机通过受控凭据管理接管；不要重新设环境变量后以为已有卷密码已更改，不要为了改密码删除卷。

停止使用 `podman compose -f infra/compose.yaml stop`；重新启动使用 `start`。禁止无意执行带卷删除的操作。生产角色拆分、备份恢复和完整迁移并发治理仍属后续发布要求，本任务不宣称生产就绪。
