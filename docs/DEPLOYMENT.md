# 部署说明

## 前置要求

请先安装 Docker 和 Docker Compose：

- Windows/macOS：安装 Docker Desktop。
- Linux：安装 Docker Engine 和 Compose v2。

验证安装：

```bash
docker --version
docker compose version
```

## Docker Compose

生产部署推荐直接使用预构建镜像：

```bash
curl -fsSL https://raw.githubusercontent.com/xpnas/inotify-go/master/docker-compose.yml -o docker-compose.yml
docker compose up -d
```

如需从源码本地构建：

```bash
docker compose -f docker-compose.build.yml up -d --build
```

启动后访问：

```text
http://localhost:9000
```

运行数据保存在 `inotify_data` Docker 卷或后端数据目录中。

常用命令：

```bash
docker compose ps
docker compose logs -f
docker compose restart
docker compose down
```

## GitHub Container Registry

`.github/workflows/ci.yml` 会在推送到 `main` 或 `master` 分支时构建并推送镜像到 GitHub Container Registry：

```text
ghcr.io/xpnas/inotify-go:latest
ghcr.io/xpnas/inotify-go:<commit-sha>
```

如果仓库或镜像是私有的，拉取前需要登录：

```bash
echo <github-token> | docker login ghcr.io -u <github-username> --password-stdin
```

## 反向代理

如果需要放在自己的反向代理后面，请把流量转发到前端服务。生产前端 Nginx 已经代理以下后端路径：

- `/api/*`
- `/Register`
- `/RegisterCheck`
- `/Ping`
- `/Healthz`
- `/Info`
- `/<32位十六进制发送key>/*`

GitHub OAuth 回调地址应配置为：

```text
https://你的域名/oauth/github/callback
```

企业微信登录回调地址为：

```text
https://你的域名/oauth/weixin/callback
```

## 备份

SQLite 数据库和 JWT 配置保存在后端数据目录：

```text
inotify.db
jwt.json
```

可以在后台 `系统状态` 页面直接下载数据库备份。

Docker Compose 部署也可以使用命令备份整个数据卷：

```bash
docker run --rm -v inotify_inotify_data:/data -v "$PWD:/backup" alpine tar czf /backup/inotify-data.tgz -C /data .
```

## 升级

```bash
docker compose pull
docker compose up -d
```

如果使用源码构建：

```bash
docker compose -f docker-compose.build.yml up -d --build
```

后端启动时会自动执行数据库迁移。
