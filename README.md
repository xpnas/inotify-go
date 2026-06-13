# Inotify

Inotify 是一个轻量级消息通知管理系统。当前项目已经重组为 Go 后端和 Vue 3 管理后台，不再包含旧版 .NET Core 与 Vue 2 代码。

## 技术栈

- Backend: Go 1.22, Gin, Gorm, SQLite
- Frontend: Vue 3, Vite, Pinia, Vue Router, Element Plus
- Deployment: Docker, Docker Compose, Nginx

## 功能

- 用户登录、JWT 鉴权、GitHub OAuth 登录
- 消息通道管理
- 当前账号 Token 与发送示例
- 当前账号历史消息查询和分页
- 系统状态、用户管理、全局参数、JWT 参数
- Bark 扫码绑定
- GET 与 POST 消息发送接口

## 通道

- 企业微信应用消息
- SMTP 邮件
- Telegram Bot
- 自定义 GET
- 自定义 POST
- 钉钉群机器人
- 飞书群机器人
- Bark

## 目录

```text
.
├── backend/              # Go API 服务
├── frontend/             # Vue 3 管理后台
├── docs/                 # 开发、部署、接口文档
├── scripts/              # 本地辅助脚本
├── docker-compose.yml    # 本地/生产 Compose 编排
└── .github/workflows/    # CI 与 Docker 镜像自动化
```

## 快速启动

### 安装 Docker

先安装 Docker 和 Docker Compose：

- Windows/macOS: 安装 Docker Desktop。
- Linux: 安装 Docker Engine，并确认 Compose v2 可用。

验证：

```bash
docker --version
docker compose version
```

### Docker Compose

```bash
cp .env.example .env
docker compose up -d --build
```

访问 `http://localhost:9000`。

默认账号：

```text
用户名: admin
密码: 123456
```

运行数据会写入 Docker volume `inotify_data`。

常用命令：

```bash
docker compose ps
docker compose logs -f
docker compose restart
docker compose down
```

升级：

```bash
docker compose pull
docker compose up -d --build
```

### GitHub Docker 自动化

项目已配置 GitHub Actions: [.github/workflows/ci.yml](.github/workflows/ci.yml)。

自动化流程：

- push 或 pull request 到 `main` / `master` 时运行后端测试。
- push 或 pull request 到 `main` / `master` 时运行前端 lint 和 build。
- push 到 `main` / `master` 时构建 Docker 镜像并推送到 GitHub Container Registry。

镜像地址格式：

```text
ghcr.io/<owner>/<repo>/backend:latest
ghcr.io/<owner>/<repo>/frontend:latest
ghcr.io/<owner>/<repo>/backend:<commit-sha>
ghcr.io/<owner>/<repo>/frontend:<commit-sha>
```

如果仓库是私有仓库，需要在 GitHub Packages 里给镜像配置访问权限，或登录 GHCR 后拉取。

### 本地开发

后端：

```bash
cd backend
go run ./cmd/inotify
```

前端：

```bash
cd frontend
npm install
npm run dev
```

前端开发服务器默认监听 `http://localhost:9000`，并把 `/api` 代理到 `http://127.0.0.1:8000`。

## 消息发送

GET 示例：

```text
GET /api/send?token=<TOKEN>&title=标题&body=内容
```

POST 示例：

```http
POST /api/send
Content-Type: application/json

{
  "token": "<TOKEN>",
  "title": "标题",
  "data": "第一行\n第二行"
}
```

更多接口说明见 [docs/API.md](docs/API.md)。

## 文档

- [开发说明](docs/DEVELOPMENT.md)
- [部署说明](docs/DEPLOYMENT.md)
- [接口说明](docs/API.md)
- [迁移说明](docs/MIGRATION.md)

## 验证

```bash
cd backend
go test ./...

cd ../frontend
npm run lint
npm run build
```
