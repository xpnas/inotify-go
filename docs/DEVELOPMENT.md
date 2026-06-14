# 开发说明

## 环境要求

- Go 1.25+
- Node.js 22+
- npm 10+
- Docker Desktop 或 Docker Engine

## 后端

```bash
cd backend
go test ./...
go run ./cmd/inotify
```

环境变量：

- `INOTIFY_ADDR`：监听地址，默认 `:8000`。
- `INOTIFY_DATA_DIR`：运行数据目录，默认 `inotify_data`。

后端主要目录：

- `backend/internal/handlers`：API 路由和处理器。
- `backend/internal/database`：数据库初始化、默认数据和迁移。
- `backend/internal/models`：数据模型。
- `backend/internal/sender`：消息通道发送实现。

## 前端

```bash
cd frontend
npm install
npm run lint
npm run dev
```

Vite 开发服务器默认运行在：

```text
http://localhost:9000
```

开发服务器会把 `/api` 代理到：

```text
http://127.0.0.1:8000
```

前端主要目录：

- `frontend/src/api`：接口封装。
- `frontend/src/views`：页面。
- `frontend/src/router`：路由。
- `frontend/src/stores`：Pinia 状态。

## 完整验证

```bash
cd backend
go test ./...

cd ../frontend
npm run lint
npm run build
```

## 项目约定

- 后端运行数据位于 `inotify_data/`，不要提交到仓库。
- 默认初始化账号为 `admin / 123456`，生产环境部署后请及时修改密码。
- 前端开发端口默认 `9000`。
- 后端开发端口默认 `8000`。
- 变更后端接口时，同步更新 `docs/API.md`。
