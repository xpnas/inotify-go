# API 文档

所有后台接口都挂在 `/api` 下。需要登录的接口通过 `X-Token` 请求头传递 JWT。

## 登录

```http
POST /api/oauth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "123456"
}
```

### 三方登录与绑定

GitHub 和企业微信登录使用独立回调页面：

- GitHub 回调：`/oauth/github/callback`
- 企业微信回调：`/oauth/weixin/callback`

GitHub 登录和绑定：

```http
GET /api/oauth/githublogin?redirectUri=https://example.com/oauth/github/callback
GET /api/oauth/githublogin?code=<CODE>&redirectUri=https://example.com/oauth/github/callback
GET /api/oauth/githubbind?redirectUri=https://example.com/oauth/github/callback
GET /api/oauth/githubbind?code=<CODE>&redirectUri=https://example.com/oauth/github/callback
POST /api/oauth/githubunbind
```

企业微信登录和绑定：

```http
GET /api/oauth/WeixinQrLogin?redirectUri=https://example.com/oauth/weixin/callback
GET /api/oauth/WeixinQrLogin?code=<CODE>
GET /api/oauth/WeixinQrBind?redirectUri=https://example.com/oauth/weixin/callback
GET /api/oauth/WeixinQrBind?code=<CODE>&redirectUri=https://example.com/oauth/weixin/callback
POST /api/oauth/WeixinQrUnbind
```

绑定和解绑接口需要 `X-Token`。

## 账号

用户修改自己的密码时需要提供旧密码：

```http
POST /api/oauth/resetPassword
X-Token: <JWT>
Content-Type: application/json

{
  "username": "admin",
  "oldPassword": "123456",
  "password": "new-password"
}
```

## 发送消息

### GET

```http
GET /api/send?token=<TOKEN>&title=标题&body=内容
```

### POST JSON

```http
POST /api/send
Content-Type: application/json

{
  "token": "<TOKEN>",
  "title": "标题",
  "data": "第一行\n第二行",
  "url": "https://example.com",
  "group": "default",
  "sound": "1107"
}
```

字段说明：

- `token`：用户发送 token。
- `key`：通道发送 key。传入 `key` 时只发送到该通道。
- `title`：消息标题。
- `body`、`data`、`content`：消息内容别名。
- `url`：可选跳转链接。
- `group`：可选分组。
- `sound`：可选 Bark 提示音。

## 历史记录

```http
GET /api/setting/GetMessageHistories?page=1&pageSize=10&title=&content=&success=true&startTime=2026-01-01&endTime=2026-01-31
X-Token: <JWT>
```

历史记录接口只返回当前账号的消息记录。`detail` 字段包含每个通道的发送摘要；失败时会尽量记录失败原因。

## 消息通道

新增或修改通道时可以通过 `config.ProxyMode` 配置代理策略：

- `no`：不使用代理。
- `global`：使用系统全局 `proxyAddress`。
- `custom`：使用当前通道的 `config.ProxyAddress`。

测试通道：

```http
POST /api/setting/TestSendAuth
X-Token: <JWT>
Content-Type: application/json

{
  "templateID": "HTTP-GET",
  "name": "test",
  "config": {
    "URL": "https://example.com/{title}/{data}",
    "ProxyMode": "global"
  }
}
```

响应示例：

```json
{
  "success": false,
  "message": "HTTP 状态码 403: forbidden",
  "statusCode": 403,
  "response": "forbidden"
}
```

## 系统诊断与备份

```http
GET /api/settingsys/Diagnostics
GET /api/settingsys/BackupDatabase
X-Token: <JWT>
```

诊断接口返回 OAuth 配置状态、回调地址、代理格式和数据目录可写性。备份接口会下载 SQLite 数据库文件。

## Bark 注册

```http
GET /Register?act=<TOKEN>
```

Bark App 可以打开或扫描这个地址。服务端会为当前 token 创建或更新 Bark 通道。
