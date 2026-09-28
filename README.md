# 应急安全检查平台

V1 工程骨架：一个 Flutter App（Android/iOS、按身份隔离工作区）、部门 Web 和 Go 单体。已实现未认证入口壳、存活/数据库就绪探针、迁移基础和身份内部授权判定；尚未接入登录、数据库业务表、行政记录或真实支付。

产品边界见[文档地图](docs/README.md)，开发纪律见[工程治理](docs/engineering.md)，当前进度见[路线图](ROADMAP.md)，契约统一放在 [constras](constras/README.md)。

## 启动

```powershell
# 先按 docs/database-development.md 启动数据库并设置 DATABASE_URL
cd backend
go run ./cmd/migrate
go run ./cmd/server
```

`GET /health/live` 只表示进程存活，不表示数据库、业务或行政依据就绪。

```powershell
cd apps/admin-web
npm install
npm run dev
```

```powershell
cd apps/mobile
flutter pub get --enforce-lockfile
flutter analyze
flutter test
```

Android/iOS 原生目录与依赖锁文件已生成，Flutter 分析和 widget 测试已通过。iOS 构建与签名需要 macOS/Xcode，不能以 Windows 分析通过代替双端构建。当前应用标识是生成器开发占位值，正式发布前需确认并替换。

## Podman

```powershell
$env:POSTGRES_PASSWORD = '<local-development-password>'
podman compose -f infra/compose.yaml up -d
podman compose -f infra/compose.yaml ps
```

仅启动本地回环地址 55472 的 PostgreSQL，使用命名数据卷。不得无意删除卷。API 已接入数据库连接与迁移版本就绪检查；对象存储仍未接入。环境、密码与实际集成检查见[数据库开发说明](docs/database-development.md)，不以探针成功宣称完整业务就绪。

## 检查

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/test/preflight.ps1
```

仅查文档时显式加 `-DocsOnly`，不等于工程验证通过。依赖锁文件已入库，不提交密钥、真实证据或模拟行政授权。GitHub origin 已建立，远程 preflight 已通过；main 要求 PR 与状态检查，后续使用任务分支。Node 复现版本见 `.nvmrc`，Web 干净安装使用 `npm ci`。实际 CI 与保护证据见[任务记录](docs/development-tasks.md)。
