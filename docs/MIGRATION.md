# 迁移说明

本仓库已经重组为新的 Go 后端和 Vue 3 前端项目。

## 已移除内容

- `Inotify/`：旧版 .NET Core 后端。
- `Inotify.Vue/`：旧版 Vue 2 管理后台。
- `sonar.sln`：旧版解决方案文件。
- `public/`：旧版截图资源。

## 当前目录结构

- `backend/`：Go API 服务。
- `frontend/`：Vue 3 + Vite 管理后台。
- `docs/`：项目文档。
- `.github/workflows/ci.yml`：测试、构建和 Docker 镜像自动化。

## 数据迁移

当前 Go 后端使用 SQLite，并在启动时通过 Gorm 自动迁移数据表结构。

默认数据目录：

- 本地开发：`backend/inotify_data` 或启动目录下的 `inotify_data`。
- Docker 部署：`/app/inotify_data`。

主要运行数据：

```text
inotify.db
jwt.json
```

从旧版项目迁移时，不建议直接复用旧库表。推荐重新部署新版后，在后台重新配置用户、消息通道、三方登录和系统参数。
