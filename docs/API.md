# API

All admin APIs are under `/api`. Authenticated requests use the `X-Token` header.

## Login

```http
POST /api/oauth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "123456"
}
```

## Send Message

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

Fields:

- `token`: user send token.
- `key`: channel key. If `key` is provided, the message is sent to that single channel.
- `title`: message title.
- `body`, `data`, `content`: message body aliases.
- `url`: optional action URL.
- `group`: optional group.
- `sound`: optional Bark sound.

## History

```http
GET /api/setting/GetMessageHistories?page=1&pageSize=10&title=&content=&success=true&startTime=2026-01-01&endTime=2026-01-31
X-Token: <JWT>
```

The history endpoint only returns messages for the current account.

## Bark Registration

```http
GET /Register?act=<TOKEN>
```

The Bark app can open or scan this URL. The server creates or updates a Bark channel for the current token.
