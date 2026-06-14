# Inotify

Inotify 是一个轻量级消息通知管理系统。当前项目已经重组为 Go 后端和 Vue 3 管理后台，不再包含旧版 .NET Core 与 Vue 2 代码。

**项目地址**：[https://github.com/xpnas/inotify-go](https://github.com/xpnas/inotify-go)

## 技术栈

- Backend: Go 1.25, Gin, Gorm, SQLite
- Frontend: Vue 3, Vite, Pinia, Vue Router, Element Plus
- Deployment: Docker, Docker Compose, Nginx

## 功能

- 用户登录、JWT 鉴权、GitHub OAuth 登录、企业微信扫码登录
- 消息通道管理（企业微信、邮件、Telegram、钉钉、飞书、Bark、WxPusher 等）
- 当前账号 Token 与发送示例
- 当前账号历史消息查询和分页
- 系统状态、用户管理、全局参数、JWT 参数
- Bark 扫码绑定
- GET 与 POST 消息发送接口

## 通道

- 企业微信应用消息（支持图片 URL 自动转图文消息）
- SMTP 邮件
- Telegram Bot（支持图片 URL 自动转 sendPhoto）
- 钉钉群机器人
- 飞书群机器人
- WxPusher（标准推送 + SPT 极简推送）
- 自定义 GET
- 自定义 POST
- Bark

## 目录

```text
.
├── backend/                  # Go API 服务
├── frontend/                 # Vue 3 管理后台
├── docs/                     # 开发、部署、接口文档
├── scripts/                  # 本地辅助脚本
├── docker-compose.yml        # 生产一键部署（拉取预构建镜像）
├── docker-compose.build.yml  # 本地从源码构建
└── .github/workflows/        # CI 与 Docker 镜像自动化
```

## 一键安装（推荐）

确保已安装 Docker 和 Docker Compose，执行：

```bash
curl -fsSL https://raw.githubusercontent.com/xpnas/inotify-go/master/docker-compose.yml -o docker-compose.yml
docker compose up -d
```

访问 `http://<服务器IP>:9000`，默认账号：

```text
用户名: admin
密码: 123456
```

如需修改端口，在启动前设置环境变量：

```bash
INOTIFY_HTTP_PORT=8080 docker compose up -d
```

升级到最新版：

```bash
docker compose pull && docker compose up -d
```

## 从源码构建

克隆仓库后使用 `docker-compose.build.yml`：

```bash
git clone https://github.com/xpnas/inotify-go.git
cd inotify-go
docker compose -f docker-compose.build.yml up -d --build
```

## 本地开发

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

## GitHub 登录配置

在 GitHub 创建 OAuth App：

```text
GitHub Settings -> Developer settings -> OAuth Apps -> New OAuth App
```

生产环境示例：

```text
Homepage URL: https://你的域名
Authorization callback URL: https://你的域名/oauth/github/callback
```

本地开发示例：

```text
Homepage URL: http://localhost:9000
Authorization callback URL: http://localhost:9000/oauth/github/callback
```

创建完成后复制 `Client ID` 和 `Client Secret`，登录 Inotify 后台，在 `系统管理 -> 全局设置 -> GitHub 登录设置` 中填写并保存。页面会显示当前应填写到 GitHub OAuth App 的 `GitHub redirect_uri`，通常为当前站点地址加 `/oauth/github/callback`。

如果服务器访问 GitHub 较慢或受限，可以在同一页面填写 `代理地址`，例如：

```text
http://127.0.0.1:7890
```

清空 `GitHub Client ID` 或 `GitHub Client Secret` 后保存，即可关闭 GitHub 登录入口。

已有账号可以登录后进入 `三方登录`，把当前 Inotify 账号绑定到 GitHub 或企业微信。绑定后再使用三方登录时，会进入这个已有账号，而不是按三方账号标识创建新账号。

## 运维与排障

- 通道新增和编辑页面支持测试发送，并显示 HTTP 状态码、响应摘要或配置缺失原因。
- 通道代理策略支持 `不使用`、`全局代理`、`自定义代理地址`。
- 历史记录会保存通道发送详情，失败时可查看失败原因摘要。
- 系统状态页面提供配置诊断、回调地址检查、数据目录可写性检查和 SQLite 数据库备份下载。
- 用户管理页面会显示 GitHub 与企业微信绑定状态。

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

## Docker 镜像

镜像由 GitHub Actions 自动构建并推送至 GitHub Container Registry，支持 `linux/amd64` 和 `linux/arm64`，前后端已打包为单一镜像：

```text
ghcr.io/xpnas/inotify-go:latest
```

## GitHub Actions

每次推送到 `main` / `master` 分支时自动：

1. 运行后端 `go test ./...`
2. 运行前端 lint 和 build
3. 构建多平台 Docker 镜像并推送到 GHCR

## 文档

- [开发说明](docs/DEVELOPMENT.md)
- [部署说明](docs/DEPLOYMENT.md)
- [接口说明](docs/API.md)
- [迁移说明](docs/MIGRATION.md)

## 验证

```bash
cd backend && go test ./...
cd ../frontend && npm run lint && npm run build
```
